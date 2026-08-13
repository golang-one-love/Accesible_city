import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@shared/api/axios'
import { Barrier } from '@entities/barrier/types'
import { useAuthStore } from '@features/auth/store'

export function VolunteerPanel() {
  const user = useAuthStore((state) => state.user)
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(true)
  const [complaintFor, setComplaintFor] = useState<Barrier | null>(null)
  const [complaintReason, setComplaintReason] = useState('')

  const isVolunteer = user?.role === 'volunteer'

  const { data: pendingBarriers, isLoading, refetch } = useQuery({
    queryKey: ['volunteer-pending'],
    queryFn: async () => {
      const response = await api.get('/barriers?status=pending&limit=100')
      return response.data.barriers as Barrier[]
    },
    enabled: isVolunteer,
  })

  const confirmMutation = useMutation({
    mutationFn: async (barrierId: string) => {
      await api.post(`/barriers/${barrierId}/confirm`)
    },
    onSuccess: () => {
      refetch()
      queryClient.invalidateQueries({ queryKey: ['barriers'] })
    },
  })

  const complainMutation = useMutation({
    mutationFn: async ({ barrierId, reason }: { barrierId: string; reason: string }) => {
      await api.post(`/barriers/${barrierId}/complaint`, { reason })
    },
    onSuccess: () => {
      setComplaintFor(null)
      setComplaintReason('')
      refetch()
      queryClient.invalidateQueries({ queryKey: ['barriers'] })
    },
  })

  if (!isVolunteer) return null

  return (
    <div className="volunteer-panel">
      <button className="volunteer-panel-toggle" onClick={() => setOpen(!open)}>
        <span>{open ? '▾' : '▸'}</span> На проверке: {pendingBarriers?.length ?? 0}
      </button>

      {open && (
        <div className="volunteer-panel-body">
          {isLoading ? (
            <div className="spinner" />
          ) : !pendingBarriers || pendingBarriers.length === 0 ? (
            <p className="empty-state">Точек на проверке нет</p>
          ) : (
            <ul className="volunteer-list">
              {pendingBarriers.map((barrier) => (
                <li key={barrier.id} className="volunteer-item">
                  <div className="volunteer-item-info">
                    <span className="volunteer-item-coords">
                      {barrier.coordinates.latitude.toFixed(5)}, {barrier.coordinates.longitude.toFixed(5)}
                    </span>
                    <span className="volunteer-item-desc">{barrier.description || 'Без описания'}</span>
                  </div>
                  <div className="volunteer-item-actions">
                    <button
                      className="btn btn-success btn-sm"
                      onClick={() => confirmMutation.mutate(barrier.id)}
                      disabled={confirmMutation.isPending}
                    >
                      Подтвердить
                    </button>
                    <button
                      className="btn btn-danger btn-sm"
                      onClick={() => {
                        setComplaintFor(barrier)
                        setComplaintReason('')
                      }}
                    >
                      Опровергнуть
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      {complaintFor && (
        <div className="modal-overlay" onClick={() => setComplaintFor(null)}>
          <div className="modal complaint-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Опровергнуть барьер</h3>
              <button className="modal-close" onClick={() => setComplaintFor(null)}>×</button>
            </div>
            <div className="modal-body">
              <p className="complaint-barrier-coords">
                {complaintFor.coordinates.latitude.toFixed(5)}, {complaintFor.coordinates.longitude.toFixed(5)}
              </p>
              <div className="form-group">
                <label htmlFor="complaint-reason">Причина</label>
                <textarea
                  id="complaint-reason"
                  rows={3}
                  value={complaintReason}
                  onChange={(e) => setComplaintReason(e.target.value)}
                  placeholder="Почему точка не является барьером?"
                />
              </div>
              {complainMutation.isError && (
                <div className="alert-error">Не удалось отправить опровержение</div>
              )}
              <div className="modal-actions">
                <button className="btn btn-secondary" onClick={() => setComplaintFor(null)}>Отмена</button>
                <button
                  className="btn btn-danger"
                  onClick={() => complainMutation.mutate({ barrierId: complaintFor.id, reason: complaintReason })}
                  disabled={complainMutation.isPending || !complaintReason.trim()}
                >
                  {complainMutation.isPending ? 'Отправка...' : 'Опровергнуть'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}