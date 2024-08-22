package mfapi

import (
	"context"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"time"
)

type mockDBClient struct {
	err error
}

// limit は無視な mock
func (m *mockDBClient) GetDetails(ctx context.Context, limit int) (details []model.Detail, err error) {
	if m.err != nil {
		return nil, m.err
	}

	t11 := time.Date(2010, 1, 1, 14, 30, 0, 0, jst)

	details = []model.Detail{
		{
			ID:         2,
			YYYYMMID:   2,
			Date:       time.Date(2010, 1, 1, 0, 0, 0, 0, jst),
			Name:       "テスト明細Y",
			Price:      1234,
			RegistDate: time.Date(2010, 1, 1, 13, 30, 0, 0, jst),
			// MawCheckDate  *time.Time `json:"maw_check_date"`
			// MawRegistDate *time.Time `json:"maw_regist_date"`
		},
		{
			ID:            1,
			YYYYMMID:      1,
			Date:          time.Date(2010, 1, 1, 0, 0, 0, 0, jst),
			Name:          "テスト明細X",
			Price:         2345,
			RegistDate:    time.Date(2010, 1, 1, 13, 30, 0, 0, jst),
			MawCheckDate:  &t11,
			MawRegistDate: &t11,
		},
	}
	return details, nil
}

func (m *mockDBClient) GetExtractRules(ctx context.Context) (er []model.ExtractRuleDB, err error) {
	if m.err != nil {
		return nil, m.err
	}

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
			FieldName:  "m_category",
			Value:      "ぶぶんいっち",
			ExactMatch: 0,
			CategoryID: 400,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}, nil
}

func (m *mockDBClient) AddExtractRule(ctx context.Context, rule openapi.RuleRequest) (ruleDB model.ExtractRuleDB, err error) {
	if m.err != nil {
		return model.ExtractRuleDB{}, m.err
	}

	return model.ExtractRuleDB{
		ID:         100, // fix value
		FieldName:  rule.FieldName,
		Value:      rule.Value,
		ExactMatch: int64(rule.ExactMatch),
		CategoryID: int64(rule.CategoryId),
		CreatedAt:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}, nil
}
