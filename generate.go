package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const mcpFileName = ".mcp.json"

// mcpFile mirrors the on-disk .mcp.json shape. Configs are kept as raw JSON so
// any server shape (stdio, sse, custom fields) round-trips with its original
// key order preserved.
type mcpFile struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// runGenerate selects servers from the catalog and writes them to ./.mcp.json,
// merging into an existing file and refusing to clobber existing servers unless
// force is set.
func runGenerate(cat *Catalog, names []string, force bool) error {
	if len(names) == 0 {
		return errors.New("no servers given — try `mcpgen list` to see what's available")
	}

	// Validate every requested name up front so we never write a partial result.
	var unknown []string
	for _, name := range names {
		if _, ok := cat.Servers[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		return fmt.Errorf("unknown server(s): %s\nrun `mcpgen list` to see available servers",
			strings.Join(unknown, ", "))
	}

	// Load existing .mcp.json if present, so we merge instead of replace.
	out := mcpFile{MCPServers: map[string]json.RawMessage{}}
	if data, err := os.ReadFile(mcpFileName); err == nil {
		if err := json.Unmarshal(data, &out); err != nil {
			return fmt.Errorf("existing %s is not valid JSON: %w", mcpFileName, err)
		}
		if out.MCPServers == nil {
			out.MCPServers = map[string]json.RawMessage{}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("reading %s: %w", mcpFileName, err)
	}

	var added, skipped []string
	for _, name := range names {
		if _, exists := out.MCPServers[name]; exists && !force {
			skipped = append(skipped, name)
			continue
		}
		out.MCPServers[name] = cat.Servers[name].Config
		added = append(added, name)
	}

	if len(skipped) > 0 {
		return fmt.Errorf("%s already defines: %s\nre-run with --force to overwrite",
			mcpFileName, strings.Join(skipped, ", "))
	}

	data, err := marshalMCPFile(out.MCPServers)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", mcpFileName, err)
	}
	if err := os.WriteFile(mcpFileName, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", mcpFileName, err)
	}

	sort.Strings(added)
	fmt.Printf("Wrote %s (%d server(s): %s)\n", mcpFileName, len(added), strings.Join(added, ", "))
	fmt.Println("Remember to replace any PLACEHOLDER values (API keys, paths, connection strings).")
	return nil
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
