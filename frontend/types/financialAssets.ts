// Mirrors internal/openapi/mfimporter-api.yaml. Decimal values stay strings.
export type AssetSource = 'sbi' | 'nrkn'
export type AssetDecimal = string | null

export interface AssetTotals {
  valuationJpy: AssetDecimal
  costJpy: AssetDecimal
  unrealizedPnlJpy: AssetDecimal
}

export interface AssetBalanceSource {
  source: AssetSource
  snapshotId: string | null
  fetchedAt: string | null
  totals: AssetTotals
}

export interface AssetBalance {
  timestamp: string | null
  periodStart: string | null
  periodEnd: string | null
  totals: AssetTotals
  sources: AssetBalanceSource[]
  missingSources: AssetSource[]
}

export interface AssetHolding extends AssetTotals {
  holdingId: string
  name: string
  section: string | null
  productCode: string | null
  compositeFigi: string | null
  referenceDate: string | null
  quantity: AssetDecimal
}

export interface AssetSnapshot {
  snapshotId: string
  source: AssetSource
  fetchedAt: string
  importedAt: string
  totals: AssetTotals
  holdings: AssetHolding[]
}

export interface AssetRow extends AssetHolding {
  source: AssetSource
  snapshotId: string
  fetchedAt: string
  importedAt: string
}
