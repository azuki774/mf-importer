package mawinter

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"strings"
	"time"

	"go.uber.org/zap"
)

type DBClient interface {
	GetCFDetails(ctx context.Context) (cfDetails []model.Detail, err error) // mawinter check未チェックのデータを取得
	// CheckCFDetail: レコードを mawinter check/regist 済とする. regist = true なら registフラグも立てる
	CheckCFDetail(ctx context.Context, cfDetail model.Detail, regist bool) (err error)
	GetExtractRules(ctx context.Context) (er []model.ExtractRuleDB, err error)
}
type MawinterClient interface {
	Regist(ctx context.Context, c model.Detail, catID int) (err error)
}

type Mawinter struct {
	Logger      *zap.Logger
	DBClient    DBClient
	MawClient   MawinterClient
	ExtractRule model.ExtractRule // 抽出するルール
	ProcessDate time.Time         // 処理するファイルの登録日を指定
	Dryrun      bool              // DBの状態を変更しない、mawinter サーバに送信しない
}

func NewMawinter(db DBClient, maw MawinterClient, dryRun bool) Mawinter {
	var mawinter Mawinter
	l := logger.NewLogger()
	mawinter.Logger = l
	mawinter.DBClient = db
	mawinter.MawClient = maw
	mawinter.Dryrun = dryRun
	return mawinter
}

// getCategoryIDwithExtractCond: 抽出条件に合う場合はカテゴリIDを出力する、そうでない場合は ok = false
func (m *Mawinter) getCategoryIDwithExtractCond(c model.Detail) (catID int, ok bool) {
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

	// Fetch Extract Rule from DB
	m.Logger.Info("fetch extract rules from DB")
	m.ExtractRule = *model.NewExtractRule()
	ers, err := m.DBClient.GetExtractRules(ctx)
	if err != nil {
		m.Logger.Error("failed to extract rules from DB", zap.Error(err))
		return err
	}
	err = m.ExtractRule.AddRule(ers)
	if err != nil {
		m.Logger.Error("failed to add extract rules", zap.Error(err))
		return err
	}

	// Fetch Details from DB
	cfDetails, err := m.DBClient.GetCFDetails(ctx)
	if err != nil {
		m.Logger.Error("failed to fetch details from DB", zap.Error(err))
		return err
	}
	m.Logger.Info("fetch records from DB complete")

	for _, c := range cfDetails {
		cLogger := m.Logger.With(
			zap.Time("Date", c.Date),
			zap.String("Name", c.Name),
			zap.Int64("Price", c.Price),
			zap.String("M_Category", c.MCategory),
		)

		// catID を取得する。ない場合は抽出対象外なので mawinter に post しない
		catID, regist := m.getCategoryIDwithExtractCond(c)

		// DryRunMode ONのときは何もせず終了
		if m.Dryrun {
			cLogger.Info("dryrun: post to mawinter", zap.Bool("regist", regist))
			continue
		}

		if regist {
			cLogger.Info("post to mawinter")
			err := m.MawClient.Regist(ctx, c, catID)
			if err != nil {
				m.Logger.Error("failed to post mawinter", zap.Error(err))
				return err
			}
			cLogger.Info("post to mawinter complete")
		}

		// detail テーブルを更新する
		err = m.DBClient.CheckCFDetail(ctx, c, regist)
		if err != nil {
			cLogger.Error("failed to insert checked histories", zap.Error(err))
			return err
		}

	}

	m.Logger.Info("Regist end")
	return nil
}
