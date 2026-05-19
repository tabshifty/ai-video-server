import { describe, expect, it, vi } from 'vitest'
import {
  buildAVManualScrapeRoute,
  buildMovieManualScrapeRoute,
  canManuallyEditVideoStatus,
  extractTvPendingDiagnostics,
  getManualVideoStatusOptions,
  getManualVideoStatusValue,
  getVideoStatusMeta,
  getVideoThumbnailPlaceholder,
  getVideoThumbnailURL,
  shouldShowVideoThumbnail,
  teardownPreviewPlayer
} from './videoList.helpers'

describe('videoList helpers', () => {
  it('includes tv_pending status label and tag type', () => {
    expect(getVideoStatusMeta('tv_pending')).toEqual({
      label: '待绑定',
      tagType: 'warning'
    })
  })

  it('returns only editable manual status options', () => {
    expect(getManualVideoStatusOptions()).toEqual([
      { value: 'uploaded', label: '已上传' },
      { value: 'scraping', label: '刮削中' },
      { value: 'tv_pending', label: '待绑定' },
      { value: 'ready', label: '可播放' },
      { value: 'failed', label: '失败' }
    ])
  })

  it('blocks manual edits for processing only', () => {
    expect(canManuallyEditVideoStatus('processing')).toBe(false)
    expect(canManuallyEditVideoStatus('ready')).toBe(true)
  })

  it('omits processing from the manual update payload value', () => {
    expect(getManualVideoStatusValue('processing')).toBe('')
    expect(getManualVideoStatusValue('ready')).toBe('ready')
  })

  it('extracts tv pending diagnostics from metadata', () => {
    const diagnostics = extractTvPendingDiagnostics({
      scrape_error: 'ambiguous tv candidate',
      scrape_stage: 'candidate_ambiguous',
      parsed_title: '三体',
      parsed_season_number: 1,
      parsed_episode_number: 2,
      candidate_count: 2,
      candidate_preview: [{ tmdb_id: 1, title: '三体' }]
    })

    expect(diagnostics).toEqual({
      error: 'ambiguous tv candidate',
      stage: 'candidate_ambiguous',
      parsedTitle: '三体',
      parsedSeasonNumber: 1,
      parsedEpisodeNumber: 2,
      candidateCount: 2,
      candidatePreview: [{ tmdb_id: 1, title: '三体' }]
    })
  })

  it('tears down the preview player and releases the media source', () => {
    const pause = vi.fn()
    const removeAttribute = vi.fn()
    const load = vi.fn()
    const player = {
      pause,
      removeAttribute,
      load,
      src: 'https://example.com/video.mp4'
    }

    teardownPreviewPlayer(player)

    expect(pause).toHaveBeenCalledTimes(1)
    expect(removeAttribute).toHaveBeenCalledWith('src')
    expect(player.src).toBe('')
    expect(load).toHaveBeenCalledTimes(1)
  })

  it('ignores null players and partial player objects', () => {
    expect(() => teardownPreviewPlayer(null)).not.toThrow()
    expect(() => teardownPreviewPlayer({ src: 'blob:1' })).not.toThrow()
  })

  it('builds the AV manual scrape route from video detail metadata', () => {
    expect(buildAVManualScrapeRoute({
      id: 'video-1',
      title: '已刮削标题',
      metadata: {
        external_id: 'MXGS-888',
        av_code: 'MXGS-888'
      }
    })).toEqual({
      path: '/av-scrape',
      query: {
        video_id: 'video-1',
        external_id: 'MXGS-888',
        title: 'MXGS-888'
      }
    })
  })

  it('builds the movie manual scrape route from video detail metadata', () => {
    expect(buildMovieManualScrapeRoute({
      id: 'video-1',
      title: '盗梦空间',
      type: 'movie',
      metadata: {
        release_date: '2010-07-16'
      }
    })).toEqual({
      path: '/scrape',
      query: {
        video_id: 'video-1',
        type: 'movie',
        title: '盗梦空间',
        year: 2010
      }
    })
  })

  it('derives movie manual scrape year from nested tmdb release date', () => {
    expect(buildMovieManualScrapeRoute({
      id: 'video-2',
      title: '星际穿越',
      type: 'movie',
      metadata: {
        tmdb: {
          release_date: '2014-11-07'
        }
      }
    }).query).toEqual({
      video_id: 'video-2',
      type: 'movie',
      title: '星际穿越',
      year: 2014
    })
  })

  it('omits invalid movie manual scrape years', () => {
    expect(buildMovieManualScrapeRoute({
      id: 'video-3',
      title: '无年份电影',
      type: 'movie',
      metadata: {
        release_date: 'unknown',
        tmdb: {
          release_date: ''
        }
      }
    })).toEqual({
      path: '/scrape',
      query: {
        video_id: 'video-3',
        type: 'movie',
        title: '无年份电影'
      }
    })
  })

  it('builds a thumbnail url only when the video has an id', () => {
    expect(getVideoThumbnailURL({ id: 'video-1' })).toBe('/api/v1/videos/video-1/thumbnail')
    expect(getVideoThumbnailURL({ id: '' })).toBe('')
  })

  it('shows thumbnails only for ready videos', () => {
    expect(shouldShowVideoThumbnail({ id: 'video-1', status: 'ready' })).toBe(true)
    expect(shouldShowVideoThumbnail({ id: 'video-1', status: 'processing' })).toBe(false)
  })

  it('returns a placeholder label for the thumbnail cell', () => {
    expect(getVideoThumbnailPlaceholder({ id: 'video-1', status: 'ready' })).toBe('暂无封面')
    expect(getVideoThumbnailPlaceholder({ id: 'video-1', status: 'processing' })).toBe('未就绪')
  })
})
