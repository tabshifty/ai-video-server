import { describe, expect, it } from 'vitest'

import { resolvePageJump } from './adminTablePagination.helpers'

describe('resolvePageJump', () => {
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
