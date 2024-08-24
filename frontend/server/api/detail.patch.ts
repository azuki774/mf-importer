export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const query = getQuery(event)
  const url = config.public.apiBaseEndpoint + "/details/" + query['id'] + "?ope=" + query['ope'];
  console.log(url)
  const result = await $fetch(url,
    {
      method: "PATCH",
    }
  )
  return result
})
