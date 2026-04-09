import { defineStore } from 'pinia'
import { loginApi, logoutApi, profileApi } from '../api/auth'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('admin_token') || '',
    role: localStorage.getItem('admin_user_role') || ''
  }),
  getters: {
    isAuthed: (state) => !!state.token,
    isAdmin: (state) => state.role === 'admin'
  },
  actions: {
    async login(form) {
      const data = await loginApi(form)
      this.token = data.access_token
      localStorage.setItem('admin_token', this.token)
      const profile = await profileApi()
      this.role = profile.role || ''
      localStorage.setItem('admin_user_role', this.role)
    },
    async logout() {
      try {
        await logoutApi()
      } catch (_) {
        // ignore logout network failures
      }
      this.token = ''
      this.role = ''
      localStorage.removeItem('admin_token')
      localStorage.removeItem('admin_user_role')
    }
  }
})
