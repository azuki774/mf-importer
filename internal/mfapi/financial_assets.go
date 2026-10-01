package mfapi

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"regexp"
	"sort"
	"strings"
	"time"
)

type financialRepository interface {
	ListFinancialSnapshots(context.Context, model.FinancialSnapshotQuery) ([]model.FinancialSnapshot, error)
	GetFinancialSnapshot(context.Context, string, int64) (model.FinancialSnapshot, error)
	GetFinancialHoldings(context.Context, string, []int64) ([]model.FinancialHolding, error)
	LatestFinancialSnapshot(context.Context, string, time.Time, bool) (*model.FinancialSnapshot, error)
}

type financialService struct {
	repo financialRepository
	now  func() time.Time
}

func (a *APIService) ListFinancialAssetSnapshots(ctx context.Context, p openapi.ListFinancialAssetSnapshotsParams) (openapi.SnapshotPage, error) {
	return (financialService{a.financialRepo, time.Now}).list(ctx, p)
}

func (a *APIService) GetFinancialAssetSnapshot(ctx context.Context, id string) (openapi.SnapshotDetail, error) {
	return (financialService{a.financialRepo, time.Now}).detail(ctx, id)
}

func (a *APIService) GetFinancialAssetBalances(ctx context.Context, p openapi.GetFinancialAssetBalancesParams) (openapi.BalancePage, error) {
	return (financialService{a.financialRepo, time.Now}).balances(ctx, p)
}

func financialSources(input *openapi.Sources) ([]string, error) {
	sources := []string{"nrkn", "sbi"}
	if input != nil {
		sources = append([]string{}, (*input)...)
	}
	if len(sources) < 1 || len(sources) > 2 {
		return nil, model.ErrInvalidFinancialAssetRequest
	}
	sort.Strings(sources)
	for i, s := range sources {
		if (s != "sbi" && s != "nrkn") || (i > 0 && sources[i-1] == s) {
			return nil, model.ErrInvalidFinancialAssetRequest
		}
	}
	return sources, nil
}

func financialLimit(p *int) (int, error) {
	if p == nil {
		return 100, nil
	}
	if *p < 1 || *p > 500 {
		return 0, model.ErrInvalidFinancialAssetRequest
	}
	return *p, nil
}

func validFinancialTime(t time.Time) bool { y := t.In(jst).Year(); return y >= 1000 && y <= 9999 }

