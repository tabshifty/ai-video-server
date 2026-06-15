import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolveRouterHistoryBase } from './historyBase'

describe('resolveRouterHistoryBase', () => {
  it('保留 Vite 生产构建的 /admin/ 基路径', () => {
    expect(resolveRouterHistoryBase('/admin/')).toBe('/admin/')
  })

  it('缺省时回落到根路径', () => {
    expect(resolveRouterHistoryBase('')).toBe('/')
    expect(resolveRouterHistoryBase()).toBe('/')
  })

  it('注册短视频审核路由并隐藏壳层重复标题', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    expect(source).toContain("import ShortReview from '../views/ShortReview.vue'")
    expect(source).toContain("{ path: '/short-review', component: ShortReview, meta: { hideShellPageHeader: true } }")
  })
})
