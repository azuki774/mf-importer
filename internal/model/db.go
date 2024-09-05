package model

import (
	"fmt"
	"mf-importer/internal/openapi"
	"mf-importer/internal/util"
	"strconv"
	"strings"
	"time"
)

// Detail: DB 兼 ユースケース用モデル
type Detail struct {
	ID            int64      `json:"id"`
	YYYYMMID      int        `json:"yyyymm_id"`
	Date          time.Time  `json:"date"`
	Name          string     `json:"name"`
	Price         int64      `json:"price"`
	FinIns        string     `json:"fin_ins"`
	LCategory     string     `json:"l_category"`
	MCategory     string     `json:"m_category"`
	RegistDate    time.Time  `json:"regist_date"`
	MawCheckDate  *time.Time `json:"maw_check_date"`
	MawRegistDate *time.Time `json:"maw_regist_date"`
	RawDate       string     `json:"raw_date"`
	RawPrice      string     `json:"raw_price"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// DB操作用
type ExtractRuleDB struct {
	ID         int64     `json:"id"`
	FieldName  string    `json:"field_name"`
	Value      string    `json:"value"`
	ExactMatch int64     `json:"exact_match"`
	CategoryID int64     `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ImportHistory struct {
	JobLabel       string    `gorm:"job_label"`
	ParsedEntryNum int64     `gorm:"parsed_entry_num"`
	NewEntryNum    int64     `gorm:"new_entry_num"`
	SrcFile        string    `gorm:"src_file"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ex: "2024/08/18"
func getDateFromCSV(rawDate string) (date time.Time, err error) {
	date, err = time.ParseInLocation("2006/01/02", rawDate, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

func getPriceFromCSV(rawPrice string) (price int64, err error) {
	// , - " を消す
	rawPrice = strings.ReplaceAll(rawPrice, ",", "")
	rawPrice = strings.ReplaceAll(rawPrice, "-", "")
	rawPrice = strings.ReplaceAll(rawPrice, `"`, "")
	price32, err := strconv.Atoi(rawPrice)
	if err != nil {
		return 0, err
	}
	return int64(price32), nil
}

func ConvCSVtoDetail(csv [][]string) (details []Detail, err error) {
	// CSV
	// 計算対象,日付,内容,金額（円）,保有金融機関,大項目,中項目,メモ,振替,ID
	// ,07/16(火),ローソン,-291,三井住友カード,食費,食料品,,,
	for i, row := range csv {
		if len(row) != 10 {
			return []Detail{}, fmt.Errorf("invalid csv: row: %d", i)
		}

		// ヘッダ行はあれば外す
		if row[0] == "計算対象" {
			continue
		}
		row := Detail{
			//// ID            int64     `json:"id"`
			YYYYMMID: len(csv) - i, // YYYYMM_id は逆順につける（日付の古い順）
			// Date          time.Time `json:"date"`
			Name: row[2],
			// Price         int64     `json:"price"`
			FinIns:     row[4],
			LCategory:  row[5],
			MCategory:  row[6],
			RegistDate: util.NowFunc(),
			//// MawCheckDate  time.Time `json:"maw_check_date"`
			//// MawRegistDate time.Time `json:"maw_regist_date"`
			RawDate:  row[1],
			RawPrice: row[3],
			//// CreatedAt     time.Time `json:"created_at"`
			//// UpdatedAt     time.Time `json:"updated_at"`
		}

		row.Date, err = getDateFromCSV(row.RawDate)
		if err != nil {
			return []Detail{}, fmt.Errorf("failed to convert date: %s", row.Name)
		}

		row.Price, err = getPriceFromCSV(row.RawPrice)
		if err != nil {
			return []Detail{}, fmt.Errorf("failed to convert price: %s", row.Name)
		}

		details = append(details, row)
	}
	return details, nil
}

func (e *ExtractRuleDB) ToExtractRule() openapi.Rule {
	return openapi.Rule{
		CategoryId: int(e.CategoryID),
		ExactMatch: int(e.ExactMatch),
		FieldName:  e.FieldName,
		Id:         int(e.ID),
		Value:      e.Value,
	}
}
