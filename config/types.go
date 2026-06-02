package config

// Store reads and writes configuration values by key. Keys map to files within the
// store's base directory.
type Store interface {
	// Load reads the value at key into target.
	Load(key string, target any) error
	// Save writes value to key. Creates parent directories as needed.
	Save(key string, value any) error
	// Delete removes the value at key. Returns nil if key does not exist.
	Delete(key string) error
	// Exists reports whether a value exists at key.
	Exists(key string) bool
}

// SecretStore is a Store with restricted file permissions (0o600). Implementations
// must ensure secrets are never written world-readable.
type SecretStore interface {
	Store
}

// Codec serializes and deserializes config values.
type Codec interface {
	// Marshal encodes v into bytes.
	Marshal(v any) ([]byte, error)
	// Unmarshal decodes data into v.
	Unmarshal(data []byte, v any) error
	// Extension returns the primary file extension for this codec (e.g. ".json",
	// ".yml"), used for writing.
	Extension() string
}

// SecretBackend stores sensitive values out of the merged config view. The file
// backend ships here (FileSecrets); a keychain-backed implementation can live in an
// addon module without changing this contract.
type SecretBackend interface {
	Load(key string, target any) error
	Save(key string, value any) error
	Delete(key string) error
}
