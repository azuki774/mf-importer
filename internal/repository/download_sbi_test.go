package repository

import "testing"

func TestIsSafeRelPath(t *testing.T) {
	tests := []struct {
		rel  string
		want bool
	}{
		{"2026/08/20260816-114651.json", true},
		{"top.json", true},
		{"../../etc/config", false},
		{"2026/../../etc/config", false},
		{"a/../b.json", false},
		{"/abs/path.json", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isSafeRelPath(tt.rel); got != tt.want {
			t.Errorf("isSafeRelPath(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}
