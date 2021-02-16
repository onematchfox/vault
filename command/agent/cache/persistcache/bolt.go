package persistcache

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	bolt "go.etcd.io/bbolt"
)

const (
	// Keep track of schema version for future migrations
	storageVersionKey = "version"
	storageVersion    = "1"

	// CacheFileName - filename for the persistent cache file
	CacheFileName = "vault-agent-cache.db"
)

// BoltStorage is a persistent cache using a bolt db. Items are organized with
// the encryption key as the top-level bucket, and then leases and tokens are
// stored in sub buckets.
type BoltStorage struct {
	db         *bolt.DB
	rootBucket string
	logger     hclog.Logger

	// Encryption interface to be named later
	// sekret Encryption
}

// BoltStorageConfig is the collection of input parameters for setting up bolt
// storage
type BoltStorageConfig struct {
	Path       string
	RootBucket string
	Logger     hclog.Logger
}

// NewBoltStorage opens a new bolt db at the specified file path and returns it.
// If the db already exists the buckets will just be created if they don't
// exist.
func NewBoltStorage(config *BoltStorageConfig) (*BoltStorage, error) {
	cachePath := filepath.Join(config.Path, CacheFileName)
	db, err := bolt.Open(cachePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		return createBoltSchema(tx, config.RootBucket)
	})
	if err != nil {
		return nil, err
	}
	bs := &BoltStorage{
		db:         db,
		rootBucket: config.RootBucket,
		logger:     config.Logger,
	}
	return bs, nil
}

func createBoltSchema(tx *bolt.Tx, rootBucketName string) error {
	root, err := tx.CreateBucketIfNotExists([]byte(rootBucketName))
	if err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", rootBucketName, err)
	}
	_, err = root.CreateBucketIfNotExists([]byte(TokenType))
	if err != nil {
		return fmt.Errorf("failed to create token sub-bucket: %w", err)
	}
	_, err = root.CreateBucketIfNotExists([]byte(AuthLeaseType))
	if err != nil {
		return fmt.Errorf("failed to create auth lease sub-bucket: %w", err)
	}
	_, err = root.CreateBucketIfNotExists([]byte(SecretLeaseType))
	if err != nil {
		return fmt.Errorf("failed to create secret lease sub-bucket: %w", err)
	}

	// check and set file version in the root bucket
	version := root.Get([]byte(storageVersionKey))
	switch {
	case version == nil:
		err = root.Put([]byte(storageVersionKey), []byte(storageVersion))
		if err != nil {
			return fmt.Errorf("failed to set storage version: %w", err)
		}
	case string(version) != storageVersion:
		return fmt.Errorf("storage migration from %s to %s not implemented", string(version), storageVersion)
	}
	return nil
}

// Set an index in bolt storage
func (b *BoltStorage) Set(id string, index []byte, indexType string) error {

	// TODO(tvoran): encrypt index here

	return b.db.Update(func(tx *bolt.Tx) error {
		top := tx.Bucket([]byte(b.rootBucket))
		if top == nil {
			return fmt.Errorf("bucket %q not found", b.rootBucket)
		}
		s := top.Bucket([]byte(indexType))
		if s == nil {
			return fmt.Errorf("bucket %q not found", indexType)
		}
		return s.Put([]byte(id), index)
	})
}

// Delete an index by id from bolt storage
func (b *BoltStorage) Delete(id string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		top := tx.Bucket([]byte(b.rootBucket))
		if top == nil {
			return fmt.Errorf("bucket %q not found", b.rootBucket)
		}
		// Since Delete returns a nil error if the key doesn't exist, just call
		// delete in both sub-buckets without checking existence first
		if err := top.Bucket([]byte(TokenType)).Delete([]byte(id)); err != nil {
			return fmt.Errorf("failed to delete %q from token bucket: %w", id, err)
		}
		if err := top.Bucket([]byte(AuthLeaseType)).Delete([]byte(id)); err != nil {
			return fmt.Errorf("failed to delete %q from auth lease bucket: %w", id, err)
		}
		if err := top.Bucket([]byte(SecretLeaseType)).Delete([]byte(id)); err != nil {
			return fmt.Errorf("failed to delete %q from secret lease bucket: %w", id, err)
		}
		b.logger.Trace("deleted index from bolt db", "id", id)
		return nil
	})
}

// GetByType returns a list of stored items of the specified type
func (b *BoltStorage) GetByType(indexType string) ([][]byte, error) {
	returnBytes := [][]byte{}

	err := b.db.View(func(tx *bolt.Tx) error {
		top := tx.Bucket([]byte(b.rootBucket))
		if top == nil {
			return fmt.Errorf("bucket %q not found", b.rootBucket)
		}
		top.Bucket([]byte(indexType)).ForEach(func(k, v []byte) error {

			// TODO(tvoran): decrypt v here

			returnBytes = append(returnBytes, v)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return returnBytes, nil
}

// Close the boltdb
func (b *BoltStorage) Close() error {
	b.logger.Trace("closing bolt db", "file", b.db.Path())
	return b.db.Close()
}

// Clear the boltdb by deleting all the root buckets and recreating the
// schema/layout
func (b *BoltStorage) Clear() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		tx.ForEach(func(name []byte, bucket *bolt.Bucket) error {
			b.logger.Trace("deleting bolt bucket", "name", name)
			if err := tx.DeleteBucket(name); err != nil {
				return err
			}
			return nil
		})
		return createBoltSchema(tx, b.rootBucket)
	})
}
