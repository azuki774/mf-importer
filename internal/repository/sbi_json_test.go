package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"mf-importer/internal/logger"
)

func TestSbiJSONOperator_GetSbiTargetFiles_Recursive(t *testing.T) {
	dir := t.TempDir()
	// create nested structure 2026/08/20260816-114651.json and 2026/09/20260901-120000.json and a csv that should be ignored
	nested1 := filepath.Join(dir, "2026", "08")
	nested2 := filepath.Join(dir, "2026", "09")
	if err := os.MkdirAll(nested1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nested2, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(nested1, "20260816-114651.json"), []byte(`{"fetched_at":"2026-08-16T12:00:00Z","status":"ok","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`))
	writeFile(t, filepath.Join(nested2, "20260901-120000.json"), []byte(`{"fetched_at":"2026-09-01T12:00:00Z","status":"ok","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`))
	// top-level json
	writeFile(t, filepath.Join(dir, "top.json"), []byte(`{"fetched_at":"2026-08-17T12:00:00Z","status":"ok","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`))
	// csv should be ignored
	writeFile(t, filepath.Join(dir, "cf.csv"), []byte(`dummy`))
	writeFile(t, filepath.Join(nested1, "ignore.txt"), []byte(`txt`))

	op := &SbiJSONOperator{Logger: logger.NewLogger()}
	files, err := op.GetSbiTargetFiles(context.Background(), dir)
	if err != nil {
		t.Fatalf("GetSbiTargetFiles: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("files = %d, want 3 (only .json recursively): %v", len(files), files)
	}
	for _, f := range files {
		if filepath.Ext(f) != ".json" {
			t.Errorf("unexpected ext %s", f)
		}
	}
}

func TestSbiJSONOperator_LoadSbiJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "20260816-114651.json")
	jsonStr := `{"fetched_at":"2026-08-16T11:46:51.908856153Z","status":"ok","schema_version":1,"nisa":{"total_jpy":600000,"prev_day_jpy":12000,"prev_day_pct":2.04,"prev_month_jpy":45000,"prev_month_pct":8.11,"pnl_jpy":150000,"pnl_pct":33.33,"domestic_stocks":{"value_jpy":100000,"pnl_jpy":20000,"pnl_pct":25,"prev_day_jpy":3000,"prev_day_pct":3.09,"prev_month_jpy":8000,"prev_month_pct":8.7,"holdings":[{"name":"ダミー銘柄A","quantity":100,"unit_cost":800,"unit_price":1000,"prev_day_jpy":30,"prev_day_pct":3.09,"pnl_jpy":20000,"pnl_pct":25,"value_jpy":100000}]},"us_stocks":{"value_jpy":200000,"pnl_jpy":50000,"pnl_pct":33.33,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":30000,"prev_month_pct":17.65,"holdings":[{"name":"ダミー米国株B","quantity":10,"unit_cost":100,"unit_price":125.5,"pnl_jpy":40000,"pnl_pct":40,"value_jpy":200000}]},"funds":{"value_jpy":300000,"pnl_jpy":80000,"pnl_pct":36.36,"prev_day_jpy":9000,"prev_day_pct":3.09,"prev_month_jpy":7000,"prev_month_pct":2.39,"holdings":[{"name":"ダミー投信C","quantity":100000,"unit_cost":2000,"unit_price":3000,"prev_day_jpy":90,"prev_day_pct":3.09,"pnl_jpy":80000,"pnl_pct":36.36,"value_jpy":300000}]}},"old_nisa":{"total_jpy":400000,"prev_day_jpy":10000,"prev_day_pct":2.56,"pnl_jpy":250000,"pnl_pct":166.67,"funds":[{"name":"ダミー投信D","quantity":150000,"unit_cost":1000,"unit_price":2666.67,"prev_day_jpy":66.67,"prev_day_pct":2.56,"pnl_jpy":250000,"pnl_pct":166.67,"value_jpy":400000}]},"cash":{"jpy":{"amount":50000,"value_jpy":50000},"usd":{"amount":500,"value_jpy":80000}},"others":{"funds":{"amount":30000,"value_jpy":30000}},"grand_total_jpy":1160000}`
	writeFile(t, path, []byte(jsonStr))

	op := &SbiJSONOperator{Logger: logger.NewLogger()}
	snap, holdings, err := op.LoadSbiJSON(context.Background(), path)
	if err != nil {
		t.Fatalf("LoadSbiJSON: %v", err)
	}
	if string(snap.Status) != "OK" {
		t.Errorf("status = %q, want OK", snap.Status)
	}
	if snap.FetchedAt.IsZero() {
		t.Error("fetched_at zero")
	}
	if len(holdings) != 4 {
		t.Errorf("holdings = %d, want 4", len(holdings))
	}
	if snap.GrandTotalJPY != 1160000 {
		t.Errorf("grand = %v", snap.GrandTotalJPY)
	}
	if holdings[0].Name != "ダミー銘柄A" {
		t.Errorf("holdings[0].Name = %q", holdings[0].Name)
	}
}
