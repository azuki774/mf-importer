package model

import (
	"mf-importer/internal/util"
	"reflect"
	"testing"
	"time"
)

func Test_getDateFromCSV(t *testing.T) {
	type args struct {
		rawDate string
	}
	tests := []struct {
		name     string
		args     args
		wantDate time.Time
		wantErr  bool
	}{
		{
			name: "normal1",
			args: args{
				rawDate: "2024/08/22",
			},
			wantDate: time.Date(2024, 8, 22, 0, 0, 0, 0, time.Local),
		},
		{
			name: "error1",
			args: args{
				rawDate: "1/30(火)",
			},
			wantDate: time.Time{},
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDate, err := getDateFromCSV(tt.args.rawDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDateFromCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDate, tt.wantDate) {
				t.Errorf("getDateFromCSV() = %v, want %v", gotDate, tt.wantDate)
			}
		})
	}
}

func Test_getPriceFromCSV(t *testing.T) {
	type args struct {
		rawPrice string
	}
	tests := []struct {
		name      string
		args      args
		wantPrice int64
		wantErr   bool
	}{
		{
			name: "-291",
			args: args{
				rawPrice: "-291",
			},
			wantPrice: int64(291),
			wantErr:   false,
		},
		{
			name: "-1,291",
			args: args{
				rawPrice: "-1,291",
			},
			wantPrice: int64(1291),
			wantErr:   false,
		},
		{
			name: "-1,a291 (error)",
			args: args{
				rawPrice: "-1,a291",
			},
			wantPrice: int64(0),
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrice, err := getPriceFromCSV(tt.args.rawPrice)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPriceFromCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrice != tt.wantPrice {
				t.Errorf("getPriceFromCSV() = %v, want %v", gotPrice, tt.wantPrice)
			}
		})
	}
}

func TestConvCSVtoDetail(t *testing.T) {
	type args struct {
		csv [][]string
	}
	tests := []struct {
		name        string
		args        args
		wantDetails []Detail
		wantErr     bool
		nowT        time.Time // Unittestの時間
	}{
		{
			name: "normal",
			args: args{
				csv: [][]string{
					{"", "2024/07/19", "はま寿司", `"-1,705"`, "三井住友カード", "食費", "外食", "", "", ""}, // YYYYMMID = 2
					{"", "2024/07/16", "ローソン", "-291", "三井住友カード", "食費", "食料品", "", "", ""},    // YYYYMMDDID = 1
				},
			},
			wantDetails: []Detail{
				{
					YYYYMMID:  2,
					Date:      time.Date(2024, 07, 19, 0, 0, 0, 0, time.Local),
					Name:      "はま寿司",
					FinIns:    "三井住友カード",
					LCategory: "食費",
					MCategory: "外食",
					// RegistDate: util.NowFunc(),
					Price:    int64(1705),
					RawDate:  "2024/07/19",
					RawPrice: `"-1,705"`,
				},
				{
					YYYYMMID:  1,
					Date:      time.Date(2024, 07, 16, 0, 0, 0, 0, time.Local),
					Name:      "ローソン",
					FinIns:    "三井住友カード",
					LCategory: "食費",
					MCategory: "食料品",
					// RegistDate: util.NowFunc(),
					Price:    int64(291),
					RawDate:  "2024/07/16",
					RawPrice: "-291",
				},
			},
			wantErr: false,
			nowT:    time.Date(2024, 07, 20, 0, 0, 0, 0, time.Local),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.NowFunc = tt.nowT.Local
			// テストデータ want の RegistDate は後から設定
			tt.wantDetails[0].RegistDate = util.NowFunc()
			tt.wantDetails[1].RegistDate = util.NowFunc()

			gotDetails, err := ConvCSVtoDetail(tt.args.csv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvCSVtoDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDetails, tt.wantDetails) {
				t.Errorf("ConvCSVtoDetail() = %v, want %v", gotDetails, tt.wantDetails)
			}
		})
	}
}