func (s financialService) list(ctx context.Context, p openapi.ListFinancialAssetSnapshotsParams) (openapi.SnapshotPage, error) {
	result := openapi.SnapshotPage{Items: []openapi.SnapshotDetail{}}
	sources, err := financialSources(p.Source)
	if err != nil {
		return result, err
	}
	limit, err := financialLimit(p.Limit)
	if err != nil {
		return result, err
	}
	if p.From != nil && !validFinancialTime(*p.From) || p.To != nil && !validFinancialTime(*p.To) || p.From != nil && p.To != nil && !p.From.Before(*p.To) {
		return result, model.ErrInvalidFinancialAssetRequest
	}
	filter := financialFilter(sources, timeKey(p.From), timeKey(p.To), "snapshots-desc", limit)
	var after financialCursor
	if p.Cursor != nil {
		after, err = decodeFinancialCursor(*p.Cursor, "snapshots", filter)
		if err != nil {
			return result, err
		}
		source, id, e := parseSnapshotID(after.ID)
		if e != nil || !containsSource(sources, source) || !validFinancialTime(after.FetchedAt) || after.FetchedAt.Nanosecond()%1000 != 0 || after.Period != "" || p.From != nil && after.FetchedAt.Before(*p.From) || p.To != nil && !after.FetchedAt.Before(*p.To) {
			return result, model.ErrInvalidFinancialAssetRequest
		}
		row, e := s.repo.GetFinancialSnapshot(ctx, source, id)
		if errors.Is(e, model.ErrRecordNotFound) {
			return result, model.ErrInvalidFinancialAssetRequest
		}
		if e != nil {
			return result, e
		}
		if !row.FetchedAt.Equal(after.FetchedAt) {
			return result, model.ErrInvalidFinancialAssetRequest
		}
	}
	var snapshots []model.FinancialSnapshot
	for _, source := range sources {
		q := model.FinancialSnapshotQuery{Source: source, From: p.From, To: p.To, Limit: limit + 1}
		if p.Cursor != nil {
			lastSource, lastID, _ := parseSnapshotID(after.ID)
			q.AfterFetchedAt = &after.FetchedAt
			switch {
			case source == lastSource:
				q.AfterID = lastID
			case source < lastSource:
				// Include every ID from lower-sorting sources at the cursor time.
				q.AfterID = -1
			}
		}
		rows, e := s.repo.ListFinancialSnapshots(ctx, q)
		if e != nil {
			return result, e
		}
		snapshots = append(snapshots, rows...)
	}
	sort.Slice(snapshots, func(i, j int) bool {
		a, b := snapshots[i], snapshots[j]
		if !a.FetchedAt.Equal(b.FetchedAt) {
			return a.FetchedAt.After(b.FetchedAt)
		}
		return snapshotID(a.Source, a.ID) > snapshotID(b.Source, b.ID)
	})
	if len(snapshots) > limit {
		snapshots = snapshots[:limit]
		last := snapshots[len(snapshots)-1]
		cursor := encodeFinancialCursor(financialCursor{Version: 1, Kind: "snapshots", Filter: filter, ID: snapshotID(last.Source, last.ID), FetchedAt: last.FetchedAt})
		result.NextCursor = &cursor
	}
	for _, source := range sources {
		ids := []int64{}
		positions := map[int64]int{}
		for i, row := range snapshots {
			if row.Source == source {
				ids = append(ids, row.ID)
				positions[row.ID] = i
			}
		}
		if len(ids) == 0 {
			continue
		}
		holdings, e := s.repo.GetFinancialHoldings(ctx, source, ids)
		if e != nil {
			return result, e
		}
		for _, h := range holdings {
			if i, ok := positions[h.SnapshotID]; ok {
				snapshots[i].Holdings = append(snapshots[i].Holdings, h)
			}
		}
	}
	for _, row := range snapshots {
		detail, err := financialDetail(row)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, detail)
	}
	return result, nil
}

func (s financialService) detail(ctx context.Context, id string) (openapi.SnapshotDetail, error) {
	source, dbID, err := parseSnapshotID(id)
	if err != nil {
		return openapi.SnapshotDetail{}, err
	}
	row, err := s.repo.GetFinancialSnapshot(ctx, source, dbID)
	if err != nil {
		return openapi.SnapshotDetail{}, err
	}
	row.Holdings, err = s.repo.GetFinancialHoldings(ctx, source, []int64{dbID})
	if err != nil {
		return openapi.SnapshotDetail{}, err
	}
	return financialDetail(row)
}

func financialTotals(row model.FinancialSnapshot) openapi.Totals {
	return openapi.Totals{ValuationJpy: row.ValuationJpy, CostJpy: row.CostJpy, UnrealizedPnlJpy: row.UnrealizedPnlJpy}
}

func financialDetail(row model.FinancialSnapshot) (openapi.SnapshotDetail, error) {
	result := openapi.SnapshotDetail{SnapshotId: snapshotID(row.Source, row.ID), Source: openapi.Source(row.Source), FetchedAt: row.FetchedAt.In(jst), ImportedAt: row.ImportedAt.In(jst), Totals: financialTotals(row), Holdings: []openapi.Holding{}}
	sort.Slice(row.Holdings, func(i, j int) bool { return row.Holdings[i].ID < row.Holdings[j].ID })
	for _, h := range row.Holdings {
		cost, err := financialHoldingCost(row.Source, h)
		if err != nil {
			return openapi.SnapshotDetail{}, err
		}
		result.Holdings = append(result.Holdings, openapi.Holding{HoldingId: fmt.Sprintf("v1:%s:%020d:%020d", row.Source, row.ID, h.ID), Section: h.Section, ProductCode: h.ProductCode, CompositeFigi: h.CompositeFigi, ReferenceDate: h.ReferenceDate, Name: h.Name, Quantity: h.Quantity, ValuationJpy: h.ValuationJpy, CostJpy: cost, UnrealizedPnlJpy: h.UnrealizedPnlJpy})
	}
	return result, nil
}

