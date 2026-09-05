package model

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestParseSbiJSON_NewFormat(t *testing.T) {
	raw, err := os.ReadFile("../../test/sbi_example_new.json")
	if err != nil {
		t.Skipf("test fixture not found: %v", err)
	}
	snap, holdings, err := ParseSbiJSON(raw)
	if err != nil {
		t.Fatalf("ParseSbiJSON: %v", err)
	}
	if snap.Status != SbiStatusOK {
		t.Errorf("status = %q, want OK", snap.Status)
	}
	if snap.SchemaVersion != 1 {
		t.Errorf("schema_version = %d, want 1", snap.SchemaVersion)
	}
	if math.Abs(snap.GrandTotalJPY-1160000) > 0.01 {
		t.Errorf("grand_total = %v, want 1160000", snap.GrandTotalJPY)
	}
	if snap.NisaTotalJPY != 600000 {
		t.Errorf("nisa total = %v", snap.NisaTotalJPY)
	}
	if len(holdings) != 4 {
		t.Fatalf("holdings = %d, want 4", len(holdings))
	}
	// holdings sections
	wantSections := map[string]int{"nisa_domestic": 1, "nisa_us": 1, "nisa_funds": 1, "old_nisa_funds": 1}
	got := map[string]int{}
	for _, h := range holdings {
		got[h.Section]++
	}
	if !reflect.DeepEqual(got, wantSections) {
		t.Errorf("sections = %v, want %v", got, wantSections)
	}
}

func TestParseSbiJSON_StatusNormalization(t *testing.T) {
	tests := []struct {
		in    string
		want  SbiStatus
		isErr bool
	}{
		{`{"fetched_at":"2026-08-16T12:00:00Z","status":"ok","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`, SbiStatusOK, false},
		{`{"fetched_at":"2026-08-16T12:00:00Z","status":"maintenance","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`, SbiStatusMaintenance, false},
		{`{"fetched_at":"2026-08-16T12:00:00Z","status":"ERROR","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`, SbiStatusError, false},
		{`{"fetched_at":"2026-08-16T12:00:00Z","status":"invalid","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`, "", true},
	}
	for _, tt := range tests {
		snap, _, err := ParseSbiJSON([]byte(tt.in))
		if (err != nil) != tt.isErr {
			t.Errorf("status %q err = %v, wantErr %v", tt.in, err, tt.isErr)
			continue
		}
		if !tt.isErr && snap.Status != tt.want {
			t.Errorf("status = %q, want %q", snap.Status, tt.want)
		}
	}
}

func TestParseSbiJSON_MissingFetchedAt(t *testing.T) {
	jsonStr := `{"status":"ok","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`
	if _, _, err := ParseSbiJSON([]byte(jsonStr)); err == nil {
		t.Error("expected error for missing fetched_at")
	}
}

func TestSbiHolding_Sections(t *testing.T) {
	// Verify holdings are correctly sectioned and serializable
	raw := `{"fetched_at":"2026-08-16T12:00:00Z","status":"ok","schema_version":1,"nisa":{"total_jpy":100,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":10,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[{"name":"A","quantity":1,"unit_cost":1,"unit_price":1,"value_jpy":10}]},"us_stocks":{"value_jpy":20,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[{"name":"B","quantity":2,"unit_cost":2,"unit_price":2,"value_jpy":20}]},"funds":{"value_jpy":30,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[{"name":"C","quantity":3,"unit_cost":3,"unit_price":3,"value_jpy":30}]}},"old_nisa":{"total_jpy":40,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[{"name":"D","quantity":4,"unit_cost":4,"unit_price":4,"value_jpy":40}]},"cash":{"jpy":{"amount":5,"value_jpy":5},"usd":{"amount":6,"value_jpy":6}},"others":{"funds":{"amount":7,"value_jpy":7}},"grand_total_jpy":158}`
	snap, holdings, err := ParseSbiJSON([]byte(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if snap.FetchedAt.IsZero() {
		t.Error("fetched_at zero")
	}
	if len(holdings) != 4 {
		t.Fatalf("holdings %d", len(holdings))
	}
	sec := map[string]bool{}
	for _, h := range holdings {
		sec[h.Section] = true
	}
	for _, s := range []string{"nisa_domestic", "nisa_us", "nisa_funds", "old_nisa_funds"} {
		if !sec[s] {
			t.Errorf("missing section %s", s)
		}
	}
	// json roundtrip
	b, _ := json.Marshal(snap)
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	if _, ok := m["fetched_at"]; !ok {
		t.Error("json missing fetched_at")
	}
	_ = time.Now()
}

func TestParseSbiJSON_ExampleNewFixture(t *testing.T) {
	// Ensure the new format fixture (if present) roundtrips
	data, err := os.ReadFile("../../test/sbi_example_new.json")
	if err != nil {
		t.Skip()
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	_, _, err = ParseSbiJSON(data)
	if err != nil {
		t.Fatalf("parse new fixture: %v", err)
	}
}
