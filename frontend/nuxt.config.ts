// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-04-03',
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss'],
  routeRules: {
    "/": { ssr: false },
  },
  // 静的生成時に useFetch の結果を payload に埋め込まない
  // (Go サーバへの取り込み結果確認ではクライアント側で /api を叩かせるため)
  experimental: {
    payloadExtraction: false,
  },
  runtimeConfig: {
    public: {
      apiBaseEndpoint: "http://127.0.0.1:8080",
    }
  }
})
