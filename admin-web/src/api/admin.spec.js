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
  batchUpdateAdminVideos,
  createAdminTvEpisode,
  createAdminTvSeason,
  createAdminTvSeries,
  deleteAdminTVAppReleaseDraft,
  deleteAdminPasswordVaultEntry,
  deleteAdminTvEpisode,
  deleteAdminTvSeason,
  deleteAdminTvSeries,
  deleteAdminImage,
  deleteAdminVideo,
  deleteAdminVideoSubtitle,
  deleteLatestOrphanFileScan,
  getAdminArchiveImportBatches,
  getAdminArchiveImportBatchDetail,
  getAdminArchiveImportFileDetail,
  getAdminPasswordVaultEntries,
  getAdminPasswordVaultPassword,
  deleteAdminArchiveImportBatch,
  deleteAdminArchiveImportGroup,
  getAdminVideoTags,
  getAdminTvSeries,
  getAdminTvSeriesDetail,
  generateAdminImage,
  createAdminArchiveImportGroup,
  getAdminImageCollections,
  getAdminImageGenerationStatus,
  getAdminImageViewBlob,
  getAdminImages,
  getAdminTVAppReleaseDetail,
  getAdminTVAppReleases,
  getAdminVideoSubtitles,
  getLatestOrphanFileScan,
  offlineAdminTVAppRelease,
  publishAdminTVAppRelease,
  processAdminArchiveImportGroup,
  rescanAdminVideoSubtitles,
  removeAdminArchiveImportGroupFiles,
  restoreAdminTVAppRelease,
  retryAdminArchiveImportExtract,
  scrapePreview,
  startOrphanFileScan,
  assignAdminArchiveImportGroupFiles,
  updateAdminArchiveImportGroup,
  updateAdminTVAppRelease,
  updateAdminArchiveImportFile,
  updateAdminPasswordVaultEntry,
  updateAdminVideoSubtitle,
  updateAdminTvEpisode,
  updateAdminTvSeason,
  updateAdminTvSeries,
  uploadAdminTVAppAPK,
  uploadAdminArchiveImport,
  createAdminPasswordVaultEntry,
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

describe('admin image library apis', () => {
  beforeEach(() => {
    get.mockReset()
    get.mockResolvedValue({ ok: true })
  })

  it('loads image assets, collections and view blobs for workbench references', async () => {
    await getAdminImages({ q: '封面', status: 'ready', active: 1, stored_mime: 'image/png,image/jpeg,image/webp' })
    await getAdminImageCollections({ active: 1 })
    await getAdminImageViewBlob('image-1', { w: 360, h: 270, fit: 'cover', q: 78 })

    expect(get).toHaveBeenCalledWith('/admin/images', {
      params: { q: '封面', status: 'ready', active: 1, stored_mime: 'image/png,image/jpeg,image/webp' }
    })
    expect(get).toHaveBeenCalledWith('/admin/image-collections', {
      params: { active: 1 }
    })
    expect(get).toHaveBeenCalledWith('/admin/images/image-1/view', {
      params: { w: 360, h: 270, fit: 'cover', q: 78 },
      responseType: 'blob'
    })
  })
})

describe('admin password vault apis', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
    remove.mockReset()
    get.mockResolvedValue({ ok: true })
    post.mockResolvedValue({ ok: true })
    put.mockResolvedValue({ ok: true })
    remove.mockResolvedValue({ ok: true })
  })

  it('requests password vault list and password reveal endpoints', async () => {
    await getAdminPasswordVaultEntries({ q: 'nas', page: 1 })
    await getAdminPasswordVaultPassword('entry-1')

    expect(get).toHaveBeenCalledWith('/admin/password-vault', {
      params: { q: 'nas', page: 1 }
    })
    expect(get).toHaveBeenCalledWith('/admin/password-vault/entry-1/password')
  })

  it('creates, updates and deletes password vault entries', async () => {
    const payload = { name: 'NAS', account: 'admin', password: 'secret', url: '', note: '' }

    await createAdminPasswordVaultEntry(payload)
    await updateAdminPasswordVaultEntry('entry-1', payload)
    await deleteAdminPasswordVaultEntry('entry-1')

    expect(post).toHaveBeenCalledWith('/admin/password-vault', payload)
    expect(put).toHaveBeenCalledWith('/admin/password-vault/entry-1', payload)
    expect(remove).toHaveBeenCalledWith('/admin/password-vault/entry-1', {
      timeout: 0
    })
  })
})

