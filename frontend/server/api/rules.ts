export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const query = getQuery(event)
  const url = config.public.apiBaseEndpoint + "/rules";
  const result = await $fetch(url,
    {
      method: "GET",
      query: { sort: query.sort, order: query.order },
    }
  )
  return result
})
