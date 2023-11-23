package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"net/http"

	"go.uber.org/zap"
)

type MawinterClient struct {
	Logger *zap.Logger
	APIURL string // mawinter-server API のエンドポイント
}

func NewMawinterClient(apiurl string) *MawinterClient {
	l := logger.NewLogger()
	return &MawinterClient{Logger: l, APIURL: apiurl}
}

func (m *MawinterClient) Regist(ctx context.Context, c model.Detail, catID int) (err error) {
	rec, err := model.NewCreateRecord(c, catID)
	if err != nil {
		return err
	}
	recB, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		m.APIURL,
		bytes.NewBuffer(recB),
	)
	if err != nil {
		return err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		return fmt.Errorf("unexpected code: %d", res.StatusCode)
	}

	m.Logger.Info("post records", zap.String("date", rec.Date), zap.Int64("category_id", rec.CategoryID), zap.Int64("price", rec.Price), zap.String("memo", rec.Memo))
	return nil
}

func (m *MawinterClient) GetMawinterWeb(ctx context.Context, yyyymm string) (recs []model.GetRecord, err error) {
	url := m.APIURL + "/" + yyyymm + "?from=mawinter-web"
	resp, err := http.Get(url)
	if err != nil {
		return []model.GetRecord{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []model.GetRecord{}, err
	}
	err = json.Unmarshal(body, &recs)
	if err != nil {
		return []model.GetRecord{}, err
	}

	return recs, nil
}
