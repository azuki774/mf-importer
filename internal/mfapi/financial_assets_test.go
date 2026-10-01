package mfapi

import (
	"context"
	"errors"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"reflect"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
)

func financialTestPtr[T any](v T) *T { return &v }
func financialTestDate(s string) *types.Date {
	v, e := time.Parse(time.DateOnly, s)
	if e != nil {
		panic(e)
	}
	return &types.Date{Time: v}
}

func TestFinancialSnapshotsPaginationAndDetail(t *testing.T) {
	s := newMockFinancialService()
	ctx := context.Background()
	p := openapi.ListFinancialAssetSnapshotsParams{
		Limit: financialTestPtr(1),
		From:  financialTestPtr(time.Date(2000, 1, 1, 0, 0, 0, 0, jst)),
		To:    financialTestPtr(time.Date(2000, 2, 1, 0, 0, 0, 0, jst)),
	}
	seen := map[string]bool{}
	previous := ""
	for i := 0; i < 4; i++ {
		page, err := s.list(ctx, p)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) != 1 {
			t.Fatal("page length")
		}
		item := page.Items[0]
		if seen[item.SnapshotId] {
			t.Fatal("duplicate snapshot")
		}
		seen[item.SnapshotId] = true
		if i == 1 && previous >= item.SnapshotId {
			t.Fatal("same timestamp ID ordering")
		}
		previous = item.SnapshotId
		detail, err := s.detail(ctx, item.SnapshotId)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(item, detail) || len(item.Holdings) != 1 {
			t.Fatal("list/detail mismatch or missing holdings")
		}
		if i < 3 && page.NextCursor == nil || i == 3 && page.NextCursor != nil {
			t.Fatal("cursor termination")
		}
		p.Cursor = page.NextCursor
	}
	if len(seen) != 4 {
		t.Fatal("missing snapshots")
	}
	_, err := s.detail(ctx, "invalid")
	if !errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
		t.Fatal("invalid id")
	}
	_, err = s.detail(ctx, snapshotID("sbi", 999))
	if !errors.Is(err, model.ErrRecordNotFound) {
		t.Fatal("missing id")
	}
}

func TestFinancialSnapshotFilterAndCursorValidation(t *testing.T) {
	s := newMockFinancialService()
	ctx := context.Background()
	page, err := s.list(ctx, openapi.ListFinancialAssetSnapshotsParams{Limit: financialTestPtr(1)})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []openapi.ListFinancialAssetSnapshotsParams{
		{Limit: financialTestPtr(2), Cursor: page.NextCursor},
		{Limit: financialTestPtr(1), Cursor: page.NextCursor, Source: financialTestPtr(openapi.Sources{"sbi"})},
		{Cursor: financialTestPtr("invalid")}, {Source: financialTestPtr(openapi.Sources{"sbi", "sbi"})},
		{Limit: financialTestPtr(0)}, {Limit: financialTestPtr(501)}, {Source: financialTestPtr(openapi.Sources{})},
	} {
		if _, err := s.list(ctx, p); !errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
			t.Fatal("invalid list accepted")
		}
	}
	t0 := s.repo.(memoryFinancialRepository).rows[0].FetchedAt
	t1 := t0.AddDate(0, 0, 2)
	p, err := s.list(ctx, openapi.ListFinancialAssetSnapshotsParams{From: &t0, To: &t1})
	if err != nil || len(p.Items) != 2 {
		t.Fatal("inclusive from/exclusive to")
	}
}

func TestFinancialSingleBalanceBoundaryAndNulls(t *testing.T) {
	s := newMockFinancialService()
	at := s.repo.(memoryFinancialRepository).rows[0].FetchedAt
	s.now = func() time.Time { return at }
	for _, p := range []openapi.GetFinancialAssetBalancesParams{{}, {At: &at}} {
		page, err := s.balances(context.Background(), p)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) != 1 || page.NextCursor != nil {
			t.Fatal("single page")
		}
		point := page.Items[0]
		if point.Timestamp == nil || point.PeriodStart != nil || len(point.MissingSources) != 0 || *point.Totals.ValuationJpy != "300.1" || point.Totals.CostJpy != nil {
			t.Fatal("single totals/boundary")
		}
	}
	before := time.Date(1999, 12, 31, 12, 0, 0, 0, jst).Add(-time.Nanosecond)
	page, err := s.balances(context.Background(), openapi.GetFinancialAssetBalancesParams{At: &before})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items[0].MissingSources) != 2 || page.Items[0].Totals.ValuationJpy != nil {
		t.Fatal("missing should be null")
	}
}

