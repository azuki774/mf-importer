// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true },
  routeRules: {
    "/": { ssr: true, prerender: true },
  },
  runtimeConfig: {
    public: { // 外部から取得するにはpublic が必要
      apiBaseEndpoint: "http://172.19.250.172:20010",
    }
  }
})
