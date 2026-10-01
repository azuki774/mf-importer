import type { AssetBalance, AssetRow, AssetSnapshot, AssetSource } from '@/types/financialAssets'

interface AssetView {
  balance: AssetBalance
  snapshots: AssetSnapshot[]
  failedSources: AssetSource[]
}

export function useFinancialAssets(sources: Ref<AssetSource[]>) {
  const view = shallowRef<AssetView | null>(null)
  const pending = ref(false)
  const error = ref(false)
  let controller: AbortController | undefined
  let generation = 0

  async function refresh(clear = false) {
    const current = ++generation
    controller?.abort()
    controller = new AbortController()
    const signal = controller.signal
    if (clear) view.value = null
    error.value = false
    if (sources.value.length === 0) {
      pending.value = false
      view.value = null
      return
    }
    pending.value = true
    const query = new URLSearchParams()
    sources.value.forEach(source => query.append('source', source))
    try {
      const response = await $fetch<{ items: AssetBalance[] }>(`/api/v2/financial-assets/balances?${query}`, {
        signal, retry: 0, timeout: 15_000,
      })
      const balance = response.items[0]
      if (!balance) throw new Error('Missing balance')
      // Resolve immutable IDs from this balance, never fetch latest details separately.
      const details = await Promise.all(balance.sources.map(async (source) => {
        if (source.snapshotId === null) return { source: source.source, snapshot: null, failed: false }
        try {
          const snapshot = await $fetch<AssetSnapshot>(`/api/v2/financial-assets/snapshots/${encodeURIComponent(source.snapshotId)}`, {
            signal, retry: 0, timeout: 15_000,
          })
          if (snapshot.snapshotId !== source.snapshotId || snapshot.source !== source.source) {
            throw new Error('Snapshot mismatch')
          }
          return { source: source.source, snapshot, failed: false }
        } catch {
          return { source: source.source, snapshot: null, failed: true }
        }
      }))
      if (current !== generation) return
      // Publish summary and details together; old details never accompany new totals.
      view.value = {
        balance,
        snapshots: details.flatMap(detail => detail.snapshot ? [detail.snapshot] : []),
        failedSources: details.filter(detail => detail.failed).map(detail => detail.source),
      }
    } catch {
      if (current === generation) error.value = true
    } finally {
      if (current === generation) pending.value = false
    }
  }

  // Client-only: never fetch or serialize financial data during static generation.
  onMounted(() => {
    watch(sources, () => refresh(true), { immediate: true })
  })
  onBeforeUnmount(() => {
    generation++
    controller?.abort()
  })

  const rows = computed<AssetRow[]>(() => view.value?.snapshots.flatMap(snapshot =>
    snapshot.holdings.map(holding => ({
      ...holding, source: snapshot.source, snapshotId: snapshot.snapshotId,
      fetchedAt: snapshot.fetchedAt, importedAt: snapshot.importedAt,
    })),
  ) ?? [])
  const complete = computed(() => !!view.value
    && view.value.failedSources.length === 0 && view.value.balance.missingSources.length === 0)

  return { view, rows, complete, pending, error, refresh }
}
