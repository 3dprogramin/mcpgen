// Package style provides terminal coloring and the ASCII banner. All helpers
// degrade to plain text when stdout is not a terminal or NO_COLOR is set.
package style

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// enabled is true when stdout is a terminal and NO_COLOR is unset, so piped or
// redirected output stays free of escape codes.
var enabled = os.Getenv("NO_COLOR") == "" && term.IsTerminal(int(os.Stdout.Fd()))

func colorize(code, s string) string {
	if !enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// Bold returns s in bold.
func Bold(s string) string { return colorize("1", s) }

// Dim returns s dimmed.
func Dim(s string) string { return colorize("2", s) }

// Red returns s in red.
func Red(s string) string { return colorize("31", s) }

// Green returns s in green.
func Green(s string) string { return colorize("32", s) }

// Yellow returns s in yellow.
func Yellow(s string) string { return colorize("33", s) }

// Cyan returns s in cyan.
func Cyan(s string) string { return colorize("36", s) }

func boldCyan(s string) string { return colorize("1;36", s) }

// bannerLines is "mcpgen" rendered in a figlet-style font.
var bannerLines = []string{
	" _ __ ___   ___ _ __   __ _  ___ _ __  ",
	"| '_ ` _ \\ / __| '_ \\ / _` |/ _ \\ '_ \\ ",
	"| | | | | | (__| |_) | (_| |  __/ | | |",
	"|_| |_| |_|\\___| .__/ \\__, |\\___|_| |_|",
	"               |_|    |___/            ",
}

// PrintBanner shows the ASCII banner, but only on an interactive terminal so it
// never pollutes piped output.
func PrintBanner() {
	if !enabled {
		return
	}
	for _, l := range bannerLines {
		fmt.Println(boldCyan(l))
	}
	fmt.Println(Dim("  generate .mcp.json from a catalog of MCP servers"))
	fmt.Println()
}
