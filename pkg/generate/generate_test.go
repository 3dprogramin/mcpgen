package generate

import (
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/3dprogramin/mcpgen/pkg/catalog"
)

func readServers(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(FileName)
	if err != nil {
		t.Fatalf("reading %s: %v", FileName, err)
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

func TestRunCreateAndMerge(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())

	if _, err := Run(cat, []Selection{{Name: "burp"}}, false); err != nil {
		t.Fatal(err)
	}
	if got := readServers(t); len(got) != 1 || got[0] != "burp" {
		t.Fatalf("after create: %v", got)
	}

	// Merge a second server, the first should remain.
	if _, err := Run(cat, []Selection{{Name: "chrome-devtools"}}, false); err != nil {
		t.Fatal(err)
	}
	if got := readServers(t); len(got) != 2 {
		t.Fatalf("after merge: %v", got)
	}
}

func TestRunConflict(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())

	if _, err := Run(cat, []Selection{{Name: "burp"}}, false); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(cat, []Selection{{Name: "burp"}}, false); err == nil {
		t.Fatal("expected conflict error without force")
	}
	if _, err := Run(cat, []Selection{{Name: "burp"}}, true); err != nil {
		t.Fatalf("force should overwrite: %v", err)
	}
}

func TestRunUnknown(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	if _, err := Run(cat, []Selection{{Name: "nope"}}, false); err == nil {
		t.Fatal("expected error for unknown server")
	}
}
