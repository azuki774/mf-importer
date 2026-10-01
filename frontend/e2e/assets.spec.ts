import { test, expect } from '@playwright/test'

const api = '/api/v2/financial-assets'
const balanceRoute = `**${api}/balances?*`
const detailRoute = `**${api}/snapshots/*`
// Uses only the fixed synthetic financial repository in the mock API.

test('ナビゲーションから統合明細を表示し、総資産と全ページの明細合計を分ける', async ({ page }) => {
  const requests: string[] = []
  page.on('request', request => {
    if (request.url().includes(api)) requests.push(request.url())
  })
  await page.goto('/')
  await page.getByRole('link', { name: '金融資産', exact: true }).click()
  await expect(page.getByRole('heading', { name: '金融資産', exact: true })).toBeVisible()
  await expect(page.getByText('64 件 / 取得済み 64 件')).toBeVisible()
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(50)
  await expect(page.getByTestId('asset-total')).toHaveText('320 円')
  await expect(page.getByTestId('holdings-total-value')).toHaveText('290 円')
  await expect(page.getByTestId('holdings-total-cost')).toHaveText('—')
  await expect(page.getByTestId('holdings-total-pnl')).toHaveText('—')
  await expect(page.getByRole('article', { name: 'SBIの資産' })).toContainText('2000/02/01 12:00:00')
  await expect(page.getByRole('article', { name: 'NRKNの資産' })).toContainText('2000/02/02 12:00:00')
  expect(requests.filter(url => url.includes('/balances?'))).toHaveLength(1)
  expect(requests.filter(url => url.includes('/snapshots/'))).toHaveLength(2)
  expect(requests.some(url => /\/snapshots(?:\?|$)/.test(url))).toBe(false)

  await page.getByLabel('商品検索').fill('ダミー共通商品')
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(3)
  await expect(page.getByTestId('holdings-total-value')).toHaveText('190 円')
  await expect(page.getByTestId('holdings-total-cost')).toHaveText('176 円')
  await expect(page.getByTestId('holdings-total-pnl')).toHaveText('+14 円')
  await expect(page.getByTestId('asset-total')).toHaveText('320 円')
  await page.screenshot({ path: 'e2e/artifacts/assets-desktop.png', fullPage: true })
})

test('明細のページ切り替え・検索・区分・並べ替えと全件表示', async ({ page }) => {
  await page.goto('/assets')
  const rows = page.getByTestId('asset-holding-row')
  await expect(rows).toHaveCount(50)
  await expect(page.getByRole('button', { name: '← 前' })).toBeDisabled()
  await page.getByRole('button', { name: '次 →' }).click()
  await expect(rows).toHaveCount(14)
  await expect(page.getByText('2 / 2')).toBeVisible()
  await expect(page.getByRole('button', { name: '次 →' })).toBeDisabled()
  await expect(page.getByTestId('holdings-total-value')).toHaveText('290 円')
  await page.getByLabel('商品検索').fill('ダミー同名商品')
  await expect(rows).toHaveCount(2)
  await expect(page.getByText('1 / 1')).toBeVisible()
  await page.getByLabel('商品検索').fill('testfigi0100')
  await expect(rows).toHaveCount(3)
  await page.getByLabel('保有区分', { exact: true }).selectOption('old_nisa_funds')
  await expect(rows).toHaveCount(1)
  await page.getByLabel('商品検索').fill('存在しないダミー商品')
  await expect(page.getByText('条件に一致する商品がありません')).toBeVisible()
  await page.getByLabel('商品検索').clear()
  await page.getByLabel('保有区分', { exact: true }).selectOption('all')
  await page.getByLabel('表示件数').selectOption('0')
  await expect(rows).toHaveCount(64)
  await page.getByRole('button', { name: /^取得価額/ }).click()
  await expect(rows.last()).toContainText('ダミー損益未取得商品')
  await page.getByRole('button', { name: /^取得価額/ }).click()
  await expect(rows.last()).toContainText('ダミー損益未取得商品')
  await expect(page.getByRole('columnheader', { name: /^取得価額/ })).toHaveAttribute('aria-sort', 'ascending')
  await expect(rows.first()).toContainText('0 円')
})

test('明細展開で未取得とゼロ、取得精度、数量、基準日を確認できる', async ({ page }) => {
  await page.goto('/assets')
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(50)
  await page.getByLabel('商品検索').fill('ダミー損益未取得商品')
  await page.getByRole('button', { name: /ダミー損益未取得商品/ }).click()
  await expect(page.getByText('0.000001', { exact: true })).toBeVisible()
  const details = page.locator('tr').filter({ has: page.locator('dt', { hasText: '数量（取得元の単位）' }) })
  await expect(details.locator('div').filter({ has: page.locator('dt', { hasText: '取得価額（円・取得精度）' }) }).getByRole('definition')).toHaveText('—')
  await page.getByLabel('商品検索').fill('ダミー識別子なし商品')
  await page.getByRole('button', { name: /ダミー識別子なし商品/ }).click()
  await expect(page.getByTestId('holdings-total-cost')).toHaveText('0 円')
  await expect(details.locator('div').filter({ has: page.locator('dt', { hasText: '評価損益率' }) }).getByRole('definition')).toHaveText('—')
  await page.getByLabel('商品検索').fill('TEST101')
  await page.getByRole('button', { name: /ダミー共通商品/ }).click()
  await expect(page.getByText('2000-02-01', { exact: true })).toBeVisible()
})

