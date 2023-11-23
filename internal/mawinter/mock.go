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

type mockMawinterClient struct {
	GetMawinterWebError error
}
type mockDBClient struct {
}

func (m *mockMawinterClient) Regist(ctx context.Context, c model.Detail, catID int) (err error) {
	return nil
}

func (m *mockMawinterClient) GetMawinterWeb(ctx context.Context, yyyymm string) (recs []model.GetRecord, err error) {
	if m.GetMawinterWebError != nil {
		return []model.GetRecord{}, m.GetMawinterWebError
	}

	recs = []model.GetRecord{
		{
			ID:           1,
			CategoryID:   100,
			CategoryName: "cat1",
			Datetime:     time.Date(2010, 1, 23, 1, 2, 3, 0, time.Local),
			From:         fromMawinterWebText,
			Price:        1234,
		},
		{
			ID:           5,
			CategoryID:   500,
			CategoryName: "cat5",
			Datetime:     time.Date(2010, 1, 24, 1, 2, 3, 0, time.Local),
			From:         fromMawinterWebText,
			Price:        5678,
		},
	}
	return recs, nil
}

func (m *mockDBClient) GetCFDetails(ctx context.Context) (cfRecords []model.Detail, err error) {
	return []model.Detail{
		{
			ID:           11,
			YYYYMMID:     1,
			Date:         time.Date(2023, 01, 01, 0, 0, 0, 0, time.Local),
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
			Date:         time.Date(2023, 01, 02, 0, 0, 0, 0, time.Local),
			RawDate:      "01/02（水）",
			Name:         "ふぃーるど５",
			Price:        1234,
			LCategory:    "大分類",
			MCategory:    "中分類",
			MawCheckDate: NULLtimeTime,
		},
		{
			ID:           100, // regist
			YYYYMMID:     2,
			Date:         time.Date(2023, 01, 02, 0, 0, 0, 0, time.Local),
			RawDate:      "01/02（水）",
			Name:         "かんぜんいっち",
			Price:        1234,
			LCategory:    "大分類",
			MCategory:    "中分類",
			MawCheckDate: NULLtimeTime,
		},
		{
			ID:           101, // already registed from mawinter-web
			YYYYMMID:     2,
			Date:         time.Date(2010, 01, 24, 0, 0, 0, 0, time.Local),
			RawDate:      "01/24（水）",
			Name:         "かんぜんいっち",
			Price:        5678,
			LCategory:    "大分類",
			MCategory:    "中分類",
			MawCheckDate: NULLtimeTime,
		},
	}, nil
}

func (m *mockDBClient) CheckCFDetail(ctx context.Context, cfDetail model.Detail, regist bool) (err error) {
	return nil
}

func (m *mockDBClient) GetExtractRules(ctx context.Context) (er []model.ExtractRuleDB, err error) {
	return []model.ExtractRuleDB{
		{
			ID:         1,
			FieldName:  "name",
			Value:      "かんぜんいっち",
			ExactMatch: 1,
			CategoryID: 100,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         2,
			FieldName:  "name",
			Value:      "ぶぶんいっち",
			ExactMatch: 0,
			CategoryID: 200,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         3,
			FieldName:  "m_category",
			Value:      "かんぜんいっち",
			ExactMatch: 1,
			CategoryID: 300,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         4,
			FieldName:  "m_category",
			Value:      "ぶぶんいっち",
			ExactMatch: 0,
			CategoryID: 400,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}, nil
}
