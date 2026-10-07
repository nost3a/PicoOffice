import { defineStore } from 'pinia'
import { apiLogin, apiRegister, apiMe } from '@/api'

// auth state: token + current user
export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('pico_token') || '',
    user: JSON.parse(localStorage.getItem('pico_user') || 'null')
  }),
  getters: {
    isLogin: (s) => !!s.token,
    isAdmin: (s) => s.user?.role === 'admin'
  },
  actions: {
    async login(form) {
      const resp = await apiLogin(form)
      this.token = resp.token
      this.user = resp.user
      localStorage.setItem('pico_token', resp.token)
      localStorage.setItem('pico_user', JSON.stringify(resp.user))
      return resp
    },
    async register(form) {
      return await apiRegister(form)
    },
    async refreshMe() {
      if (!this.token) return
      try {
        const resp = await apiMe()
        this.user = resp.user || resp
        localStorage.setItem('pico_user', JSON.stringify(this.user))
      } catch (e) {
        // 401 already handled by interceptor
      }
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('pico_token')
      localStorage.removeItem('pico_user')
    }
  }
})
