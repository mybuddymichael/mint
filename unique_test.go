package main

import (
	"testing"
)

func TestMinUniquePrefixLengths(t *testing.T) {
	tests := []struct {
		name string
		ids  []string
		want map[string]int
	}{
		{
			name: "single ID",
			ids:  []string{"mt-abc123"},
			want: map[string]int{"mt-abc123": 1},
		},
		{
			name: "two IDs different start",
			ids:  []string{"mt-abc123", "mt-xyz789"},
			want: map[string]int{"mt-abc123": 4, "mt-xyz789": 4},
		},
		{
			name: "two IDs common prefix",
			ids:  []string{"mt-8a9jfnd", "mt-8afj0qn"},
			want: map[string]int{"mt-8a9jfnd": 6, "mt-8afj0qn": 6},
		},
		{
			name: "three IDs varying prefixes",
			ids:  []string{"mt-abc", "mt-abd", "mt-xyz"},
			want: map[string]int{"mt-abc": 6, "mt-abd": 6, "mt-xyz": 4},
		},
		{
			name: "empty list",
			ids:  []string{},
			want: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MinUniquePrefixLengths(tt.ids)
			if len(got) != len(tt.want) {
				t.Errorf("MinUniquePrefixLengths() returned %d items, want %d", len(got), len(tt.want))
			}
			for id, wantLen := range tt.want {
				if gotLen, ok := got[id]; !ok {
					t.Errorf("MinUniquePrefixLengths() missing key %s", id)
				} else if gotLen != wantLen {
					t.Errorf("MinUniquePrefixLengths()[%s] = %d, want %d", id, gotLen, wantLen)
				}
			}
		})
	}
}
