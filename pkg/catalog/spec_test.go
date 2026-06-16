package catalog

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestServerSpecConfigStdio(t *testing.T) {
	spec := ServerSpec{
		Type:    "stdio",
		Command: "npx",
		Args:    []string{"-y", "some-mcp"},
		Env:     map[string]string{"TOKEN": "x"},
	}
	cfg, err := spec.Config()
	if err != nil {
		t.Fatal(err)
	}
	if keys := objKeys(t, cfg); !reflect.DeepEqual(keys, []string{"type", "command", "args", "env"}) {
		t.Errorf("key order = %v", keys)
	}
}

func TestServerSpecConfigStdioMinimal(t *testing.T) {
	cfg, err := ServerSpec{Type: "stdio", Command: "npx"}.Config()
	if err != nil {
		t.Fatal(err)
	}
	// args and env are omitted when empty.
	if keys := objKeys(t, cfg); !reflect.DeepEqual(keys, []string{"type", "command"}) {
		t.Errorf("key order = %v, want [type command]", keys)
	}
}

func TestServerSpecConfigRemote(t *testing.T) {
	for _, typ := range []string{"sse", "http"} {
		cfg, err := ServerSpec{Type: typ, URL: "http://127.0.0.1:9000"}.Config()
		if err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		var got struct {
			Type string `json:"type"`
			URL  string `json:"url"`
		}
		if err := json.Unmarshal(cfg, &got); err != nil {
			t.Fatal(err)
		}
		if got.Type != typ || got.URL != "http://127.0.0.1:9000" {
			t.Errorf("%s: got %+v", typ, got)
		}
	}
}

func TestServerSpecConfigErrors(t *testing.T) {
	cases := []ServerSpec{
		{Type: "stdio"},           // missing command
		{Type: "sse"},             // missing url
		{Type: "http"},            // missing url
		{Type: "bogus", URL: "x"}, // unknown type
	}
	for _, spec := range cases {
		if _, err := spec.Config(); err == nil {
			t.Errorf("expected error for %+v", spec)
		}
	}
}
