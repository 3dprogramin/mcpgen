// Command mcpgen generates project-local .mcp.json files from a bundled catalog
// of MCP servers, so you never have to copy-paste server configs by hand again.
package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const usage = `mcpgen — generate .mcp.json files from a catalog of MCP servers

Usage:
  mcpgen list                         List available servers
  mcpgen generate                     Pick servers interactively
  mcpgen generate <name> [name...]    Add servers to ./.mcp.json
  mcpgen generate <name> [args...]    Override/append args on a single server
  mcpgen help                         Show this help

Flags (generate):
  -f, --force   Overwrite servers that already exist in ./.mcp.json

Arg overrides (single server only):
  Pass extra args after the server name. A "--flag=value" replaces a matching
  "--flag=..." already in the config; anything else is appended. Use "--" to
  pass args that don't start with "-".

Examples:
  mcpgen list
  mcpgen generate                                       # interactive
  mcpgen generate chrome-devtools mongodb
  mcpgen generate chrome-devtools --browser-url=http://127.0.0.1:9333
  mcpgen generate burp --force
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, red("error:")+" "+err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printBanner()
		fmt.Print(usage)
		return nil
	}

	cat, err := loadCatalog()
	if err != nil {
		return err
	}

	switch args[0] {
	case "list", "ls", "-l", "--list":
		printBanner()
		return runList(cat)
	case "generate", "gen", "g":
		printBanner()
		sels, force, err := parseGenerateArgs(args[1:])
		if err != nil {
			return err
		}
		if len(sels) == 0 {
			sels, err = interactiveSelect(cat)
			if err != nil {
				return err
			}
		}
		return runGenerate(cat, sels, force)
	case "help", "-h", "--help":
		printBanner()
		fmt.Print(usage)
		return nil
	default:
		printBanner()
		fmt.Print(usage)
		return fmt.Errorf("unknown command: %q", args[0])
	}
}

// parseGenerateArgs splits generate arguments into server selections and the
// force flag. Tokens before "--" that start with "-" (other than the force flag)
// and everything after "--" are treated as arg overrides; they apply to the one
// selected server. Bare tokens are server names.
func parseGenerateArgs(args []string) (sels []selection, force bool, err error) {
	var names, extra []string
	afterSep := false
	for _, a := range args {
		switch {
		case afterSep:
			extra = append(extra, a)
		case a == "--":
			afterSep = true
		case a == "-f" || a == "--force":
			force = true
		case strings.HasPrefix(a, "-"):
			extra = append(extra, a)
		default:
			names = append(names, a)
		}
	}

	for _, n := range names {
		sels = append(sels, selection{name: n})
	}

	if len(extra) > 0 {
		if len(sels) != 1 {
			return nil, false, fmt.Errorf(
				"arg overrides apply to a single server, but %d were given; "+
					"run interactively or one server at a time", len(sels))
		}
		sels[0].overrideArgs = extra
	}
	return sels, force, nil
}

func runList(cat *Catalog) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION")
	for _, name := range cat.names() {
		fmt.Fprintf(w, "%s\t%s\n", name, cat.Servers[name].Description)
	}
	return w.Flush()
}
