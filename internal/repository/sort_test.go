package repository

import "testing"

func TestBuildDetailOrderClause(t *testing.T) {
	tests := []struct {
		name  string
		sort  string
		order string
		want  string
	}{
		{name: "useDate desc", sort: "useDate", order: "desc", want: "date DESC, ID DESC"},
		{name: "price asc", sort: "price", order: "asc", want: "price ASC, ID ASC"},
		{name: "registDate maps regist_date", sort: "registDate", order: "asc", want: "regist_date ASC, ID ASC"},
		{name: "importJudgeDate maps maw_check_date", sort: "importJudgeDate", order: "desc", want: "maw_check_date DESC, ID DESC"},
		{name: "importDate maps maw_regist_date", sort: "importDate", order: "asc", want: "maw_regist_date ASC, ID ASC"},
		{name: "name asc", sort: "name", order: "asc", want: "name ASC, ID ASC"},
		{name: "unknown sort falls back to default", sort: "evil;drop", order: "desc", want: "date DESC, ID DESC"},
		{name: "unknown order falls back to asc", sort: "name", order: "sideways", want: "name ASC, ID ASC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildDetailOrderClause(tt.sort, tt.order); got != tt.want {
				t.Errorf("buildDetailOrderClause(%q, %q) = %v, want %v", tt.sort, tt.order, got, tt.want)
			}
		})
	}
}

func TestBuildRuleOrderClause(t *testing.T) {
	tests := []struct {
		name  string
		sort  string
		order string
		want  string
	}{
		{name: "id asc single column", sort: "id", order: "asc", want: "ID ASC"},
		{name: "id desc single column", sort: "id", order: "desc", want: "ID DESC"},
		{name: "fieldName maps field_name", sort: "fieldName", order: "asc", want: "field_name ASC, ID ASC"},
		{name: "value desc", sort: "value", order: "desc", want: "value DESC, ID DESC"},
		{name: "exactMatch maps exact_match", sort: "exactMatch", order: "asc", want: "exact_match ASC, ID ASC"},
		{name: "categoryId maps category_id", sort: "categoryId", order: "desc", want: "category_id DESC, ID DESC"},
		{name: "unknown sort falls back to default", sort: "evil;drop", order: "desc", want: "ID ASC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRuleOrderClause(tt.sort, tt.order); got != tt.want {
				t.Errorf("buildRuleOrderClause(%q, %q) = %v, want %v", tt.sort, tt.order, got, tt.want)
			}
		})
	}
}
