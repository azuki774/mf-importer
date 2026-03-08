export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const url = config.apiBaseEndpoint + "/rules";
  const result = await $fetch(url,
    {
      method: "GET",
    }
  )
  return result
})
