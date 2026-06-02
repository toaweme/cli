package config

// FileSecrets returns a file-backed SecretBackend writing 0600-permission files
// under dir. A leading "~" in dir expands to the user's home directory.
func FileSecrets(dir string) SecretBackend {
	s := NewFileStore(dir)
	s.perm = 0o600
	return s
}
