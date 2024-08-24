export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const query = getQuery(event)
  const url = config.public.apiBaseEndpoint + "/rules/" + query.id;
  const result = await $fetch(url,
    {
      method: "DELETE",
    }
  )
  return result
})
