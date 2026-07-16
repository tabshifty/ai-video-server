import { describe, expect, it } from 'vitest'
import { shouldShowCrudCollectionSkeleton } from './crudCollectionState'

describe('基础 CRUD 集合加载状态', () => {
  const cases = [
    {
      name: '首次加载且没有缓存行时显示骨架',
      state: { loading: true, loaded: false, rowCount: 0 },
      expected: true
    },
    {
      name: '首次失败后重试且没有缓存行时继续显示骨架',
      state: { loading: true, loaded: true, rowCount: 0 },
      expected: true
    },
    {
      name: '后台刷新且有缓存行时保留列表',
      state: { loading: true, loaded: true, rowCount: 3 },
      expected: false
    },
    {
      name: '空闲空列表不显示骨架',
      state: { loading: false, loaded: true, rowCount: 0 },
      expected: false
    }
  ]

  it.each(cases)('$name', ({ state, expected }) => {
    expect(shouldShowCrudCollectionSkeleton(state)).toBe(expected)
  })
})
