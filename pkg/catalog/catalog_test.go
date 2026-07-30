package catalog

import (
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// isolateUserCatalog points XDG_CONFIG_HOME at an empty temp dir so tests don't
// pick up a real ~/.config/mcpgen/servers.json.
func isolateUserCatalog(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
}

// TestCatalogOrder pins the display order to servers.json (not alphabetical).
func TestCatalogOrder(t *testing.T) {
	isolateUserCatalog(t)
	cat, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"obsidian", "chrome-devtools", "mongodb", "figma-developer", "burp", "xquik"}
	if got := cat.Names(); !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %v, want %v", got, want)
	}
}

// TestUserCatalog covers saving a custom server and loading it back, merged
// after the bundled entries.
func TestUserCatalog(t *testing.T) {
	isolateUserCatalog(t)

	if _, err := SaveCustom("zz-tool", "My tool", json.RawMessage(`{"type":"stdio","command":"npx"}`)); err != nil {
		t.Fatal(err)
	}

	cat, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := cat.Servers["zz-tool"]
	if !ok {
		t.Fatal("custom server not loaded")
	}
	if !s.Custom {
		t.Error("custom server should be marked Custom")
	}
	if names := cat.Names(); names[len(names)-1] != "zz-tool" {
		t.Errorf("custom server should sort last, got %v", names)
	}

	// Saving the same name again updates rather than duplicates.
	if _, err := SaveCustom("zz-tool", "Updated", json.RawMessage(`{"type":"stdio","command":"node"}`)); err != nil {
		t.Fatal(err)
	}
	cat, _ = Load()
	if got := cat.Servers["zz-tool"].Description; got != "Updated" {
		t.Errorf("description = %q, want Updated", got)
	}
	count := 0
	for _, n := range cat.Names() {
		if n == "zz-tool" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("zz-tool appears %d times, want 1", count)
	}
}

func TestUserCatalogPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix permission bits")
	}
	isolateUserCatalog(t)

	path, err := SaveCustom("private", "Private", json.RawMessage(`{"type":"http","url":"https://example.com"}`))
	if err != nil {
		t.Fatal(err)
	}
	assertFilePermissions(t, path, 0o600)

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	assertFilePermissions(t, path, 0o644)

	if _, err := SaveCustom("private", "Updated", json.RawMessage(`{"type":"http","url":"https://example.com"}`)); err != nil {
		t.Fatal(err)
	}
	assertFilePermissions(t, path, 0o600)
}

func assertFilePermissions(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("user catalog permissions = %04o, want %04o", got, want)
	}
}

func TestXquikCatalogEntry(t *testing.T) {
	isolateUserCatalog(t)
	cat, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	server, ok := cat.Servers["xquik"]
	if !ok {
		t.Fatal("xquik server is missing")
	}
	var config struct {
		Type    string            `json:"type"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	}
	if err := json.Unmarshal(server.Config, &config); err != nil {
		t.Fatal(err)
	}
	if config.Type != "http" || config.URL != "https://xquik.com/mcp" {
		t.Errorf("xquik config = %+v", config)
	}
	if got := config.Headers["X-API-Key"]; got != "YOUR_XQUIK_API_KEY" {
		t.Errorf("xquik API key placeholder = %q", got)
	}
}

// TestCatalogValid doubles as the CI guard for servers.json: every entry must
// have a description and a config that is a JSON object.
func TestCatalogValid(t *testing.T) {
	isolateUserCatalog(t)
	cat, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cat.Servers) == 0 {
		t.Fatal("catalog is empty")
	}
	for name, s := range cat.Servers {
		if strings.TrimSpace(s.Description) == "" {
			t.Errorf("%q: empty description", name)
		}
		if len(s.Config) == 0 {
			t.Errorf("%q: missing config", name)
			continue
		}
		if _, err := parseOrderedObject(s.Config); err != nil {
			t.Errorf("%q: config is not a JSON object: %v", name, err)
		}
	}
}

// TestCatalogNoObviousSecrets guards against committing real credentials in the
// catalog instead of placeholders.
func TestCatalogNoObviousSecrets(t *testing.T) {
	banned := []string{"figd_", "sk-", "ghp_"}
	for _, b := range banned {
		if strings.Contains(string(catalogBytes), b) {
			t.Errorf("servers.json appears to contain a real secret (%q); use a placeholder", b)
		}
	}
}
