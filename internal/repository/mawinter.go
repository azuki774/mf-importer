package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type MawinterClient struct {
	Logger *zap.Logger
	APIURL string // mawinter-server API のエンドポイント
}

var mawTracer = otel.Tracer("mf-importer-maw/mawinter-client")

func NewMawinterClient(apiurl string) *MawinterClient {
	l := logger.NewLogger()
	return &MawinterClient{Logger: l, APIURL: apiurl}
}

func (m *MawinterClient) Regist(ctx context.Context, c model.Detail, catID int) (err error) {
	ctx, span := mawTracer.Start(ctx, "mawinter.api.regist", trace.WithAttributes(
		attribute.String("mawinter.api.url", m.APIURL),
		attribute.String("detail.name", c.Name),
		attribute.Int64("detail.price", c.Price),
		attribute.Int("detail.category_id", catID),
	))
	defer span.End()

	rec, err := model.NewCreateRecord(c, catID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	recB, err := json.Marshal(rec)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	req, err := http.NewRequest(
		"POST",
		m.APIURL,
		bytes.NewBuffer(recB),
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	req = req.WithContext(ctx)

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	res, err := client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	defer res.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", res.StatusCode))
	if res.StatusCode != 201 {
		err := fmt.Errorf("unexpected code: %d", res.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	m.Logger.Info("post records", zap.String("date", rec.Date), zap.Int64("category_id", rec.CategoryID), zap.Int64("price", rec.Price), zap.String("memo", rec.Memo))
	return nil
}
