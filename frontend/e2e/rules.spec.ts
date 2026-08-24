import { test, expect } from '@playwright/test'

// 一覧のデータロードが完了するまで待ってから行数を数えるためのヘルパー
async function waitForRulesLoaded(page: import('@playwright/test').Page) {
  await expect(page.getByRole('cell', { name: '携帯電話' })).toBeVisible()
  return page.locator('tbody tr').count()
}

test('ルール一覧にフィクスチャが表示される', async ({ page }) => {
  await page.goto('/rules')

  await expect(page.getByRole('heading', { name: 'ルール一覧' })).toBeVisible()

  const rows = page.locator('tbody tr')
  await expect(rows).toHaveCount(2)

  await expect(rows.nth(0)).toContainText('携帯電話')
  await expect(rows.nth(0)).toContainText('m_category')
  await expect(rows.nth(0)).toContainText('完全')
  await expect(rows.nth(1)).toContainText('WASABI')
  await expect(rows.nth(1)).toContainText('部分')

  await page.screenshot({ path: 'e2e/artifacts/rules.png', fullPage: true })
})

test('ルールを追加すると一覧に反映される', async ({ page }) => {
  await page.goto('/rules')
  const before = await waitForRulesLoaded(page)

  const form = page.locator('form')
  await form.locator('select').selectOption('name')
  await form.getByPlaceholder('値を入力').fill('ローソン')
  await form.getByPlaceholder('例: 101').fill('101')

  await page.getByRole('button', { name: '追加する' }).click()

  await expect(page.getByText('ルールを追加しました')).toBeVisible()
  const rows = page.locator('tbody tr')
  await expect(rows).toHaveCount(before + 1)
  await expect(rows.nth(before)).toContainText('ローソン')

  await page.screenshot({ path: 'e2e/artifacts/rules-added.png', fullPage: true })
})

test('ルールを削除できる', async ({ page }) => {
  await page.goto('/rules')
  const before = await waitForRulesLoaded(page)
  await expect(before).toBeGreaterThan(0)

  await page.locator('tbody tr').first().getByRole('button', { name: '削除' }).click()
  await page.getByRole('button', { name: '削除する' }).click()

  await expect(page.getByText('ルールを削除しました')).toBeVisible()
  await expect(page.locator('tbody tr')).toHaveCount(before - 1)
})

test('列ヘッダーで並び替えできる', async ({ page }) => {
  await page.goto('/rules')
  const rows = page.locator('tbody tr')
  // モックサーバーはテスト間で状態を共有するため、件数と相対順序だけで検証する。
  // 空状態は tbody に 1 行(「ルールがありません」)のみなので、2 行以上でロード完了とみなす
  await expect.poll(async () => rows.count(), { timeout: 5_000 }).toBeGreaterThan(1)
  await expect(page.getByRole('columnheader', { name: /^ID/ })).toHaveAttribute('aria-sort', 'ascending')

  const valueHeader = page.getByRole('columnheader', { name: /^値/ })

  const ascRequest = page.waitForRequest(
    (req) => req.url().includes('/api/rules?') && req.url().includes('sort=value') && req.url().includes('order=asc')
  )
  await valueHeader.click()
  await ascRequest
  await expect(valueHeader).toHaveAttribute('aria-sort', 'ascending')
  const ascFirst = await rows.first().textContent()

  const descRequest = page.waitForRequest(
    (req) => req.url().includes('/api/rules?') && req.url().includes('sort=value') && req.url().includes('order=desc')
  )
  await valueHeader.click()
  await descRequest
  await expect(valueHeader).toHaveAttribute('aria-sort', 'descending')

  // 値の重複がないフィクスチャなので、昇順と降順で先頭行が入れ替わる
  // (textContents の生読みは再取得前に競合するため、リトライ付きで検証する)
  await expect(rows.first()).not.toHaveText(ascFirst ?? '', { timeout: 10_000 })
})
