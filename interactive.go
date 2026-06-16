package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

// interactiveSelect runs the no-args generate flow: pick servers from a numbered
// list, then optionally override each selected server's args.
func interactiveSelect(cat *Catalog) ([]selection, error) {
	names := cat.names()
	if len(names) == 0 {
		return nil, fmt.Errorf("catalog is empty")
	}
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Available MCP servers:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, name := range names {
		fmt.Fprintf(w, "  %d)\t%s\t%s\n", i+1, name, cat.Servers[name].Description)
	}
	w.Flush()

	fmt.Print("\nSelect servers (e.g. 1 3, 1-3, or 'all'): ")
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return nil, err
	}
	idx, err := parseSelection(line, len(names))
	if err != nil {
		return nil, err
	}

	var sels []selection
	for _, i := range idx {
		name := names[i]
		sel := selection{name: name}

		if cur, ok := serverArgs(cat.Servers[name].Config); ok {
			fmt.Printf("\nArgs for %q\n  current: %s\n", name, strings.Join(cur, " "))
			fmt.Print("  extra/override args (e.g. --browser-url=http://127.0.0.1:9333), blank to keep: ")
			argLine, _ := reader.ReadString('\n')
			if fields := strings.Fields(argLine); len(fields) > 0 {
				sel.extraArgs = fields
			}
		}
		sels = append(sels, sel)
	}
	return sels, nil
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
