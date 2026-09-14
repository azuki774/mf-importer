package model

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestParseSbiJSON_NewFormat(t *testing.T) {
	raw, err := os.ReadFile("../../test/sbi_example_new.json")
	if err != nil {
		t.Fatal(err)
	}
	snapshot, holdings, err := ParseSbiJSON(raw)
	if err != nil {
		t.Fatalf("ParseSbiJSON: %v", err)
	}
	if snapshot.Status != SbiStatusOK || snapshot.SchemaVersion != CurrentSbiSchemaVersion {
		t.Fatalf("snapshot status/version = %q/%q", snapshot.Status, snapshot.SchemaVersion)
	}
	want := map[string]string{
		"nisa_domestic":  "DUMMY0000001",
		"nisa_us":        "DUMMY0000002",
		"nisa_funds":     "DUMMY0000003",
		"old_nisa_funds": "DUMMY0000004",
	}
	got := make(map[string]string)
	for _, holding := range holdings {
		if holding.CompositeFIGI == nil {
			t.Fatalf("holding %q has nil FIGI", holding.Name)
		}
		got[holding.Section] = *holding.CompositeFIGI
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("section FIGIs = %#v, want %#v", got, want)
	}
	requireFloat64Value(t, "grand total", snapshot.GrandTotalJPY, 1160000)
	requireFloat64Value(t, "nisa total", snapshot.NisaTotalJPY, 600000)
	for _, holding := range holdings {
		if holding.Section == "nisa_us" {
			if holding.PrevDayJPY != nil || holding.PrevDayPct != nil {
				t.Error("US holding prev-day values must be nil")
			}
		} else if holding.PrevDayJPY == nil || holding.PrevDayPct == nil {
			t.Errorf("%s holding prev-day values must be present", holding.Section)
		}
	}
}

func TestParseSbiJSON_StatusNormalization(t *testing.T) {
	for _, test := range []struct {
		status string
		want   SbiStatus
	}{
		{status: "ok", want: SbiStatusOK},
		{status: "maintenance", want: SbiStatusMaintenance},
		{status: "ERROR", want: SbiStatusError},
	} {
		t.Run(test.status, func(t *testing.T) {
			raw := strings.Replace(currentSbiOKJSON(), `"status":"ok"`, `"status":"`+test.status+`"`, 1)
			snapshot, _, err := ParseSbiJSON([]byte(raw))
			if err != nil || snapshot.Status != test.want {
				t.Fatalf("status = %q, error = %v; want %q", snapshot.Status, err, test.want)
			}
		})
	}
}

func TestParseSbiJSON_NormalizesFetchedAtToDatabasePrecision(t *testing.T) {
	snapshot, _, err := ParseSbiJSON([]byte(`{"fetched_at":"2026-08-16T11:46:51.908856153+09:00","status":"maintenance","schema_version":"2026-09-12"}`))
	if err != nil {
		t.Fatalf("ParseSbiJSON: %v", err)
	}
	if got := snapshot.FetchedAt.Location().String(); got != "Asia/Tokyo" {
		t.Fatalf("location = %q", got)
	}
	if got := snapshot.FetchedAt.Nanosecond(); got != 908856000 {
		t.Fatalf("nanosecond = %d", got)
	}
}

func TestParseSbiJSON_RejectsMissingFetchedAt(t *testing.T) {
	raw := strings.Replace(currentSbiOKJSON(), `"fetched_at":"2026-08-16T12:00:00Z",`, "", 1)
	if _, _, err := ParseSbiJSON([]byte(raw)); err == nil {
		t.Fatal("accepted missing fetched_at")
	}
}

func TestParseSbiJSON_RejectsMissingNullNumericAndUnknownSchemaVersions(t *testing.T) {
	for _, test := range []struct {
		name string
		json string
	}{
		{name: "missing", json: strings.Replace(currentSbiOKJSON(), `,"schema_version":"2026-09-12"`, "", 1)},
		{name: "null", json: strings.Replace(currentSbiOKJSON(), `"schema_version":"2026-09-12"`, `"schema_version":null`, 1)},
		{name: "numeric", json: strings.Replace(currentSbiOKJSON(), `"schema_version":"2026-09-12"`, `"schema_version":1`, 1)},
		{name: "unknown", json: strings.Replace(currentSbiOKJSON(), `"schema_version":"2026-09-12"`, `"schema_version":"2026-09-13"`, 1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := ParseSbiJSON([]byte(test.json)); err == nil {
				t.Fatal("accepted unsupported schema version")
			}
		})
	}
}

