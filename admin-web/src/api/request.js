import axios from 'axios'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || '/api/v1',
  timeout: 30000
})

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
      const err = new Error(payload.msg || 'request failed')
      err.code = payload.code
      err.data = payload.data
      throw err
    }
    return payload.data
  },
  (error) => {
    if (error?.response?.status === 401) {
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_user_role')
      if (location.pathname !== '/login') {
        location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export default request
