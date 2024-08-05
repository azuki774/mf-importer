package repository

import (
	"context"
	"errors"
	"mf-importer/internal/model"
	"net"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const DBConnectRetry = 5
const DBConnectRetryInterval = 5

var NULLtimeTime = time.Time{} // mock

type DBClient struct {
	Conn *gorm.DB
}

func NewDBRepository(host, port, user, pass, name string) (dbR *DBClient, err error) {
	addr := net.JoinHostPort(host, port)
	dsn := user + ":" + pass + "@(" + addr + ")/" + name + "?parseTime=true&loc=Local"
	var gormdb *gorm.DB
	for i := 0; i < DBConnectRetry; i++ {
		gormdb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			// Success DB connect
			break
		}
		if i == DBConnectRetry {
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
	result := d.Conn.Table("detail").Where("maw_check_date IS NULL").Find(&cfRecords)
	if result.Error != nil {
		return []model.Detail{}, result.Error
	}
	return cfRecords, nil
}

func (d *DBClient) CheckCFDetail(ctx context.Context, cfDetail model.Detail, regist bool) (err error) {
	id := cfDetail.ID
	t := time.Now().Format("2006-01-02")
	err = d.Conn.Transaction(func(tx *gorm.DB) error {
		result := tx.Table("detail").Where("ID = ?", id).Update("maw_check_date", t)
		if result.Error != nil {
			return result.Error
		}

		if regist {
			result := tx.Table("detail").Where("ID = ?", id).Update("maw_regist_date", t)
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
	result := d.Conn.Table("extract_rule").Find(&er)
	if result.Error != nil {
		return []model.ExtractRuleDB{}, result.Error
	}
	return er, nil
}

// yyyymmdd と name と price がすべて一致するものを抽出する（登録済判断に利用）
// すでに登録があれば true とする
func (d *DBClient) CheckAlreadyRegistDetail(ctx context.Context, detail model.Detail) (exists bool, err error) {
	var getDetail model.Detail
	if err = d.Conn.WithContext(ctx).Table("detail").
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
	return d.Conn.WithContext(ctx).Table("detail").Create(&detail).Error
}

func (d *DBClient) RegistDetailHistory(ctx context.Context, jobname string, parsedNum int, insertNum int) (err error) {
	importHis := model.ImportHistory{
		JobLabel:       jobname,
		ParsedEntryNum: int64(parsedNum),
		NewEntryNum:    int64(insertNum),
	}
	return d.Conn.WithContext(ctx).Table("import_history").Create(&importHis).Error
}
