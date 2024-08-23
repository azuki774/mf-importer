export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const url = config.public.apiBaseEndpoint + "/details";
  const result = await $fetch(url,
    {
      method: "GET",
    }
  )
  return result
})
