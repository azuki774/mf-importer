package repository

import (
	"testing"
	"time"
)

func TestJstBoundMicrosecondRounding(t *testing.T) {
	input := time.Date(2026, 1, 2, 3, 4, 5, 123456789, time.UTC)
	got := jstBound(input, true)
	want := time.Date(2026, 1, 2, 12, 4, 5, 123457000, time.FixedZone("JST", 9*60*60))
	if !got.Equal(want) || got.Nanosecond() != want.Nanosecond() {
		t.Fatalf("ceil boundary = %s, want %s", got, want)
	}
	got = jstBound(input, false)
	want = time.Date(2026, 1, 2, 12, 4, 5, 123456000, time.FixedZone("JST", 9*60*60))
	if !got.Equal(want) || got.Nanosecond() != want.Nanosecond() {
		t.Fatalf("floor boundary = %s, want %s", got, want)
	}
}
