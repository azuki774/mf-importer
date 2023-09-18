package repository

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"testing"
	"time"

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
		ctx   context.Context
		c     model.Detail
		catID int
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
				c: model.Detail{
					Date:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local),
					Price: 1451,
					Name:  "なまえA",
				},
				catID: 100,
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
				c: model.Detail{
					Date:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local),
					Price: 1451,
					Name:  "なまえA",
				},
				catID: 100,
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
			if err := m.Regist(tt.args.ctx, tt.args.c, tt.args.catID); (err != nil) != tt.wantErr {
				t.Errorf("MawinterClient.Regist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
