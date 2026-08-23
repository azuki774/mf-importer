import type { Rule } from '@/interfaces'

export function useRules(sortBy: Ref<string>, sortOrder: Ref<string>) {
  const { data: rawRules, refresh } = useFetch<Rule[]>(
    () => `/api/rules?sort=${sortBy.value}&order=${sortOrder.value}`,
    {
      key: () => `/api/rules-${sortBy.value}-${sortOrder.value}`,
      default: () => [],
    },
  )

  const rules = computed<Rule[]>(() => rawRules.value ?? [])

  async function addRule(params: {
    fieldName: string
    value: string
    categoryId: number
    exactMatch: boolean
  }): Promise<void> {
    await $fetch('/api/rules', {
      method: 'POST',
      body: {
        fieldName: params.fieldName,
        value: params.value,
        categoryId: params.categoryId,
        exactMatch: params.exactMatch ? 1 : 0,
      },
    })
    await refresh()
  }

  async function deleteRule(id: number): Promise<void> {
    await $fetch(`/api/rules/${id}`, { method: 'DELETE' })
    await refresh()
  }

  return { rules, refresh, addRule, deleteRule }
}
