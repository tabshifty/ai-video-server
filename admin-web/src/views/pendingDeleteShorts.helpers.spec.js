import { describe, expect, it } from 'vitest'
import {
  formatShortVideoDuration,
  isPendingDeleteShortVideo,
  resolvePendingDeleteSelectionIndex,
  resolveTotalPendingDeletePages
} from './pendingDeleteShorts.helpers'

describe('pending delete shorts helpers', () => {
  it('识别待删除短视频', () => {
    expect(isPendingDeleteShortVideo({ type: 'short', status: 'pending_delete' })).toBe(true)
    expect(isPendingDeleteShortVideo({ type: 'short', status: 'ready' })).toBe(false)
    expect(isPendingDeleteShortVideo({ type: 'movie', status: 'pending_delete' })).toBe(false)
  })

  it('计算分页和处理后选中位置', () => {
    expect(resolveTotalPendingDeletePages(0, 20)).toBe(1)
    expect(resolveTotalPendingDeletePages(41, 20)).toBe(3)
    expect(resolvePendingDeleteSelectionIndex([{ id: 1 }, { id: 2 }], 5)).toBe(1)
    expect(resolvePendingDeleteSelectionIndex([], 0)).toBe(-1)
  })

  it('格式化短视频时长', () => {
    expect(formatShortVideoDuration(0)).toBe('--:--')
    expect(formatShortVideoDuration(65)).toBe('1:05')
  })
})
