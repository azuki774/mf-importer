package model

import "strings"

const mfImportFromStr = "mf-importer"

type CreateRecord struct {
	CategoryID int64  `json:"category_id"`
	Date       string `json:"datetime"` // YYYYMMDD
	Price      int64  `json:"price"`
	From       string `json:"from"`
	Type       string `json:"type"`
	Memo       string `json:"memo"`
}

func NewCreateRecord(c Detail, catID int) (r CreateRecord, err error) {
	r = CreateRecord{
		CategoryID: int64(catID),
		Date:       strings.Replace(c.Date, "-", "", -1),
		Price:      int64(c.Price),
		From:       mfImportFromStr,
		Memo:       c.Name, // メモ欄に元々の名前を入れる
	}
	return r, nil
}
