package repository

import (
	"context"
	"encoding/csv"
	"io"
	"mf-importer/internal/model"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type DetailCSVOperator struct {
	Logger   *zap.Logger
	Encoding string // "utf8" (default) or "sjis"
}

func (d *DetailCSVOperator) newCSVReader(r io.Reader) *csv.Reader {
	if strings.EqualFold(d.Encoding, "sjis") {
		return csv.NewReader(transform.NewReader(r, japanese.ShiftJIS.NewDecoder()))
	}
	return csv.NewReader(r)
}

func (d *DetailCSVOperator) LoadCfCSV(ctx context.Context, path string) (details []model.Detail, err error) {
	file, err := os.Open(path)
	if err != nil {
		return []model.Detail{}, err
	}
	defer file.Close()

	reader := d.newCSVReader(file)
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

func (d *DetailCSVOperator) LoadBsHistoryCSV(ctx context.Context, path string) (histories []model.AssetHistory, err error) {
	file, err := os.Open(path)
	if err != nil {
		return []model.AssetHistory{}, err
	}
	defer file.Close()

	reader := d.newCSVReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return []model.AssetHistory{}, err
	}

	histories, err = model.ConvCSVtoAssetHistory(rows)
	if err != nil {
		return []model.AssetHistory{}, err
	}
	return histories, nil
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
