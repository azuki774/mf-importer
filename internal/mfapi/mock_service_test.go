package mfapi

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"

	"github.com/oapi-codegen/runtime/types"
)

const cfFixture = `"計算対象","日付","内容","金額（円）","保有金融機関","大項目","中項目","メモ","振替","ID"
"1","2024/08/22","APPLE COM BILL","-1080","三井住友カード","通信費","その他通信費","","0","x"
"1","2024/08/01","ローソン","-534","三井住友カード","食費","食料品","","0","y"
`

const unparsableFixture = `計算対象,日付,内容,金額（円）,保有金融機関,大項目,中項目,メモ,振替,削除
,06/10(水),マクドナルド,-600,三井住友カード,食費,外食,,,
`

func newTestService(t *testing.T, files map[string]string) *MockAPIService {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write fixture: %v", err)
		}
	}
	svc, err := NewMockAPIService(logger.NewLogger(), dir)
	if err != nil {
		t.Fatalf("failed to create mock service: %v", err)
	}
	return svc
}

func TestMockService_LoadsParsableCSVOnly(t *testing.T) {
	svc := newTestService(t, map[string]string{
		"cf.csv":           cfFixture,
		"cf_lastmonth.csv": unparsableFixture,
	})

	count, err := svc.GetDetailsCount(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.Count != 2 {
		t.Fatalf("expected 2 details (unparsable csv skipped), got %d", count.Count)
	}
}

func TestMockService_DetailsOrderAndIDs(t *testing.T) {
	svc := newTestService(t, map[string]string{"cf.csv": cfFixture})

	dets, err := svc.GetDetails(context.Background(), 20, 0, "useDate", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dets) != 2 {
		t.Fatalf("expected 2 details, got %d", len(dets))
	}

	if dets[0].Name != "APPLE COM BILL" || dets[0].Id != 2 {
		t.Errorf("newest first expected: %+v", dets[0])
	}
	if dets[1].Name != "ローソン" || dets[1].Id != 1 {
		t.Errorf("oldest last expected: %+v", dets[1])
	}
	wantUseDate := time.Date(2024, 8, 22, 0, 0, 0, 0, time.Local)
	if !dets[0].UseDate.Time.Equal(wantUseDate) || dets[0].Price != 1080 {
		t.Errorf("unexpected fields: %+v", dets[0])
	}
}

func TestMockService_GetDetailsPaging(t *testing.T) {
	svc := newTestService(t, map[string]string{"cf.csv": cfFixture})
	ctx := context.Background()

	page, err := svc.GetDetails(ctx, 1, 1, "useDate", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page) != 1 || page[0].Name != "ローソン" {
		t.Errorf("unexpected page: %+v", page)
	}

	page, err = svc.GetDetails(ctx, 20, 100, "useDate", "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page) != 0 {
		t.Errorf("expected empty page, got %+v", page)
	}
}

func TestMockService_ImportJudgeEveryThird(t *testing.T) {
	fixture := cfFixture + `"1","2024/07/15","セブン","-300","nanaco","食費","外食","","0","z"`
	svc := newTestService(t, map[string]string{"cf.csv": fixture})
	ctx := context.Background()

	dets, _ := svc.GetDetails(ctx, 20, 0, "useDate", "desc")
	var newest openapi.Detail
	for _, d := range dets {
		if d.Name == "APPLE COM BILL" {
			newest = d
		}
	}
	if newest.ImportJudgeDate == nil || newest.ImportDate == nil {
		t.Errorf("every third detail should have import dates: %+v", newest)
	}
}

func TestMockService_ResetImportDetails(t *testing.T) {
	fixture := cfFixture + `"1","2024/07/15","セブン","-300","nanaco","食費","外食","","0","z"`
	svc := newTestService(t, map[string]string{"cf.csv": fixture})
	ctx := context.Background()

	dets, _ := svc.GetDetails(ctx, 20, 0, "useDate", "desc")
	var targetID int
	for _, d := range dets {
		if d.Name == "APPLE COM BILL" {
			targetID = d.Id
			if d.ImportJudgeDate == nil || d.ImportDate == nil {
				t.Fatalf("precondition failed, no import dates: %+v", d)
			}
		}
	}

	if err := svc.ResetImportDetails(ctx, targetID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dets, _ = svc.GetDetails(ctx, 20, 0, "useDate", "desc")
	for _, d := range dets {
		if d.Id == targetID && (d.ImportJudgeDate != nil || d.ImportDate != nil) {
			t.Errorf("import dates should be cleared: %+v", d)
		}
	}
}

func TestMockService_RulesCRUD(t *testing.T) {
	svc := newTestService(t, map[string]string{})
	ctx := context.Background()

	rules, err := svc.GetRules(ctx, "id", "asc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 initial rules, got %d", len(rules))
	}

	rule, err := svc.GetRule(ctx, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.Value != "WASABI" || rule.ExactMatch != 0 {
		t.Errorf("unexpected rule: %+v", rule)
	}

	added, err := svc.AddRule(ctx, openapi.RuleRequest{FieldName: "name", Value: "スーパー", ExactMatch: 0, CategoryId: 101})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if added.Id != 3 {
		t.Errorf("expected id 3, got %d", added.Id)
	}

	if err := svc.DeleteRule(ctx, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := svc.DeleteRule(ctx, 999); err != model.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}

	rules, _ = svc.GetRules(ctx, "id", "asc")
	if len(rules) != 2 {
		t.Errorf("expected 2 rules after delete, got %d", len(rules))
	}
}

func seedDetailSortFixture() *MockAPIService {
	tz := time.Local
	d := func(s string) time.Time {
		t, _ := time.ParseInLocation("2006-01-02", s, tz)
		return t
	}
	tp := func(t time.Time) *time.Time { return &t }
	return &MockAPIService{
		Logger: logger.NewLogger(),
		details: []openapi.Detail{
			{Id: 1, Name: "B店", Price: 300, UseDate: typesDate("2024-08-01"), RegistDate: d("2024-08-02")},
			{Id: 2, Name: "A店", Price: 500, UseDate: typesDate("2024-08-01"), RegistDate: d("2024-08-01"), ImportJudgeDate: tp(d("2024-08-04")), ImportDate: tp(d("2024-08-05"))},
			{Id: 3, Name: "C店", Price: 100, UseDate: typesDate("2024-07-30"), RegistDate: d("2024-08-03"), ImportDate: tp(d("2024-08-05"))},
			{Id: 4, Name: "A店", Price: 400, UseDate: typesDate("2024-08-03"), RegistDate: d("2024-08-01"), ImportJudgeDate: tp(d("2024-08-06"))},
			{Id: 5, Name: "D店", Price: 500, UseDate: typesDate("2024-07-28"), RegistDate: d("2024-08-02")},
		},
	}
}

func typesDate(s string) types.Date {
	t, _ := time.ParseInLocation("2006-01-02", s, time.Local)
	return types.Date{Time: t}
}

func detailIDs(dets []openapi.Detail) []int {
	ids := make([]int, 0, len(dets))
	for _, d := range dets {
		ids = append(ids, d.Id)
	}
	return ids
}

func assertIDsEqual(t *testing.T, got []int, want []int) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("unexpected order: got %v, want %v", got, want)
	}
}

