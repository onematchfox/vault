package persistcache

// IndexType - Different index types that are used to create buckets in the bolt
// db, to make it easier to query and restore by type.
type IndexType string

const (
	// SecretLeaseType - Lease with secret info
	SecretLeaseType IndexType = "secret-lease"

	// AuthLeaseType - Lease with auth info
	AuthLeaseType = "auth-lease"

	// TokenType - Auto-auth token type
	TokenType = "token"
)

// Storage interface for persistent storage
type Storage interface {
	// Set saves an Index item in the persistent storage, with a string key,
	// []byte value, and type of Index
	Set(string, []byte, IndexType) error

	// Delete an Index item from the persistent storage
	Delete(id string) error

	// GetByType - retrieve a list of serialized Index's by type
	GetByType(IndexType) ([][]byte, error)

	// Close the persistent storage
	Close() error

	// Clear?

	// Rotate key?
}
