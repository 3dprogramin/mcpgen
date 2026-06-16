package main

import (
	"reflect"
	"testing"
)

func TestParseSelection(t *testing.T) {
	const count = 5
	tests := []struct {
		in      string
		want    []int
		wantErr bool
	}{
		{"all", []int{0, 1, 2, 3, 4}, false},
		{"ALL", []int{0, 1, 2, 3, 4}, false},
		{"1 3", []int{0, 2}, false},
		{"1,3", []int{0, 2}, false},
		{"1-3", []int{0, 1, 2}, false},
		{"3 1", []int{2, 0}, false},
		{"1 1 2", []int{0, 1}, false},
		{"  2  ", []int{1}, false},
		{"", nil, true},
		{"0", nil, true},
		{"6", nil, true},
		{"abc", nil, true},
		{"3-1", nil, true},
		{"1-", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseSelection(tt.in, count)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %v", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSelection(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
