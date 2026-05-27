package toml

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

// Codec serializes and deserializes config values as TOML.
type Codec struct{}

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

func (c *Codec) Extension() string {
	return ".toml"
}
