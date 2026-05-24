import request from './request'

export const getAdminStats = () => request.get('/admin/stats')
export const getAdminVideos = (params) => request.get('/admin/videos', { params })
export const getAdminVideoTags = (params) => request.get('/admin/video-tags', { params })
export const getAdminPopularVideoTags = (params) => request.get('/admin/video-tags/popular', { params })
export const getAdminVideoDetail = (id) => request.get(`/admin/videos/${id}`)
export const getAdminVideoSubtitles = (id) => request.get(`/admin/videos/${id}/subtitles`)
export const uploadAdminVideoSubtitle = (id, formData) =>
  request.post(`/admin/videos/${id}/subtitles/upload`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 0
  })
export const rescanAdminVideoSubtitles = (id) => request.post(`/admin/videos/${id}/subtitles/scan`)
export const updateAdminVideoSubtitle = (id, subtitleId, payload) =>
  request.put(`/admin/videos/${id}/subtitles/${subtitleId}`, payload)
export const deleteAdminVideoSubtitle = (id, subtitleId) =>
  request.delete(`/admin/videos/${id}/subtitles/${subtitleId}`)
export const getAdminVideoPlayURL = (id) => request.get(`/admin/videos/${id}/play-url`)
export const captureAdminVideoThumbnail = (id, payload) => request.post(`/admin/videos/${id}/thumbnail/capture`, payload)
export const updateAdminVideo = (id, payload) => request.put(`/admin/videos/${id}`, payload)
export const deleteAdminVideo = (id) => request.delete(`/admin/videos/${id}`)
export const batchDeleteAdminVideos = (payload) => request.post('/admin/videos/batch-delete', payload)
export const retranscodeVideo = (id) => request.post(`/admin/videos/${id}/retranscode`)
export const getAdminTvSeries = (params) => request.get('/admin/tv/series', { params })
export const getAdminTvSeriesDetail = (id) => request.get(`/admin/tv/series/${id}`)
export const createAdminTvSeries = (payload) => request.post('/admin/tv/series', payload)
export const updateAdminTvSeries = (id, payload) => request.put(`/admin/tv/series/${id}`, payload)
export const deleteAdminTvSeries = (id) => request.delete(`/admin/tv/series/${id}`)
export const createAdminTvSeason = (seriesId, payload) => request.post(`/admin/tv/series/${seriesId}/seasons`, payload)
export const updateAdminTvSeason = (id, payload) => request.put(`/admin/tv/seasons/${id}`, payload)
export const deleteAdminTvSeason = (id) => request.delete(`/admin/tv/seasons/${id}`)
export const createAdminTvEpisode = (seasonId, payload) => request.post(`/admin/tv/seasons/${seasonId}/episodes`, payload)
export const updateAdminTvEpisode = (id, payload) => request.put(`/admin/tv/episodes/${id}`, payload)
export const deleteAdminTvEpisode = (id) => request.delete(`/admin/tv/episodes/${id}`)
export const getAdminIPTVPlaylist = () => request.get('/admin/iptv/playlist')
export const uploadAdminIPTVPlaylist = (formData) =>
  request.post('/admin/iptv/playlist/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 0
  })
export const updateAdminIPTVSource = (payload) => request.put('/admin/iptv/playlist/source', payload)
export const refreshAdminIPTVPlaylist = () => request.post('/admin/iptv/playlist/refresh')

export const getAdminUsers = (params) => request.get('/admin/users', { params })
export const registerAdminUser = (payload) => request.post('/auth/register', payload)
export const updateUserRole = (id, payload) => request.put(`/admin/users/${id}/role`, payload)
export const getAdminActors = (params) => request.get('/admin/actors', { params })
export const createAdminActor = (payload) => request.post('/admin/actors', payload)
export const scrapeAdminActorPreview = (payload) => request.post('/admin/actors/scrape/preview', payload)
export const updateAdminActor = (id, payload) => request.put(`/admin/actors/${id}`, payload)
export const getAdminCollections = (params) => request.get('/admin/collections', { params })
export const createAdminCollection = (payload) => request.post('/admin/collections', payload)
export const updateAdminCollection = (id, payload) => request.put(`/admin/collections/${id}`, payload)
export const deleteAdminCollection = (id) => request.delete(`/admin/collections/${id}`)
export const uploadAdminImages = (formData) =>
  request.post('/admin/images/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 0
  })
export const checkAdminImageUpload = (payload) => request.post('/admin/images/check', payload)
export const getAdminImages = (params) => request.get('/admin/images', { params })
export const getAdminImageDetail = (id) => request.get(`/admin/images/${id}`)
export const updateAdminImage = (id, payload) => request.put(`/admin/images/${id}`, payload)
export const deleteAdminImage = (id) => request.delete(`/admin/images/${id}`)
export const getAdminImageViewBlob = (id, params) =>
  request.get(`/admin/images/${id}/view`, {
    params,
    responseType: 'blob'
  })
export const getAdminImageCollections = (params) => request.get('/admin/image-collections', { params })
export const createAdminImageCollection = (payload) => request.post('/admin/image-collections', payload)
export const updateAdminImageCollection = (id, payload) => request.put(`/admin/image-collections/${id}`, payload)
export const deleteAdminImageCollection = (id) => request.delete(`/admin/image-collections/${id}`)

export const getAdminTasks = (params) => request.get('/admin/tasks', { params })

export const systemCleanup = (payload) => request.post('/admin/system/cleanup', payload)
export const getSystemLogs = (params) => request.get('/admin/system/logs', { params })

export const scrapePreview = (payload) => request.post('/admin/scrape/preview', payload)
export const scrapeConfirm = (payload) => request.put('/admin/scrape/confirm', payload)
export const scrapeSkip = (payload) => request.put('/admin/scrape/skip', payload)
export const getAVScrapeConfig = () => request.get('/admin/av-scrape/config')
export const updateAVScrapeConfig = (payload) => request.put('/admin/av-scrape/config', payload)
export const avScrapePreview = (payload) => request.post('/admin/av-scrape/preview', payload)
export const avScrapeConfirm = (payload) => request.put('/admin/av-scrape/confirm', payload)
