package model

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExtractCondition int

type ExtractRule struct {
	FromName        map[string]int // Name -> CategoryID（完全一致）
	FromNameIn      map[string]int // Name -> CategoryID（部分一致）
	FromMCategory   map[string]int // MCategory -> CategoryID（完全一致）
	FromMCategoryIn map[string]int // MCategory -> CategoryID（部分一致）
}

// ExtractRuleCSV: extract_rule.csv を読み込む構造体（1レコード分）
type ExtractRuleCSV struct {
	FieldName        string
	Name             string
	ExtractCondition bool // 完全一致かどうか
	CategoryID       int  // 変換先の category_id
}

type CFRecord struct {
	ID                primitive.ObjectID `bson:"_id"`
	RegistID          int                `bson:"regist_id"`
	YYYYMMDD          string             `bson:"yyyymmdd"`
	Date              string             `bson:"date"` // "06/01（火）"
	Name              string             `bson:"name"`
	Price             string             `bson:"price"`
	LCategory         string             `bson:"l_category"`
	MCategory         string             `bson:"m_category"`
	MawStatusChecked  bool               `bson:"maw_status_checked"`
	MawStatusRegisted bool               `bson:"maw_status_registed"`
	// {
	// 	_id: ObjectId("6489bbe3163254689370aa32"),
	// 	regist_date: '20230614',
	// 	date: '05/21(日)',
	// 	name: 'PAYPAL *GOOGLE YOUTUBE SU',
	// 	price: '-1,180',
	// 	yyyymm_id: 58,
	// 	yyyymmdd: '20230521',
	// 	yyyymm: '202305',
	// 	fin_ins: 'XXX',
	// 	l_category: '通信費',
	// 	m_category: '情報サービス'
	// },
	CategoryID int // mawinter-server 挿入用
}

func NewExtractRule() *ExtractRule {
	e := &ExtractRule{
		FromName:        make(map[string]int, 0),
		FromNameIn:      make(map[string]int, 0),
		FromMCategory:   make(map[string]int, 0),
		FromMCategoryIn: make(map[string]int, 0),
	}
	return e
}

func (e *ExtractRule) AddRule(erc ExtractRuleCSV) (err error) {
	const fieldName = "name"
	const fieldMCategory = "m_category"

	if erc.FieldName == fieldName {
		if erc.ExtractCondition {
			// 完全一致
			e.FromName[erc.Name] = erc.CategoryID
		} else {
			// 部分一致
			e.FromNameIn[erc.Name] = erc.CategoryID
		}
	} else if erc.FieldName == fieldMCategory {
		if erc.ExtractCondition {
			// 完全一致
			e.FromMCategory[erc.Name] = erc.CategoryID
		} else {
			// 部分一致
			e.FromMCategoryIn[erc.Name] = erc.CategoryID
		}
	} else {
		return fmt.Errorf("unknown field name in extract CSV")
	}

	return nil
}
