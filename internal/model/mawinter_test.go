package model

import (
	"reflect"
	"testing"
)

func Test_convPriceForm(t *testing.T) {
	type args struct {
		orig string
	}
	tests := []struct {
		name      string
		args      args
		wantPrice int
		wantErr   bool
	}{
		{
			name:      "ok 1",
			args:      args{"-1,234"},
			wantPrice: 1234,
			wantErr:   false,
		},
		{
			name:      "ok 2",
			args:      args{"-1,234,567"},
			wantPrice: 1234567,
			wantErr:   false,
		},
		{
			name:      "error",
			args:      args{"-1,234#"},
			wantPrice: 0,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrice, err := convPriceForm(tt.args.orig)
			if (err != nil) != tt.wantErr {
				t.Errorf("convPriceForm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrice != tt.wantPrice {
				t.Errorf("convPriceForm() = %v, want %v", gotPrice, tt.wantPrice)
			}
		})
	}
}

func TestNewCreateRecord(t *testing.T) {
	type args struct {
		c CFRecord
	}
	tests := []struct {
		name    string
		args    args
		wantR   CreateRecord
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				c: CFRecord{
					YYYYMMDD:   "20230521",
					Price:      "-1,451",
					CategoryID: 100,
					Name:       "なまえA",
				},
			},
			wantR: CreateRecord{
				CategoryID: 100,
				Date:       "20230521",
				Price:      1451,
				From:       "mf-importer",
				Memo:       "なまえA",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, err := NewCreateRecord(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCreateRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("NewCreateRecord() = %v, want %v", gotR, tt.wantR)
			}
		})
	}
}
