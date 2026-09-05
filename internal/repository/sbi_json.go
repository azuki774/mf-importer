package repository

import (
	"context"
	"fmt"
	"io/fs"
	"mf-importer/internal/model"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// SbiJSONOperator は SBI JSON ファイルの読み込みと探索を担当する
type SbiJSONOperator struct {
	Logger *zap.Logger
}

// LoadSbiJSON は JSON ファイルを読み込み、SbiSnapshot と holdings に変換する
// status は内部で OK/MAINTENANCE/ERROR に正規化される
func (o *SbiJSONOperator) LoadSbiJSON(ctx context.Context, path string) (*model.SbiSnapshot, []model.SbiHolding, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read sbi json %s: %w", path, err)
	}
	snap, holdings, err := model.ParseSbiJSON(data)
	if err != nil {
		return nil, nil, fmt.Errorf("parse sbi json %s: %w", path, err)
	}
	return snap, holdings, nil
}

// GetSbiTargetFiles は inputDir 配下を再帰的に探索し、.json ファイルを返す
// SBIは s3://bucket/prefix/YYYY/MM/YYYYMMDD-HHMMSS.json のネスト構造のため WalkDir を使う
func (o *SbiJSONOperator) GetSbiTargetFiles(ctx context.Context, inputDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(d.Name()), ".json") {
			abs, err := filepath.Abs(path)
			if err != nil {
				if o.Logger != nil {
					o.Logger.Error("failed to get Abs path", zap.String("path", path), zap.Error(err))
				}
				return err
			}
			files = append(files, abs)
		}
		return nil
	})
	if err != nil {
		if o.Logger != nil {
			o.Logger.Error("failed to walk sbi input directory", zap.String("inputDir", inputDir), zap.Error(err))
		}
		return nil, fmt.Errorf("walk sbi dir %s: %w", inputDir, err)
	}
	return files, nil
}
