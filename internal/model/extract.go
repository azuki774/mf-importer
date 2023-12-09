package model

import (
	"fmt"
	"strings"
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

// そのレコードが Suica のものかを判定する
// 内容: '入 XXXX 出 YYYY'
// 保有金融機関: 'モバイルSuica (モバイルSuica ID)'
func IsSuicaDetail(d Detail) (ok bool) {
	const mobileSuicaFinIns = "モバイルSuica (モバイルSuica ID)"
	if d.FinIns != mobileSuicaFinIns {
		return false
	}
	iri_index := strings.Index(d.Name, "入")
	de_index := strings.Index(d.Name, "出")
	if iri_index == -1 || de_index == -1 {
		// 入、出が存在しない場合
		return false
	}
	if iri_index >= de_index {
		// 出xxx入yyyy と順番が逆の場合
		return false
	}
	return true
}
