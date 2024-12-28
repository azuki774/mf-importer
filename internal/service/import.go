package service

import (
	"context"
	"mf-importer/internal/model"
	"mf-importer/internal/repository"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type DBClient interface {
	CheckAlreadyRegistDetail(ctx context.Context, detail model.Detail) (exists bool, err error)
	RegistDetail(ctx context.Context, detail model.Detail) (err error)
	RegistDetailHistory(ctx context.Context, jobname string, parsedNum int, insertNum int, srcFile string) (err error)
	// そのソースファイルを取り込んだ、最後の情報をDBから取得する
	GetLastDetailHistoryWhereSrcFile(ctx context.Context, srcFile string) (parsedNum, insertedNum int, err error)
}

type DetailCSVOperator interface {
	LoadCfCSV(ctx context.Context, path string) (details []model.Detail, err error)
	GetTargetFiles(ctx context.Context, inputDir string) (targetCSVs []string, err error)
}

// cf Detail
type Importer struct {
	Logger   *zap.Logger
	DBClient DBClient
	CSVOpe   DetailCSVOperator
	InputDir string
	JobName  string
	DryRun   bool
}

func NewImporter(l *zap.Logger, DBClient DBClient, inputDir string, dryrun bool) *Importer {
	return &Importer{
		Logger:   l,
		DBClient: DBClient,
		CSVOpe:   &repository.DetailCSVOperator{Logger: l},
		InputDir: inputDir,
		DryRun:   dryrun,
		JobName:  os.Getenv("jobname"),
	}
}

func (i *Importer) Start(ctx context.Context) (err error) {
	targetCSVs, err := i.CSVOpe.GetTargetFiles(ctx, i.InputDir)
	if err != nil {
		i.Logger.Error("failed to get target CSV files", zap.Error(err))
		return err
	}
	i.Logger.Info("get target CSV files", zap.Strings("path", targetCSVs))

	for _, path := range targetCSVs {
		lf := i.Logger.With(zap.String("file", path))
		details, err := i.CSVOpe.LoadCfCSV(ctx, path)
		if err != nil {
			return err
		}

		// d.YYYYMMID とDBの挿入順 (ID) を一致させるため、逆順にする
		var revDetails []model.Detail
		for i := 0; i < len(details); i++ {
			revDetails = append(revDetails, details[len(details)-i-1])
		}

		var parsedNum int
		var insertedNum int

		for _, d := range revDetails {
			exists, err := i.DBClient.CheckAlreadyRegistDetail(ctx, d)
			if err != nil {
				return err
			}
			parsedNum += 1
			if exists {
				// 登録済なら skip
				continue
			}
			lf.Info("new data detected, insert to DB", zap.Int("yyyymm_id", d.YYYYMMID))
			if i.DryRun {
				lf.Info("however, it works as dry-run. do nothing.")
			} else if err := i.DBClient.RegistDetail(ctx, d); err != nil {
				lf.Error("failed to insert DB", zap.Error(err))
				return err
			}
			insertedNum += 1
		}

		lf.Info("insert detail sucessfully", zap.Int("parsedNum", parsedNum), zap.Int("insertedNum", insertedNum))

		// history regist
		fileName := filepath.Base(path)
		pn, _, err := i.DBClient.GetLastDetailHistoryWhereSrcFile(ctx, fileName)
		if err != nil {
			// このデータ取得失敗ではエラーにしない
			lf.Warn("failed to get last detail history", zap.Error(err))
		}
		if pn != parsedNum || err != nil {
			// 前回取り込み時と変化を検知したとき or 前回取り込みがエラーになったとき
			if err := i.DBClient.RegistDetailHistory(ctx, i.JobName, parsedNum, insertedNum, fileName); err != nil {
				lf.Error("failed to insert import history", zap.Error(err))
				return err
			}
		}

		lf.Info("insert detail history sucessfully")
	}

	return nil
}
