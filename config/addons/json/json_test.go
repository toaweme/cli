package json

import (
	"encoding/json"
	"testing"
)

func Test_Codec_Extension(t *testing.T) {
	c := New()
	if c.Extension() != ".json" {
		t.Fatalf("want .json, got %s", c.Extension())
	}
}

func Test_New_OverrideExtensions(t *testing.T) {
	c := New(".json", ".jsonc")
	if c.Extension() != ".json" {
		t.Fatalf("primary want .json, got %s", c.Extension())
	}
	exts := c.Extensions()
	if len(exts) != 2 || exts[1] != ".jsonc" {
		t.Fatalf("want [.json .jsonc], got %v", exts)
	}
}

func Test_Codec_RoundTrip(t *testing.T) {
	type cfg struct {
		Name    string `json:"name"`
		Port    int    `json:"port"`
		Verbose bool   `json:"verbose"`
	}

	c := New()
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
