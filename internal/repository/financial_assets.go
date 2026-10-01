package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mf-importer/internal/model"

	"gorm.io/gorm"
)

// Explicit table mapping prevents user-provided source names from becoming SQL identifiers.
func financialTables(source string) (string, string, bool) {
	switch source {
	case "nrkn":
		return "nrkn_snapshot", "nrkn_holding", true
	case "sbi":
		return "sbi_snapshot", "sbi_holding", true
	default:
		return "", "", false
	}
}

// These read-only rows intentionally differ from the legacy ingestion models:
// those use float64, which can lose SQL DECIMAL precision before JSON encoding.
// Reading DECIMAL as text preserves its stored value for the API's decimal-string
// fields without changing the existing ingestion behavior or models.
type financialSnapshotRow struct {
	ID            int64     `gorm:"column:id"`
	Source        string    `gorm:"column:source"`
	FetchedAt     time.Time `gorm:"column:fetched_at"`
	ImportedEpoch *int64    `gorm:"column:imported_epoch"`
	Valuation     *string   `gorm:"column:valuation"`
	Cost          *string   `gorm:"column:cost"`
	Pnl           *string   `gorm:"column:pnl"`
}

type financialHoldingRow struct {
	ID, SnapshotID                                     int64
	Name                                               string
	Section, ProductCode, CompositeFigi, ReferenceDate *string
	Quantity, Valuation, Cost, Pnl                     *string
	QuantityFloat                                      *float64
}

func snapshotSelect(source string) string {
	if source == "nrkn" {
		return "id, fetched_at, CAST(UNIX_TIMESTAMP(created_at) AS SIGNED) AS imported_epoch, CAST(grand_total_jpy AS CHAR) AS valuation, CAST(total_cost_jpy AS CHAR) AS cost, CAST(pnl_jpy AS CHAR) AS pnl"
	}
	return "id, fetched_at, CAST(UNIX_TIMESTAMP(created_at) AS SIGNED) AS imported_epoch, CAST(grand_total_jpy AS CHAR) AS valuation, NULL AS cost, NULL AS pnl"
}

func snapshotModel(row financialSnapshotRow, source string) model.FinancialSnapshot {
	result := model.FinancialSnapshot{ID: row.ID, Source: source, FetchedAt: row.FetchedAt,
		ValuationJpy: row.Valuation, CostJpy: row.Cost, UnrealizedPnlJpy: row.Pnl, Holdings: []model.FinancialHolding{}}
	if row.ImportedEpoch != nil {
		result.ImportedAt = time.Unix(*row.ImportedEpoch, 0).UTC()
	}
	return result
}

func (d *DBClient) ListFinancialSnapshots(ctx context.Context, query model.FinancialSnapshotQuery) ([]model.FinancialSnapshot, error) {
	if query.Limit < 0 || query.Offset < 0 || query.Limit == 0 && query.Offset != 0 {
		return nil, fmt.Errorf("list financial snapshots: invalid pagination")
	}
	if len(query.Sources) == 0 {
		return nil, fmt.Errorf("list financial snapshots: missing sources")
	}
	var selects []string
	var args []any
	seen := map[string]bool{}
	for _, source := range query.Sources {
		snapTable, _, ok := financialTables(source)
		if !ok || seen[source] {
			return nil, fmt.Errorf("list financial snapshots: invalid sources")
		}
		seen[source] = true
		// Both the table and source literal are restricted by financialTables.
		part := "SELECT " + snapshotSelect(source) + ", '" + source + "' AS source FROM " + snapTable + " WHERE status = ?"
		args = append(args, "OK")
		if query.From != nil {
			part += " AND fetched_at >= ?"
			args = append(args, jstBound(*query.From, true))
		}
		if query.To != nil {
			part += " AND fetched_at < ?"
			args = append(args, jstBound(*query.To, true))
		}
		selects = append(selects, part)
	}
	// The global offset belongs outside UNION ALL. Applying it separately to
	// each source would skip records and break pages spanning both sources.
	statement := "SELECT * FROM (" + strings.Join(selects, " UNION ALL ") + ") AS snapshots ORDER BY fetched_at DESC, source DESC, id DESC"
	if query.Limit > 0 {
		statement += " LIMIT ? OFFSET ?"
		args = append(args, query.Limit, query.Offset)
	}
	var rows []financialSnapshotRow
	if err := d.Conn.WithContext(ctx).Raw(statement, args...).Scan(&rows).Error; err != nil {
		return nil, dbReadError("list financial snapshots", err)
	}
	result := make([]model.FinancialSnapshot, 0, len(rows))
	for _, row := range rows {
		result = append(result, snapshotModel(row, row.Source))
	}
	return result, nil
}

