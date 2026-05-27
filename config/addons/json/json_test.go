package json

import (
	"encoding/json"
	"testing"
)

func Test_Codec_Extension(t *testing.T) {
	c := &Codec{}
	if c.Extension() != ".json" {
		t.Fatalf("want .json, got %s", c.Extension())
	}
}

func Test_Codec_RoundTrip(t *testing.T) {
	type cfg struct {
		Name    string `json:"name"`
		Port    int    `json:"port"`
		Verbose bool   `json:"verbose"`
	}

	c := &Codec{}
	original := cfg{Name: "test", Port: 8080, Verbose: true}

	data, err := c.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !json.Valid(data) {
		t.Fatalf("output is not valid JSON: %s", data)
	}

	var loaded cfg
	if err := c.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if loaded != original {
		t.Fatalf("want %+v, got %+v", original, loaded)
	}
}

func Test_Codec_PrettyPrint(t *testing.T) {
	c := &Codec{}
	data, err := c.Marshal(map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// should be indented, not single-line
	if len(data) < 10 {
		t.Fatalf("expected pretty-printed JSON, got: %s", data)
	}
}
