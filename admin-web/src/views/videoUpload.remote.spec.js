import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

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
})
