package service

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type DBClient interface{}
type Importer struct {
	Logger   *zap.Logger
	DBClient DBClient
	InputDir string
	DryRun   bool
}

func NewImporter(l *zap.Logger, DBClient DBClient, inputDir string, dryrun bool) *Importer {
	return &Importer{
		Logger:   l,
		DBClient: DBClient,
		InputDir: inputDir,
		DryRun:   dryrun,
	}
}

// inputDir から読み取り対象のファイルを取得
func (i *Importer) getTargetFiles(ctx context.Context) (targetCSVs []string, err error) {
	files, err := os.ReadDir(i.InputDir)
	if err != nil {
		i.Logger.Error("failed to get input directory", zap.Error(err))
		return []string{}, err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".csv" {
			absPath, err := filepath.Abs(f.Name())
			if err != nil {
				i.Logger.Error("failed to get Abs path", zap.Error(err))
				return []string{}, err
			}
			targetCSVs = append(targetCSVs, absPath)
		}
	}
	return targetCSVs, nil
}

func (i *Importer) Start(ctx context.Context) (err error) {
	targetCSVs, err := i.getTargetFiles(ctx)
	if err != nil {
		i.Logger.Error("failed to get target CSV files", zap.Error(err))
		return err
	}
	i.Logger.Info("get target CSV files", zap.Strings("path", targetCSVs))
	return nil
}
