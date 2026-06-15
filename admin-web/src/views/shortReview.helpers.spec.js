import { describe, expect, it, vi } from 'vitest'
import {
  buildShortReviewQuery,
  findVideoIndexByID,
  hasMoreReviewPages,
  isReviewableShortVideo,
  normalizeReviewPosition,
  readStoredShortReviewPosition,
  readStoredShortReviewSound,
  resolveInitialReviewIndex,
  resolveNextReviewStep,
  SHORT_REVIEW_POSITION_KEY,
  SHORT_REVIEW_SOUND_KEY,
  writeStoredShortReviewPosition,
  writeStoredShortReviewSound
} from './shortReview.helpers'

describe('short review helpers', () => {
  const items = [
    { id: 'video-1', type: 'short', status: 'ready' },
    { id: 'video-2', type: 'short', status: 'ready' }
  ]

  it('limits reviewable videos to ready shorts', () => {
    expect(isReviewableShortVideo({ type: 'short', status: 'ready' })).toBe(true)
    expect(isReviewableShortVideo({ type: 'movie', status: 'ready' })).toBe(false)
    expect(isReviewableShortVideo({ type: 'short', status: 'processing' })).toBe(false)
  })

  it('restores saved position when the video exists, otherwise starts from the first loaded item', () => {
    expect(findVideoIndexByID(items, 'video-2')).toBe(1)
    expect(resolveInitialReviewIndex(items, 'video-2')).toBe(1)
    expect(resolveInitialReviewIndex(items, 'missing')).toBe(0)
    expect(resolveInitialReviewIndex([], 'video-2')).toBe(-1)
  })

  it('normalizes persisted position from the current JSON format and legacy plain ids', () => {
    expect(normalizeReviewPosition('video-7')).toEqual({ videoID: 'video-7', page: 1 })
    expect(normalizeReviewPosition('{"video_id":"video-8","page":4}')).toEqual({ videoID: 'video-8', page: 4 })
    expect(normalizeReviewPosition({ videoID: 'video-9', page: '3' })).toEqual({ videoID: 'video-9', page: 3 })
  })

  it('advances inside the loaded queue before loading more or ending', () => {
    expect(resolveNextReviewStep({ currentIndex: 0, loadedCount: 2, totalCount: 5 })).toEqual({
      type: 'select',
      index: 1
    })
    expect(resolveNextReviewStep({ currentIndex: 1, loadedCount: 2, totalCount: 5 })).toEqual({
      type: 'load-more'
    })
    expect(resolveNextReviewStep({ currentIndex: 1, loadedCount: 2, totalCount: 2 })).toEqual({
      type: 'end'
    })
  })

  it('does not request another page while loading more or with no loaded items', () => {
    expect(hasMoreReviewPages({ loadedCount: 2, totalCount: 5, loadingMore: true })).toBe(false)
    expect(hasMoreReviewPages({ loadedCount: 0, totalCount: 5 })).toBe(false)
  })

  it('builds a fixed short ready query with optional keyword', () => {
    expect(buildShortReviewQuery({ page: 3, pageSize: 40, keyword: ' 猫 ' })).toEqual({
      page: 3,
      page_size: 40,
      q: '猫',
      type: 'short',
      status: 'ready'
    })
  })

  it('persists review position and sound preference defensively', () => {
    const storage = new Map()
    const adapter = {
      getItem: (key) => storage.get(key) || '',
      setItem: (key, value) => storage.set(key, value),
      removeItem: (key) => storage.delete(key)
    }

    writeStoredShortReviewPosition({ videoID: ' video-9 ', page: 3 }, adapter)
    expect(storage.get(SHORT_REVIEW_POSITION_KEY)).toBe('{"video_id":"video-9","page":3}')
    expect(readStoredShortReviewPosition(adapter)).toEqual({ videoID: 'video-9', page: 3 })

    writeStoredShortReviewPosition({ videoID: '' }, adapter)
    expect(storage.has(SHORT_REVIEW_POSITION_KEY)).toBe(false)

    writeStoredShortReviewSound(true, adapter)
    expect(storage.get(SHORT_REVIEW_SOUND_KEY)).toBe('1')
    expect(readStoredShortReviewSound(adapter)).toBe(true)
  })

  it('ignores storage failures', () => {
    const brokenStorage = {
      getItem: vi.fn(() => {
        throw new Error('blocked')
      }),
      setItem: vi.fn(() => {
        throw new Error('blocked')
      }),
      removeItem: vi.fn(() => {
        throw new Error('blocked')
      })
    }

    expect(readStoredShortReviewPosition(brokenStorage)).toEqual({ videoID: '', page: 1 })
    expect(readStoredShortReviewSound(brokenStorage)).toBe(false)
    expect(() => writeStoredShortReviewPosition('video-1', brokenStorage)).not.toThrow()
    expect(() => writeStoredShortReviewSound(true, brokenStorage)).not.toThrow()
  })
})
