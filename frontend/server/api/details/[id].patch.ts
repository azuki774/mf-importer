export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const query = getQuery(event)
  const id = event.context.params?.id
  const url = config.public.apiBaseEndpoint + "/details/" + id + "?ope=" + query['ope'];
  const result = await $fetch(url,
    {
      method: "PATCH",
    }
  )
  return result
})
