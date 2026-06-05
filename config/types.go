package config

// Store is a single config file: read or write it whole, or address a single dotted key within it.
// The file (directory, base name, codec) is fixed when the store is constructed, so callers never pass a path.
// Secrets are just a Store with restricted permissions (see FileSecrets).
type Store interface {
	// Read decodes the whole file into target. A missing file is not an error.
	Read(target any) error
	// Write persists value as the whole file.
	Write(value any) error
	// Exists reports whether the file exists.
	Exists() bool
	// Delete removes the whole file. A missing file is not an error.
	Delete() error
	// KeyRead returns the value at a dotted path within the file, or nil when absent.
	// A missing file yields a nil value and no error.
	KeyRead(key string) (any, error)
	// KeyWrite sets a single dotted path within the file (read-modify-write).
	KeyWrite(key string, value any) error
	// KeyExists reports whether a dotted path is present in the file.
	KeyExists(key string) bool
	// KeyDelete removes a single dotted path within the file (read-modify-write).
	// A missing file or absent path is not an error.
	KeyDelete(key string) error
}

// Codec serializes and deserializes config values.
type Codec interface {
	// Marshal encodes v into bytes.
	Marshal(v any) ([]byte, error)
	// Unmarshal decodes data into v.
	Unmarshal(data []byte, v any) error
	// Extension returns the primary file extension for this codec (e.g. ".json", ".yml"), used for writing.
	Extension() string
}
