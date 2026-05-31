package config

// SecretBackend stores sensitive values out of the merged config view. The file
// backend ships here; a keychain-backed implementation can live in an addon
// module without changing this contract.
type SecretBackend interface {
	Load(key string, target any) error
	Save(key string, value any) error
	Delete(key string) error
}

// FileSecrets returns a file-backed SecretBackend writing 0600-permission files
// under dir. A leading "~" in dir expands to the user's home directory.
func FileSecrets(dir string) SecretBackend {
	s := NewFileStore(dir)
	s.perm = 0o600
	return s
}
