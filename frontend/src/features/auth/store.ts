import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { api } from '@shared/api/axios'

export interface User {
  id: string
  email: string
  nickname: string
  role: string
  is_active: boolean
  created_at: string
}

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isInitialized: boolean
  setTokens: (access: string, refresh: string) => void
  setUser: (user: User) => void
  updateUser: (patch: Partial<User>) => void
  logout: () => void
  initializeAuth: () => Promise<void>
  updateAccessToken: (token: string) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isInitialized: false,

      setTokens: (access, refresh) => {
        set({ accessToken: access, refreshToken: refresh })
        api.defaults.headers.common['Authorization'] = `Bearer ${access}`
      },

      setUser: (user) => set({ user }),

      updateUser: (patch) => set({ user: { ...get().user!, ...patch } }),

      logout: () => {
        set({ user: null, accessToken: null, refreshToken: null })
        delete api.defaults.headers.common['Authorization']
      },

      initializeAuth: async () => {
        const { refreshToken } = get()
        if (!refreshToken) {
          set({ isInitialized: true })
          return
        }

        try {
          const response = await api.post('/auth/refresh', { refresh_token: refreshToken })
          set({ accessToken: response.data.access_token, refreshToken: response.data.refresh_token })
          api.defaults.headers.common['Authorization'] = `Bearer ${response.data.access_token}`
          
          const userResponse = await api.get('/auth/profile')
          set({ user: { nickname: '', ...userResponse.data }, isInitialized: true })
        } catch {
          get().logout()
          set({ isInitialized: true })
        }
      },

      updateAccessToken: (token) => {
        set({ accessToken: token })
        api.defaults.headers.common['Authorization'] = `Bearer ${token}`
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
    }
  )
)