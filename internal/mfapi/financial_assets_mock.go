package mfapi

import (
	"context"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"sort"
	"time"
)

func (m *MockAPIService) ListFinancialAssetSnapshots(ctx context.Context, p openapi.ListFinancialAssetSnapshotsParams) (openapi.SnapshotPage, error) {
	return newMockFinancialService().list(ctx, p)
}
func (m *MockAPIService) GetFinancialAssetSnapshot(ctx context.Context, id string) (openapi.SnapshotDetail, error) {
	return newMockFinancialService().detail(ctx, id)
}
func (m *MockAPIService) GetFinancialAssetBalances(ctx context.Context, p openapi.GetFinancialAssetBalancesParams) (openapi.BalancePage, error) {
	return newMockFinancialService().balances(ctx, p)
}

func newMockFinancialService() financialService {
	return financialService{repo: syntheticFinancialRepository(), now: time.Now}
}

// These fixed, hand-written snapshots are entirely synthetic. They never load
// personal financial data, and exercise the same conversion/aggregation as DB mode.
func syntheticFinancialRepository() memoryFinancialRepository {
	text := func(s string) *string { return &s }
	t0 := time.Date(2000, 1, 1, 12, 0, 0, 0, jst)
	t1 := time.Date(2000, 1, 3, 12, 0, 0, 0, jst)
	return memoryFinancialRepository{rows: []model.FinancialSnapshot{
		{ID: 1, Source: "sbi", FetchedAt: t0, ImportedAt: t0.Add(time.Hour), ValuationJpy: text("100.10"), Holdings: []model.FinancialHolding{{ID: 1, SnapshotID: 1, Name: "ダミー商品A", Section: text("nisa_funds"), CompositeFigi: text("TESTFIGI0001"), Quantity: text("2"), ValuationJpy: text("100.10"), UnrealizedPnlJpy: text("-1.00")}}},
		{ID: 1, Source: "nrkn", FetchedAt: t0, ImportedAt: t0.Add(time.Hour), ValuationJpy: text("200"), CostJpy: text("220"), UnrealizedPnlJpy: text("-20"), Holdings: []model.FinancialHolding{{ID: 1, SnapshotID: 1, Name: "ダミー商品B", ProductCode: text("TEST001"), CompositeFigi: text("TESTFIGI0002"), ReferenceDate: text("2000-01-01"), Quantity: text("3"), ValuationJpy: text("200"), CostJpy: text("220"), UnrealizedPnlJpy: text("-20")}}},
		{ID: 2, Source: "sbi", FetchedAt: t1, ImportedAt: t1.Add(time.Hour), ValuationJpy: text("110.20"), Holdings: []model.FinancialHolding{{ID: 2, SnapshotID: 2, Name: "ダミー商品A", Section: text("nisa_funds"), CompositeFigi: text("TESTFIGI0001"), Quantity: text("2"), ValuationJpy: text("110.20"), UnrealizedPnlJpy: text("9.10")}}},
		{ID: 2, Source: "nrkn", FetchedAt: t1, ImportedAt: t1.Add(time.Hour), ValuationJpy: text("210"), CostJpy: text("220"), UnrealizedPnlJpy: text("-10"), Holdings: []model.FinancialHolding{{ID: 2, SnapshotID: 2, Name: "ダミー商品B", ProductCode: text("TEST001"), CompositeFigi: text("TESTFIGI0002"), ReferenceDate: text("2000-01-03"), Quantity: text("3"), ValuationJpy: text("210"), CostJpy: text("220"), UnrealizedPnlJpy: text("-10")}}},
	}}
}

type memoryFinancialRepository struct{ rows []model.FinancialSnapshot }

func (m memoryFinancialRepository) ListFinancialSnapshots(ctx context.Context, q model.FinancialSnapshotQuery) ([]model.FinancialSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows := []model.FinancialSnapshot{}
	for _, row := range m.rows {
		if row.Source != q.Source || q.From != nil && row.FetchedAt.Before(*q.From) || q.To != nil && !row.FetchedAt.Before(*q.To) {
			continue
		}
		if q.AfterFetchedAt != nil && (row.FetchedAt.After(*q.AfterFetchedAt) || row.FetchedAt.Equal(*q.AfterFetchedAt) && q.AfterID != -1 && row.ID >= q.AfterID) {
			continue
		}
		row.Holdings = nil
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if !rows[i].FetchedAt.Equal(rows[j].FetchedAt) {
			return rows[i].FetchedAt.After(rows[j].FetchedAt)
		}
		return rows[i].ID > rows[j].ID
	})
	if q.Limit > 0 && len(rows) > q.Limit {
		rows = rows[:q.Limit]
	}
	return rows, nil
}

func (m memoryFinancialRepository) GetFinancialSnapshot(ctx context.Context, source string, id int64) (model.FinancialSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return model.FinancialSnapshot{}, err
	}
	for _, row := range m.rows {
		if row.Source == source && row.ID == id {
			row.Holdings = nil
			return row, nil
		}
	}
	return model.FinancialSnapshot{}, model.ErrRecordNotFound
}

func (m memoryFinancialRepository) GetFinancialHoldings(ctx context.Context, source string, ids []int64) ([]model.FinancialHolding, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	wanted := map[int64]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	rows := []model.FinancialHolding{}
	for _, s := range m.rows {
		if s.Source == source && wanted[s.ID] {
			rows = append(rows, s.Holdings...)
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	return rows, nil
}

func (m memoryFinancialRepository) LatestFinancialSnapshot(ctx context.Context, source string, at time.Time, inclusive bool) (*model.FinancialSnapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var result *model.FinancialSnapshot
	for _, row := range m.rows {
		if row.Source != source || row.FetchedAt.After(at) || !inclusive && row.FetchedAt.Equal(at) {
			continue
		}
		if result == nil || row.FetchedAt.After(result.FetchedAt) || row.FetchedAt.Equal(result.FetchedAt) && row.ID > result.ID {
			copy := row
			copy.Holdings = nil
			result = &copy
		}
	}
	return result, nil
}
