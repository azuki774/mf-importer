package service

import (
	"context"
	"mf-importer/internal/model"
	"time"
)

type mockDBClient struct {
	err    error
	exists bool
	itr    int
}

type mockDetailCSVOperator struct {
	err error
}

func (m *mockDBClient) CheckAlreadyRegistDetail(ctx context.Context, detail model.Detail) (exists bool, err error) {
	if m.err != nil {
		return false, m.err
	}

	// 問い合わせがあるたびに、DBに存在するパターンとないパターンを交互に返す
	m.itr += 1
	if (m.itr % 2) == 0 {
		m.exists = false
	} else {
		m.exists = true
	}
	return m.exists, nil
}

func (m *mockDBClient) RegistDetail(ctx context.Context, detail model.Detail) (err error) {
	return m.err
}

func (m *mockDBClient) RegistDetailHistory(ctx context.Context, jobname string, parsedNum int, insertNum int, srcFile string) (err error) {
	return m.err
}

func (m *mockDetailCSVOperator) LoadCfCSV(ctx context.Context, path string) (details []model.Detail, err error) {
	if m.err != nil {
		return []model.Detail{}, m.err
	}

	return []model.Detail{
		{
			YYYYMMID:   1, // YYYYMM_id は逆順につける（日付の古い順）
			Date:       time.Date(2024, 7, 1, 0, 0, 0, 0, time.Local),
			Name:       "シベリア商店",
			Price:      810,
			FinIns:     "三井住友カード",
			LCategory:  "大分類",
			MCategory:  "中分類",
			RegistDate: time.Date(2024, 7, 5, 0, 0, 0, 0, time.Local),
			RawDate:    "07/01(月)",
			RawPrice:   "-810",
		},
		{
			YYYYMMID:   2, // YYYYMM_id は逆順につける（日付の古い順）
			Date:       time.Date(2024, 7, 2, 0, 0, 0, 0, time.Local),
			Name:       "シベリア商店",
			Price:      860,
			FinIns:     "三井住友カード",
			LCategory:  "大分類",
			MCategory:  "中分類",
			RegistDate: time.Date(2024, 7, 5, 0, 0, 0, 0, time.Local),
			RawDate:    "07/02(火)",
			RawPrice:   "-860",
		},
		{
			YYYYMMID:   3, // YYYYMM_id は逆順につける（日付の古い順）
			Date:       time.Date(2024, 7, 4, 0, 0, 0, 0, time.Local),
			Name:       "シベリア商店",
			Price:      310,
			FinIns:     "三井住友カード",
			LCategory:  "大分類",
			MCategory:  "中分類",
			RegistDate: time.Date(2024, 7, 5, 0, 0, 0, 0, time.Local),
			RawDate:    "07/03(水)",
			RawPrice:   "-310",
		},
	}, nil
}

func (m *mockDetailCSVOperator) GetTargetFiles(ctx context.Context, inputDir string) (targetCSVs []string, err error) {
	if m.err != nil {
		return []string{}, m.err
	}

	return []string{"cf.csv", "cf_lastmonth.csv"}, nil
}
