export default defineEventHandler((event) => {
  const id = getRouterParam(event, 'snapshotId')
  if (!id) throw createError({ statusCode: 400, statusMessage: 'Missing snapshot ID' })
  return financialAssetsProxy(event, `snapshots/${encodeURIComponent(id)}`)
})
