package mawinter

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestMawinter_Regist(t *testing.T) {
	type fields struct {
		Logger      *zap.Logger
		DBClient    MongoDBClient
		MawClient   MawinterClient
		CSVFileOp   CSVFileOperator
		ExtractRule model.ExtractRule
		ProcessDate time.Time
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Logger:      logger.NewLogger(),
				DBClient:    &mockDBClient{},
				MawClient:   &mockMawinterClient{},
				CSVFileOp:   &mockCSVFileOperator{},
				ProcessDate: time.Now(),
			},
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mawinter{
				Logger:      tt.fields.Logger,
				DBClient:    tt.fields.DBClient,
				MawClient:   tt.fields.MawClient,
				CSVFileOp:   tt.fields.CSVFileOp,
				ExtractRule: tt.fields.ExtractRule,
				ProcessDate: tt.fields.ProcessDate,
			}
			if err := m.Regist(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Mawinter.Regist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
