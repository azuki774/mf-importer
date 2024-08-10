package mfapi

import (
	"context"
	"mf-importer/internal/model"
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
