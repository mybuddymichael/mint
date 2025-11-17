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
			ids:  []string{"mint-abc123"},
			want: map[string]int{"mint-abc123": 1},
		},
		{
			name: "two IDs different start",
			ids:  []string{"mint-abc123", "mint-xyz789"},
			want: map[string]int{"mint-abc123": 6, "mint-xyz789": 6},
		},
		{
			name: "two IDs common prefix",
			ids:  []string{"mint-8a9jfnd", "mint-8afj0qn"},
			want: map[string]int{"mint-8a9jfnd": 8, "mint-8afj0qn": 8},
		},
		{
			name: "three IDs varying prefixes",
			ids:  []string{"mint-abc", "mint-abd", "mint-xyz"},
			want: map[string]int{"mint-abc": 8, "mint-abd": 8, "mint-xyz": 6},
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
