package model

import "time"

const mfImportFromStr = "mf-importer"

type CreateRecord struct {
	CategoryID int64  `json:"category_id"`
	Date       string `json:"datetime"` // YYYYMMDD
	Price      int64  `json:"price"`
	From       string `json:"from"`
	Type       string `json:"type"`
	Memo       string `json:"memo"`
}

type GetRecord struct {
	ID           int       `json:"id"`
	CategoryID   int       `json:"category_id"`
	CategoryName string    `json:"category_name"`
	Datetime     time.Time `json:"datetime"`
	From         string    `json:"from"`
	Type         string    `json:"type"`
	Price        int       `json:"price"`
	Memo         string    `json:"memo"`
}

func NewCreateRecord(c Detail, catID int) (r CreateRecord, err error) {
	r = CreateRecord{
		CategoryID: int64(catID),
		Date:       c.Date.Format("20060102"),
		Price:      int64(c.Price),
		From:       mfImportFromStr,
		Memo:       c.Name, // メモ欄に元々の名前を入れる
	}
	return r, nil
}
