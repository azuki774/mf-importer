package model

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func nrknFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("../../test/nrkn_example.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestParseNrknJSON(t *testing.T) {
	s, h, err := ParseNrknJSON(nrknFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != NrknStatusOK || s.GrandTotalJPY != 100 || s.TotalCostJPY != 110 || s.PnlJPY != -10 {
		t.Fatal("totals were not preserved")
	}
	want := time.Date(2000, 1, 2, 12, 4, 5, 123456000, time.FixedZone("Asia/Tokyo", 9*60*60))
	if !s.FetchedAt.Equal(want) || s.FetchedAt.Hour() != 12 {
		t.Fatalf("timestamp = %v", s.FetchedAt)
	}
	if len(h) != 1 || h[0].ProductCode != "000001" || h[0].CompositeFIGI != "DUMMY0000001" || h[0].UnitPriceRaw != "*40円" || h[0].RedemptionUnitPriceRaw != "*39.5円" || h[0].Quantity != 2.5 || h[0].RedemptionUnitPrice != 39.5 || h[0].ReferenceDate != "2000-01-01" {
		t.Fatal("holding fields were not preserved")
	}
}

func TestParseNrknRequiredFields(t *testing.T) {
	var original map[string]any
	if err := json.Unmarshal(nrknFixture(t), &original); err != nil {
		t.Fatal(err)
	}
	for _, holding := range []bool{false, true} {
		fields := original
		if holding {
			fields = original["holdings"].([]any)[0].(map[string]any)
		}
		for field := range fields {
			for _, missing := range []bool{false, true} {
				var input map[string]any
				if err := json.Unmarshal(nrknFixture(t), &input); err != nil {
					t.Fatal(err)
				}
				target := input
				if holding {
					target = input["holdings"].([]any)[0].(map[string]any)
				}
				if missing {
					delete(target, field)
				} else {
					target[field] = nil
				}
				raw, _ := json.Marshal(input)
				if _, _, err := ParseNrknJSON(raw); err == nil {
					t.Errorf("accepted missing/null field %s (holding=%v)", field, holding)
				}
			}
		}
	}
}

func TestParseNrknInvalidAndFutureInputs(t *testing.T) {
	for _, test := range []struct {
		name, from, to string
		unsupported    bool
	}{
		{"future", `"2026-09-14"`, `"2099-01-01"`, true},
		{"numeric version", `"2026-09-14"`, `1`, false},
		{"error status", `"ok"`, `"error"`, false},
		{"empty name", `"ダミー商品A"`, `" "`, false},
		{"invalid FIGI", `"DUMMY0000001"`, `"invalid"`, false},
		{"invalid date", `"2000-01-01"`, `"2000-02-30"`, false},
		{"invalid timestamp", `"2000-01-02T03:04:05.123456789Z"`, `"private-invalid-time"`, false},
		{"fractional integer", `"value_jpy": 100`, `"value_jpy": 1.5`, false},
		{"overflow", `"quantity": 2.5`, `"quantity": 1e999`, false},
		{"invalid allocation", `"allocation_pct": 100`, `"allocation_pct": 101`, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := strings.Replace(string(nrknFixture(t)), test.from, test.to, 1)
			_, _, err := ParseNrknJSON([]byte(raw))
			if err == nil || errors.Is(err, ErrUnsupportedNrknSchemaVersion) != test.unsupported {
				t.Fatalf("unexpected error: %v", err)
			}
			if strings.Contains(err.Error(), "private-invalid-time") {
				t.Fatal("input leaked")
			}
		})
	}
}

func TestParseNrknProductsAndDisplayedTotals(t *testing.T) {
	var input map[string]any
	if err := json.Unmarshal(nrknFixture(t), &input); err != nil {
		t.Fatal(err)
	}
	holding := input["holdings"].([]any)[0].(map[string]any)
	input["holdings"] = []any{holding, holding}
	raw, _ := json.Marshal(input)
	if _, _, err := ParseNrknJSON(raw); err == nil {
		t.Fatal("duplicate product accepted")
	}
	copyHolding := make(map[string]any)
	for k, v := range holding {
		copyHolding[k] = v
	}
	copyHolding["product_code"] = "000002"
	input["holdings"] = []any{holding, copyHolding}
	raw, _ = json.Marshal(input)
	s, h, err := ParseNrknJSON(raw)
	if err != nil || len(h) != 2 || s.GrandTotalJPY != 100 {
		t.Fatal("distinct products sharing a FIGI or displayed totals were not preserved")
	}
	input["holdings"] = []any{}
	raw, _ = json.Marshal(input)
	if _, _, err := ParseNrknJSON(raw); err == nil {
		t.Fatal("empty holdings accepted")
	}
	input["holdings"] = []any{holding}
	for _, key := range []string{"grand_total_jpy", "total_cost_jpy", "pnl_jpy"} {
		input[key] = 0
	}
	for _, key := range []string{"quantity", "unit_price", "value_jpy", "cost_jpy", "redemption_unit_price", "redemption_value_jpy", "pnl_jpy", "allocation_pct"} {
		holding[key] = 0
	}
	raw, _ = json.Marshal(input)
	if _, _, err := ParseNrknJSON(raw); err != nil {
		t.Fatal("explicit zeros rejected", err)
	}
}
