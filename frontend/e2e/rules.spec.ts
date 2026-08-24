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