// financialHoldingCost calculates the missing SBI holding cost in API responses:
// costJpy = valuationJpy - unrealizedPnlJpy. Both operands are stored JPY values
// from the same snapshot; exact decimal arithmetic avoids float64 rounding.
// This is a calculated value, not a separately imported acquisition-cost field,
// and is not written back to the DB. Never use quantity * unit_cost here: unit
// cost may be in another currency or quoted per multiple fund units. Existing
// costs (including NRKN's imported value) take precedence; missing operands stay
// null. Snapshot totals are not derived because holdings do not cover all assets.
func financialHoldingCost(source string, h model.FinancialHolding) (*string, error) {
	if h.CostJpy != nil || source != "sbi" {
		return h.CostJpy, nil
	}
	if h.ValuationJpy == nil || h.UnrealizedPnlJpy == nil {
		return nil, nil
	}
	if !financialDecimalPattern.MatchString(*h.UnrealizedPnlJpy) {
		return nil, errors.New("invalid stored financial decimal")
	}
	negativePnl := "-" + *h.UnrealizedPnlJpy
	if strings.HasPrefix(*h.UnrealizedPnlJpy, "-") {
		negativePnl = strings.TrimPrefix(*h.UnrealizedPnlJpy, "-")
	}
	return sumFinancialDecimals([]*string{h.ValuationJpy, &negativePnl})
}

func (s financialService) balances(ctx context.Context, p openapi.GetFinancialAssetBalancesParams) (openapi.BalancePage, error) {
	result := openapi.BalancePage{Items: []openapi.BalancePoint{}}
	sources, err := financialSources(p.Source)
	if err != nil {
		return result, err
	}
	rangeMode := p.From != nil || p.To != nil
	if !rangeMode {
		if p.Interval != nil || p.Limit != nil || p.Cursor != nil {
			return result, model.ErrInvalidFinancialAssetRequest
		}
		at := s.now()
		if p.At != nil {
			at = *p.At
		}
		if !validFinancialTime(at) {
			return result, model.ErrInvalidFinancialAssetRequest
		}
		latest := map[string]*model.FinancialSnapshot{}
		for _, source := range sources {
			row, e := s.repo.LatestFinancialSnapshot(ctx, source, at, true)
			if e != nil {
				return result, e
			}
			latest[source] = row
		}
		point, e := financialBalancePoint(sources, latest)
		if e != nil {
			return result, e
		}
		timestamp := at.In(jst).Format(time.RFC3339Nano)
		point.Timestamp = &timestamp
		result.Items = append(result.Items, point)
		return result, nil
	}
	if p.At != nil || p.From == nil || p.To == nil {
		return result, model.ErrInvalidFinancialAssetRequest
	}
	from := financialDate(p.From.Time)
	to := financialDate(p.To.Time)
	if !validFinancialTime(from) || !validFinancialTime(to) || !from.Before(to) {
		return result, model.ErrInvalidFinancialAssetRequest
	}
	interval := "day"
	if p.Interval != nil {
		interval = string(*p.Interval)
	}
	if interval != "day" && interval != "month" {
		return result, model.ErrInvalidFinancialAssetRequest
	}
	limit, err := financialLimit(p.Limit)
	if err != nil {
		return result, err
	}
	filter := financialFilter(sources, from.Format(time.DateOnly), to.Format(time.DateOnly), interval, limit)
	start := from
	if p.Cursor != nil {
		c, e := decodeFinancialCursor(*p.Cursor, "balances", filter)
		if e != nil {
			return result, e
		}
		start, e = time.ParseInLocation(time.DateOnly, c.Period, jst)
		if e != nil || !start.After(from) || !start.Before(to) || c.ID != "" || !c.FetchedAt.IsZero() || interval == "month" && start.Day() != 1 {
			return result, model.ErrInvalidFinancialAssetRequest
		}
		// A continuation must start on a page boundary of this exact range.
		// Count calendar days in UTC to avoid timezone/DST duration differences.
		steps := int((time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC).Unix() - time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC).Unix()) / 86400)
		if interval == "month" {
			steps = (start.Year()-from.Year())*12 + int(start.Month()-from.Month())
		}
		if steps%limit != 0 {
			return result, model.ErrInvalidFinancialAssetRequest
		}
	}
	type period struct{ start, end time.Time }
	periods := []period{}
	for t := start; t.Before(to) && len(periods) < limit; {
		end := t.AddDate(0, 0, 1)
		if interval == "month" {
			end = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, jst)
		}
		if end.After(to) {
			end = to
		}
		periods = append(periods, period{t, end})
		t = end
	}
	pageEnd := periods[len(periods)-1].end
	if pageEnd.Before(to) {
		v := encodeFinancialCursor(financialCursor{Version: 1, Kind: "balances", Filter: filter, Period: pageEnd.Format(time.DateOnly)})
		result.NextCursor = &v
	}
	latest := map[string]*model.FinancialSnapshot{}
	events := map[string][]model.FinancialSnapshot{}
	indexes := map[string]int{}
	for _, source := range sources {
		row, e := s.repo.LatestFinancialSnapshot(ctx, source, start, false)
		if e != nil {
			return result, e
		}
		latest[source] = row
		rows, e := s.repo.ListFinancialSnapshots(ctx, model.FinancialSnapshotQuery{Source: source, From: &start, To: &pageEnd})
		if e != nil {
			return result, e
		}
		// Period aggregation consumes events oldest first; list reads are newest first.
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
		events[source] = rows
	}
	for _, period := range periods {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		for _, source := range sources {
			rows := events[source]
			i := indexes[source]
			for i < len(rows) && rows[i].FetchedAt.Before(period.end) {
				latest[source] = &rows[i]
				i++
			}
			indexes[source] = i
		}
		point, e := financialBalancePoint(sources, latest)
		if e != nil {
			return result, e
		}
		a, b := period.start.Format(time.DateOnly), period.end.Format(time.DateOnly)
		point.PeriodStart = &a
		point.PeriodEnd = &b
		result.Items = append(result.Items, point)
	}
	return result, nil
}

func financialDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, jst)
}

func financialBalancePoint(sources []string, latest map[string]*model.FinancialSnapshot) (openapi.BalancePoint, error) {
	point := openapi.BalancePoint{Sources: []openapi.BalanceSource{}, MissingSources: []openapi.Source{}}
	values, costs, pnls := []*string{}, []*string{}, []*string{}
	for _, source := range sources {
		part := openapi.BalanceSource{Source: openapi.Source(source)}
		if row := latest[source]; row != nil {
			id := snapshotID(source, row.ID)
			at := row.FetchedAt.In(jst).Format(time.RFC3339Nano)
			part.SnapshotId = &id
			part.FetchedAt = &at
			part.Totals = financialTotals(*row)
		} else {
			point.MissingSources = append(point.MissingSources, openapi.Source(source))
		}
		point.Sources = append(point.Sources, part)
		values = append(values, part.Totals.ValuationJpy)
		costs = append(costs, part.Totals.CostJpy)
		pnls = append(pnls, part.Totals.UnrealizedPnlJpy)
	}
	var err error
	if point.Totals.ValuationJpy, err = sumFinancialDecimals(values); err != nil {
		return point, err
	}
	if point.Totals.CostJpy, err = sumFinancialDecimals(costs); err != nil {
		return point, err
	}
	point.Totals.UnrealizedPnlJpy, err = sumFinancialDecimals(pnls)
	return point, err
}

var financialDecimalPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)

// Decimal strings are added exactly, never via float64. This preserves the SQL
// DECIMAL precision in both individual responses and cross-source totals.
func sumFinancialDecimals(values []*string) (*string, error) {
	total := new(big.Rat)
	scale := 0
	for _, value := range values {
		if value == nil {
			return nil, nil
		}
		if !financialDecimalPattern.MatchString(*value) {
			return nil, errors.New("invalid stored financial decimal")
		}
		if i := strings.IndexByte(*value, '.'); i >= 0 && len(*value)-i-1 > scale {
			scale = len(*value) - i - 1
		}
		r, ok := new(big.Rat).SetString(*value)
		if !ok {
			return nil, errors.New("invalid stored financial decimal")
		}
		total.Add(total, r)
	}
	v := total.FloatString(scale)
	if strings.Contains(v, ".") {
		v = strings.TrimRight(strings.TrimRight(v, "0"), ".")
	}
	if v == "-0" {
		v = "0"
	}
	return &v, nil
}
