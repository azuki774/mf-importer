package mfapi

import (
	"cmp"
	"context"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"mf-importer/internal/repository"
	"sort"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

// MockAPIService は server.APIService のインメモリ実装（DB 不要のローカル確認用）
type MockAPIService struct {
	Logger     *zap.Logger
	details    []openapi.Detail
	rules      []model.ExtractRuleDB
	nextRuleID int64
}

func NewMockAPIService(l *zap.Logger, inputDir string) (*MockAPIService, error) {
	details, err := buildMockDetails(l, inputDir)
	if err != nil {
		return nil, err
	}
	rules := initialMockRules()
	return &MockAPIService{
		Logger:     l,
		details:    details,
		rules:      rules,
		nextRuleID: int64(len(rules)) + 1,
	}, nil
}

// buildMockDetails は inputDir 配下の CSV を読み、DB 取り込み後と同じ形の Detail 一覧を作る
func buildMockDetails(l *zap.Logger, inputDir string) ([]openapi.Detail, error) {
	op := &repository.DetailCSVOperator{Logger: l, Encoding: "utf8"}
	ctx := context.Background()

	files, err := op.GetTargetFiles(ctx, inputDir)
	if err != nil {
		return nil, err
	}

	var parsed []model.Detail
	for _, f := range files {
		ds, err := op.LoadCfCSV(ctx, f)
		if err != nil {
			l.Warn("mock service skips unparsable csv", zap.String("file", f), zap.Error(err))
			continue
		}
		parsed = append(parsed, ds...)
	}

	sort.SliceStable(parsed, func(i, j int) bool { return parsed[i].Date.Before(parsed[j].Date) })

	details := make([]openapi.Detail, 0, len(parsed))
	for i, d := range parsed {
		det := openapi.Detail{
			Id:         i + 1,
			Name:       d.Name,
			Price:      int(d.Price),
			RegistDate: d.RegistDate,
			UseDate:    types.Date{Time: d.Date},
		}
		if (i+1)%3 == 0 {
			judgeDate := d.Date.AddDate(0, 0, 3)
			registDate := d.Date.AddDate(0, 0, 4)
			det.ImportJudgeDate = &judgeDate
			det.ImportDate = &registDate
		}
		details = append(details, det)
	}
	return details, nil
}

// initialMockRules は deployment/extract_rule.csv と同じ内容の固定ルール
func initialMockRules() []model.ExtractRuleDB {
	return []model.ExtractRuleDB{
		{ID: 1, FieldName: "m_category", Value: "携帯電話", ExactMatch: 1, CategoryID: 231},
		{ID: 2, FieldName: "name", Value: "WASABI", ExactMatch: 0, CategoryID: 230},
	}
}

// ソートキーは internal/repository/sort.go のホワイトリストと一致させること
var mockDetailSortKeys = map[string]struct{}{
	"useDate": {}, "name": {}, "price": {},
	"registDate": {}, "importJudgeDate": {}, "importDate": {},
}

var mockRuleSortKeys = map[string]struct{}{
	"id": {}, "fieldName": {}, "value": {}, "exactMatch": {}, "categoryId": {},
}

// normalizeMockSort: 不明なキーはテーブル既定(キー/方向)へフォールバックする。
// 有効なキーの場合、方向は order == "desc" のとき降順、それ以外は昇順(SQL の句組み立てと同一)。
func normalizeMockSort(sort string, order string, known map[string]struct{}, defKey string, defDesc bool) (string, bool) {
	if _, ok := known[sort]; ok {
		return sort, order == "desc"
	}
	return defKey, defDesc
}

func compareTimePtr(a *time.Time, b *time.Time) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return -1 // NULL は昇順で先頭(MariaDB 相当)
	case b == nil:
		return 1
	default:
		return a.Compare(*b)
	}
}

func compareDetailBy(a openapi.Detail, b openapi.Detail, key string) int {
	switch key {
	case "name":
		return strings.Compare(a.Name, b.Name)
	case "price":
		return cmp.Compare(a.Price, b.Price)
	case "registDate":
		return a.RegistDate.Compare(b.RegistDate)
	case "importJudgeDate":
		return compareTimePtr(a.ImportJudgeDate, b.ImportJudgeDate)
	case "importDate":
		return compareTimePtr(a.ImportDate, b.ImportDate)
	default: // useDate
		return a.UseDate.Time.Compare(b.UseDate.Time)
	}
}

