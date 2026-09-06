package repository

import "testing"

func TestSbiMonthPrefix(t *testing.T) {
	tests := []struct {
		name      string
		bucketDir string
		month     string
		want      string
		wantErr   bool
	}{
		{name: "valid with slash", bucketDir: "assets/", month: "202608", want: "assets/2026/08/"},
		{name: "valid without slash", bucketDir: "assets", month: "202608", want: "assets/2026/08/"},
		{name: "invalid month", bucketDir: "assets/", month: "202600", wantErr: true},
		{name: "invalid format", bucketDir: "assets/", month: "2026-08", wantErr: true},
		{name: "short format", bucketDir: "assets/", month: "2608", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sbiMonthPrefix(tt.bucketDir, tt.month)
			if (err != nil) != tt.wantErr {
				t.Fatalf("sbiMonthPrefix() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("sbiMonthPrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

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
