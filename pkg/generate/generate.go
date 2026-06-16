// Package generate writes selected catalog servers into a project-local
// .mcp.json, merging with any existing file.
package generate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/3dprogramin/mcpgen/pkg/catalog"
)

// FileName is the project-local file mcpgen reads and writes.
const FileName = ".mcp.json"

// mcpFile mirrors the on-disk .mcp.json shape. Configs are kept as raw JSON so
// any server shape (stdio, sse, custom fields) round-trips with its original
// key order preserved.
type mcpFile struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// Selection is a chosen server plus any args to apply to its config. When
// ReplaceAll is set, OverrideArgs replace the config's args entirely; otherwise
// they are merged in.
type Selection struct {
	Name         string
	OverrideArgs []string
	ReplaceAll   bool
}

// Run writes the selected servers to ./.mcp.json, merging into an existing file
// and refusing to clobber existing servers unless force is set. It returns the
// sorted names that were written.
func Run(cat *catalog.Catalog, sels []Selection, force bool) ([]string, error) {
	if len(sels) == 0 {
		return nil, errors.New("no servers given - try `mcpgen list` to see what's available")
	}

	// Validate every requested name up front so we never write a partial result.
	var unknown []string
	for _, sel := range sels {
		if _, ok := cat.Servers[sel.Name]; !ok {
			unknown = append(unknown, sel.Name)
		}
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unknown server(s): %s\nrun `mcpgen list` to see available servers",
			strings.Join(unknown, ", "))
	}

	// Load existing .mcp.json if present, so we merge instead of replace.
	out := mcpFile{MCPServers: map[string]json.RawMessage{}}
	if data, err := os.ReadFile(FileName); err == nil {
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("existing %s is not valid JSON: %w", FileName, err)
		}
		if out.MCPServers == nil {
			out.MCPServers = map[string]json.RawMessage{}
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading %s: %w", FileName, err)
	}

	var added, skipped []string
	for _, sel := range sels {
		if _, exists := out.MCPServers[sel.Name]; exists && !force {
			skipped = append(skipped, sel.Name)
			continue
		}
		cfg, err := catalog.ApplyArgs(cat.Servers[sel.Name].Config, sel.OverrideArgs, sel.ReplaceAll)
		if err != nil {
			return nil, fmt.Errorf("server %q: %w", sel.Name, err)
		}
		out.MCPServers[sel.Name] = cfg
		added = append(added, sel.Name)
	}

	if len(skipped) > 0 {
		return nil, fmt.Errorf("%s already defines: %s\nre-run with --force to overwrite",
			FileName, strings.Join(skipped, ", "))
	}

	data, err := marshalMCPFile(out.MCPServers)
	if err != nil {
		return nil, fmt.Errorf("encoding %s: %w", FileName, err)
	}
	if err := os.WriteFile(FileName, data, 0o644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", FileName, err)
	}

	sort.Strings(added)
	return added, nil
}

// marshalMCPFile renders the .mcp.json with 2-space indentation, servers sorted
// by name, and each server's config indented in place so its original key order
// is preserved.
func marshalMCPFile(servers map[string]json.RawMessage) ([]byte, error) {
	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}
	sort.Strings(names)

	var b bytes.Buffer
	b.WriteString("{\n  \"mcpServers\": {")
	for i, name := range names {
		if i > 0 {
			b.WriteByte(',')
		}
		key, err := json.Marshal(name)
		if err != nil {
			return nil, err
		}
		b.WriteString("\n    ")
		b.Write(key)
		b.WriteString(": ")

		var cfg bytes.Buffer
		// Each config sits two levels deep, so subsequent lines get a 4-space prefix.
		if err := json.Indent(&cfg, servers[name], "    ", "  "); err != nil {
			return nil, fmt.Errorf("config for %q is not valid JSON: %w", name, err)
		}
		b.Write(cfg.Bytes())
	}
	if len(names) > 0 {
		b.WriteString("\n  ")
	}
	b.WriteString("}\n}\n")
	return b.Bytes(), nil
}
