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

// NrknJSONOperator は NRKN JSON ファイルの読み込みと探索を担当する
type NrknJSONOperator struct {
	Logger *zap.Logger
}

// LoadNrknJSON は JSON ファイルを読み込み、NrknSnapshot と holdings に変換する
// status は OK のみ受け付ける
func (o *NrknJSONOperator) LoadNrknJSON(ctx context.Context, path string) (*model.NrknSnapshot, []model.NrknHolding, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read nrkn json %s: %w", path, err)
	}
	snap, holdings, err := model.ParseNrknJSON(data)
	if err != nil {
		return nil, nil, fmt.Errorf("parse nrkn json %s: %w", path, err)
	}
	return snap, holdings, nil
}

// GetNrknTargetFiles は inputDir 配下を再帰的に探索し、.json ファイルを返す
// NRKNは s3://bucket/prefix/YYYY/MM/YYYYMMDD-HHMMSS.json のネスト構造のため WalkDir を使う
func (o *NrknJSONOperator) GetNrknTargetFiles(ctx context.Context, inputDir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(inputDir, func(path string, d fs.DirEntry, err error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
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
			o.Logger.Error("failed to walk nrkn input directory", zap.String("inputDir", inputDir), zap.Error(err))
		}
		return nil, fmt.Errorf("walk nrkn dir %s: %w", inputDir, err)
	}
	return files, nil
}
