// Command mcpgen generates project-local .mcp.json files from a bundled catalog
// of MCP servers, so you never have to copy-paste server configs by hand again.
package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

const usage = `mcpgen — generate .mcp.json files from a catalog of MCP servers

Usage:
  mcpgen list                       List available servers
  mcpgen generate <name> [name...]  Add servers to ./.mcp.json
  mcpgen help                       Show this help

Flags (generate):
  -f, --force   Overwrite servers that already exist in ./.mcp.json

Examples:
  mcpgen list
  mcpgen generate chrome-devtools mongodb
  mcpgen generate burp --force
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error: "+err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(usage)
		return nil
	}

	cat, err := loadCatalog()
	if err != nil {
		return err
	}

	switch args[0] {
	case "list", "ls", "-l", "--list":
		return runList(cat)
	case "generate", "gen", "g":
		names, force := parseGenerateArgs(args[1:])
		return runGenerate(cat, names, force)
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		fmt.Print(usage)
		return fmt.Errorf("unknown command: %q", args[0])
	}
}

// parseGenerateArgs splits generate arguments into server names and the force
// flag, accepting the flag in any position.
func parseGenerateArgs(args []string) (names []string, force bool) {
	for _, a := range args {
		switch a {
		case "-f", "--force":
			force = true
		default:
			names = append(names, a)
		}
	}
	return names, force
}

func runList(cat *Catalog) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION")
	for _, name := range cat.names() {
		fmt.Fprintf(w, "%s\t%s\n", name, cat.Servers[name].Description)
	}
	return w.Flush()
}
