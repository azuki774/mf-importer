<script setup lang="ts">
import type { AssetBalance } from '@/types/financialAssets'
import { assetPnlClass, assetSourceLabel, formatAssetDateTime, formatAssetMoney } from '@/utils/financialAssetDecimal'

defineProps<{ balance: AssetBalance }>()
</script>

<template>
  <section aria-labelledby="asset-summary-heading" class="space-y-4">
    <div class="bg-white rounded-lg border border-gray-200 shadow-sm px-4 sm:px-6 py-5">
      <h2 id="asset-summary-heading" class="text-sm font-medium text-gray-600">総資産評価額</h2>
      <p data-testid="asset-total" class="mt-2 text-3xl font-semibold text-gray-900 tabular-nums break-words">
        {{ formatAssetMoney(balance.totals.valuationJpy) }}
      </p>
      <p class="mt-3 text-xs text-gray-500">各データ元の最新の正常取得データを集計しています。取得時点はデータ元ごとに異なる場合があります。</p>
      <p class="mt-1 text-xs text-gray-500">表示基準日時：{{ formatAssetDateTime(balance.timestamp) }}（日本時間）</p>
      <p v-if="balance.missingSources.length" role="status" class="mt-3 text-sm text-amber-800">
        {{ balance.missingSources.map(assetSourceLabel).join('・') }}のデータがありません。総資産は未確定です。
      </p>
    </div>

    <div class="grid gap-4 sm:grid-cols-2">
      <article v-for="source in balance.sources" :key="source.source" :aria-label="`${assetSourceLabel(source.source)}の資産`" class="bg-white rounded-lg border border-gray-200 shadow-sm p-4 sm:p-5">
        <div class="flex justify-between items-center gap-2">
          <h3 class="font-semibold text-gray-900">{{ assetSourceLabel(source.source) }}</h3>
          <span class="text-xs rounded-full px-2 py-1" :class="source.snapshotId ? 'bg-primary-50 text-primary-700' : 'bg-gray-100 text-gray-600'">
            {{ source.snapshotId ? '最新保存データ' : 'データなし' }}
          </span>
        </div>
        <dl class="mt-4 space-y-2 text-sm">
          <div class="flex justify-between gap-3"><dt class="text-gray-500">評価額</dt><dd class="tabular-nums font-medium">{{ formatAssetMoney(source.totals.valuationJpy) }}</dd></div>
          <div class="flex justify-between gap-3"><dt class="text-gray-500">取得価額</dt><dd class="tabular-nums">{{ formatAssetMoney(source.totals.costJpy) }}</dd></div>
          <div class="flex justify-between gap-3"><dt class="text-gray-500">評価損益</dt><dd class="tabular-nums" :class="assetPnlClass(source.totals.unrealizedPnlJpy)">{{ formatAssetMoney(source.totals.unrealizedPnlJpy, true) }}</dd></div>
        </dl>
        <p class="mt-4 text-xs text-gray-500">データ取得日時：{{ formatAssetDateTime(source.fetchedAt) }}（日本時間）</p>
      </article>
    </div>
    <p class="text-xs text-gray-500">「—」は未取得・不明の値です。SBI全体の取得価額・評価損益は提供されていません。金額は円単位に四捨五入して表示します。</p>
  </section>
</template>
