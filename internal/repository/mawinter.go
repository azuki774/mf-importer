package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"net/http"

	"go.uber.org/zap"
)

type MawinterClient struct {
	Logger  *zap.Logger
	PostURL string // mawinter-server Post API のエンドポイント
	GetURL  string // mawinter-server Get API のエンドポイント
}

func NewMawinterClient(posturl string) *MawinterClient {
	l := logger.NewLogger()
	return &MawinterClient{Logger: l, PostURL: posturl}
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
		m.PostURL,
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
