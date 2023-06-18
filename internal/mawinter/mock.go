package mawinter

import (
	"context"
	"mf-importer/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var forTestExtractRule model.ExtractRule = model.ExtractRule{
	FromName:        map[string]int{"かんぜんいっち": 100},
	FromNameIn:      map[string]int{"ぶぶんめい": 200},
	FromMCategory:   map[string]int{"かんぜんいっち": 300},
	FromMCategoryIn: map[string]int{"ぶぶんめい": 400},
}

type mockMawinterClient struct{}
type mockDBClient struct {
}

func (m *mockMawinterClient) Regist(ctx context.Context, c model.CFRecord) (err error) {
	return nil
}

type mockCSVFileOperator struct {
}

func (m *mockCSVFileOperator) LoadExtractCSV(path string) (es []model.ExtractRuleCSV, err error) {
	es = []model.ExtractRuleCSV{
		{
			FieldName:        "name",
			Name:             "ふぃーるど１",
			ExtractCondition: true,
			CategoryID:       100,
		},
		{
			FieldName:        "name",
			Name:             "ふぃーるど２",
			ExtractCondition: false,
			CategoryID:       200,
		},
		{
			FieldName:        "m_category",
			Name:             "ふぃーるど３",
			ExtractCondition: true,
			CategoryID:       300,
		},
		{
			FieldName:        "m_category",
			Name:             "ふぃーるど４",
			ExtractCondition: false,
			CategoryID:       400,
		},
	}
	return es, nil
}

func (m *mockDBClient) GetCFRecords(ctx context.Context) (cfRecords []model.CFRecord, err error) {
	return []model.CFRecord{
		{
			ID:        primitive.ObjectID([12]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}),
			RegistID:  1,
			YYYYMMDD:  "20230101",
			Date:      "01/01（火）",
			Name:      "ふぃーるど１",
			Price:     "-1,234",
			LCategory: "大分類",
			MCategory: "中分類",
		},
		{
			ID:        primitive.ObjectID([12]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}),
			RegistID:  2,
			YYYYMMDD:  "20230102",
			Date:      "01/02（水）",
			Name:      "ふぃーるど５",
			Price:     "-1,234",
			LCategory: "大分類",
			MCategory: "中分類",
		},
	}, nil
}
