// Package tui implements the interactive server picker and arg prompts.
package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/3dprogramin/mcpgen/pkg/catalog"
	"github.com/3dprogramin/mcpgen/pkg/generate"
	"github.com/3dprogramin/mcpgen/pkg/style"
)

// Select runs the no-args generate flow: pick servers (arrow-key checkbox on a
// TTY, numbered prompt otherwise), then optionally override each selected
// server's args.
func Select(cat *catalog.Catalog) ([]generate.Selection, error) {
	names := cat.Names()
	if len(names) == 0 {
		return nil, fmt.Errorf("catalog is empty")
	}
	reader := bufio.NewReader(os.Stdin)

	descs := make([]string, len(names))
	for i, n := range names {
		descs[i] = cat.Servers[n].Description
	}
	idx, err := pick(reader, "Select servers to add", names, descs)
	if err != nil {
		return nil, err
	}
	if len(idx) == 0 {
		return nil, fmt.Errorf("no servers selected")
	}

	var sels []generate.Selection
	for _, i := range idx {
		name := names[i]
		sel := generate.Selection{Name: name}

		if cur, ok := catalog.ServerArgs(cat.Servers[name].Config); ok {
			fmt.Printf("\n%s %s\n", style.Bold("Args for"), style.Bold(`"`+name+`"`))
			fmt.Printf("  %s %s\n", style.Dim("current:"), strings.Join(cur, " "))
			fmt.Printf("  %s ", style.Dim("new args (replaces all, blank to keep):"))
			argLine, _ := reader.ReadString('\n')
			if fields := strings.Fields(argLine); len(fields) > 0 {
				sel.OverrideArgs = fields
				sel.ReplaceAll = true
			}
		}
		sels = append(sels, sel)
	}
	fmt.Println() // separate the prompts from the result message
	return sels, nil
}

// Pick shows the filterable checkbox picker (or numbered fallback) for an
// arbitrary list and returns the selected indices. Reusable outside the catalog
// flow (e.g. choosing servers to remove).
func Pick(title string, names, descs []string) ([]int, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("nothing to choose from")
	}
	return pick(bufio.NewReader(os.Stdin), title, names, descs)
}

// pick branches between the TTY checkbox picker and the numbered fallback.
func pick(reader *bufio.Reader, title string, names, descs []string) ([]int, error) {
	if stdinIsTTY() {
		return pickServers(reader, title, names, descs)
	}
	return numberedSelect(reader, title, names, descs)
}

// numberedSelect is the non-TTY fallback: print a numbered list and read a
// selection line.
func numberedSelect(reader *bufio.Reader, title string, names, descs []string) ([]int, error) {
	fmt.Println(title + ":")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, name := range names {
		fmt.Fprintf(w, "  %d)\t%s\t%s\n", i+1, name, descs[i])
	}
	w.Flush()

	fmt.Print("\nSelect (e.g. 1 3, 1-3, or 'all'): ")
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return nil, err
	}
	return parseSelection(line, len(names))
}

// parseSelection turns "1 3", "1-3", "all" (also comma-separated) into a sorted,
// de-duplicated list of 0-based indices, validated against count.
func parseSelection(line string, count int) ([]int, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("nothing selected")
	}
	if strings.EqualFold(line, "all") {
		all := make([]int, count)
		for i := range all {
			all[i] = i
		}
		return all, nil
	}

	seen := map[int]bool{}
	var out []int
	add := func(n int) error {
		if n < 1 || n > count {
			return fmt.Errorf("%d is out of range (1-%d)", n, count)
		}
		if !seen[n-1] {
			seen[n-1] = true
			out = append(out, n-1)
		}
		return nil
	}

	fields := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == ',' })
	for _, f := range fields {
		if lo, hi, ok := strings.Cut(f, "-"); ok {
			a, err1 := strconv.Atoi(strings.TrimSpace(lo))
			b, err2 := strconv.Atoi(strings.TrimSpace(hi))
			if err1 != nil || err2 != nil || a > b {
				return nil, fmt.Errorf("invalid range %q", f)
			}
			for n := a; n <= b; n++ {
				if err := add(n); err != nil {
					return nil, err
				}
			}
			continue
		}
		n, err := strconv.Atoi(f)
		if err != nil {
			return nil, fmt.Errorf("invalid selection %q", f)
		}
		if err := add(n); err != nil {
			return nil, err
		}
	}
	return out, nil
}
