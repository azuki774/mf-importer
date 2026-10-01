import Decimal from 'decimal.js'
import type { AssetDecimal, AssetSource } from '@/types/financialAssets'

// Covers the database's DECIMAL/BIGINT ranges plus aggregation and ratios.
const Money = Decimal.clone({ precision: 80, rounding: Decimal.ROUND_HALF_UP })

export function sumAssetDecimals(values: AssetDecimal[]): AssetDecimal {
  if (values.some(value => value === null)) return null
  return values.reduce<Decimal>((sum, value) => sum.plus(value!), new Money(0)).toFixed()
}

export function assetPnlRate(cost: AssetDecimal, pnl: AssetDecimal): AssetDecimal {
  if (cost === null || pnl === null || !new Money(cost).gt(0)) return null
  return new Money(pnl).div(cost).times(100).toFixed()
}

export function compareAssetDecimals(a: AssetDecimal, b: AssetDecimal, order: 'asc' | 'desc'): number {
  // Missing values always sort last, in either direction.
  if (a === null) return b === null ? 0 : 1
  if (b === null) return -1
  return new Money(a).cmp(b) * (order === 'asc' ? 1 : -1)
}

export function formatAssetDecimal(value: AssetDecimal, places?: number, signed = false): string {
  if (value === null) return '—'
  const decimal = new Money(value)
  const rounded = places === undefined ? decimal : decimal.toDecimalPlaces(places)
  const text = rounded.toFixed(places)
  const [integer, fraction] = text.split('.')
  const grouped = integer!.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  return `${signed && rounded.isPositive() && !rounded.isZero() ? '+' : ''}${grouped}${fraction === undefined ? '' : `.${fraction}`}`
}

export function formatAssetMoney(value: AssetDecimal, signed = false): string {
  return value === null ? '—' : `${formatAssetDecimal(value, 0, signed)} 円`
}

export function assetPnlClass(value: AssetDecimal): string {
  if (value === null || new Money(value).isZero()) return 'text-gray-600'
  return new Money(value).isNegative() ? 'text-red-700' : 'text-emerald-700'
}

export function assetSourceLabel(source: AssetSource): string {
  return source === 'sbi' ? 'SBI' : 'NRKN'
}

export function assetSectionLabel(section: string | null): string {
  const labels: Record<string, string> = {
    nisa_domestic: 'NISA・国内株式',
    nisa_us: 'NISA・米国株式',
    nisa_funds: 'NISA・投資信託',
    old_nisa_funds: '旧NISA・投資信託',
  }
  return section === null ? '—' : labels[section] ?? section
}

export function formatAssetDateTime(value: string | null): string {
  if (value === null) return '—'
  return new Intl.DateTimeFormat('ja-JP', {
    timeZone: 'Asia/Tokyo', year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23',
  }).format(new Date(value))
}
