import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { api } from '@shared/api/axios'
import { ModerationRequest, ModerationStatus } from '@entities/moderation/types'
import { Barrier } from '@entities/barrier/types'

function useBarrier(barrierId: string | null) {
  return useQuery({
    queryKey: ['moderation-barrier', barrierId],
    queryFn: async () => {
      const response = await api.get(`/barriers/${barrierId}`)
      return response.data as Barrier
    },
    enabled: !!barrierId,
  })
}

function ModerationCard({ request, onReviewed }: { request: ModerationRequest; onReviewed: () => void }) {
  const { data: barrier, isLoading } = useBarrier(request.barrier_id)
  const [comment, setComment] = useState('')

  const destroy = () => {
    setComment('')
    onReviewed()
  }

  return (
    <div className="moderation-card">
      <div className="moderation-card-header">
        <span className="request-id">#{request.id.slice(0, 8)}</span>
        <span className="request-date">
          {new Date(request.created_at).toLocaleString('ru-RU')}
        </span>
      </div>

      {isLoading ? (
        <div className="spinner" />
      ) : barrier ? (
        <>
          <div className="moderation-card-info">
            <p className="moderation-card-coords">
              Ш: {barrier.coordinates.latitude.toFixed(5)}, Д: {barrier.coordinates.longitude.toFixed(5)}
            </p>
            <p className="moderation-card-desc">
              {barrier.description || 'Описание отсутствует'}
            </p>
            <p className="moderation-card-meta">
              <span className="role-badge">{barrier.type}</span>
              <span className="role-badge">severity: {barrier.severity}</span>
            </p>
          </div>

          {barrier.photos.length > 0 ? (
            <div className="moderation-photos">
              {barrier.photos.map((photo) => (
                <img
                  key={photo.id}
                  src={photo.presigned_url}
                  alt={photo.original_filename}
                  className="moderation-photo"
                />
              ))}
            </div>
          ) : (
            <p className="empty-state">Фотографии не приложены</p>
          )}

          {request.status === 'pending' && (
            <div className="moderation-card-actions">
              <div className="form-group">
                <label htmlFor={`reject-comment-${request.id}`}>
                  Комментарий к решению
                </label>
                <textarea
                  id={`reject-comment-${request.id}`}
                  rows={2}
                  value={comment}
                  onChange={(e) => setComment(e.target.value)}
                  placeholder="Причина отклонения / замечание..."
                />
              </div>
              <div className="request-actions">
                <ModerationActionButton
                  requestId={request.id}
                  action="approve"
                  comment={comment}
                  onDone={destroy}
                />
                <ModerationActionButton
                  requestId={request.id}
                  action="reject"
                  comment={comment}
                  onDone={destroy}
                />
              </div>
            </div>
          )}
        </>
      ) : (
        <p className="alert-error">Барьер не найден</p>
      )}
    </div>
  )
}

function ModerationActionButton({ requestId, action, comment, onDone }: {
  requestId: string
  action: 'approve' | 'reject'
  comment: string
  onDone: () => void
}) {
  const mutation = useMutation({
    mutationFn: async () => {
      await api.post(`/moderation/queue/${requestId}/${action}`, { comment })
    },
    onSuccess: onDone,
  })

  const isApprove = action === 'approve'
  return (
    <button
      className={`btn ${isApprove ? 'btn-success' : 'btn-danger'} btn-sm`}
      onClick={() => mutation.mutate()}
      disabled={mutation.isPending}
    >
      {mutation.isPending ? 'Сохранение...' : isApprove ? 'Одобрить' : 'Отклонить'}
    </button>
  )
}

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
                <span
                  className="request-status"
                  style={{ backgroundColor: getStatusColor(request.status) }}
                >
                  {getStatusLabel(request.status)}
                </span>
                {request.moderator_comment && (
                  <span className="request-comment">{request.moderator_comment}</span>
                )}
              </div>
              <ModerationCard request={request} onReviewed={() => refetch()} />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}