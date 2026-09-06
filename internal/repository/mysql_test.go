package repository

import "testing"

func TestBuildMySQLDSNUsesTokyoLocation(t *testing.T) {
	got := buildMySQLDSN("db.example", "3306", "test-user", "test-pass", "mfimporter")
	want := "test-user:test-pass@(db.example:3306)/mfimporter?parseTime=true&loc=Asia%2FTokyo"
	if got != want {
		t.Fatalf("DSN = %q, want %q", got, want)
	}
}
