package main

import (
	"strings"
	"testing"
)

// TestCatalogValid doubles as the CI guard for servers.json: every entry must
// have a description and a config that is a JSON object.
func TestCatalogValid(t *testing.T) {
	cat, err := loadCatalog()
	if err != nil {
		t.Fatalf("loadCatalog: %v", err)
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
