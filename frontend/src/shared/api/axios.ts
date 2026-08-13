import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth-storage')
    if (token) {
      try {
        const parsed = JSON.parse(token)
        if (parsed.state?.accessToken) {
          config.headers.Authorization = `Bearer ${parsed.state.accessToken}`
        }
      } catch {
        // ignore
      }
    }
    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true
      
      try {
        const token = localStorage.getItem('auth-storage')
        if (token) {
          const parsed = JSON.parse(token)
          if (parsed.state?.refreshToken) {
            const response = await axios.post(
              `${import.meta.env.VITE_API_BASE_URL || '/api'}/auth/refresh`,
              { refresh_token: parsed.state.refreshToken }
            )
            
            const newAccessToken = response.data.access_token
            const newRefreshToken = response.data.refresh_token
            
            const updatedStorage = {
              ...parsed,
              state: {
                ...parsed.state,
                accessToken: newAccessToken,
                refreshToken: newRefreshToken,
              },
            }
            localStorage.setItem('auth-storage', JSON.stringify(updatedStorage))
            
            originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
            return api(originalRequest)
          }
        }
      } catch {
        localStorage.removeItem('auth-storage')
        window.location.href = '/login'
      }
    }
    
    return Promise.reject(error)
  }
)