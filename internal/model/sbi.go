package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type SbiStatus string

const (
	SbiStatusOK          SbiStatus = "OK"
	SbiStatusMaintenance SbiStatus = "MAINTENANCE"
	SbiStatusError       SbiStatus = "ERROR"
)

type SbiSnapshot struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	FetchedAt     time.Time `json:"fetched_at" gorm:"uniqueIndex:uq_fetched_at"`
	Status        SbiStatus `json:"status"`
	SchemaVersion int       `json:"schema_version"`

	GrandTotalJPY float64 `json:"grand_total_jpy"`

	NisaTotalJPY     float64 `json:"nisa_total_jpy"`
	NisaPrevDayJPY   float64 `json:"nisa_prev_day_jpy"`
	NisaPrevDayPct   float64 `json:"nisa_prev_day_pct"`
	NisaPrevMonthJPY float64 `json:"nisa_prev_month_jpy"`
	NisaPrevMonthPct float64 `json:"nisa_prev_month_pct"`
	NisaPnlJPY       float64 `json:"nisa_pnl_jpy"`
	NisaPnlPct       float64 `json:"nisa_pnl_pct"`

	NisaDomesticValueJPY     float64 `json:"nisa_domestic_value_jpy"`
	NisaDomesticPnlJPY       float64 `json:"nisa_domestic_pnl_jpy"`
	NisaDomesticPnlPct       float64 `json:"nisa_domestic_pnl_pct"`
	NisaDomesticPrevDayJPY   float64 `json:"nisa_domestic_prev_day_jpy"`
	NisaDomesticPrevDayPct   float64 `json:"nisa_domestic_prev_day_pct"`
	NisaDomesticPrevMonthJPY float64 `json:"nisa_domestic_prev_month_jpy"`
	NisaDomesticPrevMonthPct float64 `json:"nisa_domestic_prev_month_pct"`

	NisaUsValueJPY     float64 `json:"nisa_us_value_jpy"`
	NisaUsPnlJPY       float64 `json:"nisa_us_pnl_jpy"`
	NisaUsPnlPct       float64 `json:"nisa_us_pnl_pct"`
	NisaUsPrevDayJPY   float64 `json:"nisa_us_prev_day_jpy"`
	NisaUsPrevDayPct   float64 `json:"nisa_us_prev_day_pct"`
	NisaUsPrevMonthJPY float64 `json:"nisa_us_prev_month_jpy"`
	NisaUsPrevMonthPct float64 `json:"nisa_us_prev_month_pct"`

	NisaFundsValueJPY     float64 `json:"nisa_funds_value_jpy"`
	NisaFundsPnlJPY       float64 `json:"nisa_funds_pnl_jpy"`
	NisaFundsPnlPct       float64 `json:"nisa_funds_pnl_pct"`
	NisaFundsPrevDayJPY   float64 `json:"nisa_funds_prev_day_jpy"`
	NisaFundsPrevDayPct   float64 `json:"nisa_funds_prev_day_pct"`
	NisaFundsPrevMonthJPY float64 `json:"nisa_funds_prev_month_jpy"`
	NisaFundsPrevMonthPct float64 `json:"nisa_funds_prev_month_pct"`

	OldNisaTotalJPY   float64 `json:"old_nisa_total_jpy"`
	OldNisaPrevDayJPY float64 `json:"old_nisa_prev_day_jpy"`
	OldNisaPrevDayPct float64 `json:"old_nisa_prev_day_pct"`
	OldNisaPnlJPY     float64 `json:"old_nisa_pnl_jpy"`
	OldNisaPnlPct     float64 `json:"old_nisa_pnl_pct"`

	CashJpyAmount   float64 `json:"cash_jpy_amount"`
	CashJpyValueJpy float64 `json:"cash_jpy_value_jpy"`
	CashUsdAmount   float64 `json:"cash_usd_amount"`
	CashUsdValueJpy float64 `json:"cash_usd_value_jpy"`

	OtherFundsAmount   float64 `json:"other_funds_amount"`
	OtherFundsValueJpy float64 `json:"other_funds_value_jpy"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SbiHolding struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	SnapshotID int64     `json:"snapshot_id" gorm:"index:idx_snapshot_id;index:idx_snapshot_section"`
	Section    string    `json:"section" gorm:"index:idx_snapshot_section"`
	Name       string    `json:"name"`
	Quantity   float64   `json:"quantity"`
	UnitCost   float64   `json:"unit_cost"`
	UnitPrice  float64   `json:"unit_price"`
	PrevDayJPY float64   `json:"prev_day_jpy"`
	PrevDayPct float64   `json:"prev_day_pct"`
	PnlJPY     float64   `json:"pnl_jpy"`
	PnlPct     float64   `json:"pnl_pct"`
	ValueJPY   float64   `json:"value_jpy"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type sbiMoney struct {
	Amount   float64 `json:"amount"`
	ValueJPY float64 `json:"value_jpy"`
}

