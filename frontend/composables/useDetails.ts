import type { ImportRecord } from '@/interfaces'

export function useDetails(perPage: Ref<number>, page: Ref<number>) {
  const detailsKey = computed(
    () => `/api/details?limit=${perPage.value}&offset=${(page.value - 1) * perPage.value}`
  )

  const { data: rawData, status, refresh } = useFetch<ImportRecord[]>(
    () => detailsKey.value,
    { key: () => `details-${page.value}-${perPage.value}` },
  )

  const records = computed<ImportRecord[]>(() => {
    if (!rawData.value) return []
    return rawData.value.map((d) => ({
      ...d,
      registDate: d.registDate.slice(0, 19),
      importJudgeDate: d.importJudgeDate?.slice(0, 19) ?? '',
      importDate: d.importDate?.slice(0, 19) ?? '',
    }))
  })

  const totalCount = ref(0)

  onMounted(async () => {
    try {
      const res = await $fetch<{ count: number }>('/api/details/count')
      totalCount.value = res.count ?? 0
    } catch {
      totalCount.value = 0
    }
  })

  const totalPages = computed(() => Math.ceil(totalCount.value / perPage.value) || 1)

  async function resetDetail(id: number): Promise<void> {
    await $fetch(`/api/detail?id=${id}&ope=reset`, { method: 'PATCH' })
    await refresh()
  }

  return { records, status, totalCount, totalPages, refresh, resetDetail }
}
