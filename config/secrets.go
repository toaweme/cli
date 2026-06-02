package config

// FileSecrets returns a file-backed secrets Store: a single 0o600-permission file
// named "secrets" under dir, where each secret is a key within it (read with
// KeyRead, written with KeyWrite). A leading "~" in dir expands to the user's home
// directory. Secrets are just a Store, so one can be passed to NewResolver like any
// other to fold its values into a command's options.
func FileSecrets(dir string, codecs ...Codec) *FileStore {
	s := NewFileStore(dir, "secrets", codecs...)
	s.perm = 0o600
	return s
}
