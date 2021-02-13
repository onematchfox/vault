package persistcache

const (
	// SecretLeaseType - Lease with secret info
	SecretLeaseType = "secret-lease"

	// AuthLeaseType - Lease with auth info
	AuthLeaseType = "auth-lease"

	// TokenType - Auto-auth token type
	TokenType = "token"
)

// Storage interface for persistent storage
type Storage interface {
	// Set saves an Index item in the persistent storage, with a string key,
	// []byte value, and type of Index
	Set(string, []byte, string) error

	// Delete an Index item from the persistent storage
	Delete(id string) error

	// GetByType - retrieve a list of serialized Index's by type
	GetByType(string) ([][]byte, error)

	// Close the persistent storage
	Close() error

	// Clear?

	// Rotate key?
}
