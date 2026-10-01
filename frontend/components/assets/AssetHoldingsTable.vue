<script setup lang="ts">
import type { AssetRow, AssetTotals } from '@/types/financialAssets'
import {
  assetPnlClass, assetPnlRate, assetSectionLabel, assetSourceLabel, compareAssetDecimals,
  formatAssetDecimal, formatAssetMoney, sumAssetDecimals,
} from '@/utils/financialAssetDecimal'

const props = defineProps<{ rows: AssetRow[], complete: boolean }>()
const search = ref('')
const section = ref('all')
const page = ref(1)
const perPage = ref(50)
const expanded = ref(new Set<string>())
type SortKey = keyof AssetTotals | 'name' | 'rate'
const sortBy = ref<SortKey>('valuationJpy')
const sortOrder = ref<'asc' | 'desc'>('desc')
const sections = computed(() => [...new Set(props.rows.map(row => row.section).filter((value): value is string => value !== null))])

const filtered = computed(() => {
  const query = search.value.trim().normalize('NFKC').toLocaleLowerCase('ja')
  return props.rows.filter(row =>
    (section.value === 'all' || (section.value === 'none' ? row.section === null : row.section === section.value))
    && (!query || [row.name, row.productCode, row.compositeFigi].some(value => value?.normalize('NFKC').toLocaleLowerCase('ja').includes(query))),
  )
})
const sorted = computed(() => [...filtered.value].sort((a, b) => {
  let comparison: number
  if (sortBy.value === 'name') {
    comparison = a.name.localeCompare(b.name, 'ja') * (sortOrder.value === 'asc' ? 1 : -1)
  } else {
    const value = (row: AssetRow) => sortBy.value === 'rate'
      ? assetPnlRate(row.costJpy, row.unrealizedPnlJpy) : row[sortBy.value as keyof AssetTotals]
    comparison = compareAssetDecimals(value(a), value(b), sortOrder.value)
  }
  return comparison || a.holdingId.localeCompare(b.holdingId)
}))
const totalPages = computed(() => perPage.value === 0 ? 1 : Math.max(1, Math.ceil(sorted.value.length / perPage.value)))
const visibleRows = computed(() => perPage.value === 0 ? sorted.value : sorted.value.slice((page.value - 1) * perPage.value, page.value * perPage.value))
const totals = computed<AssetTotals>(() => {
  const sum = (key: keyof AssetTotals) => props.complete ? sumAssetDecimals(filtered.value.map(row => row[key])) : null
  return { valuationJpy: sum('valuationJpy'), costJpy: sum('costJpy'), unrealizedPnlJpy: sum('unrealizedPnlJpy') }
})

watch([search, section, perPage, sortBy, sortOrder, () => props.rows], () => {
  page.value = 1
  expanded.value = new Set()
})
watch(sections, (values) => {
  if (section.value !== 'all' && section.value !== 'none' && !values.includes(section.value)) section.value = 'all'
})

