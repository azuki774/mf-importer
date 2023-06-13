package model

import (
	"fmt"
	"time"
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

type CFRecords struct {
	ID         int       `bson:"_id"`
	RegistID   int       `bson:"regist_id"`
	RegistDate string    `bson:"regist_date"`
	Date       time.Time `bson:"date"`
	Name       string    `bson:"name"`
	Price      string    `bson:"price"`
	FinIns     string    `bson:"fin_ins"`
	LCategory  string    `bson:"l_category"`
	MCategory  string    `bson:"m_category"`
	// _id: ObjectId("64858d14543b902bfaf7b43a"),
	// regist_id: 80,
	// regist_date: '20230611',
	// date: '05/28(日)',
	// name: 'まいばすけっと',
	// price: '-1,494',
	// fin_ins: 'JCBカード',
	// l_category: '食費',
	// m_category: '食料品'
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
