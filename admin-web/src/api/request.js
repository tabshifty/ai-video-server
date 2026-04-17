import axios from 'axios'
import { ElMessage } from 'element-plus'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api/v1',
  timeout: 30000
})

let authRedirecting = false

function isAuthPayloadError(code, msg) {
  if (code === 401 || code === 403) return true
  const text = String(msg || '').toLowerCase()
  if (text.includes('token') || text.includes('authorization') || text.includes('admin only')) {
    return true
  }
  return false
}

function handleAuthExpired(message) {
  localStorage.removeItem('admin_token')
  localStorage.removeItem('admin_user_role')
  if (location.pathname === '/login') {
    return
  }
  if (authRedirecting) {
    return
  }
  authRedirecting = true
  ElMessage.error(message || '登录状态已失效，请重新登录')
  const redirect = encodeURIComponent(location.pathname + location.search)
  setTimeout(() => {
    location.replace(`/login?redirect=${redirect}`)
  }, 100)
}

request.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

request.interceptors.response.use(
  (resp) => {
    const payload = resp.data
    if (!payload || typeof payload.code === 'undefined') {
      return resp.data
    }
    if (payload.code !== 0) {
      if (isAuthPayloadError(payload.code, payload.msg)) {
        handleAuthExpired(payload.msg)
      }
      const err = new Error(payload.msg || 'request failed')
      err.code = payload.code
      err.data = payload.data
      throw err
    }
    return payload.data
  },
  (error) => {
    if (error?.response?.status === 401) {
      const msg = error?.response?.data?.msg
      handleAuthExpired(typeof msg === 'string' && msg.trim() !== '' ? msg.trim() : '登录状态已失效，请重新登录')
    }
    return Promise.reject(error)
  }
)

export default request
