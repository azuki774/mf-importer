import type { H3Event } from 'h3'

export async function financialAssetsProxy(event: H3Event, resource: string) {
  const config = useRuntimeConfig()
  // Preserve repeated source parameters and the API's own validation rules.
  const url = `${config.public.apiBaseEndpoint}/v2/financial-assets/${resource}${getRequestURL(event).search}`
  try {
    return await $fetch(url, { retry: 0, timeout: 15_000 })
  } catch (error) {
    const status = (error as { response?: { status: number } }).response?.status
    throw createError({ statusCode: status ?? 502, statusMessage: 'Financial asset request failed' })
  }
}
