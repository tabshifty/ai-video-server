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
  deleteAdminTvEpisode,
  deleteAdminTvSeason,
  deleteAdminTvSeries,
  deleteAdminVideoSubtitle,
  getAdminVideoTags,
  getAdminTvSeries,
  getAdminTvSeriesDetail,
  getAdminVideoSubtitles,
  rescanAdminVideoSubtitles,
  scrapePreview,
  updateAdminVideoSubtitle,
  updateAdminTvEpisode,
  updateAdminTvSeason,
  updateAdminTvSeries,
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

    expect(remove).toHaveBeenCalledWith('/admin/tv/series/12')
    expect(remove).toHaveBeenCalledWith('/admin/tv/seasons/22')
    expect(remove).toHaveBeenCalledWith('/admin/tv/episodes/33')
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
    expect(remove).toHaveBeenCalledWith('/admin/videos/video-1/subtitles/subtitle-1')
  })
})

describe('video delete apis', () => {
  beforeEach(() => {
    post.mockReset()
  })

  it('requests batch delete for selected video ids', async () => {
    const payload = { video_ids: ['video-1', 'video-2'] }

    await batchDeleteAdminVideos(payload)

    expect(post).toHaveBeenCalledWith('/admin/videos/batch-delete', payload)
  })
})