type sbiHoldingJSON struct {
	Name       string  `json:"name"`
	Quantity   float64 `json:"quantity"`
	UnitCost   float64 `json:"unit_cost"`
	UnitPrice  float64 `json:"unit_price"`
	PrevDayJPY float64 `json:"prev_day_jpy"`
	PrevDayPct float64 `json:"prev_day_pct"`
	PnLJPY     float64 `json:"pnl_jpy"`
	PnLPct     float64 `json:"pnl_pct"`
	ValueJPY   float64 `json:"value_jpy"`
}

type sbiNISAItem struct {
	ValueJPY     float64          `json:"value_jpy"`
	PnLJPY       float64          `json:"pnl_jpy"`
	PnLPct       float64          `json:"pnl_pct"`
	PrevDayJPY   float64          `json:"prev_day_jpy"`
	PrevDayPct   float64          `json:"prev_day_pct"`
	PrevMonthJPY float64          `json:"prev_month_jpy"`
	PrevMonthPct float64          `json:"prev_month_pct"`
	Holdings     []sbiHoldingJSON `json:"holdings"`
}

type sbiNISA struct {
	TotalJPY     float64     `json:"total_jpy"`
	PrevDayJPY   float64     `json:"prev_day_jpy"`
	PrevDayPct   float64     `json:"prev_day_pct"`
	PrevMonthJPY float64     `json:"prev_month_jpy"`
	PrevMonthPct float64     `json:"prev_month_pct"`
	PnLJPY       float64     `json:"pnl_jpy"`
	PnLPct       float64     `json:"pnl_pct"`
	Domestic     sbiNISAItem `json:"domestic_stocks"`
	USStocks     sbiNISAItem `json:"us_stocks"`
	Funds        sbiNISAItem `json:"funds"`
}

type sbiOldNISA struct {
	TotalJPY   float64          `json:"total_jpy"`
	PrevDayJPY float64          `json:"prev_day_jpy"`
	PrevDayPct float64          `json:"prev_day_pct"`
	PnLJPY     float64          `json:"pnl_jpy"`
	PnLPct     float64          `json:"pnl_pct"`
	Funds      []sbiHoldingJSON `json:"funds"`
}

type sbiCashBalances struct {
	JPY sbiMoney `json:"jpy"`
	USD sbiMoney `json:"usd"`
}

type sbiOthers struct {
	Funds sbiMoney `json:"funds"`
}

type sbiAssets struct {
	SchemaVersion int              `json:"schema_version"`
	FetchedAt     time.Time        `json:"fetched_at"`
	Status        string           `json:"status"`
	NISA          sbiNISA          `json:"nisa"`
	OldNISA       sbiOldNISA       `json:"old_nisa"`
	Cash          *sbiCashBalances `json:"cash"`
	Others        *sbiOthers       `json:"others"`
	GrandTotalJPY float64          `json:"grand_total_jpy"`
}

