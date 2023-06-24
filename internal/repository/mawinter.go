package repository

import (
	"context"
	"mf-importer/internal/model"
)

type MawinterClient struct{}

func (m *MawinterClient) Regist(ctx context.Context, c model.CFRecord) (err error) {
	return nil
}