func (d *DBClient) GetFinancialSnapshot(ctx context.Context, source string, id int64) (model.FinancialSnapshot, error) {
	table, _, ok := financialTables(source)
	if !ok {
		return model.FinancialSnapshot{}, fmt.Errorf("get financial snapshot: unsupported source")
	}
	var row financialSnapshotRow
	err := d.Conn.WithContext(ctx).Table(table).Select(snapshotSelect(source)).Where("id = ? AND status = ?", id, "OK").Scan(&row).Error
	if err != nil {
		return model.FinancialSnapshot{}, dbReadError("get financial snapshot", err)
	}
	if row.ID == 0 {
		return model.FinancialSnapshot{}, model.ErrRecordNotFound
	}
	return snapshotModel(row, source), nil
}

func (d *DBClient) LatestFinancialSnapshot(ctx context.Context, source string, at time.Time, inclusive bool) (*model.FinancialSnapshot, error) {
	table, _, ok := financialTables(source)
	if !ok {
		return nil, fmt.Errorf("latest financial snapshot: unsupported source")
	}
	op := "<"
	if inclusive {
		op = "<="
	}
	var row financialSnapshotRow
	db := d.Conn.WithContext(ctx).Table(table).Select(snapshotSelect(source)).Where("status = ?", "OK").Where("fetched_at "+op+" ?", jstBound(at, !inclusive)).Order("fetched_at DESC, id DESC").Limit(1)
	if err := db.Scan(&row).Error; err != nil {
		return nil, dbReadError("latest financial snapshot", err)
	}
	if row.ID == 0 {
		return nil, nil
	}
	result := snapshotModel(row, source)
	return &result, nil
}

func (d *DBClient) GetFinancialHoldings(ctx context.Context, source string, ids []int64) ([]model.FinancialHolding, error) {
	_, table, ok := financialTables(source)
	if !ok {
		return nil, fmt.Errorf("get financial holdings: unsupported source")
	}
	if len(ids) == 0 {
		return []model.FinancialHolding{}, nil
	}
	selectSQL := "id, snapshot_id, name, quantity AS quantity_float, CAST(value_jpy AS CHAR) AS valuation, CAST(cost_jpy AS CHAR) AS cost, CAST(pnl_jpy AS CHAR) AS pnl, NULL AS section, product_code, composite_figi, CAST(reference_date AS CHAR) AS reference_date"
	if source == "sbi" {
		selectSQL = "id, snapshot_id, name, CAST(quantity AS CHAR) AS quantity, CAST(value_jpy AS CHAR) AS valuation, NULL AS cost, CAST(pnl_jpy AS CHAR) AS pnl, section, NULL AS product_code, composite_figi, NULL AS reference_date"
	}
	var rows []financialHoldingRow
	err := d.Conn.WithContext(ctx).Table(table).Select(selectSQL).Where("snapshot_id IN ?", ids).Order("snapshot_id ASC, id ASC").Scan(&rows).Error
	if err != nil {
		return nil, dbReadError("get financial holdings", err)
	}
	result := make([]model.FinancialHolding, 0, len(rows))
	for _, row := range rows {
		if source == "nrkn" && row.QuantityFloat != nil {
			// DOUBLE has already been rounded on storage. Preserve that float value
			// with a round-trippable non-exponent representation, not a DECIMAL cast
			// that would silently truncate very small quantities.
			value := strconv.FormatFloat(*row.QuantityFloat, 'f', -1, 64)
			row.Quantity = &value
		}
		result = append(result, model.FinancialHolding{ID: row.ID, SnapshotID: row.SnapshotID, Name: row.Name, Section: row.Section, ProductCode: row.ProductCode, CompositeFigi: row.CompositeFigi, ReferenceDate: row.ReferenceDate, Quantity: row.Quantity, ValuationJpy: row.Valuation, CostJpy: row.Cost, UnrealizedPnlJpy: row.Pnl})
	}
	return result, nil
}

// Convert instants to JST wall time. DATETIME(6) truncates, so ceil boundaries
// for >= and < preserve nanosecond semantics; > and <= use floor instead.
func jstBound(value time.Time, ceil bool) time.Time {
	t := value.In(time.FixedZone("Asia/Tokyo", 9*60*60))
	floor := t.Truncate(time.Microsecond)
	if ceil && !floor.Equal(t) {
		return floor.Add(time.Microsecond)
	}
	return floor
}

func dbReadError(operation string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrRecordNotFound
	}
	return &financialReadError{operation: operation, cause: err}
}

type financialReadError struct {
	operation string
	cause     error
}

func (e *financialReadError) Error() string { return e.operation + " failed" }
func (e *financialReadError) Unwrap() error { return e.cause }