func compareRuleBy(a openapi.Rule, b openapi.Rule, key string) int {
	switch key {
	case "fieldName":
		return strings.Compare(a.FieldName, b.FieldName)
	case "value":
		return strings.Compare(a.Value, b.Value)
	case "exactMatch":
		return cmp.Compare(a.ExactMatch, b.ExactMatch)
	case "categoryId":
		return cmp.Compare(a.CategoryId, b.CategoryId)
	default: // id
		return cmp.Compare(a.Id, b.Id)
	}
}

func (m *MockAPIService) sortedDetails(key string, desc bool) []openapi.Detail {
	out := make([]openapi.Detail, len(m.details))
	copy(out, m.details)
	sort.SliceStable(out, func(i, j int) bool {
		c := compareDetailBy(out[i], out[j], key)
		if c == 0 {
			c = cmp.Compare(out[i].Id, out[j].Id)
		}
		if desc {
			return c > 0
		}
		return c < 0
	})
	return out
}

func (m *MockAPIService) sortedRules(key string, desc bool) []openapi.Rule {
	out := make([]openapi.Rule, 0, len(m.rules))
	for i := range m.rules {
		out = append(out, m.rules[i].ToExtractRule())
	}
	sort.SliceStable(out, func(i, j int) bool {
		c := compareRuleBy(out[i], out[j], key)
		if c == 0 {
			c = cmp.Compare(out[i].Id, out[j].Id)
		}
		if desc {
			return c > 0
		}
		return c < 0
	})
	return out
}

func (m *MockAPIService) GetDetails(ctx context.Context, limit int, offset int, sort string, order string) ([]openapi.Detail, error) {
	key, desc := normalizeMockSort(sort, order, mockDetailSortKeys, "useDate", true)
	sorted := m.sortedDetails(key, desc)

	start := offset
	if start < 0 {
		start = 0
	}

	out := []openapi.Detail{}
	for i := start; i < len(sorted); i++ {
		if limit > 0 && len(out) >= limit {
			break
		}
		out = append(out, sorted[i])
	}
	return out, nil
}

func (m *MockAPIService) GetDetailsCount(ctx context.Context) (openapi.DetailsCount, error) {
	return openapi.DetailsCount{Count: len(m.details)}, nil
}

func (m *MockAPIService) ResetImportDetails(ctx context.Context, id int) error {
	for i := range m.details {
		if m.details[i].Id == id {
			m.details[i].ImportJudgeDate = nil
			m.details[i].ImportDate = nil
		}
	}
	return nil
}

func (m *MockAPIService) GetRules(ctx context.Context, sort string, order string) ([]openapi.Rule, error) {
	key, desc := normalizeMockSort(sort, order, mockRuleSortKeys, "id", false)
	return m.sortedRules(key, desc), nil
}

func (m *MockAPIService) GetRule(ctx context.Context, id int) (openapi.Rule, error) {
	for i := range m.rules {
		if int(m.rules[i].ID) == id {
			return m.rules[i].ToExtractRule(), nil
		}
	}
	return openapi.Rule{}, model.ErrRecordNotFound
}

func (m *MockAPIService) AddRule(ctx context.Context, req openapi.RuleRequest) (openapi.Rule, error) {
	ruleDB := model.ExtractRuleDB{
		ID:         m.nextRuleID,
		FieldName:  req.FieldName,
		Value:      req.Value,
		ExactMatch: int64(req.ExactMatch),
		CategoryID: int64(req.CategoryId),
	}
	m.rules = append(m.rules, ruleDB)
	m.nextRuleID += 1
	return ruleDB.ToExtractRule(), nil
}

func (m *MockAPIService) DeleteRule(ctx context.Context, id int) error {
	for i := range m.rules {
		if int(m.rules[i].ID) == id {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return nil
		}
	}
	return model.ErrRecordNotFound
}
