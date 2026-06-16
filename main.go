// Command mcpgen generates project-local .mcp.json files from a bundled catalog
// of MCP servers, so you never have to copy-paste server configs by hand again.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"text/tabwriter"

	"github.com/3dprogramin/mcpgen/pkg/catalog"
	"github.com/3dprogramin/mcpgen/pkg/generate"
	"github.com/3dprogramin/mcpgen/pkg/style"
	"github.com/3dprogramin/mcpgen/pkg/tui"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "dev"

const usage = `mcpgen - generate .mcp.json files from a catalog of MCP servers

Usage:
  mcpgen list                         List available servers
  mcpgen generate                     Pick servers interactively
  mcpgen generate <name> [name...]    Add servers to ./.mcp.json
  mcpgen generate <name> [args...]    Override/append args on a single server
  mcpgen add                          Build a custom server interactively
  mcpgen version                      Show the version
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
		fmt.Fprintln(os.Stderr, style.Red("error:")+" "+err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		style.PrintBanner()
		fmt.Print(usage)
		return nil
	}

	switch args[0] {
	case "version", "-v", "--version":
		fmt.Printf("mcpgen %s\n", versionString())
		return nil
	case "help", "-h", "--help":
		style.PrintBanner()
		fmt.Print(usage)
		return nil
	}

	cat, err := catalog.Load()
	if err != nil {
		return err
	}

	switch args[0] {
	case "list", "ls", "-l", "--list":
		style.PrintBanner()
		return runList(cat)
	case "generate", "gen", "g":
		style.PrintBanner()
		return runGenerate(cat, args[1:])
	case "add":
		style.PrintBanner()
		return runAdd(args[1:])
	default:
		style.PrintBanner()
		fmt.Print(usage)
		return fmt.Errorf("unknown command: %q", args[0])
	}
}

func runGenerate(cat *catalog.Catalog, args []string) error {
	sels, force, err := parseGenerateArgs(args)
	if err != nil {
		return err
	}
	if len(sels) == 0 {
		sels, err = tui.Select(cat)
		if err != nil {
			return err
		}
	}

	added, err := generate.Run(cat, sels, force)
	if err != nil {
		return err
	}

	printWrote(added)
	fmt.Println(style.Yellow("Remember to replace any PLACEHOLDER values (API keys, paths, connection strings)."))
	return nil
}

// runAdd builds a custom server interactively and writes it to ./.mcp.json.
func runAdd(args []string) error {
	force := false
	for _, a := range args {
		if a == "-f" || a == "--force" {
			force = true
		}
	}

	cs, err := tui.AddCustom()
	if err != nil {
		return err
	}
	added, err := generate.Write(map[string]json.RawMessage{cs.Name: cs.Config}, force)
	if err != nil {
		return err
	}
	printWrote(added)

	if cs.SaveToCatalog {
		path, err := catalog.SaveCustom(cs.Name, cs.Description, cs.Config)
		if err != nil {
			return fmt.Errorf("saving to catalog: %w", err)
		}
		fmt.Printf("%s Saved %q to your catalog (%s)\n", style.Green("✓"), cs.Name, path)
	}
	return nil
}

func printWrote(added []string) {
	fmt.Printf("%s Wrote %s (%d server(s): %s)\n",
		style.Green("✓"), style.Bold(generate.FileName), len(added), strings.Join(added, ", "))
}

// parseGenerateArgs splits generate arguments into server selections and the
// force flag. Tokens before "--" that start with "-" (other than the force flag)
// and everything after "--" are treated as arg overrides; they apply to the one
// selected server. Bare tokens are server names.
func parseGenerateArgs(args []string) (sels []generate.Selection, force bool, err error) {
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
		sels = append(sels, generate.Selection{Name: n})
	}

	if len(extra) > 0 {
		if len(sels) != 1 {
			return nil, false, fmt.Errorf(
				"arg overrides apply to a single server, but %d were given; "+
					"run interactively or one server at a time", len(sels))
		}
		sels[0].OverrideArgs = extra
	}
	return sels, force, nil
}

// versionString resolves the version, preferring the ldflags value and falling
// back to the module version recorded by `go install`.
func versionString() string {
	if version != "dev" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return version
}

func runList(cat *catalog.Catalog) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION")
	for _, name := range cat.Names() {
		s := cat.Servers[name]
		desc := s.Description
		if s.Custom {
			desc += style.Dim(" (custom)")
		}
		fmt.Fprintf(w, "%s\t%s\n", name, desc)
	}
	return w.Flush()
}
