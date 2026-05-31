package yaml

import (
	"gopkg.in/yaml.v3"
)

// defaultExtensions are the extensions a YAML codec recognizes when reading. The
// first (".yml") is the primary extension used for writing and for the --format name.
var defaultExtensions = []string{".yml", ".yaml"}

// Codec serializes and deserializes config values as YAML. It recognizes one or
// more file extensions; the first is the primary, used for output.
type Codec struct {
	exts []string
}

// New returns a YAML codec. Pass extensions to override the defaults (".yml",
// ".yaml"); the first becomes the primary extension used for output. With no args
// it recognizes both ".yml" and ".yaml".
func New(exts ...string) *Codec {
	if len(exts) == 0 {
		exts = defaultExtensions
	}
	return &Codec{exts: exts}
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (c *Codec) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
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