test('データ元の選択をURLへ反映し、直接アクセスでも復元する', async ({ page }) => {
  await page.goto('/assets?source=nrkn')
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(4)
  await expect(page.getByTestId('asset-total')).toHaveText('70 円')
  await expect(page.getByRole('checkbox', { name: 'SBI', exact: true })).not.toBeChecked()
  await page.getByRole('checkbox', { name: 'SBI', exact: true }).check()
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(50)
  await page.reload()
  await expect(page.getByText('64 件 / 取得済み 64 件')).toBeVisible()
  await page.getByRole('checkbox', { name: 'SBI', exact: true }).uncheck()
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(4)
  await page.getByRole('checkbox', { name: 'NRKN', exact: true }).uncheck()
  await expect(page.getByText('データ元を選択してください。')).toBeVisible()
  await expect(page.getByTestId('asset-total')).toHaveCount(0)
})

test('片方・両方にデータがない場合をゼロと区別する', async ({ page }) => {
  let at = '1999-12-31T12:00:00+09:00'
  await page.route(balanceRoute, async (route) => {
    const url = new URL(route.request().url())
    url.searchParams.set('at', at)
    const response = await route.fetch({ url: url.toString() })
    await route.fulfill({ response })
  })
  await page.goto('/assets')
  await expect(page.getByText('NRKNのデータがありません。総資産は未確定です。')).toBeVisible()
  await expect(page.getByTestId('asset-total')).toHaveText('—')
  await expect(page.getByRole('article', { name: 'SBIの資産' })).toContainText('0 円')
  await expect(page.getByTestId('holdings-total-value')).toHaveText('—')
  at = '1999-12-30T12:00:00+09:00'
  await page.getByRole('button', { name: '再読み込み' }).click()
  await expect(page.getByText('取り込み済みの資産データがありません。')).toBeVisible()
})

test('一部の明細取得失敗では部分集計せず、再試行で回復する', async ({ page }) => {
  let fail = true
  await page.route(detailRoute, async (route) => {
    if (fail && decodeURIComponent(route.request().url()).includes(':sbi:')) {
      await route.fulfill({ status: 500, json: { error: 'synthetic failure' } })
    } else await route.continue()
  })
  await page.goto('/assets')
  await expect(page.getByRole('alert')).toContainText('SBIの保有商品を取得できませんでした')
  await expect(page.getByTestId('asset-total')).toHaveText('320 円')
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(4)
  await expect(page.getByTestId('holdings-total-value')).toHaveText('—')
  fail = false
  await page.getByRole('button', { name: '再試行' }).click()
  await expect(page.getByText('64 件 / 取得済み 64 件')).toBeVisible()
  await expect(page.getByRole('alert')).toHaveCount(0)
})

test('残高の更新失敗は前回表示を保持し、条件変更時には消す', async ({ page }) => {
  await page.goto('/assets')
  await expect(page.getByTestId('asset-total')).toHaveText('320 円')
  await page.route(balanceRoute, route => route.fulfill({ status: 500, json: { error: 'synthetic failure' } }))
  await page.getByRole('button', { name: '再読み込み' }).click()
  await expect(page.getByRole('alert')).toContainText('更新に失敗しました。前回の表示を保持しています。')
  await expect(page.getByTestId('asset-total')).toHaveText('320 円')
  await page.getByRole('checkbox', { name: 'SBI', exact: true }).uncheck()
  await expect(page.getByRole('alert')).toContainText('資産データを取得できませんでした。')
  await expect(page.getByTestId('asset-total')).toHaveCount(0)
  await page.unroute(balanceRoute)
  await page.getByRole('button', { name: '再試行' }).click()
  await expect(page.getByTestId('asset-total')).toHaveText('70 円')
})

test('条件変更前の遅い明細応答が新しい結果を上書きしない', async ({ page }) => {
  let release!: () => void
  const gate = new Promise<void>(resolve => { release = resolve })
  let entered!: () => void
  const requested = new Promise<void>(resolve => { entered = resolve })
  await page.route(detailRoute, async (route) => {
    if (decodeURIComponent(route.request().url()).includes(':sbi:')) {
      const response = await route.fetch()
      entered()
      await gate
      await route.fulfill({ response }).catch(() => {}) // The old request may have been aborted.
    } else await route.continue()
  })
  await page.goto('/assets')
  await requested
  await page.getByRole('checkbox', { name: 'SBI', exact: true }).uncheck()
  await expect(page.getByTestId('asset-total')).toHaveText('70 円')
  release()
  await expect(page.getByTestId('asset-holding-row')).toHaveCount(4)
  await expect(page.getByRole('article', { name: 'SBIの資産' })).toHaveCount(0)
})

test('スマートフォンで一覧・明細展開を操作できる', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/assets')
  await expect(page.getByTestId('asset-total')).toHaveText('320 円')
  await page.getByLabel('商品検索').fill('ダミー共通商品')
  await page.getByRole('button', { name: /ダミー共通商品/ }).first().click()
  await expect(page.getByText('数量（取得元の単位）')).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true)
  await page.screenshot({ path: 'e2e/artifacts/assets-mobile.png', fullPage: true })
})
