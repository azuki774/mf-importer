import { test, expect } from '@playwright/test'
import {
  assetPnlRate, compareAssetDecimals, formatAssetDecimal, formatAssetMoney, sumAssetDecimals,
} from '../utils/financialAssetDecimal'

test('合成値の小数・大きな整数を丸めずに集計し、表示時だけ丸める', () => {
  expect(sumAssetDecimals(['0.1', '0.2'])).toBe('0.3')
  expect(sumAssetDecimals(['9007199254740993', '0.01'])).toBe('9007199254740993.01')
  expect(formatAssetMoney('9007199254740993.51')).toBe('9,007,199,254,740,994 円')
  expect(formatAssetDecimal('0.000001')).toBe('0.000001')
  expect(formatAssetMoney('-0.1', true)).toBe('0 円')
  expect(formatAssetMoney('1.5', true)).toBe('+2 円')
  expect(formatAssetMoney('-1.5', true)).toBe('-2 円')
  expect(sumAssetDecimals(['-0.1', '0.1'])).toBe('0')
})

test('未取得値・ゼロと損益率の計算条件を区別する', () => {
  expect(sumAssetDecimals(['1', null])).toBeNull()
  expect(sumAssetDecimals([])).toBe('0')
  expect(formatAssetMoney(null)).toBe('—')
  expect(formatAssetMoney('0')).toBe('0 円')
  expect(assetPnlRate('0', '1')).toBeNull()
  expect(assetPnlRate('-1', '1')).toBeNull()
  expect(assetPnlRate(null, '1')).toBeNull()
  expect(assetPnlRate('1', null)).toBeNull()
  expect(formatAssetDecimal(assetPnlRate('3', '1'), 2, true)).toBe('+33.33')
  expect(assetPnlRate('2', '-1')).toBe('-50')
})

test('金額は数値順に比較し、未取得値は昇順・降順とも末尾にする', () => {
  const values = [null, '10', '2', '9007199254740993', '9007199254740992.99', '-1', '0']
  expect([...values].sort((a, b) => compareAssetDecimals(a, b, 'asc'))).toEqual([
    '-1', '0', '2', '10', '9007199254740992.99', '9007199254740993', null,
  ])
  expect([...values].sort((a, b) => compareAssetDecimals(a, b, 'desc'))).toEqual([
    '9007199254740993', '9007199254740992.99', '10', '2', '0', '-1', null,
  ])
})
