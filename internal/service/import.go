package service

import "go.uber.org/zap"

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
