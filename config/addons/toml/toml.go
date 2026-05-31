package toml

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

// defaultExtensions are the extensions a TOML codec recognizes when reading. The
// first (".toml") is the primary extension used for writing and for the --format name.
var defaultExtensions = []string{".toml"}

// Codec serializes and deserializes config values as TOML. It recognizes one or
// more file extensions; the first is the primary, used for output.
type Codec struct {
	exts []string
}

// New returns a TOML codec. Pass extensions to override the default (".toml"); the
// first becomes the primary extension used for output. With no args it recognizes
// ".toml".
func New(exts ...string) *Codec {
	if len(exts) == 0 {
		exts = defaultExtensions
	}
	return &Codec{exts: exts}
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Codec) Unmarshal(data []byte, v any) error {
	return toml.Unmarshal(data, v)
}

// Extension returns the primary extension, used for output.
func (c *Codec) Extension() string {
	if len(c.exts) == 0 {
		return defaultExtensions[0]
	}
	return c.exts[0]
}

// Extensions returns every extension this codec recognizes when reading.
func (c *Codec) Extensions() []string {
	if len(c.exts) == 0 {
		return defaultExtensions
	}
	return c.exts
}
