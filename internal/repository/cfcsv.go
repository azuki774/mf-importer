package repository

import (
	"context"
	"encoding/csv"
	"mf-importer/internal/model"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type DetailCSVOperator struct {
	Logger *zap.Logger
}

func (d *DetailCSVOperator) LoadCfCSV(ctx context.Context, path string) (details []model.Detail, err error) {
	file, err := os.Open(path)
	if err != nil {
		return []model.Detail{}, err
	}
	defer file.Close()

	// SJIS -> UTF8
	sjisDecoder := japanese.ShiftJIS.NewDecoder()
	utf8Reader := transform.NewReader(file, sjisDecoder)
	reader := csv.NewReader(utf8Reader)
	rows, err := reader.ReadAll()
	if err != nil {
		return []model.Detail{}, err
	}

	details, err = model.ConvCSVtoDetail(rows)
	if err != nil {
		return []model.Detail{}, err
	}
	return details, nil
}

// inputDir から読み取り対象のファイルを取得
func (d *DetailCSVOperator) GetTargetFiles(ctx context.Context, inputDir string) (targetCSVs []string, err error) {
	files, err := os.ReadDir(inputDir)
	if err != nil {
		d.Logger.Error("failed to get input directory", zap.Error(err))
		return []string{}, err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".csv" {
			curDir := filepath.Join(inputDir, f.Name())
			absPath, err := filepath.Abs(curDir)
			if err != nil {
				d.Logger.Error("failed to get Abs path", zap.Error(err))
				return []string{}, err
			}
			targetCSVs = append(targetCSVs, absPath)
		}
	}
	return targetCSVs, nil
}
