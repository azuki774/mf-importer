export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const id = event.context.params?.id
  const url = config.public.apiBaseEndpoint + "/rules/" + id;
  const result = await $fetch(url,
    {
      method: "DELETE",
    }
  )
  return result
})
