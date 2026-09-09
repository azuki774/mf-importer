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

	GrandTotalJPY *float64 `json:"grand_total_jpy"`

	NisaTotalJPY     *float64 `json:"nisa_total_jpy"`
	NisaPrevDayJPY   *float64 `json:"nisa_prev_day_jpy"`
	NisaPrevDayPct   *float64 `json:"nisa_prev_day_pct"`
	NisaPrevMonthJPY *float64 `json:"nisa_prev_month_jpy"`
	NisaPrevMonthPct *float64 `json:"nisa_prev_month_pct"`
	NisaPnlJPY       *float64 `json:"nisa_pnl_jpy"`
	NisaPnlPct       *float64 `json:"nisa_pnl_pct"`

	NisaDomesticValueJPY     *float64 `json:"nisa_domestic_value_jpy"`
	NisaDomesticPnlJPY       *float64 `json:"nisa_domestic_pnl_jpy"`
	NisaDomesticPnlPct       *float64 `json:"nisa_domestic_pnl_pct"`
	NisaDomesticPrevDayJPY   *float64 `json:"nisa_domestic_prev_day_jpy"`
	NisaDomesticPrevDayPct   *float64 `json:"nisa_domestic_prev_day_pct"`
	NisaDomesticPrevMonthJPY *float64 `json:"nisa_domestic_prev_month_jpy"`
	NisaDomesticPrevMonthPct *float64 `json:"nisa_domestic_prev_month_pct"`

	NisaUsValueJPY     *float64 `json:"nisa_us_value_jpy"`
	NisaUsPnlJPY       *float64 `json:"nisa_us_pnl_jpy"`
	NisaUsPnlPct       *float64 `json:"nisa_us_pnl_pct"`
	NisaUsPrevDayJPY   *float64 `json:"nisa_us_prev_day_jpy"`
	NisaUsPrevDayPct   *float64 `json:"nisa_us_prev_day_pct"`
	NisaUsPrevMonthJPY *float64 `json:"nisa_us_prev_month_jpy"`
	NisaUsPrevMonthPct *float64 `json:"nisa_us_prev_month_pct"`

	NisaFundsValueJPY     *float64 `json:"nisa_funds_value_jpy"`
	NisaFundsPnlJPY       *float64 `json:"nisa_funds_pnl_jpy"`
	NisaFundsPnlPct       *float64 `json:"nisa_funds_pnl_pct"`
	NisaFundsPrevDayJPY   *float64 `json:"nisa_funds_prev_day_jpy"`
	NisaFundsPrevDayPct   *float64 `json:"nisa_funds_prev_day_pct"`
	NisaFundsPrevMonthJPY *float64 `json:"nisa_funds_prev_month_jpy"`
	NisaFundsPrevMonthPct *float64 `json:"nisa_funds_prev_month_pct"`

	OldNisaTotalJPY   *float64 `json:"old_nisa_total_jpy"`
	OldNisaPrevDayJPY *float64 `json:"old_nisa_prev_day_jpy"`
	OldNisaPrevDayPct *float64 `json:"old_nisa_prev_day_pct"`
	OldNisaPnlJPY     *float64 `json:"old_nisa_pnl_jpy"`
	OldNisaPnlPct     *float64 `json:"old_nisa_pnl_pct"`

	CashJpyAmount   *float64 `json:"cash_jpy_amount"`
	CashJpyValueJpy *float64 `json:"cash_jpy_value_jpy"`
	CashUsdAmount   *float64 `json:"cash_usd_amount"`
	CashUsdValueJpy *float64 `json:"cash_usd_value_jpy"`

	OtherFundsAmount   *float64 `json:"other_funds_amount"`
	OtherFundsValueJpy *float64 `json:"other_funds_value_jpy"`

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
	PrevDayJPY *float64  `json:"prev_day_jpy"`
	PrevDayPct *float64  `json:"prev_day_pct"`
	PnlJPY     float64   `json:"pnl_jpy"`
	PnlPct     float64   `json:"pnl_pct"`
	ValueJPY   float64   `json:"value_jpy"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type sbiMoney struct {
	Amount   *float64 `json:"amount"`
	ValueJPY *float64 `json:"value_jpy"`
}

type sbiHoldingJSON struct {
	Name       string   `json:"name"`
	Quantity   float64  `json:"quantity"`
	UnitCost   float64  `json:"unit_cost"`
	UnitPrice  float64  `json:"unit_price"`
	PrevDayJPY *float64 `json:"prev_day_jpy"`
	PrevDayPct *float64 `json:"prev_day_pct"`
	PnLJPY     float64  `json:"pnl_jpy"`
	PnLPct     float64  `json:"pnl_pct"`
	ValueJPY   float64  `json:"value_jpy"`
}

type sbiNISAItem struct {
	ValueJPY     *float64         `json:"value_jpy"`
	PnLJPY       *float64         `json:"pnl_jpy"`
	PnLPct       *float64         `json:"pnl_pct"`
	PrevDayJPY   *float64         `json:"prev_day_jpy"`
	PrevDayPct   *float64         `json:"prev_day_pct"`
	PrevMonthJPY *float64         `json:"prev_month_jpy"`
	PrevMonthPct *float64         `json:"prev_month_pct"`
	Holdings     []sbiHoldingJSON `json:"holdings"`
}

type sbiNISA struct {
	TotalJPY     *float64    `json:"total_jpy"`
	PrevDayJPY   *float64    `json:"prev_day_jpy"`
	PrevDayPct   *float64    `json:"prev_day_pct"`
	PrevMonthJPY *float64    `json:"prev_month_jpy"`
	PrevMonthPct *float64    `json:"prev_month_pct"`
	PnLJPY       *float64    `json:"pnl_jpy"`
	PnLPct       *float64    `json:"pnl_pct"`
	Domestic     sbiNISAItem `json:"domestic_stocks"`
	USStocks     sbiNISAItem `json:"us_stocks"`
	Funds        sbiNISAItem `json:"funds"`
}

type sbiOldNISA struct {
	TotalJPY   *float64         `json:"total_jpy"`
	PrevDayJPY *float64         `json:"prev_day_jpy"`
	PrevDayPct *float64         `json:"prev_day_pct"`
	PnLJPY     *float64         `json:"pnl_jpy"`
	PnLPct     *float64         `json:"pnl_pct"`
	Funds      []sbiHoldingJSON `json:"funds"`
}

type sbiCashBalances struct {
	JPY *sbiMoney `json:"jpy"`
	USD *sbiMoney `json:"usd"`
}

type sbiOthers struct {
	Funds *sbiMoney `json:"funds"`
}

type sbiAssets struct {
	SchemaVersion int              `json:"schema_version"`
	FetchedAt     time.Time        `json:"fetched_at"`
	Status        string           `json:"status"`
	NISA          *sbiNISA         `json:"nisa"`
	OldNISA       *sbiOldNISA      `json:"old_nisa"`
	Cash          *sbiCashBalances `json:"cash"`
	Others        *sbiOthers       `json:"others"`
	GrandTotalJPY *float64         `json:"grand_total_jpy"`
}

var sbiTokyo = time.FixedZone("Asia/Tokyo", 9*60*60)

// NormalizeSbiFetchedAt converts a scraper timestamp to the precision stored by
// the sbi_snapshot DATETIME(6) column and gives it a canonical Tokyo location.
func NormalizeSbiFetchedAt(t time.Time) time.Time {
	return t.In(sbiTokyo).Truncate(time.Microsecond)
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
	// TODO: Reject unsupported schema versions once the compatibility policy is defined.

	var cashJPYAmount, cashJPYValue, cashUsdAmount, cashUsdValue, otherFundsAmount, otherFundsValue *float64
	if dto.Cash != nil {
		if dto.Cash.JPY != nil {
			cashJPYAmount = dto.Cash.JPY.Amount
			cashJPYValue = dto.Cash.JPY.ValueJPY
		}
		if dto.Cash.USD != nil {
			cashUsdAmount = dto.Cash.USD.Amount
			cashUsdValue = dto.Cash.USD.ValueJPY
		}
	}
	if dto.Others != nil && dto.Others.Funds != nil {
		otherFundsAmount = dto.Others.Funds.Amount
		otherFundsValue = dto.Others.Funds.ValueJPY
	}

	var nisa sbiNISA
	if dto.NISA != nil && status == SbiStatusOK {
		nisa = *dto.NISA
	}
	var oldNISA sbiOldNISA
	if dto.OldNISA != nil {
		oldNISA = *dto.OldNISA
	}
	grandTotalJPY := dto.GrandTotalJPY
	if status != SbiStatusOK {
		// The current producer marks maintenance when the NISA section is
		// unavailable, so non-OK grand totals are not complete snapshots.
		grandTotalJPY = nil
	}

	snap := &SbiSnapshot{
		FetchedAt:     NormalizeSbiFetchedAt(dto.FetchedAt),
		Status:        status,
		SchemaVersion: dto.SchemaVersion,
		GrandTotalJPY: grandTotalJPY,

		NisaTotalJPY:     nisa.TotalJPY,
		NisaPrevDayJPY:   nisa.PrevDayJPY,
		NisaPrevDayPct:   nisa.PrevDayPct,
		NisaPrevMonthJPY: nisa.PrevMonthJPY,
		NisaPrevMonthPct: nisa.PrevMonthPct,
		NisaPnlJPY:       nisa.PnLJPY,
		NisaPnlPct:       nisa.PnLPct,

		NisaDomesticValueJPY:     nisa.Domestic.ValueJPY,
		NisaDomesticPnlJPY:       nisa.Domestic.PnLJPY,
		NisaDomesticPnlPct:       nisa.Domestic.PnLPct,
		NisaDomesticPrevDayJPY:   nisa.Domestic.PrevDayJPY,
		NisaDomesticPrevDayPct:   nisa.Domestic.PrevDayPct,
		NisaDomesticPrevMonthJPY: nisa.Domestic.PrevMonthJPY,
		NisaDomesticPrevMonthPct: nisa.Domestic.PrevMonthPct,

		NisaUsValueJPY:     nisa.USStocks.ValueJPY,
		NisaUsPnlJPY:       nisa.USStocks.PnLJPY,
		NisaUsPnlPct:       nisa.USStocks.PnLPct,
		NisaUsPrevDayJPY:   nisa.USStocks.PrevDayJPY,
		NisaUsPrevDayPct:   nisa.USStocks.PrevDayPct,
		NisaUsPrevMonthJPY: nisa.USStocks.PrevMonthJPY,
		NisaUsPrevMonthPct: nisa.USStocks.PrevMonthPct,

		NisaFundsValueJPY:     nisa.Funds.ValueJPY,
		NisaFundsPnlJPY:       nisa.Funds.PnLJPY,
		NisaFundsPnlPct:       nisa.Funds.PnLPct,
		NisaFundsPrevDayJPY:   nisa.Funds.PrevDayJPY,
		NisaFundsPrevDayPct:   nisa.Funds.PrevDayPct,
		NisaFundsPrevMonthJPY: nisa.Funds.PrevMonthJPY,
		NisaFundsPrevMonthPct: nisa.Funds.PrevMonthPct,

		OldNisaTotalJPY:   oldNISA.TotalJPY,
		OldNisaPrevDayJPY: oldNISA.PrevDayJPY,
		OldNisaPrevDayPct: oldNISA.PrevDayPct,
		OldNisaPnlJPY:     oldNISA.PnLJPY,
		OldNisaPnlPct:     oldNISA.PnLPct,

		CashJpyAmount:   cashJPYAmount,
		CashJpyValueJpy: cashJPYValue,
		CashUsdAmount:   cashUsdAmount,
		CashUsdValueJpy: cashUsdValue,

		OtherFundsAmount:   otherFundsAmount,
		OtherFundsValueJpy: otherFundsValue,
	}

	var holdings []SbiHolding
	if dto.NISA != nil {
		for _, h := range dto.NISA.Domestic.Holdings {
			holdings = append(holdings, holdingToModel(h, "nisa_domestic"))
		}
		for _, h := range dto.NISA.USStocks.Holdings {
			holdings = append(holdings, holdingToModel(h, "nisa_us"))
		}
		for _, h := range dto.NISA.Funds.Holdings {
			holdings = append(holdings, holdingToModel(h, "nisa_funds"))
		}
	}
	if dto.OldNISA != nil {
		for _, h := range dto.OldNISA.Funds {
			holdings = append(holdings, holdingToModel(h, "old_nisa_funds"))
		}
	}

	return snap, holdings, nil
}

func holdingToModel(h sbiHoldingJSON, section string) SbiHolding {
	prevDayJPY := h.PrevDayJPY
	prevDayPct := h.PrevDayPct
	if section == "nisa_us" {
		// Schema version 1 cannot provide per-holding US prev-day values.
		prevDayJPY = nil
		prevDayPct = nil
	}
	return SbiHolding{
		Section:    section,
		Name:       h.Name,
		Quantity:   h.Quantity,
		UnitCost:   h.UnitCost,
		UnitPrice:  h.UnitPrice,
		PrevDayJPY: prevDayJPY,
		PrevDayPct: prevDayPct,
		PnlJPY:     h.PnLJPY,
		PnlPct:     h.PnLPct,
		ValueJPY:   h.ValueJPY,
	}
}