describe('archive import apis', () => {
  beforeEach(() => {
    get.mockReset()
    post.mockReset()
    put.mockReset()
    remove.mockReset()
    get.mockResolvedValue({ ok: true })
    post.mockResolvedValue({ ok: true })
    put.mockResolvedValue({ ok: true })
    remove.mockResolvedValue({ ok: true })
  })

  it('loads archive batches and file detail endpoints', async () => {
    await getAdminArchiveImportBatches({ page: 2, page_size: 20 })
    await getAdminArchiveImportBatchDetail('batch-1')
    await getAdminArchiveImportFileDetail('file-1')

    expect(get).toHaveBeenCalledWith('/admin/archive-import/batches', {
      params: { page: 2, page_size: 20 }
    })
    expect(get).toHaveBeenCalledWith('/admin/archive-import/batches/batch-1')
    expect(get).toHaveBeenCalledWith('/admin/archive-import/files/file-1')
  })

  it('uploads archive imports and processes batches or files', async () => {
    const formData = new FormData()
    const payload = { password: 'secret', encoding_mode: 'auto' }

    await uploadAdminArchiveImport(formData)
    await deleteAdminArchiveImportBatch('batch-1')
    await updateAdminArchiveImportFile('file-1', { title: '标题' })
    await retryAdminArchiveImportExtract('batch-1', payload)

    expect(post).toHaveBeenCalledWith('/admin/archive-import/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0
    })
    expect(remove).toHaveBeenCalledWith('/admin/archive-import/batches/batch-1', {
      timeout: 0
    })
    expect(put).toHaveBeenCalledWith('/admin/archive-import/files/file-1', { title: '标题' })
    expect(post).toHaveBeenCalledWith('/admin/archive-import/batches/batch-1/retry-extract', payload)
  })

  it('manages archive import groups within a batch', async () => {
    const groupPayload = {
      name: '第一组',
      note: '说明',
      file_ids: ['file-1']
    }
    const updatePayload = {
      name: '第一组',
      note: '更新说明',
      title: '组标题',
      description: '组说明',
      tags: ['tag-a'],
      video_type: 'short',
      video_collection_ids: ['collection-1'],
      image_collection_ids: ['image-collection-1']
    }
    const filePayload = { file_ids: ['file-1', 'file-2'] }

    await createAdminArchiveImportGroup('batch-1', groupPayload)
    await updateAdminArchiveImportGroup('group-1', updatePayload)
    await assignAdminArchiveImportGroupFiles('group-1', filePayload)
    await removeAdminArchiveImportGroupFiles('batch-1', filePayload)
    await processAdminArchiveImportGroup('group-1')
    await deleteAdminArchiveImportGroup('group-1')

    expect(post).toHaveBeenCalledWith('/admin/archive-import/batches/batch-1/groups', groupPayload)
    expect(put).toHaveBeenCalledWith('/admin/archive-import/groups/group-1', updatePayload)
    expect(post).toHaveBeenCalledWith('/admin/archive-import/groups/group-1/files', filePayload)
    expect(post).toHaveBeenCalledWith('/admin/archive-import/batches/batch-1/groups/remove-files', filePayload)
    expect(post).toHaveBeenCalledWith('/admin/archive-import/groups/group-1/process')
    expect(remove).toHaveBeenCalledWith('/admin/archive-import/groups/group-1', {
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
    await getAdminTVAppReleases({ q: '84', status: 'draft', client_type: 'android_tv' })
    await getAdminTVAppReleaseDetail(12)

    expect(get).toHaveBeenCalledWith('/admin/tv-app/releases', {
      params: { q: '84', status: 'draft', client_type: 'android_tv' }
    })
    expect(get).toHaveBeenCalledWith('/admin/tv-app/releases/12')
  })

  it('uploads tv apk with multipart and timeout disabled', async () => {
    const formData = new FormData()

    await uploadAdminTVAppAPK(formData, 'android_phone')

    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/upload', formData, {
      params: { client_type: 'android_phone' },
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0
    })
  })

  it('updates publish lifecycle for tv app release', async () => {
    const payload = { release_notes: '修复 IPTV 闪退', remarks: '推荐升级' }

    await updateAdminTVAppRelease(12, payload, 'android_tv')
    await publishAdminTVAppRelease(12, payload, 'android_tv')
    await offlineAdminTVAppRelease(12, 'android_tv')
    await restoreAdminTVAppRelease(12, 'android_tv')
    await deleteAdminTVAppReleaseDraft(12, 'android_tv')

    expect(put).toHaveBeenCalledWith('/admin/tv-app/releases/12', payload, {
      params: { client_type: 'android_tv' }
    })
    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/12/publish', payload, {
      params: { client_type: 'android_tv' }
    })
    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/12/offline', null, {
      params: { client_type: 'android_tv' }
    })
    expect(post).toHaveBeenCalledWith('/admin/tv-app/releases/12/restore', null, {
      params: { client_type: 'android_tv' }
    })
    expect(remove).toHaveBeenCalledWith('/admin/tv-app/releases/12', {
      params: { client_type: 'android_tv' },
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

describe('video batch update apis', () => {
  beforeEach(() => {
    put.mockReset()
    put.mockResolvedValue({ ok: true })
  })

  it('requests batch update for selected video ids', async () => {
    const payload = {
      video_ids: ['video-1', 'video-2'],
      update_title: true,
      title: '统一标题'
    }

    await batchUpdateAdminVideos(payload)

    expect(put).toHaveBeenCalledWith('/admin/videos/batch-update', payload, {
      timeout: 0
    })
  })
})
