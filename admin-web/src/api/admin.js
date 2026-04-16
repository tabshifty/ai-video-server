import request from './request'

export const getAdminStats = () => request.get('/admin/stats')
export const getAdminVideos = (params) => request.get('/admin/videos', { params })
export const getAdminVideoDetail = (id) => request.get(`/admin/videos/${id}`)
export const getAdminVideoPlayURL = (id) => request.get(`/admin/videos/${id}/play-url`)
export const updateAdminVideo = (id, payload) => request.put(`/admin/videos/${id}`, payload)
export const deleteAdminVideo = (id) => request.delete(`/admin/videos/${id}`)
export const retranscodeVideo = (id) => request.post(`/admin/videos/${id}/retranscode`)

export const getAdminUsers = (params) => request.get('/admin/users', { params })
export const updateUserRole = (id, payload) => request.put(`/admin/users/${id}/role`, payload)
export const getAdminActors = (params) => request.get('/admin/actors', { params })
export const createAdminActor = (payload) => request.post('/admin/actors', payload)
export const updateAdminActor = (id, payload) => request.put(`/admin/actors/${id}`, payload)

export const getAdminTasks = (params) => request.get('/admin/tasks', { params })

export const systemCleanup = (payload) => request.post('/admin/system/cleanup', payload)
export const getSystemLogs = (params) => request.get('/admin/system/logs', { params })

export const scrapePreview = (payload) => request.post('/admin/scrape/preview', payload)
export const scrapeConfirm = (payload) => request.put('/admin/scrape/confirm', payload)
