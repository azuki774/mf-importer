package repository

import (
	"encoding/csv"
	"io"
	"mf-importer/internal/model"
	"os"
	"strconv"
)

const LabelFieldName = "フィールド名"

func LoadExtractCSV(path string) (es []model.ExtractRuleCSV, err error) {
	f, err := os.Open(path)
	if err != nil {
		return []model.ExtractRuleCSV{}, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	for {
		r, err := reader.Read()
		if err == io.EOF {
			break // read end
		}
		if err != nil {
			return []model.ExtractRuleCSV{}, err
		}

		// 型変換だけして読み込み
		var e model.ExtractRuleCSV
		// ラベル行は pass する
		if r[0] == LabelFieldName {
			continue
		}

		e.FieldName = r[0]
		e.Name = r[1]
		e.ExtractCondition, err = strconv.ParseBool(r[2])
		if err != nil {
			return []model.ExtractRuleCSV{}, err
		}

		e.CategoryID, err = strconv.Atoi(r[3])
		if err != nil {
			return []model.ExtractRuleCSV{}, err
		}

		es = append(es, e)
	}

	return es, nil
}
