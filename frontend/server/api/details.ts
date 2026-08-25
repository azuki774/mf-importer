export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const query = getQuery(event)
  const url = config.public.apiBaseEndpoint + "/details"
  const result = await $fetch(url,
    {
      method: "GET",
      query: { limit: query.limit, offset: query.offset, sort: query.sort, order: query.order },
    }
  )
  return result
})