func TestFinancialDailyAndMonthlyBalances(t *testing.T) {
	s := newMockFinancialService()
	ctx := context.Background()
	p := openapi.GetFinancialAssetBalancesParams{From: financialTestDate("2000-01-01"), To: financialTestDate("2000-01-05"), Limit: financialTestPtr(2)}
	first, err := s.balances(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Items) != 2 || first.NextCursor == nil || *first.Items[1].Totals.ValuationJpy != "300.1" {
		t.Fatal("daily carry")
	}
	p.Cursor = first.NextCursor
	second, err := s.balances(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Items) != 2 || second.NextCursor != nil || *second.Items[0].Totals.ValuationJpy != "320.2" || *second.Items[0].PeriodStart != "2000-01-03" {
		t.Fatal("daily continuation")
	}
	p.From = financialTestDate("2000-01-02")
	if _, err := s.balances(ctx, p); !errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
		t.Fatal("changed range cursor accepted")
	}
	month, err := s.balances(ctx, openapi.GetFinancialAssetBalancesParams{From: financialTestDate("2000-01-02"), To: financialTestDate("2000-02-03"), Interval: financialTestPtr(openapi.Month)})
	if err != nil {
		t.Fatal(err)
	}
	if len(month.Items) != 2 || *month.Items[0].PeriodStart != "2000-01-02" || *month.Items[0].PeriodEnd != "2000-02-01" || *month.Items[1].PeriodEnd != "2000-02-03" {
		t.Fatal("clipped calendar month")
	}
}

func TestFinancialPeriodBoundaryAndLatestNull(t *testing.T) {
	s := newMockFinancialService()
	repo := s.repo.(memoryFinancialRepository)
	// Hand-written boundary rows: the later snapshot deliberately has no value.
	at := time.Date(2000, 1, 2, 0, 0, 0, 0, jst)
	repo.rows = []model.FinancialSnapshot{{ID: 1, Source: "sbi", FetchedAt: at.Add(-time.Hour), ValuationJpy: financialTestPtr("1.25")}, {ID: 2, Source: "sbi", FetchedAt: at}}
	s.repo = repo
	p := openapi.GetFinancialAssetBalancesParams{Source: financialTestPtr(openapi.Sources{"sbi"}), From: financialTestDate("2000-01-01"), To: financialTestDate("2000-01-03")}
	page, err := s.balances(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	if *page.Items[0].Totals.ValuationJpy != "1.25" || page.Items[1].Totals.ValuationJpy != nil || len(page.Items[1].MissingSources) != 0 {
		t.Fatal("period end exclusive; null must not be filled from older row")
	}
}

func TestFinancialBalanceValidation(t *testing.T) {
	s := newMockFinancialService()
	date := financialTestDate("2000-01-01")
	for _, p := range []openapi.GetFinancialAssetBalancesParams{
		{From: date}, {To: date}, {From: date, To: date},
		{At: financialTestPtr(time.Now()), From: date, To: financialTestDate("2000-01-02")},
		{Limit: financialTestPtr(1)}, {Interval: financialTestPtr(openapi.Day)}, {Cursor: financialTestPtr("x")},
		{From: date, To: financialTestDate("2000-01-02"), Interval: financialTestPtr(openapi.GetFinancialAssetBalancesParamsInterval("week"))},
	} {
		if _, err := s.balances(context.Background(), p); !errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
			t.Fatal("invalid balance accepted")
		}
	}
}

