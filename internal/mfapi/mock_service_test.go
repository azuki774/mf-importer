package mfapi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
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

	dets, err := svc.GetDetails(context.Background(), 20, 0)
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

	page, err := svc.GetDetails(ctx, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page) != 1 || page[0].Name != "ローソン" {
		t.Errorf("unexpected page: %+v", page)
	}

	page, err = svc.GetDetails(ctx, 20, 100)
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

	dets, _ := svc.GetDetails(ctx, 20, 0)
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

	dets, _ := svc.GetDetails(ctx, 20, 0)
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

	dets, _ = svc.GetDetails(ctx, 20, 0)
	for _, d := range dets {
		if d.Id == targetID && (d.ImportJudgeDate != nil || d.ImportDate != nil) {
			t.Errorf("import dates should be cleared: %+v", d)
		}
	}
}

func TestMockService_RulesCRUD(t *testing.T) {
	svc := newTestService(t, map[string]string{})
	ctx := context.Background()

	rules, err := svc.GetRules(ctx)
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

	rules, _ = svc.GetRules(ctx)
	if len(rules) != 2 {
		t.Errorf("expected 2 rules after delete, got %d", len(rules))
	}
}
