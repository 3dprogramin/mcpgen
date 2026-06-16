package tui

import (
	"reflect"
	"testing"
)

func TestMatchServers(t *testing.T) {
	names := []string{"burp", "chrome-devtools", "mongodb", "obsidian"}
	descs := []string{
		"Burp Suite SSE bridge",
		"Chrome DevTools driver",
		"query and manage a MongoDB database",
		"read and write notes",
	}
	tests := []struct {
		query string
		want  []int
	}{
		{"", []int{0, 1, 2, 3}},
		{"mongo", []int{2}},
		{"MONGO", []int{2}},    // case-insensitive
		{"chrome", []int{1}},   // name match
		{"database", []int{2}}, // description match
		{"  notes ", []int{3}}, // trimmed
		{"b", []int{0, 2, 3}},  // burp, mongodb, obsidian all contain "b"
		{"zzz", []int{}},       // no match
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := matchServers(names, descs, tt.query)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("matchServers(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
