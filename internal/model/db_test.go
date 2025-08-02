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

func Test_convertAmountFromCSV(t *testing.T) {
	type args struct {
		rawAmount string
	}
	tests := []struct {
		name       string
		args       args
		wantAmount int
		wantErr    bool
	}{
		{
			name: "simple number",
			args: args{
				rawAmount: "1000",
			},
			wantAmount: 1000,
			wantErr:    false,
		},
		{
			name: "number with comma",
			args: args{
				rawAmount: "1,000,000",
			},
			wantAmount: 1000000,
			wantErr:    false,
		},
		{
			name: "quoted number",
			args: args{
				rawAmount: `"8,839,399"`,
			},
			wantAmount: 8839399,
			wantErr:    false,
		},
		{
			name: "invalid number",
			args: args{
				rawAmount: "abc",
			},
			wantAmount: 0,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAmount, err := convertAmountFromCSV(tt.args.rawAmount)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertAmountFromCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAmount != tt.wantAmount {
				t.Errorf("convertAmountFromCSV() = %v, want %v", gotAmount, tt.wantAmount)
			}
		})
	}
}

func TestConvCSVtoAssetHistory(t *testing.T) {
	type args struct {
		csv [][]string
	}
	tests := []struct {
		name          string
		args          args
		wantHistories []AssetHistory
		wantErr       bool
	}{
		{
			name: "normal",
			args: args{
				csv: [][]string{
					{"日付", "総額", "現金・預金・投資信託", "債券", "その他金融資産", "ポイント", "詳細"}, // ヘッダー行
					{"2025-08-02", `"5,000,000"`, `"3,500,000"`, "0", `"1,400,000"`, `"100,000"`, "テスト詳細"},
					{"2025-08-01", `"4,800,000"`, `"3,200,000"`, "0", `"1,500,000"`, `"100,000"`, "テスト詳細"},
				},
			},
			wantHistories: []AssetHistory{
				{
					Date:        time.Date(2025, 8, 2, 0, 0, 0, 0, time.Local),
					TotalAmount: 5000000,
					CashDeposit: 3500000,
					Bonds:       0,
					OtherAssets: 1400000,
					Points:      100000,
					Details:     "テスト詳細",
				},
				{
					Date:        time.Date(2025, 8, 1, 0, 0, 0, 0, time.Local),
					TotalAmount: 4800000,
					CashDeposit: 3200000,
					Bonds:       0,
					OtherAssets: 1500000,
					Points:      100000,
					Details:     "テスト詳細",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid column count",
			args: args{
				csv: [][]string{
					{"2025-08-02", `"8,839,399"`, `"6,960,146"`}, // 列数が足りない
				},
			},
			wantHistories: []AssetHistory{},
			wantErr:       true,
		},
		{
			name: "invalid date format",
			args: args{
				csv: [][]string{
					{"invalid-date", `"5,000,000"`, `"3,500,000"`, "0", `"1,400,000"`, `"100,000"`, "テスト詳細"},
				},
			},
			wantHistories: []AssetHistory{},
			wantErr:       true,
		},
		{
			name: "month end date format",
			args: args{
				csv: [][]string{
					{"日付", "総額", "現金・預金・投資信託", "債券", "その他金融資産", "ポイント", "詳細"}, // ヘッダー行
					{"2025-05月末", `"6,000,000"`, `"4,000,000"`, "0", `"1,800,000"`, `"200,000"`, "5月末テスト詳細"},
					{"2024-02月末", `"5,500,000"`, `"3,800,000"`, "0", `"1,600,000"`, `"100,000"`, "2月末テスト詳細"},
				},
			},
			wantHistories: []AssetHistory{
				{
					Date:        time.Date(2025, 5, 31, 0, 0, 0, 0, time.Local), // 2025年5月末
					TotalAmount: 6000000,
					CashDeposit: 4000000,
					Bonds:       0,
					OtherAssets: 1800000,
					Points:      200000,
					Details:     "5月末テスト詳細",
				},
				{
					Date:        time.Date(2024, 2, 29, 0, 0, 0, 0, time.Local), // 2024年2月末（うるう年）
					TotalAmount: 5500000,
					CashDeposit: 3800000,
					Bonds:       0,
					OtherAssets: 1600000,
					Points:      100000,
					Details:     "2月末テスト詳細",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHistories, err := ConvCSVtoAssetHistory(tt.args.csv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvCSVtoAssetHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotHistories, tt.wantHistories) {
				t.Errorf("ConvCSVtoAssetHistory() = %v, want %v", gotHistories, tt.wantHistories)
			}
		})
	}
}
