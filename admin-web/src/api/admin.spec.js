import { describe, expect, it, vi, beforeEach } from 'vitest'

const { post, get, put, remove } = vi.hoisted(() => ({
  post: vi.fn(),
  get: vi.fn(),
  put: vi.fn(),
  remove: vi.fn()
}))

vi.mock('./request', () => ({
  default: {
    post,
    get,
    put,
    delete: remove
  }
}))

import {
  batchDeleteAdminVideos,
  createAdminTvEpisode,
  createAdminTvSeason,
  createAdminTvSeries,
  deleteAdminTVAppReleaseDraft,
  deleteAdminTvEpisode,
  deleteAdminTvSeason,
  deleteAdminTvSeries,
  deleteAdminImage,
  deleteAdminVideo,
  deleteAdminVideoSubtitle,
  deleteLatestOrphanFileScan,
  getAdminVideoTags,
  getAdminTvSeries,
  getAdminTvSeriesDetail,
  generateAdminImage,
  getAdminImageGenerationStatus,
  getAdminTVAppReleaseDetail,
  getAdminTVAppReleases,
  getAdminVideoSubtitles,
  getLatestOrphanFileScan,
  offlineAdminTVAppRelease,
  publishAdminTVAppRelease,
  rescanAdminVideoSubtitles,
  restoreAdminTVAppRelease,
  scrapePreview,
  startOrphanFileScan,
  updateAdminTVAppRelease,
  updateAdminVideoSubtitle,
  updateAdminTvEpisode,
  updateAdminTvSeason,
  updateAdminTvSeries,
  uploadAdminTVAppAPK,
  uploadAdminVideoSubtitle,
  uploadAdminImages
} from './admin'

describe('uploadAdminImages', () => {
  beforeEach(() => {
    post.mockReset()
    post.mockResolvedValue({ ok: true })
  })

  it('disables request timeout for long-running image uploads', async () => {
    const formData = new FormData()

    await uploadAdminImages(formData)

    expect(post).toHaveBeenCalledWith('/admin/images/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0
    })
  })
})

describe('image generation apis', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    get.mockResolvedValue({ ok: true })
    post.mockResolvedValue({ ok: true })
  })

  it('loads redacted image generation status and disables timeout for generation', async () => {
    const payload = { prompt: '生成海报', n: 1 }

    await getAdminImageGenerationStatus()
    await generateAdminImage(payload)

    expect(get).toHaveBeenCalledWith('/admin/image-generation/status')
    expect(post).toHaveBeenCalledWith('/admin/image-generation/generate', payload, {
      timeout: 0
    })
  })
})

describe('scrape preview apis', () => {
  beforeEach(() => {
    post.mockReset()
    post.mockResolvedValue({ ok: true })
  })

  it('disables request timeout for external scrape preview', async () => {
    const payload = { type: 'tv', title: 'The Lead' }

    await scrapePreview(payload)

    expect(post).toHaveBeenCalledWith('/admin/scrape/preview', payload, {
      timeout: 0
    })
  })
})

describe('orphan file scan apis', () => {
  beforeEach(() => {
    post.mockReset()
    get.mockReset()
    remove.mockReset()
    post.mockResolvedValue({ ok: true })
    get.mockResolvedValue({ ok: true })
    remove.mockResolvedValue({ ok: true })
  })

  it('starts, loads and deletes orphan file scans', async () => {
    await startOrphanFileScan()
    await getLatestOrphanFileScan()
    await deleteLatestOrphanFileScan()

    expect(post).toHaveBeenCalledWith('/admin/system/orphan-files/scan')
    expect(get).toHaveBeenCalledWith('/admin/system/orphan-files/latest')
    expect(remove).toHaveBeenCalledWith('/admin/system/orphan-files/latest', {
      timeout: 0
    })
  })
})

describe('video tag apis', () => {
  beforeEach(() => {
    get.mockReset()
  })

  it('requests admin video tags with query params', async () => {
    await getAdminVideoTags({ q: '动作', limit: 12 })

    expect(get).toHaveBeenCalledWith('/admin/video-tags', {
      params: { q: '动作', limit: 12 }
    })
  })
})

