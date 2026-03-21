package discogs

import (
	"testing"
)

func TestSplitTerm(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"Funk / Soul", []string{"Funk", "Soul"}},
		{"Folk, World, & Country", []string{"Folk", "World", "Country"}},
		{"RnB/Swing", []string{"RnB", "Swing"}},
		{"Rock", []string{"Rock"}},
		{"Jazz-Rock", []string{"Jazz-Rock"}}, // hyphen is not a split char
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := splitTerm(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitTerm(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitTerm(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}
