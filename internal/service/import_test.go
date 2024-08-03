package service

import (
	"context"
	"errors"
	"mf-importer/internal/logger"
	"testing"

	"go.uber.org/zap"
)

var l *zap.Logger

func init() {
	l = logger.NewLogger()
}

func TestImporter_Start(t *testing.T) {
	type fields struct {
		Logger   *zap.Logger
		DBClient DBClient
		CSVOpe   DetailCSVOperator
		InputDir string
		JobName  string
		DryRun   bool
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
				Logger:   l,
				DBClient: &mockDBClient{},
				CSVOpe:   &mockDetailCSVOperator{},
				InputDir: "cf.csv",
				JobName:  "jobname",
				DryRun:   false,
			},
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
		{
			name: "DB error",
			fields: fields{
				Logger: l,
				DBClient: &mockDBClient{
					err: errors.New("error"),
				},
				CSVOpe:   &mockDetailCSVOperator{},
				InputDir: "cf.csv",
				JobName:  "jobname",
				DryRun:   false,
			},
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
		{
			name: "csv error",
			fields: fields{
				Logger:   l,
				DBClient: &mockDBClient{},
				CSVOpe: &mockDetailCSVOperator{
					err: errors.New("error"),
				},
				InputDir: "cf.csv",
				JobName:  "jobname",
				DryRun:   false,
			},
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Importer{
				Logger:   tt.fields.Logger,
				DBClient: tt.fields.DBClient,
				CSVOpe:   tt.fields.CSVOpe,
				InputDir: tt.fields.InputDir,
				JobName:  tt.fields.JobName,
				DryRun:   tt.fields.DryRun,
			}
			if err := i.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Importer.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
