package catalog

import (
	"encoding/json"
	"fmt"
)

// TransportTypes lists the MCP server transport types mcpgen can build, most
// common first.
var TransportTypes = []string{"stdio", "sse", "http"}

// ServerSpec describes a custom MCP server to assemble into a config.
type ServerSpec struct {
	Type    string            // "stdio", "sse" or "http"
	Command string            // stdio: executable (e.g. "npx")
	Args    []string          // stdio: command arguments
	Env     map[string]string // stdio: environment variables (optional)
	URL     string            // sse/http: endpoint URL
	Headers map[string]string // sse/http: request headers (optional)
}

// Field order in these structs is the on-disk key order produced by json.Marshal.
type stdioConfig struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

type remoteConfig struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Config validates the spec and assembles its JSON config in conventional key
// order (type first).
func (s ServerSpec) Config() (json.RawMessage, error) {
	switch s.Type {
	case "stdio":
		if s.Command == "" {
			return nil, fmt.Errorf("a stdio server needs a command")
		}
		return json.Marshal(stdioConfig{Type: "stdio", Command: s.Command, Args: s.Args, Env: s.Env})
	case "sse", "http":
		if s.URL == "" {
			return nil, fmt.Errorf("a %s server needs a url", s.Type)
		}
		return json.Marshal(remoteConfig{Type: s.Type, URL: s.URL, Headers: s.Headers})
	default:
		return nil, fmt.Errorf("unknown server type %q (want one of %v)", s.Type, TransportTypes)
	}
}