func TestParseSbiJSON_RequiresCurrentSchemaVersionForEveryStatus(t *testing.T) {
	for _, status := range []string{"ok", "maintenance", "error"} {
		t.Run(status, func(t *testing.T) {
			raw := strings.Replace(currentSbiOKJSON(), `"status":"ok"`, `"status":"`+status+`"`, 1)
			raw = strings.Replace(raw, `"schema_version":"2026-09-12"`, `"schema_version":"2026-09-13"`, 1)
			if _, _, err := ParseSbiJSON([]byte(raw)); err == nil {
				t.Fatal("accepted unsupported schema version")
			}
		})
	}
}

func TestParseSbiJSON_RejectsInvalidStatusWithoutEchoingInput(t *testing.T) {
	const sensitiveStatus = "synthetic-status-value"
	_, _, err := ParseSbiJSON([]byte(`{"fetched_at":"2026-08-16T12:00:00Z","status":"` + sensitiveStatus + `","schema_version":"2026-09-12"}`))
	if err == nil || strings.Contains(err.Error(), sensitiveStatus) {
		t.Fatalf("error = %v, must not echo status", err)
	}
}

func TestParseSbiJSON_OKRejectsMissingRequiredValues(t *testing.T) {
	raw := `{"fetched_at":"2026-08-16T12:00:00Z","status":"ok","schema_version":"2026-09-12","grand_total_jpy":0}`
	_, _, err := ParseSbiJSON([]byte(raw))
	if err == nil || !strings.Contains(err.Error(), "nisa") || !strings.Contains(err.Error(), "cash") {
		t.Fatalf("error = %v, want missing field paths", err)
	}
}

func TestParseSbiJSON_OKAcceptsExplicitZeroValues(t *testing.T) {
	snapshot, holdings, err := ParseSbiJSON([]byte(currentSbiOKJSON()))
	if err != nil {
		t.Fatalf("ParseSbiJSON: %v", err)
	}
	requireFloat64Value(t, "grand total", snapshot.GrandTotalJPY, 0)
	requireFloat64Value(t, "nisa total", snapshot.NisaTotalJPY, 0)
	requireFloat64Value(t, "cash JPY amount", snapshot.CashJpyAmount, 0)
	if len(holdings) != 0 {
		t.Fatalf("holdings = %d, want 0", len(holdings))
	}
}

