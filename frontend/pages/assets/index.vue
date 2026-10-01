<script setup lang="ts">
import type { AssetSource } from '@/types/financialAssets'
import { assetSourceLabel } from '@/utils/financialAssetDecimal'

useHead({ title: '金融資産 | mf-importer' })
const route = useRoute()
const router = useRouter()
const sourceOptions: AssetSource[] = ['sbi', 'nrkn']
const sources = computed<AssetSource[]>({
  get() {
    const query = route.query.source
    if (query === undefined) return [...sourceOptions]
    const requested = Array.isArray(query) ? query : [query]
    return sourceOptions.filter(source => requested.includes(source))
  },
  set(value) {
    void router.replace({ query: { ...route.query, source: value.length ? value : 'none' } })
  },
})
function toggleSource(source: AssetSource) {
  sources.value = sources.value.includes(source) ? sources.value.filter(value => value !== source) : [...sources.value, source]
}
const { view, rows, complete, pending, error, refresh } = useFinancialAssets(sources)
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 class="text-xl font-semibold text-gray-900">金融資産</h1>
        <p class="mt-1 text-sm text-gray-500">SBI・NRKNの最新保存データと保有商品</p>
      </div>
      <button class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50" :disabled="pending || !sources.length" @click="refresh()">
        {{ pending ? '読み込み中…' : '再読み込み' }}
      </button>
    </div>
    <fieldset class="flex flex-wrap gap-4 text-sm">
      <legend class="mb-2 font-medium text-gray-700">対象データ元</legend>
      <label v-for="source in sourceOptions" :key="source" class="inline-flex items-center gap-2 cursor-pointer">
        <input type="checkbox" :checked="sources.includes(source)" class="h-4 w-4 accent-primary-600" @change="toggleSource(source)">
        {{ assetSourceLabel(source) }}
      </label>
    </fieldset>

    <p v-if="!sources.length" role="status" class="rounded-lg bg-white border border-gray-200 p-6 text-sm text-gray-600">データ元を選択してください。</p>
    <div v-if="error" role="alert" class="rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-800">
      <p>{{ view ? '更新に失敗しました。前回の表示を保持しています。' : '資産データを取得できませんでした。' }}</p>
      <button class="mt-2 underline font-medium" :disabled="pending" @click="refresh()">再試行</button>
    </div>
    <div v-if="pending && !view" role="status" class="bg-white border border-gray-200 rounded-lg p-6 animate-pulse motion-reduce:animate-none">
      <p class="text-sm text-gray-600">資産データを読み込んでいます…</p>
      <div class="mt-4 h-10 w-48 bg-gray-100 rounded" /><div class="mt-4 h-24 bg-gray-100 rounded" />
    </div>
    <div v-if="view" class="space-y-6" :aria-busy="pending">
      <p v-if="pending" role="status" class="text-sm text-gray-500">更新中です。前回の表示を保持しています。</p>
      <AssetsAssetSummary :balance="view.balance" />
      <div v-if="view.failedSources.length" role="alert" class="rounded-lg bg-amber-50 border border-amber-200 p-4 text-sm text-amber-900">
        {{ view.failedSources.map(assetSourceLabel).join('・') }}の保有商品を取得できませんでした。取得できた明細のみ表示しています。
        <button class="ml-2 underline font-medium" :disabled="pending" @click="refresh()">再試行</button>
      </div>
      <p v-if="view.balance.sources.every(source => source.snapshotId === null)" role="status" class="text-sm text-gray-600">取り込み済みの資産データがありません。</p>
      <AssetsAssetHoldingsTable :rows="rows" :complete="complete" />
    </div>
  </div>
</template>
