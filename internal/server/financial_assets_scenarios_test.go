package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"mf-importer/internal/mfapi"
	"mf-importer/internal/openapi"
)

func getFinancialMockJSON[T any](t *testing.T, r http.Handler, path string) T {
	t.Helper()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("unexpected status/content type for %s: %d", path, rec.Code)
	}
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if _, ok := envelope["nextCursor"]; ok {
		t.Fatal("offset-based response must not include nextCursor")
	}
	return out
}

func requireFinancialDecimal(t *testing.T, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Fatalf("unexpected synthetic decimal, want %s", want)
	}
}

// Exercise the actual mock service through both API mounts used by the UI.
// All expected amounts and dates below belong to the hand-written mock dataset.
func TestFinancialAssetMockUIScenarios(t *testing.T) {
	for _, prefix := range []string{"", "/api"} {
		t.Run("mount="+prefix, func(t *testing.T) {
			r := chi.NewRouter()
			gw := &apigateway{APIService: &mfapi.MockAPIService{}}
			if prefix == "" {
				registerAPI(gw, r)
			} else {
				r.Route(prefix, func(sub chi.Router) { registerAPI(gw, sub) })
			}
			base := prefix + "/v2/financial-assets"

			t.Run("latest balance and complete holdings", func(t *testing.T) {
				page := getFinancialMockJSON[openapi.BalancePage](t, r, base+"/balances")
				if len(page.Items) != 1 {
					t.Fatal("expected one current balance")
				}
				point := page.Items[0]
				requireFinancialDecimal(t, point.Totals.ValuationJpy, "320.25")
				if point.Totals.CostJpy != nil || point.Totals.UnrealizedPnlJpy != nil || len(point.MissingSources) != 0 || len(point.Sources) != 2 {
					t.Fatal("unknown aggregate metrics must remain null even with both sources")
				}
				details := map[openapi.Source]openapi.SnapshotDetail{}
				for _, source := range point.Sources {
					if source.SnapshotId == nil || source.FetchedAt == nil {
						t.Fatal("missing balance provenance")
					}
					detail := getFinancialMockJSON[openapi.SnapshotDetail](t, r, base+"/snapshots/"+url.PathEscape(*source.SnapshotId))
					fetchedAt, err := time.Parse(time.RFC3339Nano, *source.FetchedAt)
					if err != nil || detail.SnapshotId != *source.SnapshotId || detail.Source != source.Source || !detail.FetchedAt.Equal(fetchedAt) || !reflect.DeepEqual(detail.Totals, source.Totals) {
						t.Fatal("balance and selected detail disagree")
					}
					details[source.Source] = detail
				}
				sbi, nrkn := details["sbi"], details["nrkn"]
				if len(sbi.Holdings) != 60 || len(nrkn.Holdings) != 4 || !sbi.FetchedAt.Before(nrkn.FetchedAt) {
					t.Fatal("expected full holdings at different source timestamps")
				}
				// Three positions in one product must survive across sources and sections.
				a, b, c := sbi.Holdings[0], sbi.Holdings[1], nrkn.Holdings[0]
				if a.CompositeFigi == nil || !reflect.DeepEqual(a.CompositeFigi, b.CompositeFigi) || !reflect.DeepEqual(a.CompositeFigi, c.CompositeFigi) || a.Section == nil || b.Section == nil || *a.Section == *b.Section || a.HoldingId == b.HoldingId || a.HoldingId == c.HoldingId {
					t.Fatal("shared product must retain separate position identities")
				}
				requireFinancialDecimal(t, a.CostJpy, "100.15")
				requireFinancialDecimal(t, b.CostJpy, "32")
				requireFinancialDecimal(t, c.CostJpy, "44")
				// Same names with different identifiers must remain distinguishable.
				if sbi.Holdings[2].Name != nrkn.Holdings[1].Name || reflect.DeepEqual(sbi.Holdings[2].CompositeFigi, nrkn.Holdings[1].CompositeFigi) {
					t.Fatal("missing same-name different-FIGI scenario")
				}
				if nrkn.Holdings[2].Name != nrkn.Holdings[3].Name || nrkn.Holdings[2].CompositeFigi != nil || nrkn.Holdings[3].CompositeFigi != nil || nrkn.Holdings[2].ProductCode == nil || nrkn.Holdings[3].ProductCode == nil || *nrkn.Holdings[2].ProductCode == *nrkn.Holdings[3].ProductCode {
					t.Fatal("missing same-name product-code fallback scenario")
				}
				zero, unknown := sbi.Holdings[3], sbi.Holdings[4]
				if zero.CompositeFigi != nil || zero.ProductCode != nil || unknown.CostJpy != nil || unknown.UnrealizedPnlJpy != nil {
					t.Fatal("missing identifiers/metrics must remain null")
				}
				requireFinancialDecimal(t, zero.ValuationJpy, "0")
				requireFinancialDecimal(t, zero.CostJpy, "0")
				requireFinancialDecimal(t, zero.UnrealizedPnlJpy, "0")
				requireFinancialDecimal(t, unknown.Quantity, "0.000001")
				requireFinancialDecimal(t, nrkn.Holdings[2].CostJpy, "0")
				// Validate JSON primitives as well: zero is a decimal string, unknown
				// is an explicit null, not an omitted field or a numeric zero.
				raw := getFinancialMockJSON[struct {
					Holdings []map[string]json.RawMessage `json:"holdings"`
				}](t, r, base+"/snapshots/"+url.PathEscape(sbi.SnapshotId))
				if string(raw.Holdings[3]["costJpy"]) != `"0"` || string(raw.Holdings[4]["costJpy"]) != "null" || string(raw.Holdings[4]["unrealizedPnlJpy"]) != "null" {
					t.Fatal("zero/null JSON representation changed")
				}
			})

			t.Run("missing sources and empty holdings", func(t *testing.T) {
				for _, tc := range []struct {
					at      string
					missing []openapi.Source
				}{
					{"1999-12-31T11:59:59+09:00", []openapi.Source{"nrkn", "sbi"}},
					{"1999-12-31T12:00:00+09:00", []openapi.Source{"nrkn"}},
				} {
					page := getFinancialMockJSON[openapi.BalancePage](t, r, base+"/balances?at="+url.QueryEscape(tc.at))
					point := page.Items[0]
					if !reflect.DeepEqual(point.MissingSources, tc.missing) || point.Totals.ValuationJpy != nil {
						t.Fatal("missing sources must not be treated as zero")
					}
					for _, source := range point.Sources {
						if source.Source == "sbi" && len(tc.missing) == 1 {
							requireFinancialDecimal(t, source.Totals.ValuationJpy, "0")
							if source.SnapshotId == nil {
								t.Fatal("zero-valued source is available")
							}
							detail := getFinancialMockJSON[openapi.SnapshotDetail](t, r, base+"/snapshots/"+url.PathEscape(*source.SnapshotId))
							if detail.Holdings == nil || len(detail.Holdings) != 0 {
								t.Fatal("empty holdings must be an array")
							}
						} else if source.SnapshotId != nil || source.FetchedAt != nil || source.Totals.ValuationJpy != nil {
							t.Fatal("unavailable source has fabricated data")
						}
					}
				}
				filtered := getFinancialMockJSON[openapi.BalancePage](t, r, base+"/balances?source=nrkn")
				if len(filtered.Items[0].Sources) != 1 || len(filtered.Items[0].MissingSources) != 0 {
					t.Fatal("source filter ignored")
				}
				requireFinancialDecimal(t, filtered.Items[0].Totals.ValuationJpy, "70")
				requireFinancialDecimal(t, filtered.Items[0].Totals.CostJpy, "75")
				requireFinancialDecimal(t, filtered.Items[0].Totals.UnrealizedPnlJpy, "-5")
			})

			t.Run("snapshot pagination keeps every holding", func(t *testing.T) {
				path := base + "/snapshots?source=sbi&source=nrkn&limit=2"
				next := path
				seen := map[string]bool{}
				var previous *openapi.SnapshotDetail
				var sizes []int
				for n := 0; ; n++ {
					if n >= 4 {
						t.Fatal("pagination did not terminate")
					}
					page := getFinancialMockJSON[openapi.SnapshotPage](t, r, next)
					if len(page.Items) == 0 || len(page.Items) > 2 {
						t.Fatal("unexpected snapshot page size")
					}
					for _, item := range page.Items {
						if seen[item.SnapshotId] || previous != nil && (item.FetchedAt.After(previous.FetchedAt) || item.FetchedAt.Equal(previous.FetchedAt) && item.SnapshotId >= previous.SnapshotId) {
							t.Fatal("duplicate or incorrectly ordered snapshot")
						}
						seen[item.SnapshotId] = true
						detail := getFinancialMockJSON[openapi.SnapshotDetail](t, r, base+"/snapshots/"+url.PathEscape(item.SnapshotId))
						if !reflect.DeepEqual(item, detail) {
							t.Fatal("list truncated or changed snapshot holdings")
						}
						sizes = append(sizes, len(item.Holdings))
						copy := item
						previous = &copy
					}
					if len(page.Items) < 2 {
						break
					}
					next = path + "&offset=" + strconv.Itoa((n+1)*2)
				}
				if !reflect.DeepEqual(sizes, []int{4, 60, 1, 1, 1, 1, 0}) {
					t.Fatal("missing snapshots or holdings")
				}
				empty := getFinancialMockJSON[openapi.SnapshotPage](t, r, base+"/snapshots?to=1999-12-30T00:00:00Z")
				if empty.Items == nil || len(empty.Items) != 0 {
					t.Fatal("empty history must terminate with an empty array")
				}
			})

			t.Run("daily continuation and monthly carry forward", func(t *testing.T) {
				path := base + "/balances?from=2000-01-31&to=2000-02-04&interval=day&limit=2"
				first := getFinancialMockJSON[openapi.BalancePage](t, r, path)
				if len(first.Items) != 2 {
					t.Fatal("missing first daily page")
				}
				second := getFinancialMockJSON[openapi.BalancePage](t, r, path+"&offset=2")
				if len(second.Items) != 2 {
					t.Fatal("missing second daily page")
				}
				end := getFinancialMockJSON[openapi.BalancePage](t, r, path+"&offset=4")
				if end.Items == nil || len(end.Items) != 0 {
					t.Fatal("daily pagination did not terminate")
				}
				points := append(first.Items, second.Items...)
				for i, want := range []string{"320.2", "460.25", "320.25", "320.25"} {
					requireFinancialDecimal(t, points[i].Totals.ValuationJpy, want)
					if points[i].PeriodStart == nil || points[i].PeriodEnd == nil || points[i].Timestamp != nil || i > 0 && *points[i-1].PeriodEnd != *points[i].PeriodStart {
						t.Fatal("daily periods overlap or have gaps")
					}
				}
				month := getFinancialMockJSON[openapi.BalancePage](t, r, base+"/balances?from=2000-01-01&to=2000-04-01&interval=month")
				if len(month.Items) != 3 {
					t.Fatal("unexpected monthly page")
				}
				for i, want := range []string{"320.2", "320.25", "320.25"} {
					requireFinancialDecimal(t, month.Items[i].Totals.ValuationJpy, want)
				}
				if !reflect.DeepEqual(month.Items[1].Sources, month.Items[2].Sources) {
					t.Fatal("carry forward must preserve source provenance")
				}
			})

			t.Run("offset validation and retired cursor", func(t *testing.T) {
				for _, path := range []string{
					"/snapshots?offset=-1", "/snapshots?offset=nope", "/snapshots?offset=1.5",
					"/snapshots?offset=", "/snapshots?offset=0&offset=1", "/snapshots?offset=9223372036854775808",
					"/snapshots?cursor=old", "/balances?cursor=old", "/balances?offset=0",
					"/balances?at=2000-01-01T00:00:00Z&offset=1",
					"/balances?from=2000-01-01&to=2000-01-05&offset=-1",
					"/balances?from=2000-01-01&to=2000-01-05&offset=0&offset=1",
					"/balances?from=2000-01-01&to=2000-01-05&offset=nope",
					"/balances?from=2000-01-01&to=2000-01-05&cursor=old",
				} {
					rec := httptest.NewRecorder()
					r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, base+path, nil))
					if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
						t.Fatalf("expected JSON 400 for %s", path)
					}
				}
				first := getFinancialMockJSON[openapi.SnapshotPage](t, r, base+"/snapshots?limit=2")
				zero := getFinancialMockJSON[openapi.SnapshotPage](t, r, base+"/snapshots?limit=2&offset=0")
				if !reflect.DeepEqual(first, zero) {
					t.Fatal("omitted offset must default to zero")
				}
				end := getFinancialMockJSON[openapi.SnapshotPage](t, r, base+"/snapshots?limit=2&offset=7")
				if end.Items == nil || len(end.Items) != 0 {
					t.Fatal("offset at end must return an empty array")
				}
			})
		})
	}
}