function toggleSort(key: SortKey) {
  if (sortBy.value === key) sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  else { sortBy.value = key; sortOrder.value = key === 'name' ? 'asc' : 'desc' }
}
function ariaSort(key: SortKey): 'none' | 'ascending' | 'descending' {
  return sortBy.value !== key ? 'none' : sortOrder.value === 'asc' ? 'ascending' : 'descending'
}
function indicator(key: SortKey) {
  return sortBy.value === key ? sortOrder.value === 'asc' ? '▲' : '▼' : ''
}
function toggleExpanded(id: string) {
  const next = new Set(expanded.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expanded.value = next
}
</script>

<template>
  <section aria-labelledby="asset-holdings-heading" class="bg-white rounded-lg border border-gray-200 shadow-sm">
    <div class="p-4 sm:p-6 border-b border-gray-200 space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h2 id="asset-holdings-heading" class="text-base font-semibold text-gray-900">保有商品</h2>
        <p class="text-sm text-gray-500" role="status">{{ filtered.length }} 件 / 取得済み {{ rows.length }} 件</p>
      </div>
      <div class="flex flex-col sm:flex-row sm:items-end gap-3">
        <div class="flex-1">
          <label for="asset-search" class="block text-xs text-gray-600 mb-1">商品検索</label>
          <input id="asset-search" v-model="search" type="search" placeholder="商品名・商品コード・FIGI" class="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-2 focus:ring-primary-500 focus:outline-none">
        </div>
        <div>
          <label for="asset-section" class="block text-xs text-gray-600 mb-1">保有区分</label>
          <select id="asset-section" v-model="section" class="w-full border border-gray-300 rounded-md px-3 py-2 bg-white text-sm">
            <option value="all">すべて</option>
            <option v-for="value in sections" :key="value" :value="value">{{ assetSectionLabel(value) }}</option>
            <option value="none">区分なし（NRKNなど）</option>
          </select>
        </div>
        <div>
          <label for="asset-per-page" class="block text-xs text-gray-600 mb-1">表示件数</label>
          <select id="asset-per-page" v-model.number="perPage" class="w-full border border-gray-300 rounded-md px-3 py-2 bg-white text-sm">
            <option :value="50">50件</option><option :value="100">100件</option><option :value="0">全件</option>
          </select>
        </div>
      </div>
      <p class="text-xs text-gray-500">商品名を押すと数量・基準日・取得精度の金額などを表示します。</p>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full text-sm" aria-label="保有商品一覧">
        <thead class="bg-gray-50 text-xs text-gray-600">
          <tr>
            <th scope="col" class="p-3 text-left" :aria-sort="ariaSort('name')"><button class="whitespace-nowrap" @click="toggleSort('name')">商品名 {{ indicator('name') }}</button></th>
            <th scope="col" class="p-3 text-left">データ元</th>
            <th scope="col" class="hidden lg:table-cell p-3 text-left">保有区分</th>
            <th scope="col" class="p-3 text-right" :aria-sort="ariaSort('valuationJpy')"><button class="whitespace-nowrap" @click="toggleSort('valuationJpy')">評価額 {{ indicator('valuationJpy') }}</button></th>
            <th scope="col" class="hidden md:table-cell p-3 text-right" :aria-sort="ariaSort('costJpy')"><button class="whitespace-nowrap" @click="toggleSort('costJpy')">取得価額 {{ indicator('costJpy') }}</button></th>
            <th scope="col" class="p-3 text-right" :aria-sort="ariaSort('unrealizedPnlJpy')"><button class="whitespace-nowrap" @click="toggleSort('unrealizedPnlJpy')">評価損益 {{ indicator('unrealizedPnlJpy') }}</button></th>
            <th scope="col" class="hidden xl:table-cell p-3 text-right" :aria-sort="ariaSort('rate')"><button class="whitespace-nowrap" @click="toggleSort('rate')">損益率 {{ indicator('rate') }}</button></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100">
          <template v-for="row in visibleRows" :key="row.holdingId">
            <tr data-testid="asset-holding-row" class="hover:bg-gray-50">
              <th scope="row" class="p-3 text-left font-medium min-w-32 max-w-xs break-words">
                <button class="text-primary-700 hover:text-primary-900 text-left" :aria-expanded="expanded.has(row.holdingId)" :aria-controls="`holding-${row.holdingId}`" @click="toggleExpanded(row.holdingId)">
                  <span aria-hidden="true">{{ expanded.has(row.holdingId) ? '▾' : '▸' }}</span> {{ row.name }}
                </button>
              </th>
              <td class="p-3 text-xs text-gray-600">{{ assetSourceLabel(row.source) }}</td>
              <td class="hidden lg:table-cell p-3 text-xs text-gray-500">{{ assetSectionLabel(row.section) }}</td>
              <td class="p-3 text-right tabular-nums whitespace-nowrap">{{ formatAssetMoney(row.valuationJpy) }}</td>
              <td class="hidden md:table-cell p-3 text-right tabular-nums whitespace-nowrap">{{ formatAssetMoney(row.costJpy) }}</td>
              <td class="p-3 text-right tabular-nums whitespace-nowrap" :class="assetPnlClass(row.unrealizedPnlJpy)">{{ formatAssetMoney(row.unrealizedPnlJpy, true) }}</td>
              <td class="hidden xl:table-cell p-3 text-right tabular-nums whitespace-nowrap" :class="assetPnlClass(row.unrealizedPnlJpy)">{{ formatAssetDecimal(assetPnlRate(row.costJpy, row.unrealizedPnlJpy), 2, true) }}{{ assetPnlRate(row.costJpy, row.unrealizedPnlJpy) === null ? '' : ' %' }}</td>
            </tr>
            <tr v-if="expanded.has(row.holdingId)" :id="`holding-${row.holdingId}`" class="bg-gray-50">
              <td colspan="7" class="p-4 sm:px-6"><AssetsAssetHoldingDetails :row="row" /></td>
            </tr>
          </template>
          <tr v-if="!visibleRows.length"><td colspan="7" class="p-8 text-center text-gray-500">{{ rows.length ? '条件に一致する商品がありません' : '表示できる保有商品がありません' }}</td></tr>
        </tbody>
      </table>
    </div>
    <div class="border-t border-gray-200 p-4 sm:px-6 space-y-4">
      <div aria-label="表示対象の明細合計">
        <h3 class="text-sm font-medium text-gray-800">表示対象の明細合計（{{ filtered.length }}件・全ページ）</h3>
        <p v-if="!complete" class="mt-1 text-xs text-amber-800">一部のデータがないため、明細合計は未確定です。</p>
        <dl class="mt-3 grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm">
          <div><dt class="text-xs text-gray-500">評価額</dt><dd data-testid="holdings-total-value" class="mt-1 tabular-nums">{{ formatAssetMoney(totals.valuationJpy) }}</dd></div>
          <div><dt class="text-xs text-gray-500">取得価額</dt><dd data-testid="holdings-total-cost" class="mt-1 tabular-nums">{{ formatAssetMoney(totals.costJpy) }}</dd></div>
          <div><dt class="text-xs text-gray-500">評価損益</dt><dd data-testid="holdings-total-pnl" class="mt-1 tabular-nums" :class="assetPnlClass(totals.unrealizedPnlJpy)">{{ formatAssetMoney(totals.unrealizedPnlJpy, true) }}</dd></div>
        </dl>
        <p class="mt-3 text-xs text-gray-500">検索条件に一致する明細の合計です。総資産には明細に含まれない残高もあるため、総資産とは一致しない場合があります。未取得値を含む項目は「—」になります。</p>
      </div>
      <Pagination :page="page" :total-pages="totalPages" @update:page="page = $event" />
    </div>
  </section>
</template>
