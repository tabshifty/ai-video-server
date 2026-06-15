import { describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import {
  buildAVManualScrapeRoute,
  buildMovieManualScrapeRoute,
  canManuallyEditVideoStatus,
  extractTvPendingDiagnostics,
  getManualVideoStatusOptions,
  getManualVideoStatusValue,
  getVideoStatusMeta,
  getVideoThumbnailPlaceholder,
  isStaleDetailRequest,
  nextDetailRequestToken,
  getVideoThumbnailURL,
  shouldShowVideoThumbnail,
  subtitleUploadAccept,
  teardownPreviewPlayer
} from './videoList.helpers'

describe('videoList helpers', () => {
  it('includes tv_pending status label and tag type', () => {
    expect(getVideoStatusMeta('tv_pending')).toEqual({
      label: '待绑定',
      tagType: 'warning'
    })
    expect(getVideoStatusMeta('av_scrape_pending')).toEqual({
      label: '欧美 AV 待确认',
      tagType: 'warning'
    })
  })

  it('returns only editable manual status options', () => {
    expect(getManualVideoStatusOptions()).toEqual([
      { value: 'uploaded', label: '已上传' },
      { value: 'scraping', label: '刮削中' },
      { value: 'tv_pending', label: '待绑定' },
      { value: 'av_scrape_pending', label: '欧美 AV 待确认' },
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

  it('keeps detail request token stable for the active detail request', () => {
    const requestToken = nextDetailRequestToken(4)

    expect(requestToken).toBe(5)
    expect(isStaleDetailRequest(5, requestToken)).toBe(false)
    expect(isStaleDetailRequest(6, requestToken)).toBe(true)
  })

  it('does not invalidate the active detail request while opening the drawer', () => {
    const source = readFileSync(new URL('./VideoList.vue', import.meta.url), 'utf8')
    const showDetailBlock = source
      .split('async function showDetail(row)')[1]
      .split('async function refreshPlayURL')[0]
    const refreshPlayURLBlock = source
      .split('async function refreshPlayURL')[1]
      .split('function resetSubtitleState')[0]

    expect(showDetailBlock).toContain('nextDetailRequestToken(')
    expect(showDetailBlock).toContain('handleDetailClose({ invalidateToken: false })')
    expect(showDetailBlock).toContain('resetDetailState({ invalidateToken: false })')
    expect(showDetailBlock).toContain('isStaleDetailRequest(detailRequestToken.value, requestToken)')
    expect(refreshPlayURLBlock).toContain('handleDetailClose({ invalidateToken: false })')
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

  it('allows ASS and SSA subtitle uploads alongside SRT and VTT', () => {
    expect(subtitleUploadAccept.split(',')).toEqual(['.srt', '.vtt', '.ass', '.ssa'])
  })

  it('keeps current page shift selection logic scoped to visible rows', () => {
    const source = readFileSync(new URL('./VideoList.vue', import.meta.url), 'utf8')
    expect(source).toContain('function onRowSelectionSelect(selection, row)')
    expect(source).toContain('if (shiftKeyPressed.value && selectionAnchorIndex.value >= 0 && currentIndex >= 0)')
    expect(source).toContain('applySelectionByIDs(Array.from(selectedSet), selectionAnchorIndex.value)')
    expect(source).toContain('clearSelection()')
  })

  it('keeps batch actions mutually exclusive and drawer preload guarded', () => {
    const source = readFileSync(new URL('./VideoList.vue', import.meta.url), 'utf8')
    expect(source).toContain('const batchActionBusy = computed(() => deletingBatch.value || updatingBatch.value)')
    expect(source).toContain("if (selectedRows.value.length === 0 || batchActionBusy.value)")
    expect(source).toContain('const preloadResults = await Promise.allSettled([')
    expect(source).toContain("ElMessage.warning(`${failedTargets.join('、')}选项加载失败，可稍后重试`)")
  })
})