func ParseSbiJSON(data []byte) (*SbiSnapshot, []SbiHolding, error) {
	var dto sbiAssets
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, nil, fmt.Errorf("unmarshal sbi json: %w", err)
	}

	status := SbiStatus(strings.ToUpper(strings.TrimSpace(dto.Status)))
	if status == "" {
		status = SbiStatusOK
	}
	switch status {
	case SbiStatusOK, SbiStatusMaintenance, SbiStatusError:
	default:
		return nil, nil, fmt.Errorf("invalid status %q: want OK|MAINTENANCE|ERROR", dto.Status)
	}

	if dto.FetchedAt.IsZero() {
		return nil, nil, fmt.Errorf("fetched_at is required")
	}

	var cashJPYAmount, cashJPYValue, cashUsdAmount, cashUsdValue, otherFundsAmount, otherFundsValue float64
	if dto.Cash != nil {
		cashJPYAmount = dto.Cash.JPY.Amount
		cashJPYValue = dto.Cash.JPY.ValueJPY
		cashUsdAmount = dto.Cash.USD.Amount
		cashUsdValue = dto.Cash.USD.ValueJPY
	}
	if dto.Others != nil {
		otherFundsAmount = dto.Others.Funds.Amount
		otherFundsValue = dto.Others.Funds.ValueJPY
	}

	snap := &SbiSnapshot{
		FetchedAt:     dto.FetchedAt,
		Status:        status,
		SchemaVersion: dto.SchemaVersion,
		GrandTotalJPY: dto.GrandTotalJPY,

		NisaTotalJPY:     dto.NISA.TotalJPY,
		NisaPrevDayJPY:   dto.NISA.PrevDayJPY,
		NisaPrevDayPct:   dto.NISA.PrevDayPct,
		NisaPrevMonthJPY: dto.NISA.PrevMonthJPY,
		NisaPrevMonthPct: dto.NISA.PrevMonthPct,
		NisaPnlJPY:       dto.NISA.PnLJPY,
		NisaPnlPct:       dto.NISA.PnLPct,

		NisaDomesticValueJPY:     dto.NISA.Domestic.ValueJPY,
		NisaDomesticPnlJPY:       dto.NISA.Domestic.PnLJPY,
		NisaDomesticPnlPct:       dto.NISA.Domestic.PnLPct,
		NisaDomesticPrevDayJPY:   dto.NISA.Domestic.PrevDayJPY,
		NisaDomesticPrevDayPct:   dto.NISA.Domestic.PrevDayPct,
		NisaDomesticPrevMonthJPY: dto.NISA.Domestic.PrevMonthJPY,
		NisaDomesticPrevMonthPct: dto.NISA.Domestic.PrevMonthPct,

		NisaUsValueJPY:     dto.NISA.USStocks.ValueJPY,
		NisaUsPnlJPY:       dto.NISA.USStocks.PnLJPY,
		NisaUsPnlPct:       dto.NISA.USStocks.PnLPct,
		NisaUsPrevDayJPY:   dto.NISA.USStocks.PrevDayJPY,
		NisaUsPrevDayPct:   dto.NISA.USStocks.PrevDayPct,
		NisaUsPrevMonthJPY: dto.NISA.USStocks.PrevMonthJPY,
		NisaUsPrevMonthPct: dto.NISA.USStocks.PrevMonthPct,

		NisaFundsValueJPY:     dto.NISA.Funds.ValueJPY,
		NisaFundsPnlJPY:       dto.NISA.Funds.PnLJPY,
		NisaFundsPnlPct:       dto.NISA.Funds.PnLPct,
		NisaFundsPrevDayJPY:   dto.NISA.Funds.PrevDayJPY,
		NisaFundsPrevDayPct:   dto.NISA.Funds.PrevDayPct,
		NisaFundsPrevMonthJPY: dto.NISA.Funds.PrevMonthJPY,
		NisaFundsPrevMonthPct: dto.NISA.Funds.PrevMonthPct,

		OldNisaTotalJPY:   dto.OldNISA.TotalJPY,
		OldNisaPrevDayJPY: dto.OldNISA.PrevDayJPY,
		OldNisaPrevDayPct: dto.OldNISA.PrevDayPct,
		OldNisaPnlJPY:     dto.OldNISA.PnLJPY,
		OldNisaPnlPct:     dto.OldNISA.PnLPct,

		CashJpyAmount:   cashJPYAmount,
		CashJpyValueJpy: cashJPYValue,
		CashUsdAmount:   cashUsdAmount,
		CashUsdValueJpy: cashUsdValue,

		OtherFundsAmount:   otherFundsAmount,
		OtherFundsValueJpy: otherFundsValue,
	}

	var holdings []SbiHolding
	for _, h := range dto.NISA.Domestic.Holdings {
		holdings = append(holdings, holdingToModel(h, "nisa_domestic"))
	}
	for _, h := range dto.NISA.USStocks.Holdings {
		holdings = append(holdings, holdingToModel(h, "nisa_us"))
	}
	for _, h := range dto.NISA.Funds.Holdings {
		holdings = append(holdings, holdingToModel(h, "nisa_funds"))
	}
	for _, h := range dto.OldNISA.Funds {
		holdings = append(holdings, holdingToModel(h, "old_nisa_funds"))
	}

	return snap, holdings, nil
}

func holdingToModel(h sbiHoldingJSON, section string) SbiHolding {
	return SbiHolding{
		Section:    section,
		Name:       h.Name,
		Quantity:   h.Quantity,
		UnitCost:   h.UnitCost,
		UnitPrice:  h.UnitPrice,
		PrevDayJPY: h.PrevDayJPY,
		PrevDayPct: h.PrevDayPct,
		PnlJPY:     h.PnLJPY,
		PnlPct:     h.PnLPct,
		ValueJPY:   h.ValueJPY,
	}
}
