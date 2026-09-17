package main

import (
	"testing"
	"time"
)

func TestResolveNrknMonth(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		now     time.Time
		want    string
		wantErr bool
	}{
		{name: "current month", now: time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC), want: "202609"},
		{name: "Tokyo month boundary", now: time.Date(2026, 8, 31, 16, 30, 0, 0, time.UTC), want: "202609"},
		{name: "explicit month", input: "202608", now: time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC), want: "202608"},
		{name: "hyphenated month", input: "2026-08", wantErr: true},
		{name: "invalid month", input: "202613", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveNrknMonth(tt.input, tt.now)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveNrknMonth() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Fatalf("resolveNrknMonth() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNrknImportMode(t *testing.T) {
	if got, err := nrknImportMode("/tmp/nrkn", ""); err != nil || got != nrknLocalMode {
		t.Fatalf("local mode = %q, error = %v", got, err)
	}
	if got, err := nrknImportMode("", "202608"); err != nil || got != nrknS3Mode {
		t.Fatalf("S3 mode = %q, error = %v", got, err)
	}
	if _, err := nrknImportMode("/tmp/nrkn", "202608"); err == nil {
		t.Fatal("local mode with month error = nil")
	}
}