func TestMockService_DetailsSortKeys(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		order    string
		wantAsc  []int
		wantDesc []int
	}{
		{name: "useDate tie keeps ID asc", sort: "useDate", order: "asc", wantAsc: []int{5, 3, 1, 2, 4}, wantDesc: []int{4, 2, 1, 3, 5}},
		{name: "name asc", sort: "name", order: "asc", wantAsc: []int{2, 4, 1, 3, 5}, wantDesc: []int{5, 3, 1, 4, 2}},
		{name: "price ties keep ID direction", sort: "price", order: "asc", wantAsc: []int{3, 1, 4, 2, 5}, wantDesc: []int{5, 2, 4, 1, 3}},
		{name: "registDate", sort: "registDate", order: "asc", wantAsc: []int{2, 4, 1, 5, 3}, wantDesc: []int{3, 5, 1, 4, 2}},
		{name: "importJudgeDate nils first asc / last desc", sort: "importJudgeDate", order: "asc", wantAsc: []int{1, 3, 5, 2, 4}, wantDesc: []int{4, 2, 5, 3, 1}},
		{name: "importDate nils first asc / last desc", sort: "importDate", order: "asc", wantAsc: []int{1, 4, 5, 2, 3}, wantDesc: []int{3, 2, 5, 4, 1}},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name+"/asc", func(t *testing.T) {
			svc := seedDetailSortFixture()
			dets, err := svc.GetDetails(ctx, 20, 0, tt.sort, "asc")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertIDsEqual(t, detailIDs(dets), tt.wantAsc)
		})
		t.Run(tt.name+"/desc", func(t *testing.T) {
			svc := seedDetailSortFixture()
			dets, err := svc.GetDetails(ctx, 20, 0, tt.sort, "desc")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertIDsEqual(t, detailIDs(dets), tt.wantDesc)
		})
	}

	t.Run("unknown sort falls back to useDate desc", func(t *testing.T) {
		svc := seedDetailSortFixture()
		dets, err := svc.GetDetails(ctx, 20, 0, "evil;drop", "desc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIDsEqual(t, detailIDs(dets), []int{4, 2, 1, 3, 5})
	})
	t.Run("unknown order falls back to asc", func(t *testing.T) {
		svc := seedDetailSortFixture()
		dets, err := svc.GetDetails(ctx, 20, 0, "name", "sideways")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIDsEqual(t, detailIDs(dets), []int{2, 4, 1, 3, 5})
	})
}

