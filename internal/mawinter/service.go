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
	GetCFDetails(ctx context.Context) (cfDetails []model.Detail, err error)  // mawinter check未チェックのデータを取得
	CheckCFDetail(ctx context.Context, cfDetail model.Detail) (err error)    // これらのレコードを mawinter check済とする
	RegistedCFDetail(ctx context.Context, cfDetail model.Detail) (err error) // これらのレコードを mawinter regist済とする
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

	// Fetch Details from DB
	cfDetails, err := m.DBClient.GetCFDetails(ctx)
	if err != nil {
		m.Logger.Error("failed to fetch details from DB", zap.Error(err))
		return err
	}
	m.Logger.Info("fetch records from DB complete")

	for _, c := range cfDetails {
		catID, ok := m.getCategoryIDwithExtractCond(c)
		if !ok {
			// categoryID がない -> 抽出条件がないとき
			continue
		}

		err = m.DBClient.CheckCFDetail(ctx, c) // checked
		if err != nil {
			m.Logger.Error("failed to insert checked histories", zap.Error(err))
			return err
		}

		// mawinter 登録対象のとき

		// DryRunMode ONのときは何もせず終了
		if m.Dryrun {
			// comment
			continue
		}

		m.Logger.Info("post to mawinter")
		err := m.MawClient.Regist(ctx, c, catID)
		if err != nil {
			m.Logger.Error("failed to post mawinter", zap.Error(err))
			return err
		}
		m.Logger.Info("post to mawinter complete")

		m.Logger.Info("record post mawinter history")

		err = m.DBClient.RegistedCFDetail(ctx, c) // registed
		if err != nil {
			m.Logger.Error("failed to insert registed histories", zap.Error(err))
			return err
		}

		m.Logger.Info("record post mawinter history complete")
	}

	m.Logger.Info("Regist end")
	return nil
}
