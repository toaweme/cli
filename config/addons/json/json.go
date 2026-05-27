package json

import (
	"encoding/json"
)

// Codec serializes and deserializes config values as JSON.
type Codec struct{}

func (c *Codec) Marshal(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func (c *Codec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (c *Codec) Extension() string {
	return ".json"
}
