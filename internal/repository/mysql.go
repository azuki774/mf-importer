package repository

import (
	"context"
	"errors"
	"mf-importer/internal/model"
	"mf-importer/internal/openapi"
	"net"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const DBConnectRetry = 5
const DBConnectRetryInterval = 5

var NULLtimeTime = time.Time{} // mock

const tableNameDetail = "detail"
const tableNameExtractRule = "extract_rule"
const tableNameImportHistory = "import_history"

type DBClient struct {
	Conn *gorm.DB
}

func NewDBRepository(host, port, user, pass, name string) (dbR *DBClient, err error) {
	addr := net.JoinHostPort(host, port)
	dsn := user + ":" + pass + "@(" + addr + ")/" + name + "?parseTime=true&loc=Local"
	var gormdb *gorm.DB

	for i := 0; i < DBConnectRetry; i++ {
		gormdb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err == nil {
			// Success DB connect
			break
		}
		if i == (DBConnectRetry - 1) {
			return nil, err
		}

		time.Sleep(DBConnectRetryInterval * time.Second)
	}

	return &DBClient{Conn: gormdb}, nil
}
func (d *DBClient) CloseDB() (err error) {
	dbconn, err := d.Conn.DB()
	if err != nil {
		return err
	}
	return dbconn.Close()
}

func (d *DBClient) GetCFDetails(ctx context.Context) (cfRecords []model.Detail, err error) {
	result := d.Conn.Table(tableNameDetail).Where("maw_check_date IS NULL").Find(&cfRecords)
	if result.Error != nil {
		return []model.Detail{}, result.Error
	}
	return cfRecords, nil
}

func (d *DBClient) CheckCFDetail(ctx context.Context, cfDetail model.Detail, regist bool) (err error) {
	id := cfDetail.ID
	t := time.Now().Format("2006-01-02")
	err = d.Conn.Transaction(func(tx *gorm.DB) error {
		result := tx.Table(tableNameDetail).Where("ID = ?", id).Update("maw_check_date", t)
		if result.Error != nil {
			return result.Error
		}

		if regist {
			result := tx.Table(tableNameDetail).Where("ID = ?", id).Update("maw_regist_date", t)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
func (d *DBClient) GetExtractRules(ctx context.Context) (er []model.ExtractRuleDB, err error) {
	result := d.Conn.Table(tableNameExtractRule).Find(&er)
	if result.Error != nil {
		return []model.ExtractRuleDB{}, result.Error
	}
	return er, nil
}

func (d *DBClient) GetExtractRule(ctx context.Context, id int) (er model.ExtractRuleDB, err error) {
	result := d.Conn.Table(tableNameExtractRule).Where("ID = ?", id).First(&er)
	if result.Error != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ExtractRuleDB{}, model.ErrRecordNotFound
		}
		return model.ExtractRuleDB{}, result.Error
	}
	return er, nil
}

// yyyymmdd と name と price がすべて一致するものを抽出する（登録済判断に利用）
// すでに登録があれば true とする
func (d *DBClient) CheckAlreadyRegistDetail(ctx context.Context, detail model.Detail) (exists bool, err error) {
	var getDetail model.Detail
	if err = d.Conn.WithContext(ctx).Table(tableNameDetail).
		Where("date = ?", detail.Date.Format("2006-01-02")). // DB には日付しか登録しないため変形
		Where("name = ?", detail.Name).
		Where("price = ?", detail.Price).
		First(&getDetail).Error; err != nil {
		// 何らかのエラー or データなし
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// データなし
			return false, nil
		} else {
			// 何らかのエラー
			return false, err
		}
	}
	// データあり
	return true, nil
}
func (d *DBClient) RegistDetail(ctx context.Context, detail model.Detail) (err error) {
	return d.Conn.WithContext(ctx).Table(tableNameDetail).Create(&detail).Error
}

func (d *DBClient) RegistDetailHistory(ctx context.Context, jobname string, parsedNum int, insertNum int, srcFile string) (err error) {
	importHis := model.ImportHistory{
		JobLabel:       jobname,
		ParsedEntryNum: int64(parsedNum),
		NewEntryNum:    int64(insertNum),
		SrcFile:        srcFile,
	}
	return d.Conn.WithContext(ctx).Table(tableNameImportHistory).Create(&importHis).Error
}

func (d *DBClient) GetLastDetailHistoryWhereSrcFile(ctx context.Context, srcFile string) (parsedNum, insertedNum int, err error) {
	var ih model.ImportHistory
	// src_file が一致する最新のデータ1件
	err = d.Conn.WithContext(ctx).Table(tableNameImportHistory).Where("src_file = ?", srcFile).Order("ID desc").First(&ih).Error
	if err != nil {
		return 0, 0, err
	}
	return int(ih.ParsedEntryNum), int(ih.NewEntryNum), nil
}

func (d *DBClient) GetLastDetailHistoryWhereJobLabel(ctx context.Context, jobLabel string) (model.ImportHistory, error) {
	var ih model.ImportHistory
	// job_label が一致する最新のデータ1件
	err := d.Conn.WithContext(ctx).Table(tableNameImportHistory).Where("job_label = ?", jobLabel).Order("ID desc").First(&ih).Error
	if err != nil {
		return model.ImportHistory{}, err
	}
	return ih, nil
}

func (d *DBClient) GetDetails(ctx context.Context, limit int) (details []model.Detail, err error) {
	err = d.Conn.WithContext(ctx).Table(tableNameDetail).Order("ID desc").Limit(limit).Find(&details).Error
	return details, err
}

func (d *DBClient) ResetImportDetails(ctx context.Context, id int) (err error) {
	err = d.Conn.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Table(tableNameDetail).Where("id = ?", id).Update("maw_check_date", gorm.Expr("NULL")).Update("maw_regist_date", gorm.Expr("NULL")).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Table(tableNameDetail).Where("id = ?", id).Update("maw_regist_date", gorm.Expr("NULL")).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

func (d *DBClient) AddExtractRule(ctx context.Context, rule openapi.RuleRequest) (ruleDB model.ExtractRuleDB, err error) {
	ruleDB = model.ExtractRuleDB{
		// ID
		FieldName:  rule.FieldName,
		Value:      rule.Value,
		ExactMatch: int64(rule.ExactMatch),
		CategoryID: int64(rule.CategoryId),
		// CreatedAt
		// UpdatedAt
	}

	result := d.Conn.Table(tableNameExtractRule).Create(&ruleDB)
	if result.Error != nil {
		return model.ExtractRuleDB{}, result.Error
	}
	return ruleDB, nil
}

func (d *DBClient) DeleteExtractRule(ctx context.Context, id int) (err error) {
	result := d.Conn.Table(tableNameExtractRule).Delete(&model.ExtractRuleDB{}, id)
	if result.Error != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ErrRecordNotFound
		}
		return result.Error
	}
	return nil
}
