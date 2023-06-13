package mawinter

import (
	"context"
	"mf-importer/internal/model"
)

type mockMawinterClient struct{}

func (m *mockMawinterClient) Regist(ctx context.Context, r model.CreateRecord) (err error) {
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
