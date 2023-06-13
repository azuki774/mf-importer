package model

import "fmt"

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
