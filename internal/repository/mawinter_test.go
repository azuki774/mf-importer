package repository

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"testing"

	"github.com/jarcoal/httpmock"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("POST", "http://localhost:8080/v2/record",
		httpmock.NewStringResponder(201, "created"),
	)
	httpmock.RegisterResponder("POST", "http://localhost:8081/v2/record",
		httpmock.NewStringResponder(500, "error"),
	)
	m.Run()
}

func TestMawinterClient_Regist(t *testing.T) {
	type fields struct {
		Logger  *zap.Logger
		PostURL string
	}
	type args struct {
		ctx context.Context
		c   model.CFRecord
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
				Logger:  logger.NewLogger(),
				PostURL: "http://localhost:8080/v2/record",
			},
			args: args{
				ctx: context.Background(),
				c: model.CFRecord{
					YYYYMMDD:   "20230521",
					Price:      "-1,451",
					CategoryID: 100,
					Name:       "なまえA",
				},
			},
			wantErr: false,
		},
		{
			name: "unexpected error",
			fields: fields{
				Logger:  logger.NewLogger(),
				PostURL: "http://localhost:8081/v2/record",
			},
			args: args{
				ctx: context.Background(),
				c: model.CFRecord{
					YYYYMMDD:   "20230521",
					Price:      "-1,451",
					CategoryID: 100,
					Name:       "なまえA",
				},
			},
			wantErr: true,
		},
		{
			name: "error data",
			fields: fields{
				Logger:  logger.NewLogger(),
				PostURL: "http://localhost:8080/v2/record",
			},
			args: args{
				ctx: context.Background(),
				c: model.CFRecord{
					YYYYMMDD:   "20230521",
					Price:      "eeeeeeee-1,451",
					CategoryID: 100,
					Name:       "なまえA",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MawinterClient{
				Logger:  tt.fields.Logger,
				PostURL: tt.fields.PostURL,
			}
			if err := m.Regist(tt.args.ctx, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("MawinterClient.Regist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
