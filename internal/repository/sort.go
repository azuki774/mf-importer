package repository

// API ソートキー -> DB カラム名のホワイトリスト(SQL インジェクション対策)
var detailSortColumns = map[string]string{
	"useDate":         "date",
	"name":            "name",
	"price":           "price",
	"registDate":      "regist_date",
	"importJudgeDate": "maw_check_date",
	"importDate":      "maw_regist_date",
}

var ruleSortColumns = map[string]string{
	"id":         "ID",
	"fieldName":  "field_name",
	"value":      "value",
	"exactMatch": "exact_match",
	"categoryId": "category_id",
}

const (
	detailDefaultSort = "useDate"
	detailDefaultDir  = "DESC"
	ruleDefaultSort   = "id"
	ruleDefaultDir    = "ASC"
)

// buildOrderClause: ホワイトリスト照合の上、第2ソートキー(ID)付き ORDER BY 句を返す。
// 不明な sort は各テーブルのデフォルト句へ、不明な order は昇順へフォールバックする。
func buildOrderClause(columns map[string]string, defSort string, defDir string, sort string, order string) string {
	col, ok := columns[sort]
	if !ok {
		col = columns[defSort]
		if col == "ID" {
			return col + " " + defDir
		}
		return col + " " + defDir + ", ID " + defDir
	}

	dir := "ASC"
	if order == "desc" {
		dir = "DESC"
	}
	if col == "ID" {
		return col + " " + dir
	}
	return col + " " + dir + ", ID " + dir
}

func buildDetailOrderClause(sort string, order string) string {
	return buildOrderClause(detailSortColumns, detailDefaultSort, detailDefaultDir, sort, order)
}

func buildRuleOrderClause(sort string, order string) string {
	return buildOrderClause(ruleSortColumns, ruleDefaultSort, ruleDefaultDir, sort, order)
}
