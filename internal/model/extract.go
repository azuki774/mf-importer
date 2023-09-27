package model

import (
	"fmt"
)

type ExtractCondition int

type ExtractRule struct {
	FromName        map[string]int // Name -> CategoryID（完全一致）
	FromNameIn      map[string]int // Name -> CategoryID（部分一致）
	FromMCategory   map[string]int // MCategory -> CategoryID（完全一致）
	FromMCategoryIn map[string]int // MCategory -> CategoryID（部分一致）
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

func (e *ExtractRule) AddRule(ers []ExtractRuleDB) (err error) {
	const fieldName = "name"
	const fieldMCategory = "m_category"

	for _, er := range ers {
		if er.FieldName == fieldName {
			if er.ExactMatch == 1 {
				// 完全一致
				e.FromName[er.Value] = int(er.CategoryID)
			} else {
				// 部分一致
				e.FromNameIn[er.Value] = int(er.CategoryID)
			}
		} else if er.FieldName == fieldMCategory {
			if er.ExactMatch == 1 {
				// 完全一致
				e.FromMCategory[er.Value] = int(er.CategoryID)
			} else {
				// 部分一致
				e.FromMCategoryIn[er.Value] = int(er.CategoryID)
			}
		} else {
			return fmt.Errorf("unknown field name")
		}
	}

	return nil
}
