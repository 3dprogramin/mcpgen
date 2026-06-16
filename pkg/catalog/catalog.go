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

// Catalog is the full set of known servers, keyed by server name. order holds
// the server names in the order they appear in servers.json.
type Catalog struct {
	Servers map[string]Server `json:"servers"`
	order   []string
}

// Load parses the embedded servers.json, preserving the file's server order.
func Load() (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(catalogBytes, &c); err != nil {
		return nil, fmt.Errorf("parsing embedded catalog: %w", err)
	}
	order, err := serverOrder(catalogBytes)
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
