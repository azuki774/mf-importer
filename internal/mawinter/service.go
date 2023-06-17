package mawinter

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"time"

	"go.uber.org/zap"
)

type CSVFileOperator interface {
	LoadExtractCSV(path string) (es []model.ExtractRuleCSV, err error)
}
type MongoDBClient interface {
	GetCFRecords(ctx context.Context) (cfRecords []model.CFRecords, err error)
}
type MawinterClient interface {
	Regist(ctx context.Context, r model.CreateRecord) (err error)
}

const csvFilePath = "/extract_rule.csv"

type Mawinter struct {
	Logger      *zap.Logger
	DBClient    MongoDBClient
	MawClient   MawinterClient
	CSVFileOp   CSVFileOperator
	ExtractRule model.ExtractRule // 抽出するルール
	ProcessDate time.Time         // 処理するファイルの登録日を指定
}

func NewMawinter(db MongoDBClient, csv CSVFileOperator) Mawinter {
	var mawinter Mawinter
	l := logger.NewLogger()
	mawinter.Logger = l
	mawinter.DBClient = db
	mawinter.CSVFileOp = csv
	return mawinter
}

func (m *Mawinter) Regist(ctx context.Context) (err error) {
	m.Logger.Info("Regist start")

	m.Logger.Info("extract rule CSV load")
	es, err := m.CSVFileOp.LoadExtractCSV(csvFilePath)
	if err != nil {
		m.Logger.Error("failed to load extract CSV", zap.Error(err))
		return err
	}

	rule := model.NewExtractRule()
	for _, e := range es {
		err = rule.AddRule(e)
		if err != nil {
			m.Logger.Error("failed to add rule", zap.Error(err))
			return err
		}
		m.Logger.Info("add rule",
			zap.String("field_name", e.FieldName),
			zap.String("name", e.Name),
			zap.Bool("extract_condition", e.ExtractCondition),
			zap.Int("category_id", e.CategoryID),
		)
	}
	m.Logger.Info("extract rule CSV load complete")

	m.Logger.Info("fetch records from DB")
	cfRecs, err := m.DBClient.GetCFRecords(ctx)
	if err != nil {
		m.Logger.Error("failed to fetch records from DB", zap.Error(err))
		return err
	}
	m.Logger.Info("fetch records from DB complete", zap.Int("unregisted records", len(cfRecs)))

	m.Logger.Info("extract data and convert to mawinter model")
	var rs []model.CreateRecord

	m.Logger.Info("extract data and convert to mawinter model complete")

	m.Logger.Info("post to mawinter")
	for _, r := range rs {
		err := m.MawClient.Regist(ctx, r)
		if err != nil {
			m.Logger.Error("failed to insert", zap.Error(err))
			return err
		}
	}
	m.Logger.Info("post to mawinter complete")

	m.Logger.Info("record post mawinter history")

	m.Logger.Info("record post mawinter history complete")

	m.Logger.Info("Regist end")
	return nil
}
