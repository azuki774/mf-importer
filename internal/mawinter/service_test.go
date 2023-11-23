package mawinter

import (
	"context"
	"errors"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestMawinter_Regist(t *testing.T) {
	type fields struct {
		Logger      *zap.Logger
		DBClient    DBClient
		MawClient   MawinterClient
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
				ProcessDate: time.Now(),
				Dryrun:      true,
			},
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
		{
			name: "error (GET mawinter)",
			fields: fields{
				Logger:      logger.NewLogger(),
				DBClient:    &mockDBClient{},
				MawClient:   &mockMawinterClient{GetMawinterWebError: errors.New("error")},
				ProcessDate: time.Now(),
			},
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mawinter{
				Logger:      tt.fields.Logger,
				DBClient:    tt.fields.DBClient,
				MawClient:   tt.fields.MawClient,
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
		DBClient    DBClient
		MawClient   MawinterClient
		ExtractRule model.ExtractRule
		ProcessDate time.Time
		Dryrun      bool
	}
	type args struct {
		c model.Detail
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
				c: model.Detail{
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
				c: model.Detail{
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
				c: model.Detail{
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
				c: model.Detail{
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
				c: model.Detail{
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

func Test_judgeAlreadyRegisted(t *testing.T) {
	type args struct {
		dr     model.Detail
		alrecs []model.GetRecord
	}
	tests := []struct {
		name          string
		args          args
		wantDuplicate bool
	}{
		{
			name: "false case (Date)",
			args: args{
				dr: model.Detail{
					Date:  time.Date(2010, 1, 26, 0, 0, 0, 0, time.Local),
					Price: 1234,
				},
				alrecs: []model.GetRecord{
					{
						ID:           1,
						CategoryID:   100,
						CategoryName: "cat1",
						Datetime:     time.Date(2010, 1, 23, 1, 2, 3, 0, time.Local),
						From:         fromMawinterWebText,
						Price:        1234,
					},
					{
						ID:           5,
						CategoryID:   500,
						CategoryName: "cat5",
						Datetime:     time.Date(2010, 1, 24, 1, 2, 3, 0, time.Local),
						From:         fromMawinterWebText,
						Price:        5678,
					},
				},
			},
			wantDuplicate: false,
		},
		{
			name: "false case (Price)",
			args: args{
				dr: model.Detail{
					Date:  time.Date(2010, 1, 23, 1, 2, 3, 0, time.Local),
					Price: 12345,
				},
				alrecs: []model.GetRecord{
					{
						ID:           1,
						CategoryID:   100,
						CategoryName: "cat1",
						Datetime:     time.Date(2010, 1, 23, 1, 2, 3, 0, time.Local),
						From:         fromMawinterWebText,
						Price:        1234,
					},
					{
						ID:           5,
						CategoryID:   500,
						CategoryName: "cat5",
						Datetime:     time.Date(2010, 1, 24, 1, 2, 3, 0, time.Local),
						From:         fromMawinterWebText,
						Price:        5678,
					},
				},
			},
			wantDuplicate: false,
		},
		{
			name: "true case",
			args: args{
				dr: model.Detail{
					Date:  time.Date(2010, 1, 24, 0, 0, 0, 0, time.Local),
					Price: 5678,
				},
				alrecs: []model.GetRecord{
					{
						ID:           1,
						CategoryID:   100,
						CategoryName: "cat1",
						Datetime:     time.Date(2010, 1, 23, 1, 2, 3, 0, time.Local),
						From:         fromMawinterWebText,
						Price:        1234,
					},
					{
						ID:           5,
						CategoryID:   500,
						CategoryName: "cat5",
						Datetime:     time.Date(2010, 1, 24, 1, 2, 3, 0, time.Local),
						From:         fromMawinterWebText,
						Price:        5678,
					},
				},
			},
			wantDuplicate: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotDuplicate := judgeAlreadyRegisted(tt.args.dr, tt.args.alrecs); gotDuplicate != tt.wantDuplicate {
				t.Errorf("judgeAlreadyRegisted() = %v, want %v", gotDuplicate, tt.wantDuplicate)
			}
		})
	}
}
