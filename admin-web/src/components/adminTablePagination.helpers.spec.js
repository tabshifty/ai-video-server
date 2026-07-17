import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

import AdminTablePagination from './AdminTablePagination.vue'
import { resolvePageJump } from './adminTablePagination.helpers'

const source = readFileSync(new URL('./AdminTablePagination.vue', import.meta.url), 'utf8')
const style = source.match(/<style scoped>([\s\S]*?)<\/style>/)?.[1] || ''

function findRule(styleSource, selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = styleSource.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`))

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

describe('resolvePageJump', () => {
  it('通过真实 SFC 编译并在小于 1024px 时保持跳页输入和按钮至少 44px 高', () => {
    const mediaStart = style.indexOf('@media (max-width: 63.9375rem)')

    expect(AdminTablePagination).toBeTruthy()
    expect(mediaStart).toBeGreaterThan(-1)
    const mobileStyle = style.slice(mediaStart)
    expect(findRule(mobileStyle, '.admin-table-pagination__jump :deep(.el-input__wrapper)')).toContain('min-height: 44px')
    expect(findRule(mobileStyle, '.admin-table-pagination__jump :deep(.el-button)')).toContain('min-height: 44px')
  })

  it('在移动端让超宽分页从左侧开始并在组件内部水平滚动', () => {
    const mediaStart = style.indexOf('@media (max-width: 63.9375rem)')

    expect(mediaStart).toBeGreaterThan(-1)
    const mobileStyle = style.slice(mediaStart)
    const paginationRule = findRule(mobileStyle, '.admin-table-pagination')
    expect(paginationRule).toContain('justify-content: flex-start')
    expect(paginationRule).toContain('overflow-x: auto')
    expect(paginationRule).toContain('overscroll-behavior-inline: contain')
  })

  it('只在共享分页根节点内为上一页和下一页补齐中文名称与原生提示', () => {
    expect(source).toContain('ref="paginationRef"')
    expect(source).toContain("['.btn-prev', '上一页']")
    expect(source).toContain("['.btn-next', '下一页']")
    expect(source).toContain("button.setAttribute('aria-label', label)")
    expect(source).toContain("button.setAttribute('title', label)")
    expect(source).toContain('paginationRef.value?.$el')
    expect(source).not.toContain('MutationObserver')
    expect(source).not.toContain('document.querySelector')
  })

  it('为共享跳页输入提供可访问名称', () => {
    expect(source).toContain('aria-label="跳转页码"')
  })

  it('jumps to the requested page when the input is valid', () => {
    expect(
      resolvePageJump('5', {
        currentPage: 1,
        pageSize: 20,
        total: 240
      })
    ).toEqual({
      disabled: false,
      shouldJump: true,
      page: 5,
      displayValue: '5'
    })
  })

  it('clamps values smaller than the first page', () => {
    expect(
      resolvePageJump('0', {
        currentPage: 4,
        pageSize: 20,
        total: 240
      })
    ).toEqual({
      disabled: false,
      shouldJump: true,
      page: 1,
      displayValue: '1'
    })
  })

  it('clamps values larger than the last page', () => {
    expect(
      resolvePageJump('99', {
        currentPage: 2,
        pageSize: 20,
        total: 87
      })
    ).toEqual({
      disabled: false,
      shouldJump: true,
      page: 5,
      displayValue: '5'
    })
  })

  it('keeps the current page for empty or invalid values', () => {
    expect(
      resolvePageJump('  ', {
        currentPage: 3,
        pageSize: 20,
        total: 87
      })
    ).toEqual({
      disabled: false,
      shouldJump: false,
      page: 3,
      displayValue: '3'
    })

    expect(
      resolvePageJump('abc', {
        currentPage: 3,
        pageSize: 20,
        total: 87
      })
    ).toEqual({
      disabled: false,
      shouldJump: false,
      page: 3,
      displayValue: '3'
    })
  })

  it('disables jump when there are no records', () => {
    expect(
      resolvePageJump('2', {
        currentPage: 1,
        pageSize: 20,
        total: 0
      })
    ).toEqual({
      disabled: true,
      shouldJump: false,
      page: 1,
      displayValue: '1'
    })
  })
})
