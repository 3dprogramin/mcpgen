package tui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/3dprogramin/mcpgen/pkg/catalog"
	"github.com/3dprogramin/mcpgen/pkg/style"
)

// CustomServer is the result of the interactive add flow.
type CustomServer struct {
	Name          string
	Config        json.RawMessage
	SaveToCatalog bool   // also persist to the user catalog
	Description   string // used when SaveToCatalog is true
}

// AddCustom interactively builds a custom MCP server and asks whether to also
// save it to the user catalog for reuse.
func AddCustom() (CustomServer, error) {
	r := bufio.NewReader(os.Stdin)

	name := readLine(r, style.Bold("Server name")+": ")
	if name == "" {
		return CustomServer{}, fmt.Errorf("a name is required")
	}
	if strings.ContainsAny(name, " \t") {
		return CustomServer{}, fmt.Errorf("server name must not contain spaces")
	}

	spec := catalog.ServerSpec{Type: chooseType(r)}
	switch spec.Type {
	case "stdio":
		spec.Command = readDefault(r, style.Bold("Command"), "npx")
		spec.Args = strings.Fields(readLine(r, style.Dim("Args (space-separated, blank for none): ")))
		spec.Env = readKV(r, style.Dim("Env vars (KEY=VALUE ..., blank for none): "))
	default: // sse, http
		spec.URL = readLine(r, style.Bold("URL")+": ")
		spec.Headers = readKV(r, style.Dim("Headers (KEY=VALUE ..., blank for none): "))
	}

	cfg, err := spec.Config()
	if err != nil {
		return CustomServer{}, err
	}

	out := CustomServer{Name: name, Config: cfg}
	if confirm(r, fmt.Sprintf("Save %q to your catalog for reuse?", name)) {
		out.SaveToCatalog = true
		out.Description = readDefault(r, style.Bold("Description"), name)
	}

	fmt.Println() // separate the prompts from the result message
	return out, nil
}

// confirm asks a yes/no question, defaulting to no.
func confirm(r *bufio.Reader, question string) bool {
	answer := strings.ToLower(readLine(r, question+" [y/N]: "))
	return answer == "y" || answer == "yes"
}

// chooseType prompts for a transport type, defaulting to the first one (stdio).
func chooseType(r *bufio.Reader) string {
	fmt.Println(style.Bold("Type:"))
	for i, t := range catalog.TransportTypes {
		suffix := ""
		if i == 0 {
			suffix = style.Dim(" (default)")
		}
		fmt.Printf("  %d) %s%s\n", i+1, t, suffix)
	}
	for {
		line := readLine(r, "Choose [1]: ")
		if line == "" {
			return catalog.TransportTypes[0]
		}
		if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(catalog.TransportTypes) {
			return catalog.TransportTypes[n-1]
		}
		fmt.Println(style.Dim("  please enter a number from the list"))
	}
}

func readLine(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func readDefault(r *bufio.Reader, label, def string) string {
	if v := readLine(r, label+" ["+def+"]: "); v != "" {
		return v
	}
	return def
}

func readKV(r *bufio.Reader, prompt string) map[string]string {
	line := readLine(r, prompt)
	if line == "" {
		return nil
	}
	m := map[string]string{}
	for _, tok := range strings.Fields(line) {
		if k, v, ok := strings.Cut(tok, "="); ok && k != "" {
			m[k] = v
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}
