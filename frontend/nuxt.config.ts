// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss'],
  routeRules: {
    "/": { ssr: false },
  },
  runtimeConfig: {
    public: {
      apiBaseEndpoint: "http://127.0.0.1:8080",
    }
  }
})
