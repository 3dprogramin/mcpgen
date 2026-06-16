package main

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

func readServers(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(mcpFileName)
	if err != nil {
		t.Fatalf("reading %s: %v", mcpFileName, err)
	}
	var f mcpFile
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	var names []string
	for n := range f.MCPServers {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func TestRunGenerateCreateAndMerge(t *testing.T) {
	cat, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())

	if err := runGenerate(cat, []selection{{name: "burp"}}, false); err != nil {
		t.Fatal(err)
	}
	if got := readServers(t); len(got) != 1 || got[0] != "burp" {
		t.Fatalf("after create: %v", got)
	}

	// Merge a second server, the first should remain.
	if err := runGenerate(cat, []selection{{name: "chrome-devtools"}}, false); err != nil {
		t.Fatal(err)
	}
	if got := readServers(t); len(got) != 2 {
		t.Fatalf("after merge: %v", got)
	}
}

func TestRunGenerateConflict(t *testing.T) {
	cat, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())

	if err := runGenerate(cat, []selection{{name: "burp"}}, false); err != nil {
		t.Fatal(err)
	}
	if err := runGenerate(cat, []selection{{name: "burp"}}, false); err == nil {
		t.Fatal("expected conflict error without force")
	}
	if err := runGenerate(cat, []selection{{name: "burp"}}, true); err != nil {
		t.Fatalf("force should overwrite: %v", err)
	}
}

func TestRunGenerateUnknown(t *testing.T) {
	cat, err := loadCatalog()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	if err := runGenerate(cat, []selection{{name: "nope"}}, false); err == nil {
		t.Fatal("expected error for unknown server")
	}
}
