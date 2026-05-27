```go
// Store reads and writes configuration values by key.
// Keys map to files within the store's base directory.
type Store interface {
    Load(key string, target any) error
    Save(key string, value any) error
    Delete(key string) error
    Exists(key string) bool
}

// SecretStore is a Store with restricted file permissions (0o600).
type SecretStore interface {
    Store
}
```
