package yaml

import (
	"strings"
	"testing"
)

func Test_Codec_Extension(t *testing.T) {
	c := &Codec{}
	if c.Extension() != ".yaml" {
		t.Fatalf("want .yaml, got %s", c.Extension())
	}
}

func Test_Codec_RoundTrip(t *testing.T) {
	type cfg struct {
		Name    string `yaml:"name"`
		Port    int    `yaml:"port"`
		Verbose bool   `yaml:"verbose"`
	}

	c := &Codec{}
	original := cfg{Name: "test", Port: 8080, Verbose: true}

	data, err := c.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "name: test") {
		t.Fatalf("expected yaml output, got: %s", content)
	}

	var loaded cfg
	if err := c.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if loaded != original {
		t.Fatalf("want %+v, got %+v", original, loaded)
	}
}
