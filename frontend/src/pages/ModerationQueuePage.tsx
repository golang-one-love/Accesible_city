import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { api } from '@shared/api/axios'

interface ModerationRequest {
  id: string
  barrier_id: string
  reporter_id: string
  status: 'pending' | 'approved' | 'rejected'
  moderator_id: string | null
  moderator_comment: string
  created_at: string
  updated_at: string
  reviewed_at: string | null
}

type ModerationStatus = 'pending' | 'approved' | 'rejected'

export function ModerationQueuePage() {
  const [statusFilter, setStatusFilter] = useState<ModerationStatus | 'all'>('all')

  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['moderation-queue', statusFilter],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (statusFilter !== 'all') params.append('status', statusFilter)
      params.append('limit', '50')
      const response = await api.get(`/moderation/queue?${params.toString()}`)
      return response.data
    },
  })

  const approveMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/moderation/queue/${id}/approve`, { comment: '' })
    },
    onSuccess: () => refetch(),
  })

  const rejectMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/moderation/queue/${id}/reject`, { comment: '' })
    },
    onSuccess: () => refetch(),
  })

  const getStatusLabel = (status: ModerationStatus) => ({
    pending: 'На рассмотрении',
    approved: 'Одобрен',
    rejected: 'Отклонен',
  }[status] || status)

  const getStatusColor = (status: ModerationStatus) => ({
    pending: '#eab308',
    approved: '#22c55e',
    rejected: '#ef4444',
  }[status] || '#64748b')

  if (isLoading) return <div className="loading">Загрузка...</div>
  if (error) return <div className="error">Ошибка загрузки</div>

  return (
    <div className="moderation-page">
      <div className="page-header">
        <h1>Очередь модерации</h1>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as any)}
          className="filter-select"
        >
          <option value="all">Все</option>
          <option value="pending">На рассмотрении</option>
          <option value="approved">Одобренные</option>
          <option value="rejected">Отклоненные</option>
        </select>
      </div>

      {data?.requests.length === 0 ? (
        <div className="empty-state">
          <p>Заявок нет</p>
        </div>
      ) : (
        <div className="moderation-list">
          {data?.requests.map((request: ModerationRequest) => (
            <div key={request.id} className="moderation-item">
              <div className="request-info">
                <span className="request-id">#{request.id.slice(0, 8)}</span>
                <span 
                  className="request-status" 
                  style={{ backgroundColor: getStatusColor(request.status) }}
                >
                  {getStatusLabel(request.status)}
                </span>
                <span className="request-date">
                  {new Date(request.created_at).toLocaleString('ru-RU')}
                </span>
              </div>
              <div className="request-actions">
                {request.status === 'pending' && (
                  <>
                    <button
                      className="btn btn-success btn-sm"
                      onClick={() => approveMutation.mutate(request.id)}
                      disabled={approveMutation.isPending}
                    >
                      Одобрить
                    </button>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => rejectMutation.mutate(request.id)}
                      disabled={rejectMutation.isPending}
                    >
                      Отклонить
                    </button>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}