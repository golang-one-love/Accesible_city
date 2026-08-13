import { useState, useEffect } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { api } from '@shared/api/axios'

interface Notification {
  id: string
  type: string
  title: string
  message: string
  payload: Record<string, string>
  is_read: boolean
  created_at: string
  read_at: string | null
}

const typeLabels: Record<string, string> = {
  barrier_approved: 'Барьер одобрен',
  barrier_rejected: 'Барьер отклонен',
  barrier_resolved: 'Барьер устранен',
  barrier_nearby: 'Барьер рядом',
  route_updated: 'Маршрут перестроен',
  new_poi: 'Новая точка',
}

const typeIcons: Record<string, string> = {
  barrier_approved: '✅',
  barrier_rejected: '❌',
  barrier_resolved: '🟢',
  barrier_nearby: '⚠️',
  route_updated: '🧭',
  new_poi: '📍',
}

export function NotificationsPage() {
  const [filter, setFilter] = useState<'all' | 'unread'>('all')

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['notifications', filter],
    queryFn: async () => {
      const params = new URLSearchParams()
      params.append('limit', '50')
      const response = await api.get(`/notifications?${params.toString()}`)
      return response.data
    },
  })

  const markAsReadMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.post('/notifications/read', { notification_id: id })
    },
    onSuccess: () => refetch(),
  })

  const markAllAsReadMutation = useMutation({
    mutationFn: async () => {
      await api.post('/notifications/read-all')
    },
    onSuccess: () => refetch(),
  })

  const unreadCountQuery = useQuery({
    queryKey: ['notifications-unread-count'],
    queryFn: async () => {
      const response = await api.get('/notifications/unread-count')
      return response.data.count
    },
    refetchInterval: 30000,
  })

  useEffect(() => {
    const token = localStorage.getItem('auth-storage')
    if (token) {
      try {
        const parsed = JSON.parse(token)
        if (parsed.state?.accessToken) {
          const wsProto = location.protocol === 'https:' ? 'wss' : 'ws'
          const wsUrl = `${wsProto}://${location.host}/api/notifications/ws?token=${parsed.state.accessToken}`
          const websocket = new WebSocket(wsUrl)
          websocket.onmessage = (event) => {
            const msg = JSON.parse(event.data)
            if (msg.type === 'notification') {
              refetch()
            }
          }
          return () => websocket.close()
        }
      } catch {
        // ignore
      }
    }
  }, [refetch])

  const filteredNotifications = data?.notifications.filter((n: Notification) => 
    filter === 'all' || !n.is_read
  ) || []

  if (isLoading) return <div className="loading">Загрузка...</div>

  return (
    <div className="notifications-page">
      <div className="page-header">
        <h1>Уведомления</h1>
        <div className="header-actions">
          <select value={filter} onChange={(e) => setFilter(e.target.value as any)} className="filter-select">
            <option value="all">Все</option>
            <option value="unread">Непрочитанные</option>
          </select>
          {unreadCountQuery.data && unreadCountQuery.data > 0 && (
            <button 
              className="btn btn-primary btn-sm" 
              onClick={() => markAllAsReadMutation.mutate()}
              disabled={markAllAsReadMutation.isPending}
            >
              Прочитать все ({unreadCountQuery.data})
            </button>
          )}
        </div>
      </div>

      {filteredNotifications.length === 0 ? (
        <div className="empty-state">
          <p>Уведомлений нет</p>
        </div>
      ) : (
        <div className="notifications-list">
          {filteredNotifications.map((notification: Notification) => (
            <div 
              key={notification.id} 
              className={`notification-item ${!notification.is_read ? 'unread' : ''}`}
            >
              <div className="notification-icon">
                {typeIcons[notification.type] || '🔔'}
              </div>
              <div className="notification-content">
                <div className="notification-header">
                  <h4>{typeLabels[notification.type] || notification.title}</h4>
                  <span className="notification-time">
                    {new Date(notification.created_at).toLocaleString('ru-RU')}
                  </span>
                </div>
                <p className="notification-message">{notification.message}</p>
                {notification.payload.barrier_id && (
                  <p className="notification-meta">Барьер: {notification.payload.barrier_id.slice(0, 8)}...</p>
                )}
              </div>
              {!notification.is_read && (
                <button
                  className="btn btn-sm btn-primary"
                  onClick={() => markAsReadMutation.mutate(notification.id)}
                  disabled={markAsReadMutation.isPending}
                >
                  Прочитать
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}