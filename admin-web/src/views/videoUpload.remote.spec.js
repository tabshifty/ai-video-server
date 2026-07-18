import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import * as remoteOptions from './videoUpload.remote'
import {
  createRemoteSuggestionLoader,
  mergeRemoteStringOptions,
  mergeRemoteValueOptions
} from './videoUpload.remote'

describe('mergeRemoteStringOptions', () => {
  it('deduplicates incoming string options case-insensitively', () => {
    const got = mergeRemoteStringOptions(['剧情', '动作'], [' 动作 ', '爱情', 'ACTION'])

    expect(got).toEqual(['剧情', '动作', '爱情', 'ACTION'])
  })
})

describe('mergeRemoteValueOptions', () => {
  it('preserves existing selected options while merging latest remote candidates', () => {
    const got = mergeRemoteValueOptions(
      [{ value: 'existing', label: '已选' }],
      [
        { value: 'new', label: '新结果' },
        { value: 'existing', label: '已选（更新标签）' }
      ]
    )

    expect(got).toEqual([
      { value: 'existing', label: '已选（更新标签）' },
      { value: 'new', label: '新结果' }
    ])
  })
})

describe('filterRemoteOptionsByValues', () => {
  it('只保留当前已选的字符串候选并规范化空白和大小写', () => {
    expect(typeof remoteOptions.filterRemoteOptionsByValues).toBe('function')
    expect(remoteOptions.filterRemoteOptionsByValues(
      [' 剧情 ', '动作', '爱情'],
      [' 动作 ']
    )).toEqual(['动作'])
  })

  it('按值字段保留当前已选对象并忽略空值', () => {
    const selected = { value: 'collection-b', label: '乙' }

    expect(remoteOptions.filterRemoteOptionsByValues(
      [
        { value: 'collection-a', label: '甲' },
        selected,
        { value: '', label: '无效' }
      ],
      [' COLLECTION-B '],
      (item) => item.value
    )).toEqual([selected])
    expect(remoteOptions.filterRemoteOptionsByValues(
      [selected],
      ['collection-b'],
      (item) => item.value
    )[0]).toBe(selected)
  })

  it('在候选或已选值缺省时返回空数组', () => {
    expect(remoteOptions.filterRemoteOptionsByValues()).toEqual([])
    expect(remoteOptions.filterRemoteOptionsByValues([], undefined)).toEqual([])
    expect(remoteOptions.filterRemoteOptionsByValues(undefined, [])).toEqual([])
  })
})

describe('createRemoteSuggestionLoader', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('loads default suggestions for empty query after debounce', async () => {
    const fetcher = vi.fn().mockResolvedValue(['热门标签'])
    const loadingStates = []
    let options = ['已选标签']

    const loadSuggestions = createRemoteSuggestionLoader({
      delay: 120,
      fetcher,
      getOptions: () => options,
      setOptions: (next) => {
        options = next
      },
      setLoading: (next) => {
        loadingStates.push(next)
      },
      mergeOptions: mergeRemoteStringOptions
    })

    loadSuggestions('')
    await vi.advanceTimersByTimeAsync(120)
    await Promise.resolve()

    expect(fetcher).toHaveBeenCalledWith('')
    expect(options).toEqual(['已选标签', '热门标签'])
    expect(loadingStates).toEqual([true, false])
  })

  it('keeps only the latest remote result when requests resolve out of order', async () => {
    const resolvers = []
    const fetcher = vi.fn().mockImplementation(
      (query) =>
        new Promise((resolve) => {
          resolvers.push({ query, resolve })
        })
    )
    let options = []

    const loadSuggestions = createRemoteSuggestionLoader({
      delay: 50,
      fetcher,
      getOptions: () => options,
      setOptions: (next) => {
        options = next
      },
      setLoading: () => {},
      mergeOptions: mergeRemoteStringOptions
    })

    loadSuggestions('动')
    await vi.advanceTimersByTimeAsync(50)

    loadSuggestions('动作')
    await vi.advanceTimersByTimeAsync(50)

    expect(fetcher.mock.calls).toEqual([['动'], ['动作']])

    resolvers[0].resolve(['动作片'])
    await Promise.resolve()
    expect(options).toEqual([])

    resolvers[1].resolve(['动作'])
    await Promise.resolve()
    expect(options).toEqual(['动作'])
  })

  it('filters stale unselected history while retaining selected labels through query changes', async () => {
    const selected = { value: 'selected-old', label: '已选旧名称' }
    const staleHistory = { value: 'history-a', label: '未选历史 A' }
    const fetcher = vi.fn().mockImplementation((query) => {
      const results = {
        A: [{ value: 'incoming-a', label: '当前 A' }],
        B: [{ value: 'incoming-b', label: '当前 B' }],
        '': [{ value: 'incoming-default', label: '默认候选' }]
      }
      return Promise.resolve(results[query])
    })
    const snapshots = []
    let options = [selected, staleHistory]
    const selectedValues = ['selected-old']

    const loadSuggestions = createRemoteSuggestionLoader({
      delay: 80,
      fetcher,
      getOptions: () => remoteOptions.filterRemoteOptionsByValues(
        options,
        selectedValues,
        (item) => item.value
      ),
      setOptions: (next) => {
        options = next
        snapshots.push(next)
      },
      setLoading: () => {},
      mergeOptions: mergeRemoteValueOptions
    })

    for (const query of ['A', 'B', '']) {
      loadSuggestions(query)
      await vi.advanceTimersByTimeAsync(80)
    }

    expect(fetcher).toHaveBeenCalledWith('A')
    expect(fetcher).toHaveBeenCalledWith('B')
    expect(fetcher).toHaveBeenCalledWith('')
    expect(snapshots).toHaveLength(3)
    expect(snapshots).toEqual([
      [selected, { value: 'incoming-a', label: '当前 A' }],
      [selected, { value: 'incoming-b', label: '当前 B' }],
      [selected, { value: 'incoming-default', label: '默认候选' }]
    ])
    for (const snapshot of snapshots) {
      expect(snapshot).not.toContain(staleHistory)
      expect(snapshot).toContainEqual({ value: 'selected-old', label: '已选旧名称' })
    }
  })
})
