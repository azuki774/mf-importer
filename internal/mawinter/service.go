package mawinter

import (
	"context"
	"mf-importer/internal/logger"
	"mf-importer/internal/model"
	"strings"
	"time"

	"go.uber.org/zap"
)

const category_train = 270

type DBClient interface {
	GetCFDetails(ctx context.Context) (cfDetails []model.Detail, err error) // mawinter check未チェックのデータを取得
	// CheckCFDetail: レコードを mawinter check/regist 済とする. regist = true なら registフラグも立てる
	CheckCFDetail(ctx context.Context, cfDetail model.Detail, regist bool) (err error)
	GetExtractRules(ctx context.Context) (er []model.ExtractRuleDB, err error)
}
type MawinterClient interface {
	Regist(ctx context.Context, c model.Detail, catID int) (err error)
	GetMawinterWeb(ctx context.Context, yyyymm string) (recs []model.GetRecord, err error) // mawinter-web から登録されたデータをYYYYMM単位で取得
}

const fromMawinterWebText = "mawinter-web"

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

	// 重複判定に利用するため、mawinter から登録済のデータで、from="mawinter-web" なものを取ってくる
	yyyymm := time.Now().Format("200601")
	yyyymmLastMonth := time.Now().AddDate(0, -1, 0).Format("200601")
	mawrecs, err := m.MawClient.GetMawinterWeb(ctx, yyyymm)
	if err != nil {
		m.Logger.Error("failed to records from mawinter", zap.Error(err))
		return err
	}
	mrc2, err := m.MawClient.GetMawinterWeb(ctx, yyyymmLastMonth)
	if err != nil {
		m.Logger.Error("failed to records from mawinter", zap.Error(err))
		return err
	}
	mawrecs = append(mawrecs, mrc2...)
	m.Logger.Info(
		"fetch records from mawinter (from = mawinter-web, this month and lastmonth)",
		zap.Int("num", len(mawrecs)),
	)

	for _, c := range cfDetails {
		cLogger := m.Logger.With(
			zap.Time("Date", c.Date),
			zap.String("Name", c.Name),
			zap.Int64("Price", c.Price),
			zap.String("M_Category", c.MCategory),
		)

		// catID を取得する。ない場合は抽出対象外なので mawinter に post しない
		catID, regist := m.getCategoryIDwithExtractCond(c)

		// suica は別方法で抽出する
		if !regist {
			isSuica := model.IsSuicaDetail(c)
			if isSuica {
				catID = category_train
				regist = true
			}
		}

		// DryRunMode ONのときは何もせず終了
		if m.Dryrun {
			cLogger.Info("dryrun: post to mawinter", zap.Bool("regist", regist))
			continue
		}

		if regist {
			// 既に mawinter-web に登録済のデータでなければ送信する
			if !judgeAlreadyRegisted(c, mawrecs) {
				cLogger.Info("post to mawinter")
				err := m.MawClient.Regist(ctx, c, catID)
				if err != nil {
					m.Logger.Error("failed to post mawinter", zap.Error(err))
					return err
				}
				cLogger.Info("post to mawinter complete")
			} else {
				// 既に登録済の場合
				// detail テーブル更新の際には、maw_regist_date には日付を入れないので regist フラグを下ろす
				cLogger.Warn("this record may be already registed", zap.Int64("ID", c.ID), zap.Time("Date", c.Date))
				regist = false
			}

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

// これから登録しようとしているデータがリスト内にあるかを確認する
// 判定条件: 日時の yyyymmdd が一致 && price が一致
func judgeAlreadyRegisted(dr model.Detail, alrecs []model.GetRecord) (duplicate bool) {
	for _, r := range alrecs {
		// 金額が違う場合はそのレコードは false
		if r.Price != int(dr.Price) {
			continue
		}

		// YYYYMMDD が違う場合はそのレコードは false
		if r.Datetime.Format("20060102") != dr.Date.Format("20060102") {
			continue
		}

		// レコード一致判定
		return true
	}
	return false
}
