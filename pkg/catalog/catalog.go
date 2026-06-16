// Package catalog holds the bundled list of known MCP servers (embedded from
// servers.json) and the logic for customizing a server's args.
package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

//go:embed servers.json
var catalogBytes []byte

// Server is one entry in the catalog: a human-readable description plus the raw
// MCP server config that gets written to .mcp.json. Custom is true for entries
// loaded from the user's catalog rather than the bundled one.
type Server struct {
	Description string          `json:"description"`
	Config      json.RawMessage `json:"config"`
	Custom      bool            `json:"-"`
}

// Catalog is the full set of known servers, keyed by server name. order holds
// the server names in display order (bundled order, then user additions).
type Catalog struct {
	Servers map[string]Server `json:"servers"`
	order   []string
}

// Load parses the bundled catalog and merges the user catalog
// (~/.config/mcpgen/servers.json) on top, preserving display order.
func Load() (*Catalog, error) {
	c, err := parseCatalog(catalogBytes)
	if err != nil {
		return nil, fmt.Errorf("parsing embedded catalog: %w", err)
	}

	path, err := userCatalogPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, fmt.Errorf("reading user catalog %s: %w", path, err)
	}
	user, err := parseCatalog(data)
	if err != nil {
		return nil, fmt.Errorf("user catalog %s: %w", path, err)
	}
	for _, name := range user.order {
		s := user.Servers[name]
		s.Custom = true
		if _, exists := c.Servers[name]; !exists {
			c.order = append(c.order, name)
		}
		c.Servers[name] = s
	}
	return c, nil
}

// parseCatalog parses a catalog file's bytes into a Catalog with file order.
func parseCatalog(data []byte) (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Servers == nil {
		c.Servers = map[string]Server{}
	}
	order, err := serverOrder(data)
	if err != nil {
		return nil, fmt.Errorf("reading catalog order: %w", err)
	}
	c.order = order
	return &c, nil
}

// serverOrder extracts the server names under "servers" in file order.
func serverOrder(data []byte) ([]string, error) {
	top, err := parseOrderedObject(data)
	if err != nil {
		return nil, err
	}
	raw, ok := top.get("servers")
	if !ok {
		return nil, nil
	}
	servers, err := parseOrderedObject(raw)
	if err != nil {
		return nil, err
	}
	order := make([]string, len(servers))
	for i, m := range servers {
		order[i] = m.Key
	}
	return order, nil
}

// Names returns the catalog server names in the order defined by servers.json,
// falling back to sorted order if the file order is unavailable.
func (c *Catalog) Names() []string {
	if len(c.order) == len(c.Servers) {
		return append([]string(nil), c.order...)
	}
	out := make([]string, 0, len(c.Servers))
	for name := range c.Servers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
