export default defineEventHandler(async () => {
  const config = useRuntimeConfig()
  const url = config.apiBaseEndpoint + "/details/count"
  const result = await $fetch<{ count: number }>(url, { method: "GET" })
  return result
})
