export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const url = config.public.apiBaseEndpoint + "/rules";
  const result = await $fetch(url,
    {
      method: "GET",
    }
  )
  return result
})
