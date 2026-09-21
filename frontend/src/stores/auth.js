import { defineStore } from 'pinia'
import api from '../services/api'
import { hasAnyPermission, isStaffUser, canAccessTab } from '../utils/access'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('cbt_user') || 'null'),
    profile: JSON.parse(localStorage.getItem('cbt_profile') || 'null'),
    token: localStorage.getItem('cbt_token') || null,
    loading: false,
    error: null,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token && !!state.user,
    role: (state) => state.user?.role || null,
    isStudent: (state) => state.user?.role === 'SISWA',
    isStaff: (state) => isStaffUser(state.user),
    permissions: (state) => state.user?.permissions || [],
    hasAnyPermission: (state) => (perms) => hasAnyPermission(state.user, perms),
    canAccessTab: (state) => (tab) => canAccessTab(state.user, tab),
    hasPermission: (state) => (perm) => {
      if (!state.user) return false
      if (state.user.role === 'ADMIN') return true
      const perms = state.user?.permissions || []
      if (perms.includes('*')) return true
      return perms.includes(perm)
    },
    canReadAllQuestions: (state) => {
      if (!state.user) return false
      if (state.user.role === 'ADMIN') return true
      const perms = state.user?.permissions || []
      return perms.includes('*') || perms.includes('questions:read_all')
    },
  },
  actions: {
    async login(username, password) {
      this.loading = true
      this.error = null
      try {
        const res = await api.post('/auth/login', { username, password })
        const data = res.data.data
        this.token = data.token
        this.user = data.user
        this.profile = data.student_profile || null

        localStorage.setItem('cbt_token', this.token)
        localStorage.setItem('cbt_user', JSON.stringify(this.user))
        if (this.profile) {
          localStorage.setItem('cbt_profile', JSON.stringify(this.profile))
        }
        return data
      } catch (err) {
        this.error = err.response?.data?.message || 'Gagal login, periksa koneksi dan kredensial Anda.'
        throw err
      } finally {
        this.loading = false
      }
    },

    async fetchMe() {
      if (!this.token) return
      try {
        const res = await api.get('/auth/me')
        this.user = res.data.data.user
        this.profile = res.data.data.student_profile || null
        localStorage.setItem('cbt_user', JSON.stringify(this.user))
        if (this.profile) {
          localStorage.setItem('cbt_profile', JSON.stringify(this.profile))
        }
      } catch (err) {
        console.error('Failed to refresh user profile', err)
      }
    },

    async logout() {
      try {
        await api.post('/auth/logout')
      } catch (e) {
        // ignore
      }
      this.user = null
      this.profile = null
      this.token = null
      localStorage.removeItem('cbt_token')
      localStorage.removeItem('cbt_user')
      localStorage.removeItem('cbt_profile')
    }
  }
})
