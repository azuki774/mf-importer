package mawinter

import (
	"context"
	"mf-importer/internal/model"
	"time"
)

var NULLtimeTime = time.Time{}

var forTestExtractRule model.ExtractRule = model.ExtractRule{
	FromName:        map[string]int{"かんぜんいっち": 100},
	FromNameIn:      map[string]int{"ぶぶんめい": 200},
	FromMCategory:   map[string]int{"かんぜんいっち": 300},
	FromMCategoryIn: map[string]int{"ぶぶんめい": 400},
}

type mockMawinterClient struct{}
type mockDBClient struct {
}

func (m *mockMawinterClient) Regist(ctx context.Context, c model.Detail, catID int) (err error) {
	return nil
}

func (m *mockDBClient) GetCFDetails(ctx context.Context) (cfRecords []model.Detail, err error) {
	return []model.Detail{
		{
			ID:           11,
			YYYYMMID:     1,
			Date:         "2023-01-01",
			RawDate:      "01/01（火）",
			Name:         "ふぃーるど１",
			Price:        1234,
			LCategory:    "大分類",
			MCategory:    "中分類",
			MawCheckDate: NULLtimeTime,
		},
		{
			ID:           12,
			YYYYMMID:     2,
			Date:         "2023-01-02",
			RawDate:      "01/02（水）",
			Name:         "ふぃーるど５",
			Price:        1234,
			LCategory:    "大分類",
			MCategory:    "中分類",
			MawCheckDate: NULLtimeTime,
		},
	}, nil
}

func (m *mockDBClient) CheckCFDetail(ctx context.Context, cfDetail model.Detail) (err error) {
	return nil
}

func (m *mockDBClient) RegistedCFDetail(ctx context.Context, cfDetail model.Detail) (err error) {
	return nil
}

func (m *mockDBClient) GetExtractRules(ctx context.Context) (er []model.ExtractRuleDB, err error) {
	return []model.ExtractRuleDB{}, nil // TODO
}
