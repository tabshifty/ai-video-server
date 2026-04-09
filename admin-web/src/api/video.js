import request from './request'

export const checkUpload = (payload) => request.post('/upload/check', payload)
export const uploadSingle = (formData, onUploadProgress, signal) =>
  request.post('/upload', formData, {
    signal,
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress
  })

export const uploadInit = (payload) => request.post('/upload/init', payload)

export const uploadChunk = (sessionId, chunkIndex, chunkBlob, signal) =>
  request.put(`/upload/chunk?session_id=${encodeURIComponent(sessionId)}&chunk_index=${chunkIndex}`, chunkBlob, {
    signal,
    headers: { 'Content-Type': 'application/octet-stream' }
  })

export const uploadComplete = (sessionId) => request.post('/upload/complete', { session_id: sessionId })
export const uploadAbort = (sessionId) => request.delete(`/upload/session/${sessionId}`)
