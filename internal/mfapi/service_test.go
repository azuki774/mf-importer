package mfapi

import (
	"context"
	"errors"
	"mf-importer/internal/logger"
	"mf-importer/internal/openapi"
	"reflect"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

var l *zap.Logger
var t11 time.Time

func init() {
	l = logger.NewLogger()
	t11 = time.Date(2010, 1, 1, 14, 30, 0, 0, jst)
}
func TestAPIService_GetDetails(t *testing.T) {
	type fields struct {
		Logger *zap.Logger
		Repo   DBRepository
	}
	type args struct {
		ctx   context.Context
		limit int
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantDets []openapi.Detail
		wantErr  bool
	}{
		{
			name: "ok",
			fields: fields{
				Logger: l,
				Repo:   &mockDBClient{},
			},
			args: args{
				ctx:   context.Background(),
				limit: 5,
			},
			wantDets: []openapi.Detail{
				{
					Id:         2,
					Name:       "テスト明細Y",
					Price:      1234,
					RegistDate: time.Date(2010, 1, 1, 13, 30, 0, 0, jst),
					UseDate:    types.Date{Time: time.Date(2010, 1, 1, 0, 0, 0, 0, jst)},
				},
				{
					Id:              1,
					Name:            "テスト明細X",
					Price:           2345,
					RegistDate:      time.Date(2010, 1, 1, 13, 30, 0, 0, jst),
					UseDate:         types.Date{Time: time.Date(2010, 1, 1, 0, 0, 0, 0, jst)},
					ImportDate:      &t11,
					ImportJudgeDate: &t11,
				},
			},
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				Logger: l,
				Repo:   &mockDBClient{err: errors.New("error")},
			},
			args: args{
				ctx:   context.Background(),
				limit: 5,
			},
			wantDets: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &APIService{
				Logger: tt.fields.Logger,
				Repo:   tt.fields.Repo,
			}
			gotDets, err := a.GetDetails(tt.args.ctx, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIService.GetDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDets, tt.wantDets) {
				t.Errorf("APIService.GetDetails() = %v, want %v", gotDets, tt.wantDets)
			}
		})
	}
}

func TestAPIService_GetRules(t *testing.T) {
	type fields struct {
		Logger *zap.Logger
		Repo   DBRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []openapi.Rule
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Logger: l,
				Repo:   &mockDBClient{},
			},
			args: args{ctx: context.Background()},
			want: []openapi.Rule{
				{
					Id:         1,
					FieldName:  "name",
					Value:      "かんぜんいっち",
					ExactMatch: 1,
					CategoryId: 100,
				},
				{
					Id:         2,
					FieldName:  "m_category",
					Value:      "ぶぶんいっち",
					ExactMatch: 0,
					CategoryId: 400,
				},
			},
		},
		{
			name: "error",
			fields: fields{
				Logger: l,
				Repo:   &mockDBClient{err: errors.New("error")},
			},
			args:    args{ctx: context.Background()},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &APIService{
				Logger: tt.fields.Logger,
				Repo:   tt.fields.Repo,
			}
			got, err := a.GetRules(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("APIService.GetRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("APIService.GetRules() = %v, want %v", got, tt.want)
			}
		})
	}
}
