import type { Rule } from '@/interfaces'

export function useRules() {
  const { data: rawRules, refresh } = useFetch<Rule[]>('/api/rules', {
    key: '/api/rules',
    default: () => [],
  })

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
    await $fetch(`/api/rules?id=${id}`, { method: 'DELETE' })
    await refresh()
  }

  return { rules, refresh, addRule, deleteRule }
}
