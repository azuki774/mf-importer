package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const CurrentNrknSchemaVersion = "2026-09-14"
const NrknStatusOK = "OK"

var ErrUnsupportedNrknSchemaVersion = errors.New("unsupported NRKN schema version")

type NrknSnapshot struct {
	ID            int64     `json:"-" gorm:"primaryKey;autoIncrement"`
	SchemaVersion string    `json:"schema_version"`
	FetchedAt     time.Time `json:"fetched_at"`
	Status        string    `json:"status"`
	GrandTotalJPY int64     `json:"grand_total_jpy"`
	TotalCostJPY  int64     `json:"total_cost_jpy"`
	PnlJPY        int64     `json:"pnl_jpy"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

type NrknHolding struct {
	ID                     int64     `json:"-" gorm:"primaryKey;autoIncrement"`
	SnapshotID             int64     `json:"-"`
	ProductCode            string    `json:"product_code"`
	CompositeFIGI          string    `json:"composite_figi"`
	Name                   string    `json:"name"`
	Category               string    `json:"category"`
	Quantity               float64   `json:"quantity"`
	UnitPrice              float64   `json:"unit_price"`
	ValueJPY               int64     `json:"value_jpy"`
	CostJPY                int64     `json:"cost_jpy"`
	RedemptionUnitPrice    float64   `json:"redemption_unit_price"`
	RedemptionValueJPY     int64     `json:"redemption_value_jpy"`
	PnlJPY                 int64     `json:"pnl_jpy"`
	ReferenceDate          string    `json:"reference_date"`
	AllocationPct          float64   `json:"allocation_pct"`
	UnitPriceRaw           string    `json:"unit_price_raw"`
	RedemptionUnitPriceRaw string    `json:"redemption_unit_price_raw"`
	CreatedAt              time.Time `json:"-"`
	UpdatedAt              time.Time `json:"-"`
}

func NormalizeNrknFetchedAt(t time.Time) time.Time {
	return t.In(time.FixedZone("Asia/Tokyo", 9*60*60)).Truncate(time.Microsecond)
}

// ParseNrknJSON preserves displayed totals and prices without recalculating them.
// Decode errors deliberately omit input values, which can contain private data.
func ParseNrknJSON(data []byte) (*NrknSnapshot, []NrknHolding, error) {
	var envelope struct {
		SchemaVersion *string `json:"schema_version"`
	}
	if json.Unmarshal(data, &envelope) != nil || envelope.SchemaVersion == nil {
		return nil, nil, errors.New("invalid NRKN JSON or missing string schema_version")
	}
	if *envelope.SchemaVersion != CurrentNrknSchemaVersion {
		return nil, nil, ErrUnsupportedNrknSchemaVersion
	}
	if err := requireNrknFields(data, "fetched_at", "status", "grand_total_jpy", "total_cost_jpy", "pnl_jpy", "holdings"); err != nil {
		return nil, nil, err
	}
	var dto struct {
		NrknSnapshot
		Holdings []json.RawMessage `json:"holdings"`
	}
	if json.Unmarshal(data, &dto) != nil {
		return nil, nil, errors.New("invalid NRKN snapshot field type or value")
	}
	dto.Status = strings.ToUpper(strings.TrimSpace(dto.Status))
	if dto.Status != NrknStatusOK || dto.FetchedAt.IsZero() || dto.FetchedAt.Year() < 1000 || dto.FetchedAt.Year() > 9999 {
		return nil, nil, errors.New("NRKN requires OK status and a valid fetched_at")
	}
	if len(dto.Holdings) == 0 {
		return nil, nil, errors.New("NRKN holdings must not be empty")
	}
	holdings := make([]NrknHolding, 0, len(dto.Holdings))
	seen := make(map[string]bool)
	for index, raw := range dto.Holdings {
		if err := requireNrknFields(raw, "product_code", "composite_figi", "name", "category", "quantity", "unit_price", "value_jpy", "cost_jpy", "redemption_unit_price", "redemption_value_jpy", "pnl_jpy", "reference_date", "allocation_pct", "unit_price_raw", "redemption_unit_price_raw"); err != nil {
			return nil, nil, fmt.Errorf("holdings[%d]: %w", index, err)
		}
		var h NrknHolding
		if json.Unmarshal(raw, &h) != nil {
			return nil, nil, fmt.Errorf("holdings[%d]: invalid field type or value", index)
		}
		for _, value := range []string{h.ProductCode, h.Name, h.Category, h.UnitPriceRaw, h.RedemptionUnitPriceRaw} {
			if strings.TrimSpace(value) == "" {
				return nil, nil, fmt.Errorf("holdings[%d]: empty required text", index)
			}
		}
		if len(h.ProductCode) > 64 || !validCompositeFIGI(h.CompositeFIGI) {
			return nil, nil, fmt.Errorf("holdings[%d]: invalid product_code or composite_figi", index)
		}
		date, err := time.Parse("2006-01-02", h.ReferenceDate)
		if err != nil || date.Year() < 1000 || h.AllocationPct < 0 || h.AllocationPct > 100 {
			return nil, nil, fmt.Errorf("holdings[%d]: invalid reference_date or allocation_pct", index)
		}
		if seen[h.ProductCode] {
			return nil, nil, fmt.Errorf("holdings[%d]: duplicate product_code", index)
		}
		seen[h.ProductCode] = true
		holdings = append(holdings, h)
	}
	dto.FetchedAt = NormalizeNrknFetchedAt(dto.FetchedAt)
	return &dto.NrknSnapshot, holdings, nil
}

func requireNrknFields(raw []byte, fields ...string) error {
	var values map[string]json.RawMessage
	if json.Unmarshal(raw, &values) != nil {
		return errors.New("invalid NRKN JSON object")
	}
	for _, field := range fields {
		value, ok := values[field]
		if !ok || strings.TrimSpace(string(value)) == "null" {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}
