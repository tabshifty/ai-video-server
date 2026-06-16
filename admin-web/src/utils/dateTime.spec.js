import { describe, expect, it } from 'vitest'
import { formatAdminDateTime } from './dateTime'

describe('admin date time formatter', () => {
  it('formats absolute times as YYYY/MM/DD HH:mm:ss in local time', () => {
    const date = new Date(2026, 5, 16, 16, 44, 32)

    expect(formatAdminDateTime(date)).toBe('2026/06/16 16:44:32')
    expect(formatAdminDateTime(date.getTime())).toBe('2026/06/16 16:44:32')
    expect(formatAdminDateTime(Math.floor(date.getTime() / 1000))).toBe('2026/06/16 16:44:32')
    expect(formatAdminDateTime(String(date.getTime()))).toBe('2026/06/16 16:44:32')
    expect(formatAdminDateTime(String(Math.floor(date.getTime() / 1000)))).toBe('2026/06/16 16:44:32')
  })

  it('uses the configured fallback for blank or invalid values', () => {
    expect(formatAdminDateTime('')).toBe('--')
    expect(formatAdminDateTime(undefined)).toBe('--')
    expect(formatAdminDateTime('bad-value')).toBe('--')
    expect(formatAdminDateTime(new Date('bad-value'))).toBe('--')
    expect(formatAdminDateTime(null, '暂无')).toBe('暂无')
  })
})
