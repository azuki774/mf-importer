package model

type CreateRecord struct {
	CategoryID int64  `json:"category_id"`
	Date       string `json:"date"` // YYYYMMDD
	Price      int64  `json:"price"`
	From       string `json:"from"`
	Type       string `json:"type"`
	Memo       string `json:"memo"`
}
