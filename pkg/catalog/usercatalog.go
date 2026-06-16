package catalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// entry is a catalog entry written to the user catalog, in on-disk key order.
type entry struct {
	Description string          `json:"description"`
	Config      json.RawMessage `json:"config"`
}

// userCatalogPath returns the path to the user's writable catalog,
// $XDG_CONFIG_HOME/mcpgen/servers.json (defaulting to ~/.config).
func userCatalogPath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "mcpgen", "servers.json"), nil
}

// SaveCustom adds (or updates) a server in the user catalog and returns the file
// path written. Existing top-level keys and server order are preserved.
func SaveCustom(name, description string, config json.RawMessage) (string, error) {
	path, err := userCatalogPath()
	if err != nil {
		return "", err
	}

	var top orderedObject
	if data, err := os.ReadFile(path); err == nil {
		if top, err = parseOrderedObject(data); err != nil {
			return "", fmt.Errorf("user catalog %s is not valid JSON: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

	var servers orderedObject
	if raw, ok := top.get("servers"); ok && len(raw) > 0 {
		if servers, err = parseOrderedObject(raw); err != nil {
			return "", fmt.Errorf("user catalog %s has an invalid \"servers\": %w", path, err)
		}
	}

	entryRaw, err := json.Marshal(entry{Description: description, Config: config})
	if err != nil {
		return "", err
	}
	servers.put(name, entryRaw)

	serversRaw, err := json.Marshal(servers)
	if err != nil {
		return "", err
	}
	top.put("servers", serversRaw)

	compact, err := json.Marshal(top)
	if err != nil {
		return "", err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, compact, "", "  "); err != nil {
		return "", err
	}
	pretty.WriteByte('\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, pretty.Bytes(), 0o644); err != nil {
		return "", err
	}
	return path, nil
}
