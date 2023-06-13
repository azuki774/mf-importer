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
	GetCFRecords(ctx context.Context, registDate string) (cfRecords []model.CFRecords, err error)
}
type MawinterClient interface {
	Regist(ctx context.Context, r model.CreateRecord) (err error)
}

type Mawinter struct {
	Logger      *zap.Logger
	DBClient    MongoDBClient
	MawClient   MawinterClient
	CSVFileOp   CSVFileOperator
	ExtractRule model.ExtractRule // 抽出するルール
	ProcessDate time.Time         // 処理するファイルの登録日を指定
}

func NewMawinter() Mawinter {
	var mawinter Mawinter
	l := logger.NewLogger()
	mawinter.Logger = l
	return mawinter
}

func (m *Mawinter) Regist(ctx context.Context) (err error) {
	m.Logger.Info("Regist start")

	m.Logger.Info("extract rule CSV load")
	rule := model.NewExtractRule()
	es, err := m.CSVFileOp.LoadExtractCSV("")
	if err != nil {
		m.Logger.Error("failed to load extract CSV", zap.Error(err))
		return err
	}

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

	m.Logger.Info("fetch records from DB complete")

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
