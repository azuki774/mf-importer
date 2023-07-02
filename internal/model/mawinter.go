package model

import (
	"strconv"
	"strings"
)

const mfImportFromStr = "mf-importer"

type CreateRecord struct {
	CategoryID int64  `json:"category_id"`
	Date       string `json:"datetime"` // YYYYMMDD
	Price      int64  `json:"price"`
	From       string `json:"from"`
	Type       string `json:"type"`
	Memo       string `json:"memo"`
}

func convPriceForm(orig string) (price int, err error) {
	// '-1,180' -> '1180'
	orig = strings.Replace(orig, ",", "", -1)
	orig = strings.Replace(orig, "-", "", -1)
	price, err = strconv.Atoi(orig)
	return price, err
}

func NewCreateRecord(c CFRecord) (r CreateRecord, err error) {
	// yyyymmdd: '20230521',
	// price: '-1,180',
	// CategoryID: 100
	price, err := convPriceForm(c.Price)
	if err != nil {
		return CreateRecord{}, err
	}

	r = CreateRecord{
		CategoryID: int64(c.CategoryID),
		Date:       c.YYYYMMDD,
		Price:      int64(price),
		From:       mfImportFromStr,
		Memo:       c.Name, // メモ欄に元々の名前を入れる
	}

	return r, nil
}