describe('tv series apis', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
    remove.mockReset()
  })

  it('requests admin tv series list with query params', async () => {
    await getAdminTvSeries({ q: '雾城', active: 1 })

    expect(get).toHaveBeenCalledWith('/admin/tv/series', {
      params: { q: '雾城', active: 1 }
    })
  })

  it('requests admin tv series detail by id', async () => {
    await getAdminTvSeriesDetail(12)

    expect(get).toHaveBeenCalledWith('/admin/tv/series/12')
  })

  it('creates and updates admin tv series resources', async () => {
    const payload = { title: '雾城档案' }

    await createAdminTvSeries(payload)
    await updateAdminTvSeries(12, payload)

    expect(post).toHaveBeenCalledWith('/admin/tv/series', payload)
    expect(put).toHaveBeenCalledWith('/admin/tv/series/12', payload)
  })

  it('creates and updates admin tv season resources', async () => {
    const payload = { season_number: 2 }

    await createAdminTvSeason(12, payload)
    await updateAdminTvSeason(22, payload)

    expect(post).toHaveBeenCalledWith('/admin/tv/series/12/seasons', payload)
    expect(put).toHaveBeenCalledWith('/admin/tv/seasons/22', payload)
  })

  it('creates and updates admin tv episode resources', async () => {
    const payload = { episode_number: 6 }

    await createAdminTvEpisode(22, payload)
    await updateAdminTvEpisode(33, payload)

    expect(post).toHaveBeenCalledWith('/admin/tv/seasons/22/episodes', payload)
    expect(put).toHaveBeenCalledWith('/admin/tv/episodes/33', payload)
  })

  it('deletes admin tv resources', async () => {
    await deleteAdminTvSeries(12)
    await deleteAdminTvSeason(22)
    await deleteAdminTvEpisode(33)

    expect(remove).toHaveBeenCalledWith('/admin/tv/series/12', {
      timeout: 0
    })
    expect(remove).toHaveBeenCalledWith('/admin/tv/seasons/22', {
      timeout: 0
    })
    expect(remove).toHaveBeenCalledWith('/admin/tv/episodes/33', {
      timeout: 0
    })
  })
})

describe('tv app release apis', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
    remove.mockReset()
  })

  it('requests tv app release list and detail', async () => {
    await getAdminTVAppReleases({ q: '84', status: 'draft' })
    await getAdminTVAppReleaseDetail(12)

    expect(get).toHaveBeenCalledWith('/admin/tv-app/releases', {
      params: { q: '84', status: 'draft' }
    })
    expect(get).toHaveBeenCalledWith('/admin/tv-app/releases/12')
  })

  it('uploads tv apk with multipart and timeout disabled', async () => {
    const formData = new FormData()

    await uploadAdminTVAppAPK(formData)

    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0
    })
  })

  it('updates publish lifecycle for tv app release', async () => {
    const payload = { release_notes: '修复 IPTV 闪退', remarks: '推荐升级' }

    await updateAdminTVAppRelease(12, payload)
    await publishAdminTVAppRelease(12, payload)
    await offlineAdminTVAppRelease(12)
    await restoreAdminTVAppRelease(12)
    await deleteAdminTVAppReleaseDraft(12)

    expect(put).toHaveBeenCalledWith('/admin/tv-app/releases/12', payload)
    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/12/publish', payload)
    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/12/offline')
    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/12/restore')
    expect(remove).toHaveBeenCalledWith('/admin/tv-app/releases/12', {
      timeout: 0
    })
  })
})

describe('video subtitle apis', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
    remove.mockReset()
  })

  it('requests subtitle list and rescan by video id', async () => {
    await getAdminVideoSubtitles('video-1')
    await rescanAdminVideoSubtitles('video-1')

    expect(get).toHaveBeenCalledWith('/admin/videos/video-1/subtitles')
    expect(post).toHaveBeenCalledWith('/admin/videos/video-1/subtitles/scan')
  })

  it('uploads subtitle files with multipart form data', async () => {
    const formData = new FormData()

    await uploadAdminVideoSubtitle('video-1', formData)

    expect(post).toHaveBeenCalledWith('/admin/videos/video-1/subtitles/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0
    })
  })

  it('updates and deletes subtitle records', async () => {
    const payload = { language_code: 'zh-CN', is_default: true }

    await updateAdminVideoSubtitle('video-1', 'subtitle-1', payload)
    await deleteAdminVideoSubtitle('video-1', 'subtitle-1')

    expect(put).toHaveBeenCalledWith('/admin/videos/video-1/subtitles/subtitle-1', payload)
    expect(remove).toHaveBeenCalledWith('/admin/videos/video-1/subtitles/subtitle-1', {
      timeout: 0
    })
  })
})

describe('video delete apis', () => {
  beforeEach(() => {
    post.mockReset()
    remove.mockReset()
  })

  it('requests batch delete for selected video ids', async () => {
    const payload = { video_ids: ['video-1', 'video-2'] }

    await batchDeleteAdminVideos(payload)

    expect(post).toHaveBeenCalledWith('/admin/videos/batch-delete', payload, {
      timeout: 0
    })
  })

  it('disables timeout for destructive delete requests', async () => {
    await deleteAdminVideo('video-1')
    await deleteAdminImage('image-1')
    await deleteAdminVideoSubtitle('video-1', 'subtitle-1')

    expect(remove).toHaveBeenCalledWith('/admin/videos/video-1', {
      timeout: 0
    })
    expect(remove).toHaveBeenCalledWith('/admin/images/image-1', {
      timeout: 0
    })
    expect(remove).toHaveBeenCalledWith('/admin/videos/video-1/subtitles/subtitle-1', {
      timeout: 0
    })
  })
})
