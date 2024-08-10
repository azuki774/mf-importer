package mfapi

import (
	"context"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"mf-importer/internal/repository"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

var jst *time.Location

func init() {
	j, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	jst = j
}

type DBRepository interface {
	GetDetails(ctx context.Context, limit int) (details []model.Detail, err error)
}

type APIService struct {
	Logger *zap.Logger
	Repo   DBRepository
}

func NewAPIService(l *zap.Logger, db *repository.DBClient) (ap *APIService) {
	return &APIService{Logger: l, Repo: db}
}

func (a *APIService) GetDetails(ctx context.Context, limit int) (dets []openapi.Detail, err error) {
	ds, err := a.Repo.GetDetails(ctx, limit)
	if err != nil {
		a.Logger.Error("failed to get Details from DB", zap.Error(err))
		return nil, err
	}

	// DB model -> API model
	for _, d := range ds {
		det := openapi.Detail{
			Id:              int(d.ID),
			ImportDate:      d.MawRegistDate,
			ImportJudgeDate: d.MawCheckDate,
			Name:            d.Name,
			Price:           int(d.Price),
			RegistDate:      d.RegistDate,
			UseDate:         types.Date{Time: d.Date},
		}
		dets = append(dets, det)
	}

	return dets, nil
}
