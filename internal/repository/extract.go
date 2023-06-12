package repository

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// ExtractRuleCSV: extract_rule.csv を読み込む構造体（1レコード分）
type ExtractRuleCSV struct {
	FieldName        string
	Name             string
	ExtractCondition bool // 完全一致かどうか
	CategoryID       int  // 変換先の category_id
}

func LoadExtractCSV(path string) (es []ExtractRuleCSV, err error) {
	f, err := os.Open(path)
	if err != nil {
		return []ExtractRuleCSV{}, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break // read end
		}
		if err != nil {
			return []ExtractRuleCSV{}, err
		}

		// ラベル行は pass する

		// 型変換だけして読み込み
		fmt.Println(record)
	}

	return []ExtractRuleCSV{}, nil
}
