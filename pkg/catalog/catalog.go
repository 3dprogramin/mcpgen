// Package catalog holds the bundled list of known MCP servers (embedded from
// servers.json) and the logic for customizing a server's args.
package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed servers.json
var catalogBytes []byte

// Server is one entry in the bundled catalog: a human-readable description plus
// the raw MCP server config that gets written to .mcp.json.
type Server struct {
	Description string          `json:"description"`
	Config      json.RawMessage `json:"config"`
}

// Catalog is the full set of known servers, keyed by server name.
type Catalog struct {
	Servers map[string]Server `json:"servers"`
}

// Load parses the embedded servers.json.
func Load() (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(catalogBytes, &c); err != nil {
		return nil, fmt.Errorf("parsing embedded catalog: %w", err)
	}
	return &c, nil
}

// Names returns the catalog server names in sorted order.
func (c *Catalog) Names() []string {
	out := make([]string, 0, len(c.Servers))
	for name := range c.Servers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
