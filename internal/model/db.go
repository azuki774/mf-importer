package model

import "time"

// Detail: DB 兼 ユースケース用モデル
type Detail struct {
	ID            int64     `json:"id"`
	YYYYMMID      int       `json:"yyyymm_id"`
	Date          time.Time `json:"date"`
	Name          string    `json:"name"`
	Price         int64     `json:"price"`
	FinIns        string    `json:"fin_ins"`
	LCategory     string    `json:"l_category"`
	MCategory     string    `json:"m_category"`
	RegistDate    time.Time `json:"regist_date"`
	MawCheckDate  time.Time `json:"maw_check_date"`
	MawRegistDate time.Time `json:"maw_regist_date"`
	RawDate       string    `json:"raw_date"`
	RawPrice      string    `json:"raw_price"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ExtractRuleDB struct {
	ID         int64     `json:"id"`
	FieldName  string    `json:"field_name"`
	Value      string    `json:"value"`
	ExactMatch int64     `json:"exact_match"`
	CategoryID int64     `json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
