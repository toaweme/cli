package yaml

import (
	"gopkg.in/yaml.v3"
)

// Codec serializes and deserializes config values as YAML.
type Codec struct{}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (c *Codec) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

func (c *Codec) Extension() string {
	return ".yaml"
}
