import { Routes, Route, Navigate, Outlet } from 'react-router-dom'
import { useEffect } from 'react'
import { MapPage } from '@pages/MapPage'
import { RouteBuilderPage } from '@pages/RouteBuilderPage'
import { BarrierFormPage } from '@pages/BarrierFormPage'
import { ModerationQueuePage } from '@pages/ModerationQueuePage'
import { PoiDetailPage } from '@pages/PoiDetailPage'
import { NotificationsPage } from '@pages/NotificationsPage'
import { LoginPage } from '@pages/LoginPage'
import { RegisterPage } from '@pages/RegisterPage'
import { ProfilePage } from '@pages/ProfilePage'
import { Header } from '@widgets/Header'
import { useAuthStore } from '@features/auth/store'
import { ProtectedRoute } from '@shared/ui/ProtectedRoute'

export function App() {
  const { initializeAuth } = useAuthStore()
  
  useEffect(() => {
    initializeAuth()
  }, [initializeAuth])

  return (
    <div className="app">
      <Header />
      <main className="main-content">
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          
          <Route element={<ProtectedRoute><Outlet /></ProtectedRoute>}>
            <Route path="/" element={<MapPage />} />
            <Route path="/route" element={<RouteBuilderPage />} />
            <Route path="/barrier/new" element={<BarrierFormPage />} />
            <Route path="/barrier/:id" element={<BarrierFormPage />} />
            <Route path="/moderation" element={<ModerationQueuePage />} />
            <Route path="/poi/:id" element={<PoiDetailPage />} />
            <Route path="/notifications" element={<NotificationsPage />} />
            <Route path="/profile" element={<ProfilePage />} />
          </Route>
          
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  )
}