package main

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// colorEnabled is true when stdout is a terminal and NO_COLOR is unset, so piped
// or redirected output stays free of escape codes.
var colorEnabled = os.Getenv("NO_COLOR") == "" && term.IsTerminal(int(os.Stdout.Fd()))

func colorize(code, s string) string {
	if !colorEnabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func bold(s string) string     { return colorize("1", s) }
func dim(s string) string      { return colorize("2", s) }
func red(s string) string      { return colorize("31", s) }
func green(s string) string    { return colorize("32", s) }
func yellow(s string) string   { return colorize("33", s) }
func cyan(s string) string     { return colorize("36", s) }
func boldCyan(s string) string { return colorize("1;36", s) }

// bannerLines is "mcpgen" rendered in a figlet-style font.
var bannerLines = []string{
	" _ __ ___   ___ _ __   __ _  ___ _ __  ",
	"| '_ ` _ \\ / __| '_ \\ / _` |/ _ \\ '_ \\ ",
	"| | | | | | (__| |_) | (_| |  __/ | | |",
	"|_| |_| |_|\\___| .__/ \\__, |\\___|_| |_|",
	"               |_|    |___/            ",
}

// printBanner shows the ASCII banner, but only on an interactive terminal so it
// never pollutes piped output.
func printBanner() {
	if !colorEnabled {
		return
	}
	for _, l := range bannerLines {
		fmt.Println(boldCyan(l))
	}
	fmt.Println(dim("  generate .mcp.json from a catalog of MCP servers"))
	fmt.Println()
}
