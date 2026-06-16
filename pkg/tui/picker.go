package tui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"

	"github.com/3dprogramin/mcpgen/pkg/style"
)

// stdinIsTTY reports whether both stdin and stdout are interactive terminals,
// which is required for the arrow-key picker.
func stdinIsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// matchServers returns the indices of servers whose name or description contains
// query (case-insensitive). An empty query matches everything.
func matchServers(names, descs []string, query string) []int {
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]int, 0, len(names))
	for i := range names {
		if q == "" ||
			strings.Contains(strings.ToLower(names[i]), q) ||
			strings.Contains(strings.ToLower(descs[i]), q) {
			out = append(out, i)
		}
	}
	return out
}

const pickerHint = "type to filter · ↑/↓ move · space toggle · ^A all · enter confirm · esc quit"

// pickServers shows an interactive, filterable checkbox list and returns the
// selected indices (into names). Typing filters; ↑/↓ move; space toggles; ctrl-a
// toggles all matches; enter confirms; esc / ctrl-c aborts.
func pickServers(reader *bufio.Reader, title string, names, descs []string) ([]int, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, oldState)

	out := os.Stdout
	selected := make([]bool, len(names))
	query := ""
	filtered := matchServers(names, descs, query)
	active := 0 // index into filtered
	offset := 0 // top of the visible window into filtered

	nameW := 0
	for _, n := range names {
		if len(n) > nameW {
			nameW = len(n)
		}
	}
	const prefixW = 2 + 3 + 1 // "> " + "[ ]" + " "

	first := true
	prevLines := 0
	draw := func() {
		if !first {
			fmt.Fprintf(out, "\x1b[%dA\x1b[0J", prevLines) // up to top of block, clear below
		}
		first = false

		width, height := 80, 24
		if w, h, e := term.GetSize(fd); e == nil {
			if w > 0 {
				width = w
			}
			if h > 0 {
				height = h
			}
		}
		// Reserve header lines (optional title + Filter + hint) + 1 status line,
		// plus a spare so the final newline never scrolls the block out of place.
		header := 3
		if title != "" {
			header = 4
		}
		maxRows := height - header - 1
		if maxRows < 1 {
			maxRows = 1
		}

		// Keep active in range and scrolled into view.
		if active >= len(filtered) {
			active = len(filtered) - 1
		}
		if active < 0 {
			active = 0
		}
		if active < offset {
			offset = active
		}
		if active >= offset+maxRows {
			offset = active - maxRows + 1
		}
		if offset < 0 {
			offset = 0
		}

		lines := 0
		if title != "" {
			fmt.Fprintf(out, "\r\x1b[2K%s\r\n", style.Bold(title))
			lines++
		}
		fmt.Fprintf(out, "\r\x1b[2K%s %s\r\n", style.Bold("Filter:"), query+"▌")
		lines++
		fmt.Fprintf(out, "\r\x1b[2K%s\r\n", style.Dim(truncate(pickerHint, width-1)))
		lines++

		if len(filtered) == 0 {
			fmt.Fprintf(out, "\r\x1b[2K  %s\r\n", style.Dim("(no matches)"))
			lines++
		}
		end := offset + maxRows
		if end > len(filtered) {
			end = len(filtered)
		}
		for vi := offset; vi < end; vi++ {
			i := filtered[vi]
			pointer := "  "
			if vi == active {
				pointer = style.Cyan("> ")
			}
			box := "[ ]"
			if selected[i] {
				box = style.Green("[x]")
			}
			body := fmt.Sprintf("%-*s  %s", nameW, names[i], descs[i])
			fmt.Fprintf(out, "\r\x1b[2K%s%s %s\r\n", pointer, box, truncate(body, width-1-prefixW))
			lines++
		}

		fmt.Fprintf(out, "\r\x1b[2K%s\r\n", style.Dim(statusLine(selected, filtered, offset, end)))
		lines++
		prevLines = lines
	}

	refilter := func() {
		filtered = matchServers(names, descs, query)
		active, offset = 0, 0
	}
	move := func(d int) {
		if len(filtered) > 0 {
			active = (active + d + len(filtered)) % len(filtered)
		}
	}

	draw()
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		switch {
		case b == 3: // ctrl-c
			fmt.Fprint(out, "\r\n")
			return nil, errors.New("aborted")
		case b == '\r' || b == '\n':
			fmt.Fprint(out, "\r\n")
			return collectSelected(selected), nil
		case b == 127 || b == 8: // backspace
			if query != "" {
				r := []rune(query)
				query = string(r[:len(r)-1])
				refilter()
			}
		case b == 1: // ctrl-a: toggle all currently filtered
			allOn := true
			for _, i := range filtered {
				if !selected[i] {
					allOn = false
					break
				}
			}
			for _, i := range filtered {
				selected[i] = !allOn
			}
		case b == ' ':
			if len(filtered) > 0 {
				selected[filtered[active]] = !selected[filtered[active]]
			}
		case b == 0x1b: // escape sequence (arrows) or lone esc (abort)
			b2, err := reader.ReadByte()
			if err != nil || b2 != '[' {
				fmt.Fprint(out, "\r\n")
				return nil, errors.New("aborted")
			}
			switch b3, _ := reader.ReadByte(); b3 {
			case 'A':
				move(-1)
			case 'B':
				move(1)
			}
		case b > 32 && b < 127: // printable: add to filter
			query += string(rune(b))
			refilter()
		default:
			continue // ignore other control bytes without redrawing
		}
		draw()
	}
}

func statusLine(selected []bool, filtered []int, offset, end int) string {
	n := 0
	for _, s := range selected {
		if s {
			n++
		}
	}
	if len(filtered) == 0 {
		return fmt.Sprintf("%d selected", n)
	}
	return fmt.Sprintf("%d selected · showing %d-%d of %d", n, offset+1, end, len(filtered))
}

func collectSelected(selected []bool) []int {
	var idx []int
	for i, s := range selected {
		if s {
			idx = append(idx, i)
		}
	}
	return idx
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
