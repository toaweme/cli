package json

import (
	"encoding/json"
)

// defaultExtensions are the extensions a JSON codec recognizes when reading.
// The first (".json") is the primary extension used for writing and for the --help-format name.
var defaultExtensions = []string{".json"}

// Codec serializes and deserializes config values as JSON.
// It recognizes one or more file extensions; the first is the primary, used for output.
type Codec struct {
	exts []string
}

// New returns a JSON codec. Pass extensions to override the default (".json");
// the first becomes the primary extension used for output. With no args it recognizes ".json".
func New(exts ...string) *Codec {
	if len(exts) == 0 {
		exts = defaultExtensions
	}
	return &Codec{exts: exts}
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (c *Codec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
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