func TestFinancialExactDecimalAddition(t *testing.T) {
	for _, tc := range []struct{ a, b, want string }{{"0.1", "0.2", "0.3"}, {"999999999999.99", "0.01", "1000000000000"}, {"-0.10", "0.10", "0"}} {
		v, err := sumFinancialDecimals([]*string{&tc.a, &tc.b})
		if err != nil || v == nil || *v != tc.want {
			t.Fatal("decimal precision")
		}
	}
	if v, e := sumFinancialDecimals([]*string{nil, financialTestPtr("1")}); v != nil || e != nil {
		t.Fatal("null sum")
	}
}

func TestFinancialHoldingCostCalculation(t *testing.T) {
	// Hand-written values exercise gains, losses, zero and precision; none are
	// taken from actual holdings or downloaded financial data.
	for _, tc := range []struct {
		name, source           string
		value, pnl, cost, want *string
	}{
		{"gain", "sbi", financialTestPtr("120.25"), financialTestPtr("20.10"), nil, financialTestPtr("100.15")},
		{"loss", "sbi", financialTestPtr("80.25"), financialTestPtr("-20.10"), nil, financialTestPtr("100.35")},
		{"zero", "sbi", financialTestPtr("0.00"), financialTestPtr("0.00"), nil, financialTestPtr("0")},
		{"precision", "sbi", financialTestPtr("999999999999.99"), financialTestPtr("-0.01"), nil, financialTestPtr("1000000000000")},
		{"missing value", "sbi", nil, financialTestPtr("1"), nil, nil},
		{"missing pnl", "sbi", financialTestPtr("1"), nil, nil, nil},
		{"imported zero preserved", "nrkn", financialTestPtr("120"), financialTestPtr("20"), financialTestPtr("0"), financialTestPtr("0")},
		{"imported cost preserved", "nrkn", financialTestPtr("120"), financialTestPtr("20"), financialTestPtr("99"), financialTestPtr("99")},
		{"NRKN not inferred", "nrkn", financialTestPtr("120"), financialTestPtr("20"), nil, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := financialHoldingCost(tc.source, model.FinancialHolding{ValuationJpy: tc.value, UnrealizedPnlJpy: tc.pnl, CostJpy: tc.cost})
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatal("incorrect calculated cost")
			}
		})
	}
	for _, h := range []model.FinancialHolding{
		{ValuationJpy: financialTestPtr("invalid"), UnrealizedPnlJpy: financialTestPtr("1")},
		{ValuationJpy: financialTestPtr("1"), UnrealizedPnlJpy: financialTestPtr("invalid")},
	} {
		if _, err := financialHoldingCost("sbi", h); err == nil {
			t.Fatal("invalid decimal accepted")
		}
	}
}

