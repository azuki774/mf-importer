import { test, expect } from '@playwright/test'

// fixture: ../test/cf.csv (200件, 2024-01-01..2024-07-18)
// 表示件数の既定は 20 なので 10 ページ構成になる
// fixture の 3 件に 1 件ルールにより id が 3 の倍数のみ import 済み日付を持つ

test('取り込み履歴テーブルにモックデータが表示される', async ({ page }) => {
  await page.goto('/')

  await expect(page.getByRole('heading', { name: 'MoneyForward CSV 取り込み' })).toBeVisible()
  await expect(page.getByText('全 200 件')).toBeVisible()

  const rows = page.locator('tbody tr')
  await expect(rows).toHaveCount(20)

  await expect(rows.first()).toContainText('テスト明細 200')
  await expect(rows.first()).toContainText('3,087')
  await expect(rows.nth(19)).toContainText('テスト明細 181')

  await expect(rows.first().locator('td').nth(4)).toHaveText('—')
  await expect(rows.nth(1).locator('td').nth(4)).toHaveText('—')
  await expect(rows.nth(2).locator('td').nth(4)).not.toHaveText('—')

  await page.screenshot({ path: 'e2e/artifacts/index-history.png', fullPage: true })
})

test('表示件数を変更すると limit を変えて再取得する', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('全 200 件')).toBeVisible()

  const requestPromise = page.waitForRequest((req) =>
    req.url().includes('/api/details?') && req.url().includes('limit=10')
  )
  await page.locator('select').selectOption('10件')
  const request = await requestPromise

  expect(request.url()).toContain('offset=0')
  await expect(page.locator('tbody tr')).toHaveCount(10)
  await expect(page.getByText('1 / 20')).toBeVisible()

  await page.locator('select').selectOption('50件')
  await expect(page.locator('tbody tr')).toHaveCount(50)
  await expect(page.getByText('1 / 4')).toBeVisible()
})

test('ページネーションで前後ページへ移動できる', async ({ page }) => {
  await page.goto('/')
  const rows = page.locator('tbody tr')
  await expect(rows).toHaveCount(20)

  const prev = page.getByRole('button', { name: '← 前' })
  const next = page.getByRole('button', { name: '次 →' })
  await expect(prev).toBeDisabled()
  await expect(next).toBeEnabled()

  await next.click()
  await expect(page.getByText('2 / 10')).toBeVisible()
  await expect(rows).toHaveCount(20)
  await expect(rows.first()).toContainText('テスト明細 180')
  await expect(prev).toBeEnabled()

  await prev.click()
  await expect(page.getByText('1 / 10')).toBeVisible()
  await expect(rows.first()).toContainText('テスト明細 200')
  await expect(prev).toBeDisabled()

  // 最終ページ (50件表示で 4 ページ目) で次ボタンが無効になることも確認
  await page.locator('select').selectOption('50件')
  await expect(page.getByText('1 / 4')).toBeVisible()
  await next.click()
  await next.click()
  await next.click()
  await expect(page.getByText('4 / 4')).toBeVisible()
  await expect(rows).toHaveCount(50)
  await expect(rows.first()).toContainText('テスト明細 050')
  await expect(next).toBeDisabled()
  await expect(prev).toBeEnabled()
})

test('再判定操作で import 日付がクリアされる', async ({ page }) => {
  await page.goto('/')
  const judgedRow = page.locator('tbody tr').nth(2) // id 198 (3 の倍数で判定済み)
  await expect(judgedRow).toContainText('テスト明細 198')
  await expect(judgedRow.locator('td').nth(4)).not.toHaveText('—')

  await judgedRow.getByRole('button', { name: '再判定' }).click()
  await page.getByRole('button', { name: '再判定する' }).click()

  await expect(page.getByText('再判定対象に設定しました')).toBeVisible()
  await expect(judgedRow.locator('td').nth(4)).toHaveText('—', { timeout: 10_000 })
})
