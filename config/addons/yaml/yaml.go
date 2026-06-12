// Package yaml provides a YAML config codec for the cli config addon system.
package yaml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// defaultExtensions are the extensions a YAML codec recognizes when reading.
// The first (".yml") is the primary extension used for writing and for the --help-format name.
var defaultExtensions = []string{".yml", ".yaml"}

// Codec serializes and deserializes config values as YAML.
// It recognizes one or more file extensions; the first is the primary, used for output.
type Codec struct {
	exts []string
}

// New returns a YAML codec. Pass extensions to override the defaults (".yml", ".yaml");
// the first becomes the primary extension used for output.
// With no args it recognizes both ".yml" and ".yaml".
func New(exts ...string) *Codec {
	if len(exts) == 0 {
		exts = defaultExtensions
	}
	return &Codec{exts: exts}
}

// Marshal encodes v as YAML.
func (c *Codec) Marshal(v any) ([]byte, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}
	return data, nil
}

// Unmarshal decodes YAML data into v.
func (c *Codec) Unmarshal(data []byte, v any) error {
	if err := yaml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}
	return nil
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
