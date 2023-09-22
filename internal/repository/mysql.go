package repository

import (
	"context"
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
	// TODO: regist フラグを見るようにする
	result := d.Conn.Table("detail").Where("ID = ?", id).Update("maw_check_date", time.Now())
	if result.Error != nil {
		return result.Error
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
