import { defineConfig } from '@playwright/test'

// ローカル検証用: モック API + 静的ビルド済みフロントエンドを単一プロセスで起動して試験する
// 事前準備: make api-bin と cd frontend && npm run generate が必要 (make e2e がまとめて実行する)
const port = process.env.E2E_PORT ?? '8090'
const baseURL = `http://127.0.0.1:${port}`

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  workers: 1,
  outputDir: './test-results',
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL,
    locale: 'ja-JP',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  webServer: {
    command: `../build/bin/mf-importer-api mock --input-dir ../test --static-dir .output/public --addr 127.0.0.1:${port}`,
    url: `${baseURL}/api/details/count`,
    reuseExistingServer: !process.env.CI,
    timeout: 15_000,
  },
})
