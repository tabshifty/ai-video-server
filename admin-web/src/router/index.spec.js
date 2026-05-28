import { describe, expect, it } from 'vitest'
import { resolveRouterHistoryBase } from './historyBase'

describe('resolveRouterHistoryBase', () => {
  it('保留 Vite 生产构建的 /admin/ 基路径', () => {
    expect(resolveRouterHistoryBase('/admin/')).toBe('/admin/')
  })

  it('缺省时回落到根路径', () => {
    expect(resolveRouterHistoryBase('')).toBe('/')
    expect(resolveRouterHistoryBase()).toBe('/')
  })
})
