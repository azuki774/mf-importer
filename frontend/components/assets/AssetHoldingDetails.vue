<script setup lang="ts">
import type { AssetRow } from '@/types/financialAssets'
import { assetPnlRate, assetSectionLabel, formatAssetDateTime, formatAssetDecimal } from '@/utils/financialAssetDecimal'

defineProps<{ row: AssetRow }>()
</script>

<template>
  <dl class="grid gap-x-6 gap-y-3 sm:grid-cols-2 text-xs text-gray-700 break-words">
    <div><dt class="text-gray-500">保有区分</dt><dd>{{ assetSectionLabel(row.section) }}</dd></div>
    <div><dt class="text-gray-500">数量（取得元の単位）</dt><dd>{{ formatAssetDecimal(row.quantity) }}</dd></div>
    <div><dt class="text-gray-500">評価額（円・取得精度）</dt><dd>{{ formatAssetDecimal(row.valuationJpy) }}</dd></div>
    <div><dt class="text-gray-500">取得価額（円・取得精度）</dt><dd>{{ formatAssetDecimal(row.costJpy) }}</dd></div>
    <div><dt class="text-gray-500">評価損益（円・取得精度）</dt><dd>{{ formatAssetDecimal(row.unrealizedPnlJpy, undefined, true) }}</dd></div>
    <div><dt class="text-gray-500">評価損益率</dt><dd>{{ formatAssetDecimal(assetPnlRate(row.costJpy, row.unrealizedPnlJpy), 2, true) }}{{ assetPnlRate(row.costJpy, row.unrealizedPnlJpy) === null ? '' : ' %' }}</dd></div>
    <div><dt class="text-gray-500">商品コード</dt><dd>{{ row.productCode ?? '—' }}</dd></div>
    <div><dt class="text-gray-500">FIGI</dt><dd>{{ row.compositeFigi ?? '—' }}</dd></div>
    <div><dt class="text-gray-500">基準日</dt><dd>{{ row.referenceDate ?? '—' }}</dd></div>
    <div><dt class="text-gray-500">データ取得日時（日本時間）</dt><dd>{{ formatAssetDateTime(row.fetchedAt) }}</dd></div>
    <div><dt class="text-gray-500">取り込み日時（日本時間）</dt><dd>{{ formatAssetDateTime(row.importedAt) }}</dd></div>
  </dl>
  <p v-if="row.source === 'sbi'" class="mt-3 text-xs text-gray-500">SBIの明細取得価額は評価額から評価損益を引いて算出しています。どちらかが未取得の場合は「—」です。</p>
</template>
