package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"unicode/utf8"

	"golang.org/x/term"
)

// stdinIsTTY reports whether both stdin and stdout are interactive terminals,
// which is required for the arrow-key picker.
func stdinIsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// pickServers shows an interactive checkbox list and returns the selected
// indices. Up/down (or k/j) move, space toggles, a toggles all, enter confirms,
// q or ctrl-c aborts.
func pickServers(reader *bufio.Reader, names, descs []string) ([]int, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, oldState)

	out := os.Stdout
	selected := make([]bool, len(names))
	active := 0

	nameW := 0
	for _, n := range names {
		if len(n) > nameW {
			nameW = len(n)
		}
	}

	fmt.Fprint(out, "Select MCP servers — ↑/↓ move · space toggle · a all · enter confirm · q quit\r\n")

	first := true
	draw := func() {
		if !first {
			fmt.Fprintf(out, "\x1b[%dA", len(names))
		}
		first = false
		width := 80
		if w, _, e := term.GetSize(fd); e == nil && w > 0 {
			width = w
		}
		for i, n := range names {
			pointer := "  "
			if i == active {
				pointer = "> "
			}
			box := "[ ]"
			if selected[i] {
				box = "[x]"
			}
			line := fmt.Sprintf("%s%s %-*s  %s", pointer, box, nameW, n, descs[i])
			fmt.Fprintf(out, "\r\x1b[2K%s\r\n", truncate(line, width-1))
		}
	}

	move := func(delta int) {
		active = (active + delta + len(names)) % len(names)
	}

	draw()
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		switch b {
		case 3, 'q': // ctrl-c / q
			fmt.Fprint(out, "\r\n")
			return nil, errors.New("aborted")
		case '\r', '\n':
			fmt.Fprint(out, "\r\n")
			var idx []int
			for i, s := range selected {
				if s {
					idx = append(idx, i)
				}
			}
			return idx, nil
		case ' ':
			selected[active] = !selected[active]
		case 'a':
			allOn := true
			for _, s := range selected {
				if !s {
					allOn = false
					break
				}
			}
			for i := range selected {
				selected[i] = !allOn
			}
		case 'k':
			move(-1)
		case 'j':
			move(1)
		case 0x1b: // escape sequence, e.g. arrow keys "\x1b[A"
			if b2, _ := reader.ReadByte(); b2 != '[' {
				continue
			}
			switch b3, _ := reader.ReadByte(); b3 {
			case 'A':
				move(-1)
			case 'B':
				move(1)
			}
		}
		draw()
	}
}

// truncate shortens s to at most max runes, adding an ellipsis when cut.
func truncate(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	r := []rune(s)
	if max == 1 {
		return string(r[:1])
	}
	return string(r[:max-1]) + "…"
}
