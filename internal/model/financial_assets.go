package model

import "time"

type FinancialSnapshot struct {
	ID                                      int64
	Source                                  string
	FetchedAt, ImportedAt                   time.Time
	ValuationJpy, CostJpy, UnrealizedPnlJpy *string
	Holdings                                []FinancialHolding
}

type FinancialHolding struct {
	ID, SnapshotID                                     int64
	Name                                               string
	Section, ProductCode, CompositeFigi, ReferenceDate *string
	Quantity, ValuationJpy, CostJpy, UnrealizedPnlJpy  *string
}

type FinancialSnapshotQuery struct {
	Sources       []string
	From, To      *time.Time
	Limit, Offset int // Limit == 0 is an unpaged internal read (Offset must be zero).
}
