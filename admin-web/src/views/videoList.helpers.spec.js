import { describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import {
  buildAVManualScrapeRoute,
  buildMovieManualScrapeRoute,
  buildStuckScrapeRoute,
  canPreviewVideoStatus,
  canManuallyEditVideoStatus,
  createVideoBuiltInViews,
  extractTvPendingDiagnostics,
  getManualVideoStatusOptions,
  getManualVideoStatusValue,
  getVideoStatusMeta,
  getVideoThumbnailPlaceholder,
  isStaleDetailRequest,
  nextDetailRequestToken,
  normalizeVideoViewSnapshot,
  getVideoThumbnailURL,
  shouldShowVideoThumbnail,
  shouldShowStuckScrapeAction,
  subtitleUploadAccept,
  teardownPreviewPlayer
} from './videoList.helpers'

describe('videoList helpers', () => {
  it('只创建 API 已支持的内置视频视图且不共享列引用', () => {
    const columns = ['title', 'thumbnail', 'status', 'operations']
    const views = createVideoBuiltInViews(columns)

    expect(views.map((item) => [item.id, item.label, item.snapshot.status])).toEqual([
      ['builtin-all', '全部视频', ''],
      ['builtin-processing', '处理中', 'processing'],
      ['builtin-failed', '失败', 'failed']
    ])
    expect(views.every((item) => item.builtIn === true)).toBe(true)
    expect(views[0].snapshot.columns).not.toBe(columns)
    expect(views[0].snapshot.columns).not.toBe(views[1].snapshot.columns)

    views[0].snapshot.columns.pop()
    expect(columns).toEqual(['title', 'thumbnail', 'status', 'operations'])
    expect(views[1].snapshot.columns).toEqual(columns)
  })

  it('规范视频视图快照并排除分页、选择和 Drawer 状态', () => {
    const allowed = ['title', 'thumbnail', 'status', 'operations']
    const defaults = ['title', 'thumbnail', 'status', 'operations']
    const sourceColumns = ['status', 'unknown', 'status', 'title']

    const snapshot = normalizeVideoViewSnapshot({
      q: 'A',
      type: 'movie',
      status: 'ready',
      page: 9,
      selectedRows: ['video-1'],
      detailVisible: true,
      columns: sourceColumns
    }, allowed, defaults)

    expect(snapshot).toEqual({
      q: 'A',
      type: 'movie',
      status: 'ready',
      columns: ['status', 'title', 'operations']
    })
    expect(snapshot.columns).not.toBe(sourceColumns)
    expect(snapshot).not.toHaveProperty('page')
    expect(snapshot).not.toHaveProperty('selectedRows')
    expect(snapshot).not.toHaveProperty('detailVisible')
  })

  it('缺省快照复制默认列并始终保留操作列', () => {
    const allowed = ['title', 'thumbnail', 'status', 'operations']
    const defaults = ['title', 'thumbnail', 'status', 'operations']
    const snapshot = normalizeVideoViewSnapshot({}, allowed, defaults)

    expect(snapshot).toEqual({ q: '', type: '', status: '', columns: defaults })
    expect(snapshot.columns).not.toBe(defaults)
    expect(normalizeVideoViewSnapshot({ columns: ['title'] }, allowed, defaults).columns).toEqual([
      'title',
      'operations'
    ])
  })

  it('includes non-ready workflow status labels and tag types', () => {
    expect(getVideoStatusMeta('tv_pending')).toEqual({
      label: '待绑定',
      tagType: 'warning'
    })
    expect(getVideoStatusMeta('av_scrape_pending')).toEqual({
      label: '欧美 AV 待确认',
      tagType: 'warning'
    })
    expect(getVideoStatusMeta('pending_delete')).toEqual({
      label: '待删除',
      tagType: 'danger'
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

  it('blocks manual edits for processing and pending delete', () => {
    expect(canManuallyEditVideoStatus('processing')).toBe(false)
    expect(canManuallyEditVideoStatus('pending_delete')).toBe(false)
    expect(canManuallyEditVideoStatus('ready')).toBe(true)
  })

  it('omits locked statuses from the manual update payload value', () => {
    expect(getManualVideoStatusValue('processing')).toBe('')
    expect(getManualVideoStatusValue('pending_delete')).toBe('')
    expect(getManualVideoStatusValue('ready')).toBe('ready')
  })

  it('allows admin preview for ready and pending delete videos only', () => {
    expect(canPreviewVideoStatus('ready')).toBe(true)
    expect(canPreviewVideoStatus('pending_delete')).toBe(true)
    expect(canPreviewVideoStatus('processing')).toBe(false)
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

  it('shows the stuck-scrape action only for scraping movie/av/episode', () => {
    expect(shouldShowStuckScrapeAction('scraping', 'av')).toBe(true)
    expect(shouldShowStuckScrapeAction('scraping', 'movie')).toBe(true)
    expect(shouldShowStuckScrapeAction('scraping', 'episode')).toBe(true)

    expect(shouldShowStuckScrapeAction('uploaded', 'av')).toBe(false)
    expect(shouldShowStuckScrapeAction('ready', 'av')).toBe(false)
    expect(shouldShowStuckScrapeAction('av_scrape_pending', 'av')).toBe(false)
    expect(shouldShowStuckScrapeAction('scraping', 'short')).toBe(false)
    expect(shouldShowStuckScrapeAction('scraping', '')).toBe(false)
    expect(shouldShowStuckScrapeAction('', 'av')).toBe(false)
  })

  it('builds a stuck-scrape route per video type', () => {
    expect(buildStuckScrapeRoute({
      id: 'av-1',
      type: 'av',
      title: 'SSIS-123',
      metadata: { external_id: 'SSIS-123' }
    })).toEqual(buildAVManualScrapeRoute({
      id: 'av-1',
      type: 'av',
      title: 'SSIS-123',
      metadata: { external_id: 'SSIS-123' }
    }))

    expect(buildStuckScrapeRoute({
      id: 'movie-1',
      type: 'movie',
      title: '盗梦空间'
    })).toEqual(buildMovieManualScrapeRoute({
      id: 'movie-1',
      type: 'movie',
      title: '盗梦空间'
    }))

    expect(buildStuckScrapeRoute({
      id: 'ep-1',
      type: 'episode',
      title: '三体 S01E02'
    })).toEqual({
      path: '/scrape',
      query: {
        video_id: 'ep-1',
        type: 'tv',
        title: '三体 S01E02'
      }
    })
  })

  it('builds a stuck-scrape episode route with parsed season/episode when available', () => {
    expect(buildStuckScrapeRoute(
      { id: 'ep-1', type: 'episode', title: '三体' },
      { parsedTitle: '三体', parsedSeasonNumber: 1, parsedEpisodeNumber: 2 }
    )).toEqual({
      path: '/scrape',
      query: {
        video_id: 'ep-1',
        type: 'tv',
        title: '三体',
        season_number: 1,
        episode_number: 2
      }
    })
  })

  it('returns null for unknown stuck-scrape video types', () => {
    expect(buildStuckScrapeRoute({ id: 'x', type: 'short', title: 't' })).toBeNull()
    expect(buildStuckScrapeRoute({ id: 'x', type: '', title: 't' })).toBeNull()
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
