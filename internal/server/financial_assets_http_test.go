package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"mf-importer/internal/mfapi"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
)

type financialAssetHTTPService struct {
	APIService
	err    error
	called bool
}

func (s *financialAssetHTTPService) ListFinancialAssetSnapshots(context.Context, openapi.ListFinancialAssetSnapshotsParams) (openapi.SnapshotPage, error) {
	s.called = true
	return openapi.SnapshotPage{}, s.err
}
func (s *financialAssetHTTPService) GetFinancialAssetSnapshot(context.Context, string) (openapi.SnapshotDetail, error) {
	s.called = true
	return openapi.SnapshotDetail{}, s.err
}
func (s *financialAssetHTTPService) GetFinancialAssetBalances(context.Context, openapi.GetFinancialAssetBalancesParams) (openapi.BalancePage, error) {
	s.called = true
	return openapi.BalancePage{}, s.err
}

func TestFinancialAssetHTTPResponses(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		err        error
		status     int
		called     bool
		body       string
	}{
		{"ok", "/financial-assets/snapshots", nil, 200, true, ""},
		{"invalid", "/financial-assets/snapshots", model.ErrInvalidFinancialAssetRequest, 400, true, "invalid financial asset request"},
		{"missing", "/financial-assets/snapshots/fake-id", model.ErrRecordNotFound, 404, true, "financial asset snapshot not found"},
		{"failure", "/financial-assets/balances", errors.New("sensitive internal detail"), 500, true, "internal server error"},
		{"unknown", "/financial-assets/snapshots?unexpected=x", nil, 400, false, "invalid request parameters"},
		{"duplicate", "/financial-assets/snapshots?limit=1&limit=2", nil, 400, false, "invalid request parameters"},
		{"duplicate_source", "/financial-assets/snapshots?source=sbi&source=sbi", nil, 400, false, "invalid request parameters"},
		{"empty", "/financial-assets/snapshots?cursor=", nil, 400, false, "invalid request parameters"},
		{"bad_bind", "/financial-assets/snapshots?limit=nope", nil, 400, false, "invalid request parameters"},
		{"wrong_endpoint_parameter", "/financial-assets/snapshots?at=2000-01-01T00:00:00Z", nil, 400, false, "invalid request parameters"},
		{"detail_parameter", "/financial-assets/snapshots/fake-id?source=sbi", nil, 400, false, "invalid request parameters"},
		{"invalid_encoding", "/financial-assets/balances?unknown=%zz", nil, 400, false, "invalid request parameters"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &financialAssetHTTPService{err: tc.err}
			r := chi.NewRouter()
			registerAPI(&apigateway{APIService: svc}, r)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest("GET", tc.path, nil))
			if rec.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", rec.Code, tc.status, rec.Body.String())
			}
			if svc.called != tc.called {
				t.Fatalf("called %v, want %v", svc.called, tc.called)
			}
			if tc.status != 200 && (!strings.Contains(rec.Body.String(), tc.body) || strings.Contains(rec.Body.String(), "sensitive internal detail")) {
				t.Fatalf("unexpected safe error body: %s", rec.Body.String())
			}
			if tc.status == 200 && !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
				t.Fatal("missing JSON content type")
			}
		})
	}
}

func TestFinancialAssetHTTPErrorsOnStaticAPIMount(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api", func(sub chi.Router) { registerAPI(&apigateway{APIService: &financialAssetHTTPService{}}, sub) })
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/api/financial-assets/balances?source=bad", nil))
	if rec.Code != 400 || !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
}

func TestFinancialAssetHTTPWithMockService(t *testing.T) {
	r := chi.NewRouter()
	registerAPI(&apigateway{APIService: &mfapi.MockAPIService{}}, r)
	get := func(path string, status int, out any) {
		t.Helper()
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != status || !strings.Contains(rec.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("unexpected status/content type for %s", path)
		}
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatal(err)
		}
	}
	var page openapi.SnapshotPage
	get("/financial-assets/snapshots?limit=1", 200, &page)
	if len(page.Items) != 1 || len(page.Items[0].Holdings) != 1 || page.NextCursor == nil {
		t.Fatal("missing detailed snapshot page")
	}
	var detail openapi.SnapshotDetail
	get("/financial-assets/snapshots/"+page.Items[0].SnapshotId, 200, &detail)
	a, _ := json.Marshal(detail)
	b, _ := json.Marshal(page.Items[0])
	if string(a) != string(b) {
		t.Fatal("list/detail inconsistent")
	}
	get("/financial-assets/snapshots?limit=1&cursor="+url.QueryEscape(*page.NextCursor), 200, &page)
	for _, query := range []string{"", "?at=2000-01-01T12:00:00%2B09:00", "?from=2000-01-01&to=2000-01-05", "?from=2000-01-02&to=2000-02-03&interval=month"} {
		var balances openapi.BalancePage
		get("/financial-assets/balances"+query, 200, &balances)
		if len(balances.Items) == 0 {
			t.Fatal("empty mock balance")
		}
	}
	for _, query := range []string{"?from=2000-01-01", "?from=2000-01-01&to=2000-01-01", "?at=2000-01-01T00:00:00Z&interval=day", "?limit=0"} {
		var e openapi.ApiError
		get("/financial-assets/balances"+query, 400, &e)
	}
}
