package service

import (
	"context"
	"errors"
	"fmt"
	"mf-importer/internal/model"
	"sort"

	"go.uber.org/zap"
)

// NrknDBClient persists one parsed NRKN snapshot and its holdings atomically.
type NrknDBClient interface {
	ImportNrknSnapshot(context.Context, *model.NrknSnapshot, []model.NrknHolding) (inserted bool, err error)
}

// NrknJSONOperator discovers and parses NRKN JSON files.
type NrknJSONOperator interface {
	GetNrknTargetFiles(context.Context, string) ([]string, error)
	LoadNrknJSON(context.Context, string) (*model.NrknSnapshot, []model.NrknHolding, error)
}

type NrknImportResult struct {
	Files       int
	Inserted    int
	Skipped     int
	Unsupported int
}

type NrknImporter struct {
	Logger   *zap.Logger
	DBClient NrknDBClient
	Operator NrknJSONOperator
}

func NewNrknImporter(l *zap.Logger, dbClient NrknDBClient, operator NrknJSONOperator) *NrknImporter {
	return &NrknImporter{Logger: l, DBClient: dbClient, Operator: operator}
}

func (i *NrknImporter) Start(ctx context.Context, inputDir string) (NrknImportResult, error) {
	files, err := i.Operator.GetNrknTargetFiles(ctx, inputDir)
	if err != nil {
		return NrknImportResult{}, fmt.Errorf("get nrkn target files: %w", err)
	}
	if len(files) == 0 {
		return NrknImportResult{}, fmt.Errorf("no NRKN JSON files found in %s", inputDir)
	}
	sort.Strings(files)

	var result NrknImportResult
	for _, path := range files {
		result.Files++
		snapshot, holdings, err := i.Operator.LoadNrknJSON(ctx, path)
		if err != nil {
			if errors.Is(err, model.ErrUnsupportedNrknSchemaVersion) {
				result.Unsupported++
				if i.Logger != nil {
					i.Logger.Warn("skip NRKN JSON with unsupported schema version", zap.String("path", path))
				}
				continue
			}
			return result, fmt.Errorf("load NRKN JSON %s: %w", path, err)
		}
		if snapshot == nil {
			return result, fmt.Errorf("load NRKN JSON %s: empty snapshot", path)
		}
		inserted, err := i.DBClient.ImportNrknSnapshot(ctx, snapshot, holdings)
		if err != nil {
			return result, fmt.Errorf("import NRKN JSON %s: %w", path, err)
		}
		if inserted {
			result.Inserted++
		} else {
			result.Skipped++
		}
	}
	return result, nil
}