func seedRuleSortFixture() *MockAPIService {
	return &MockAPIService{
		Logger: logger.NewLogger(),
		rules: []model.ExtractRuleDB{
			{ID: 1, FieldName: "b", Value: "x", ExactMatch: 1, CategoryID: 200},
			{ID: 2, FieldName: "a", Value: "x", ExactMatch: 0, CategoryID: 210},
			{ID: 3, FieldName: "c", Value: "y", ExactMatch: 0, CategoryID: 205},
		},
		nextRuleID: 4,
	}
}

func TestMockService_RulesSortKeys(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		order    string
		wantAsc  []int
		wantDesc []int
	}{
		{name: "id", sort: "id", order: "asc", wantAsc: []int{1, 2, 3}, wantDesc: []int{3, 2, 1}},
		{name: "fieldName", sort: "fieldName", order: "asc", wantAsc: []int{2, 1, 3}, wantDesc: []int{3, 1, 2}},
		{name: "value tie keeps ID direction", sort: "value", order: "asc", wantAsc: []int{1, 2, 3}, wantDesc: []int{3, 2, 1}},
		{name: "exactMatch", sort: "exactMatch", order: "asc", wantAsc: []int{2, 3, 1}, wantDesc: []int{1, 3, 2}},
		{name: "categoryId", sort: "categoryId", order: "asc", wantAsc: []int{1, 3, 2}, wantDesc: []int{2, 3, 1}},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name+"/asc", func(t *testing.T) {
			svc := seedRuleSortFixture()
			rules, err := svc.GetRules(ctx, tt.sort, "asc")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertIDsEqual(t, ruleIDs(rules), tt.wantAsc)
		})
		t.Run(tt.name+"/desc", func(t *testing.T) {
			svc := seedRuleSortFixture()
			rules, err := svc.GetRules(ctx, tt.sort, "desc")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertIDsEqual(t, ruleIDs(rules), tt.wantDesc)
		})
	}

	t.Run("unknown sort falls back to id asc", func(t *testing.T) {
		svc := seedRuleSortFixture()
		rules, err := svc.GetRules(ctx, "hack", "weird")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIDsEqual(t, ruleIDs(rules), []int{1, 2, 3})
	})
}

func ruleIDs(rules []openapi.Rule) []int {
	ids := make([]int, 0, len(rules))
	for _, r := range rules {
		ids = append(ids, r.Id)
	}
	return ids
}
