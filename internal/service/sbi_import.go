package service

import (
	"context"
	"fmt"
	"mf-importer/internal/model"
	"sort"

	"go.uber.org/zap"
)

// SbiDBClient persists one parsed SBI snapshot and its holdings atomically.
type SbiDBClient interface {
	ImportSbiSnapshot(context.Context, *model.SbiSnapshot, []model.SbiHolding) (inserted bool, err error)
}

// SbiJSONOperator discovers and parses SBI JSON files.
type SbiJSONOperator interface {
	GetSbiTargetFiles(context.Context, string) ([]string, error)
	LoadSbiJSON(context.Context, string) (*model.SbiSnapshot, []model.SbiHolding, error)
}

type SbiImportResult struct {
	Files    int
	Inserted int
	Skipped  int
}

type SbiImporter struct {
	Logger   *zap.Logger
	DBClient SbiDBClient
	Operator SbiJSONOperator
}

func NewSbiImporter(l *zap.Logger, dbClient SbiDBClient, operator SbiJSONOperator) *SbiImporter {
	return &SbiImporter{Logger: l, DBClient: dbClient, Operator: operator}
}

func (i *SbiImporter) Start(ctx context.Context, inputDir string) (SbiImportResult, error) {
	files, err := i.Operator.GetSbiTargetFiles(ctx, inputDir)
	if err != nil {
		return SbiImportResult{}, fmt.Errorf("get sbi target files: %w", err)
	}
	if len(files) == 0 {
		return SbiImportResult{}, fmt.Errorf("no SBI JSON files found in %s", inputDir)
	}
	sort.Strings(files)

	var result SbiImportResult
	for _, path := range files {
		snapshot, holdings, err := i.Operator.LoadSbiJSON(ctx, path)
		if err != nil {
			return result, fmt.Errorf("load SBI JSON %s: %w", path, err)
		}
		if snapshot == nil {
			return result, fmt.Errorf("load SBI JSON %s: empty snapshot", path)
		}
		result.Files++
		if snapshot.Status != model.SbiStatusOK {
			holdings = nil
		}
		inserted, err := i.DBClient.ImportSbiSnapshot(ctx, snapshot, holdings)
		if err != nil {
			return result, fmt.Errorf("import SBI JSON %s: %w", path, err)
		}
		if inserted {
			result.Inserted++
		} else {
			result.Skipped++
		}
	}
	return result, nil
}
