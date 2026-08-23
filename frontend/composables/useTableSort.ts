export type SortOrder = 'asc' | 'desc'

export function useTableSort(
  defaultBy: string,
  defaultOrder: SortOrder,
  initialDirections: Record<string, SortOrder> = {},
) {
  const sortBy = ref(defaultBy)
  const sortOrder = ref<SortOrder>(defaultOrder)

  function toggleSort(key: string) {
    if (sortBy.value === key) {
      sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
      return
    }
    sortBy.value = key
    sortOrder.value = initialDirections[key] ?? 'asc'
  }

  return { sortBy, sortOrder, toggleSort }
}
