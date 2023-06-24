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
		Dryrun      bool
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
		{
			name: "ok (dryrun)",
			fields: fields{
				Logger:      logger.NewLogger(),
				DBClient:    &mockDBClient{},
				MawClient:   &mockMawinterClient{},
				CSVFileOp:   &mockCSVFileOperator{},
				ProcessDate: time.Now(),
				Dryrun:      true,
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
				Dryrun:      tt.fields.Dryrun,
			}
			if err := m.Regist(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Mawinter.Regist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMawinter_getCategoryIDwithExtractCond(t *testing.T) {
	type fields struct {
		Logger      *zap.Logger
		DBClient    MongoDBClient
		MawClient   MawinterClient
		CSVFileOp   CSVFileOperator
		ExtractRule model.ExtractRule
		ProcessDate time.Time
		Dryrun      bool
	}
	type args struct {
		c model.CFRecord
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantCatID int
		wantOk    bool
	}{
		{
			name: "FromName",
			fields: fields{
				Logger:      logger.NewLogger(),
				ExtractRule: forTestExtractRule,
			},
			args: args{
				c: model.CFRecord{
					Name:      "かんぜんいっち",
					MCategory: "",
				},
			},
			wantCatID: 100,
			wantOk:    true,
		},
		{
			name: "FromNameIn",
			fields: fields{
				Logger:      logger.NewLogger(),
				ExtractRule: forTestExtractRule,
			},
			args: args{
				c: model.CFRecord{
					Name:      "ねーむぶぶんめい",
					MCategory: "",
				},
			},
			wantCatID: 200,
			wantOk:    true,
		},
		{
			name: "FromMCategory",
			fields: fields{
				Logger:      logger.NewLogger(),
				ExtractRule: forTestExtractRule,
			},
			args: args{
				c: model.CFRecord{
					Name:      "",
					MCategory: "かんぜんいっち",
				},
			},
			wantCatID: 300,
			wantOk:    true,
		},
		{
			name: "FromMCategoryIn",
			fields: fields{
				Logger:      logger.NewLogger(),
				ExtractRule: forTestExtractRule,
			},
			args: args{
				c: model.CFRecord{
					Name:      "",
					MCategory: "AAAぶぶんめいかてごりー",
				},
			},
			wantCatID: 400,
			wantOk:    true,
		},
		{
			name: "NotFound",
			fields: fields{
				Logger:      logger.NewLogger(),
				ExtractRule: forTestExtractRule,
			},
			args: args{
				c: model.CFRecord{
					Name:      "のっとふぁうんど",
					MCategory: "のっとふぁうんど",
				},
			},
			wantCatID: 0,
			wantOk:    false,
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
				Dryrun:      tt.fields.Dryrun,
			}
			gotCatID, gotOk := m.getCategoryIDwithExtractCond(tt.args.c)
			if gotCatID != tt.wantCatID {
				t.Errorf("Mawinter.getCategoryIDwithExtractCond() gotCatID = %v, want %v", gotCatID, tt.wantCatID)
			}
			if gotOk != tt.wantOk {
				t.Errorf("Mawinter.getCategoryIDwithExtractCond() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
