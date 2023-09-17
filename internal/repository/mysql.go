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

func (m *DBClient) GetCFDetails(ctx context.Context) (cfRecords []model.Detail, err error) {
	// mock
	return []model.Detail{
		{
			ID:           11,
			YYYYMMID:     1,
			Date:         "2023-01-01",
			RawDate:      "01/01（火）",
			Name:         "ふぃーるど１",
			Price:        1234,
			LCategory:    "大分類",
			MCategory:    "中分類",
			MawCheckDate: NULLtimeTime,
		},
		{
			ID:           12,
			YYYYMMID:     2,
			Date:         "2023-01-02",
			RawDate:      "01/02（水）",
			Name:         "ふぃーるど５",
			Price:        1234,
			LCategory:    "大分類",
			MCategory:    "中分類",
			MawCheckDate: NULLtimeTime,
		},
	}, nil
}

func (m *DBClient) CheckCFDetail(ctx context.Context, cfDetail model.Detail) (err error) {
	// mock
	return nil
}

func (m *DBClient) RegistedCFDetail(ctx context.Context, cfDetail model.Detail) (err error) {
	// mock
	return nil
}

func (m *DBClient) GetExtractRules(ctx context.Context) (er []model.ExtractRuleDB, err error) {
	// mock
	return []model.ExtractRuleDB{}, nil // TODO
}
