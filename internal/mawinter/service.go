package mawinter

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"strings"
	"time"

	"go.uber.org/zap"
)

type CSVFileOperator interface {
	LoadExtractCSV(path string) (es []model.ExtractRuleCSV, err error)
}
type MongoDBClient interface {
	GetCFRecords(ctx context.Context) (cfRecords []model.CFRecord, err error)
	CheckCFRecords(ctx context.Context, cfRecords []model.CFRecord) (err error)    // これらのレコードを mawinter check済とする
	RegistedCFRecords(ctx context.Context, cfRecords []model.CFRecord) (err error) // これらのレコードを mawinter regist済とする
}
type MawinterClient interface {
	Regist(ctx context.Context, c model.CFRecord) (err error)
}

const csvFilePath = "/extract_rule.csv"

type Mawinter struct {
	Logger      *zap.Logger
	DBClient    MongoDBClient
	MawClient   MawinterClient
	CSVFileOp   CSVFileOperator
	ExtractRule model.ExtractRule // 抽出するルール
	ProcessDate time.Time         // 処理するファイルの登録日を指定
	Dryrun      bool              // DBの状態を変更しない、mawinter サーバに送信しない
}

func NewMawinter(db MongoDBClient, csv CSVFileOperator, maw MawinterClient, dryRun bool) Mawinter {
	var mawinter Mawinter
	l := logger.NewLogger()
	mawinter.Logger = l
	mawinter.DBClient = db
	mawinter.CSVFileOp = csv
	mawinter.MawClient = maw
	mawinter.Dryrun = dryRun
	return mawinter
}

// getCategoryIDwithExtractCond: 抽出条件に合う場合はカテゴリIDを出力する、そうでない場合は ok = false
func (m *Mawinter) getCategoryIDwithExtractCond(c model.CFRecord) (catID int, ok bool) {
	// FromName        map[string]int // Name -> CategoryID（完全一致）
	catID, ok = m.ExtractRule.FromName[c.Name]
	if ok {
		return catID, true
	}

	// FromNameIn      map[string]int // Name -> CategoryID（部分一致）
	for ruleName, catID := range m.ExtractRule.FromNameIn {
		// 1ルールずつ条件にあうか見ていく
		if strings.Contains(c.Name, ruleName) {
			return catID, true
		}
	}

	// FromMCategory   map[string]int // MCategory -> CategoryID（完全一致）
	catID, ok = m.ExtractRule.FromMCategory[c.MCategory]
	if ok {
		return catID, true
	}

	// FromMCategoryIn map[string]int // MCategory -> CategoryID（部分一致）
	for ruleName, catID := range m.ExtractRule.FromMCategoryIn {
		// 1ルールずつ条件にあうか見ていく
		if strings.Contains(c.MCategory, ruleName) {
			return catID, true
		}
	}

	return 0, false
}

func (m *Mawinter) Regist(ctx context.Context) (err error) {
	m.Logger.Info("Regist start")

	m.Logger.Info("extract rule CSV load")
	es, err := m.CSVFileOp.LoadExtractCSV(csvFilePath)
	if err != nil {
		m.Logger.Error("failed to load extract CSV", zap.Error(err))
		return err
	}

	m.ExtractRule = *model.NewExtractRule()
	for _, e := range es {
		err = m.ExtractRule.AddRule(e)
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
	cfCheckedRecs, err := m.DBClient.GetCFRecords(ctx)
	if err != nil {
		m.Logger.Error("failed to fetch records from DB", zap.Error(err))
		return err
	}
	m.Logger.Info("fetch records from DB complete", zap.Int("unchecked_records", len(cfCheckedRecs)))

	m.Logger.Info("extract data and convert to mawinter model")
	var cfRegistedRecs []model.CFRecord // cfCheckedRecs から抽出条件にあうものを入れる
	for _, c := range cfCheckedRecs {
		catID, ok := m.getCategoryIDwithExtractCond(c)
		if !ok {
			continue
		}

		// 抽出条件にあうものは categoryID をセットして追加
		c.CategoryID = catID
		cfRegistedRecs = append(cfRegistedRecs, c)
		m.Logger.Info("extract data", zap.String("name", c.Name), zap.String("yyyymmdd", c.YYYYMMDD), zap.String("m_category", c.MCategory), zap.String("price", c.Price))
	}
	m.Logger.Info("extract data and convert to mawinter model complete", zap.Int("extracted_records", len(cfRegistedRecs)))

	if m.Dryrun {
		m.Logger.Info("Regist dry-run end")
		return nil
	}

	m.Logger.Info("post to mawinter")
	for _, c := range cfRegistedRecs {
		err := m.MawClient.Regist(ctx, c)
		if err != nil {
			m.Logger.Error("failed to insert", zap.Error(err))
			return err
		}
		m.Logger.Info("post records", zap.String("yyyymmdd", c.YYYYMMDD), zap.String("price", c.Price), zap.String("memo", c.Name))
	}
	m.Logger.Info("post to mawinter complete")

	m.Logger.Info("record post mawinter history")
	err = m.DBClient.CheckCFRecords(ctx, cfCheckedRecs) // checked
	if err != nil {
		m.Logger.Error("failed to insert checked histories", zap.Error(err))
		return err
	}

	err = m.DBClient.RegistedCFRecords(ctx, cfRegistedRecs) // registed
	if err != nil {
		m.Logger.Error("failed to insert registed histories", zap.Error(err))
		return err
	}

	m.Logger.Info("record post mawinter history complete")

	m.Logger.Info("Regist end", zap.Int("add_checked_record", len(cfCheckedRecs)), zap.Int("add_registed_record", len(cfRegistedRecs)))
	return nil
}
