import request from './request'

export const checkUpload = (payload) => request.post('/upload/check', payload)
// 命中已有视频后显式追加本次选择的合集关系；/upload/check 本身保持只读。
export const attachUploadCollections = (payload) => request.post('/upload/attach', payload)
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
    timeout: 0,
    headers: { 'Content-Type': 'application/octet-stream' }
  })

export const uploadComplete = (sessionId) => request.post(
  '/upload/complete',
  { session_id: sessionId },
  { timeout: 0 }
)
export const uploadAbort = (sessionId) => request.delete(`/upload/session/${sessionId}`)
