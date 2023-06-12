package model

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
