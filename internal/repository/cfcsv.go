package repository

import (
	"context"
	"encoding/csv"
	"mf-importer/internal/model"
	"os"
)

func LoadCfCSV(ctx context.Context, path string) (details []model.Detail, err error) {
	file, err := os.Open(path)
	if err != nil {
		return []model.Detail{}, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	rows, err := r.ReadAll()
	if err != nil {
		return []model.Detail{}, err
	}

	details, err = model.ConvCSVtoDetail(rows)
	if err != nil {
		return []model.Detail{}, err
	}
	return details, nil
}