func TestParseSbiJSON_StoresAndValidatesCompositeFIGI(t *testing.T) {
	holding := syntheticHoldingJSON("ダミー銘柄A", "DUMMY0000001")
	raw := withDomesticHolding(currentSbiOKJSON(), holding)
	_, holdings, err := ParseSbiJSON([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSbiJSON: %v", err)
	}
	if len(holdings) != 1 || holdings[0].CompositeFIGI == nil || *holdings[0].CompositeFIGI != "DUMMY0000001" {
		t.Fatalf("holdings = %#v, want one synthetic FIGI", holdings)
	}

	for _, figi := range []string{"dummy0000001", "DUMMY00000001", "DUMMY000000-1"} {
		raw := withDomesticHolding(currentSbiOKJSON(), syntheticHoldingJSON("ダミー銘柄A", figi))
		if _, _, err := ParseSbiJSON([]byte(raw)); err == nil {
			t.Errorf("FIGI %q was accepted", figi)
		}
	}
	validHolding := syntheticHoldingJSON("ダミー銘柄A", "DUMMY0000001")
	for _, holdingJSON := range []string{
		strings.Replace(validHolding, `,"composite_figi":"DUMMY0000001"`, "", 1),
		strings.Replace(validHolding, `"composite_figi":"DUMMY0000001"`, `"composite_figi":null`, 1),
	} {
		_, _, err := ParseSbiJSON([]byte(withDomesticHolding(currentSbiOKJSON(), holdingJSON)))
		if err == nil || !strings.Contains(err.Error(), "composite_figi") {
			t.Errorf("error = %v, want composite_figi validation", err)
		}
	}
}

func TestParseSbiJSON_RejectsDuplicateCompositeFIGIWithinSection(t *testing.T) {
	holding := syntheticHoldingJSON("ダミー銘柄A", "DUMMY0000001")
	raw := withDomesticHolding(currentSbiOKJSON(), holding+","+syntheticHoldingJSON("ダミー銘柄B", "DUMMY0000001"))
	if _, _, err := ParseSbiJSON([]byte(raw)); err == nil {
		t.Fatal("accepted duplicate FIGI in one section")
	}
}

func TestParseSbiJSON_AllowsSameCompositeFIGIInDifferentSections(t *testing.T) {
	holding := syntheticHoldingJSON("ダミー銘柄A", "DUMMY0000001")
	raw := withDomesticHolding(currentSbiOKJSON(), holding)
	raw = withNextHolding(raw, syntheticHoldingJSON("ダミー銘柄B", "DUMMY0000001"))
	_, holdings, err := ParseSbiJSON([]byte(raw))
	if err != nil || len(holdings) != 2 || holdings[0].Section == holdings[1].Section {
		t.Fatalf("holdings = %#v, error = %v; want same FIGI in separate sections", holdings, err)
	}
}

func TestParseSbiJSON_OKRejectsMissingHoldingValues(t *testing.T) {
	raw := withDomesticHolding(currentSbiOKJSON(), `{"name":"ダミー銘柄A"}`)
	_, _, err := ParseSbiJSON([]byte(raw))
	if err == nil || !strings.Contains(err.Error(), "nisa.domestic_stocks.holdings[0]") {
		t.Fatalf("error = %v, want holding path", err)
	}
}

func TestParseSbiJSON_NonOKMarksAllValuesUnavailable(t *testing.T) {
	for _, status := range []string{"maintenance", "error"} {
		t.Run(status, func(t *testing.T) {
			raw := strings.Replace(currentSbiOKJSON(), `"status":"ok"`, `"status":"`+status+`"`, 1)
			snapshot, holdings, err := ParseSbiJSON([]byte(raw))
			if err != nil {
				t.Fatalf("ParseSbiJSON: %v", err)
			}
			value := reflect.ValueOf(snapshot).Elem()
			pointerType := reflect.TypeOf((*float64)(nil))
			for index := 0; index < value.NumField(); index++ {
				if value.Type().Field(index).Type == pointerType && !value.Field(index).IsNil() {
					t.Errorf("%s must be nil for %s", value.Type().Field(index).Name, status)
				}
			}
			if len(holdings) != 0 {
				t.Fatalf("holdings = %d, want 0", len(holdings))
			}
		})
	}
}

func TestSbiHolding_JSONRoundTrip(t *testing.T) {
	figi := "DUMMY0000001"
	holding := SbiHolding{Section: "nisa_domestic", CompositeFIGI: &figi, Name: "ダミー銘柄A"}
	data, err := json.Marshal(holding)
	if err != nil {
		t.Fatal(err)
	}
	var got SbiHolding
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.CompositeFIGI == nil || *got.CompositeFIGI != figi {
		t.Fatalf("round trip FIGI = %#v", got.CompositeFIGI)
	}
}

func requireFloat64Value(t *testing.T, name string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s = nil, want %v", name, want)
	}
	if math.Abs(*got-want) > 0.01 {
		t.Errorf("%s = %v, want %v", name, *got, want)
	}
}

func currentSbiOKJSON() string {
	return `{"fetched_at":"2026-08-16T12:00:00Z","status":"ok","schema_version":"2026-09-12","nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"pnl_jpy":0,"pnl_pct":0,"domestic_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"us_stocks":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]},"funds":{"value_jpy":0,"pnl_jpy":0,"pnl_pct":0,"prev_day_jpy":0,"prev_day_pct":0,"prev_month_jpy":0,"prev_month_pct":0,"holdings":[]}},"old_nisa":{"total_jpy":0,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"funds":[]},"cash":{"jpy":{"amount":0,"value_jpy":0},"usd":{"amount":0,"value_jpy":0}},"others":{"funds":{"amount":0,"value_jpy":0}},"grand_total_jpy":0}`
}

func syntheticHoldingJSON(name, figi string) string {
	return `{"name":"` + name + `","composite_figi":"` + figi + `","quantity":1,"unit_cost":1,"unit_price":1,"prev_day_jpy":0,"prev_day_pct":0,"pnl_jpy":0,"pnl_pct":0,"value_jpy":1}`
}

func withDomesticHolding(raw, holdings string) string {
	return strings.Replace(raw, `"holdings":[]`, `"holdings":[`+holdings+`]`, 1)
}

func withNextHolding(raw, holding string) string {
	return strings.Replace(raw, `"holdings":[]`, `"holdings":[`+holding+`]`, 1)
}
