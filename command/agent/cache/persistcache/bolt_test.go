package persistcache

import (
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/ory/dockertest/v3/docker/pkg/ioutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBolt_SetGet(t *testing.T) {
	path, err := ioutils.TempDir("", "bolt-test")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	b, err := NewBoltStorage(&BoltStorageConfig{
		Path:       path,
		RootBucket: "test",
		Logger:     hclog.Default(),
	})
	require.NoError(t, err)

	secrets, err := b.GetByType(SecretLeaseType)
	assert.NoError(t, err)
	require.Len(t, secrets, 0)

	err = b.Set("test1", []byte("hello"), SecretLeaseType)
	assert.NoError(t, err)
	secrets, err = b.GetByType(SecretLeaseType)
	assert.NoError(t, err)
	require.Len(t, secrets, 1)
	assert.Equal(t, []byte("hello"), secrets[0])
}

func TestBoltDelete(t *testing.T) {
	path, err := ioutils.TempDir("", "bolt-test")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	b, err := NewBoltStorage(&BoltStorageConfig{
		Path:       path,
		RootBucket: "test",
		Logger:     hclog.Default(),
	})
	require.NoError(t, err)

	err = b.Set("secret-test1", []byte("hello1"), SecretLeaseType)
	require.NoError(t, err)
	err = b.Set("secret-test2", []byte("hello2"), SecretLeaseType)
	require.NoError(t, err)

	secrets, err := b.GetByType(SecretLeaseType)
	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.ElementsMatch(t, [][]byte{[]byte("hello1"), []byte("hello2")}, secrets)

	err = b.Delete("secret-test1")
	require.NoError(t, err)
	secrets, err = b.GetByType(SecretLeaseType)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, []byte("hello2"), secrets[0])
}

func TestBoltClear(t *testing.T) {
	path, err := ioutils.TempDir("", "bolt-test")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	b, err := NewBoltStorage(&BoltStorageConfig{
		Path:       path,
		RootBucket: "test",
		Logger:     hclog.Default(),
	})
	require.NoError(t, err)

	// Populate the bolt db
	err = b.Set("secret-test1", []byte("hello"), SecretLeaseType)
	require.NoError(t, err)
	secrets, err := b.GetByType(SecretLeaseType)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, []byte("hello"), secrets[0])

	err = b.Set("auth-test1", []byte("hello"), AuthLeaseType)
	require.NoError(t, err)
	auths, err := b.GetByType(AuthLeaseType)
	require.NoError(t, err)
	assert.Len(t, auths, 1)
	assert.Equal(t, []byte("hello"), auths[0])

	err = b.Set("token-test1", []byte("hello"), TokenType)
	require.NoError(t, err)
	tokens, err := b.GetByType(TokenType)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, []byte("hello"), tokens[0])

	// Clear the bolt db, and check that it's indeed clear
	err = b.Clear()
	require.NoError(t, err)
	secrets, err = b.GetByType(SecretLeaseType)
	require.NoError(t, err)
	assert.Len(t, secrets, 0)
	auths, err = b.GetByType(AuthLeaseType)
	require.NoError(t, err)
	assert.Len(t, auths, 0)
	tokens, err = b.GetByType(TokenType)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}
