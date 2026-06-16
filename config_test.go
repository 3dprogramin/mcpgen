package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMergeArgs(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		extra    []string
		want     []string
	}{
		{
			name:     "replace joined flag",
			existing: []string{"-y", "chrome-devtools-mcp@latest", "--browser-url=http://127.0.0.1:9222"},
			extra:    []string{"--browser-url=http://127.0.0.1:9333"},
			want:     []string{"-y", "chrome-devtools-mcp@latest", "--browser-url=http://127.0.0.1:9333"},
		},
		{
			name:     "replace split flag value",
			existing: []string{"-y", "mongodb-mcp-server", "--connectionString", "mongodb://old"},
			extra:    []string{"--connectionString=mongodb://new"},
			want:     []string{"-y", "mongodb-mcp-server", "--connectionString", "mongodb://new"},
		},
		{
			name:     "append new flag",
			existing: []string{"-y", "pkg", "--browser-url=http://x"},
			extra:    []string{"--headless"},
			want:     []string{"-y", "pkg", "--browser-url=http://x", "--headless"},
		},
		{
			name:     "skip duplicate bare token",
			existing: []string{"--headless"},
			extra:    []string{"--headless"},
			want:     []string{"--headless"},
		},
		{
			name:     "append flag with no existing match",
			existing: []string{"-y", "pkg"},
			extra:    []string{"--port=8080"},
			want:     []string{"-y", "pkg", "--port=8080"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeArgs(append([]string(nil), tt.existing...), tt.extra)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func objKeys(t *testing.T, raw json.RawMessage) []string {
	t.Helper()
	obj, err := parseOrderedObject(raw)
	if err != nil {
		t.Fatalf("parseOrderedObject: %v", err)
	}
	keys := make([]string, len(obj))
	for i, p := range obj {
		keys[i] = p.Key
	}
	return keys
}

func TestApplyArgsMergePreservesOrder(t *testing.T) {
	config := json.RawMessage(`{"type":"stdio","command":"npx","args":["-y","chrome-devtools-mcp@latest","--browser-url=http://127.0.0.1:9222"],"env":{}}`)
	got, err := applyArgs(config, []string{"--browser-url=http://127.0.0.1:9333"}, false)
	if err != nil {
		t.Fatal(err)
	}
	if keys := objKeys(t, got); !reflect.DeepEqual(keys, []string{"type", "command", "args", "env"}) {
		t.Errorf("key order changed: %v", keys)
	}
	args, _ := serverArgs(got)
	want := []string{"-y", "chrome-devtools-mcp@latest", "--browser-url=http://127.0.0.1:9333"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestApplyArgsReplaceAll(t *testing.T) {
	config := json.RawMessage(`{"type":"stdio","command":"npx","args":["-y","old"],"env":{}}`)
	override := []string{"-y", "new-pkg", "--flag=1"}
	got, err := applyArgs(config, override, true)
	if err != nil {
		t.Fatal(err)
	}
	args, _ := serverArgs(got)
	if !reflect.DeepEqual(args, override) {
		t.Errorf("args = %v, want %v", args, override)
	}
}

func TestApplyArgsEmptyOverrideUnchanged(t *testing.T) {
	config := json.RawMessage(`{"type":"sse","url":"http://x"}`)
	got, err := applyArgs(config, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(config) {
		t.Errorf("config changed: %s", got)
	}
}

func TestApplyArgsNoArgsField(t *testing.T) {
	config := json.RawMessage(`{"type":"sse","url":"http://x"}`)
	if _, err := applyArgs(config, []string{"--flag=1"}, false); err == nil {
		t.Error("expected error when overriding args on a server without args")
	}
}

func TestServerArgs(t *testing.T) {
	if args, ok := serverArgs(json.RawMessage(`{"command":"npx","args":["-y","pkg"]}`)); !ok || len(args) != 2 {
		t.Errorf("serverArgs stdio = %v, %v", args, ok)
	}
	if _, ok := serverArgs(json.RawMessage(`{"type":"sse","url":"http://x"}`)); ok {
		t.Error("serverArgs should be false for sse config without args")
	}
}
