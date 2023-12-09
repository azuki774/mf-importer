package model

import "testing"

func TestIsSuicaDetail(t *testing.T) {
	type args struct {
		d Detail
	}
	tests := []struct {
		name   string
		args   args
		wantOk bool
	}{
		{
			name: "suica",
			args: args{
				d: Detail{
					Name:   "入 東京 出 大阪",
					FinIns: "モバイルSuica (モバイルSuica ID)",
				},
			},
			wantOk: true,
		},
		{
			name: "no suica(fin_ins)",
			args: args{
				d: Detail{
					Name:   "入 東京 出 大阪",
					FinIns: "suicaじゃない",
				},
			},
			wantOk: false,
		},
		{
			name: "no suica(name) 1",
			args: args{
				d: Detail{
					Name:   "出 東京 入 大阪",
					FinIns: "モバイルSuica (モバイルSuica ID)",
				},
			},
			wantOk: false,
		},
		{
			name: "no suica(name) 2",
			args: args{
				d: Detail{
					Name:   "物販",
					FinIns: "モバイルSuica (モバイルSuica ID)",
				},
			},
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := IsSuicaDetail(tt.args.d); gotOk != tt.wantOk {
				t.Errorf("IsSuicaDetail() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