func TestFinancialCalculatedCostConsistentAcrossListAndDetail(t *testing.T) {
	s := newMockFinancialService()
	p, err := s.list(context.Background(), openapi.ListFinancialAssetSnapshotsParams{
		Source: financialTestPtr(openapi.Sources{"sbi"}),
		From:   financialTestPtr(time.Date(2000, 1, 1, 0, 0, 0, 0, jst)),
		To:     financialTestPtr(time.Date(2000, 2, 1, 0, 0, 0, 0, jst)),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Items) == 0 {
		t.Fatal("missing synthetic snapshots")
	}
	for _, row := range p.Items {
		if row.Totals.CostJpy != nil || row.Totals.UnrealizedPnlJpy != nil {
			t.Fatal("snapshot totals must not be inferred from holdings")
		}
		detail, err := s.detail(context.Background(), row.SnapshotId)
		if err != nil || !reflect.DeepEqual(detail, row) {
			t.Fatal("list/detail cost mismatch")
		}
		if len(row.Holdings) == 0 {
			t.Fatal("missing synthetic holdings")
		}
		for _, h := range row.Holdings {
			if h.CostJpy == nil || *h.CostJpy != "101.1" {
				t.Fatal("missing calculated holding cost")
			}
		}
	}
	// Derivation affects only API output, not stored read models.
	for _, row := range s.repo.(memoryFinancialRepository).rows {
		if row.Source == "sbi" {
			for _, h := range row.Holdings {
				if h.CostJpy != nil {
					t.Fatal("mutated stored holding cost")
				}
			}
		}
	}
}

func TestFinancialCursorPositionValidation(t *testing.T) {
	s := newMockFinancialService()
	ctx := context.Background()
	page, err := s.list(ctx, openapi.ListFinancialAssetSnapshotsParams{Limit: financialTestPtr(1)})
	if err != nil {
		t.Fatal(err)
	}
	filter := financialFilter([]string{"nrkn", "sbi"}, "", "", "", 1)
	cursor, err := decodeFinancialCursor(*page.NextCursor, "snapshots", filter)
	if err != nil {
		t.Fatal(err)
	}
	cursor.FetchedAt = cursor.FetchedAt.Add(time.Hour)
	_, err = s.list(ctx, openapi.ListFinancialAssetSnapshotsParams{Limit: financialTestPtr(1), Cursor: financialTestPtr(encodeFinancialCursor(cursor))})
	if !errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
		t.Fatal("timestamp/ID mismatch accepted")
	}
	cursor.ID = snapshotID("nrkn", 999)
	_, err = s.list(ctx, openapi.ListFinancialAssetSnapshotsParams{Limit: financialTestPtr(1), Cursor: financialTestPtr(encodeFinancialCursor(cursor))})
	if !errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
		t.Fatal("nonexistent cursor row accepted")
	}
	p := openapi.GetFinancialAssetBalancesParams{From: financialTestDate("2000-01-01"), To: financialTestDate("2000-01-10"), Limit: financialTestPtr(2)}
	balances, err := s.balances(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	filter = financialFilter([]string{"nrkn", "sbi"}, "2000-01-01", "2000-01-10", "day", 2)
	cursor, err = decodeFinancialCursor(*balances.NextCursor, "balances", filter)
	if err != nil {
		t.Fatal(err)
	}
	cursor.Period = "2000-01-02"
	p.Cursor = financialTestPtr(encodeFinancialCursor(cursor))
	if _, err := s.balances(ctx, p); !errors.Is(err, model.ErrInvalidFinancialAssetRequest) {
		t.Fatal("non-page boundary accepted")
	}
}

func TestFinancialMockTotalsDoNotComeFromHoldings(t *testing.T) {
	s := newMockFinancialService()
	ctx := context.Background()
	at := time.Date(2000, 2, 3, 0, 0, 0, 0, jst)
	page, err := s.balances(ctx, openapi.GetFinancialAssetBalancesParams{At: &at})
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range page.Items[0].Sources {
		if source.SnapshotId == nil {
			t.Fatal("missing synthetic source")
		}
		detail, err := s.detail(ctx, *source.SnapshotId)
		if err != nil {
			t.Fatal(err)
		}
		var values, costs, pnls []*string
		for _, h := range detail.Holdings {
			values = append(values, h.ValuationJpy)
			costs = append(costs, h.CostJpy)
			pnls = append(pnls, h.UnrealizedPnlJpy)
		}
		value, err := sumFinancialDecimals(values)
		if err != nil {
			t.Fatal(err)
		}
		cost, err := sumFinancialDecimals(costs)
		if err != nil {
			t.Fatal(err)
		}
		pnl, err := sumFinancialDecimals(pnls)
		if err != nil {
			t.Fatal(err)
		}
		if source.Source == "sbi" {
			if value == nil || *value != "220.25" || detail.Totals.ValuationJpy == nil || *detail.Totals.ValuationJpy != "250.25" {
				t.Fatal("SBI summary must include the synthetic balance absent from holdings")
			}
			if cost != nil || pnl != nil || detail.Totals.CostJpy != nil || detail.Totals.UnrealizedPnlJpy != nil {
				t.Fatal("unknown metrics must not become partial totals")
			}
		} else if value == nil || *value != "70" || cost == nil || *cost != "75" || pnl == nil || *pnl != "-5" {
			t.Fatal("unexpected complete NRKN holding totals")
		}
	}
}
