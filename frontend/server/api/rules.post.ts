export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  const url = config.public.apiBaseEndpoint + "/rules";
  const reqbody = await readBody(event);
  console.log(reqbody)
  const result = await $fetch(url,
    {
      method: "POST",
      body: reqbody
    }
  )
  return result
})
