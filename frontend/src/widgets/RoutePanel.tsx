import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '@shared/api/axios'
import { Coordinates, MobilityProfile } from '@entities/route/types'

const MOBILITY_PROFILES: { value: MobilityProfile; label: string; icon: string }[] = [
  { value: 'wheelchair', label: 'Инвалидное кресло', icon: '♿' },
  { value: 'stroller', label: 'Коляска', icon: '👶' },
  { value: 'elderly', label: 'Пожилой человек', icon: '👴' },
  { value: 'default', label: 'Обычный пешеход', icon: '🚶' },
]

export function RoutePanel({
  clickPoint,
  onPointSelected,
}: {
  clickPoint: Coordinates | null
  onPointSelected: () => void
}) {
  const [start, setStart] = useState<Coordinates | null>(null)
  const [finish, setFinish] = useState<Coordinates | null>(null)
  const [profile, setProfile] = useState<MobilityProfile>('default')
  const [isBuilding, setIsBuilding] = useState(false)

  const buildRouteMutation = useMutation({
    mutationFn: async (params: { start: Coordinates; finish: Coordinates; mobility_profile: MobilityProfile }) => {
      const response = await api.post('/routes/build', params)
      return response.data
    },
    onSuccess: (data) => {
      console.log('Route built:', data)
      setIsBuilding(false)
    },
    onError: (error) => {
      console.error('Route error:', error)
      setIsBuilding(false)
    },
  })

  const handleBuildRoute = () => {
    if (!start || !finish) return
    setIsBuilding(true)
    buildRouteMutation.mutate({ start, finish, mobility_profile: profile })
  }

  return (
    <div className="route-panel">
      <div className="route-panel-header">
        <h3>🧭 Построить маршрут</h3>
      </div>

      {clickPoint && (
        <div className="route-pick-point">
          <p className="route-hint">Точка на карте: {clickPoint.latitude}, {clickPoint.longitude}</p>
          <div className="click-point-actions">
            <button
              className="btn btn-primary btn-sm"
              onClick={() => {
                setStart(clickPoint)
                onPointSelected()
              }}
            >
              Отсюда
            </button>
            <button
              className="btn btn-success btn-sm"
              onClick={() => {
                setFinish(clickPoint)
                onPointSelected()
              }}
            >
              Сюда
            </button>
            <button className="btn btn-secondary btn-sm" onClick={onPointSelected}>
              Отмена
            </button>
          </div>
        </div>
      )}

      <div className="route-form">
        <div className="form-group">
          <label>Точка начала</label>
          {start ? (
            <div className="route-point-set">
              <span className="route-point-coords">{start.latitude}, {start.longitude}</span>
              <button className="btn btn-sm btn-secondary" onClick={() => setStart(null)}>
                Очистить
              </button>
            </div>
          ) : (
            <p className="route-hint">Кликните по карте и выберите «Отсюда»</p>
          )}
        </div>

        <div className="form-group">
          <label>Точка конца</label>
          {finish ? (
            <div className="route-point-set">
              <span className="route-point-coords">{finish.latitude}, {finish.longitude}</span>
              <button className="btn btn-sm btn-secondary" onClick={() => setFinish(null)}>
                Очистить
              </button>
            </div>
          ) : (
            <p className="route-hint">Кликните по карте и выберите «Сюда»</p>
          )}
        </div>

        <div className="form-group">
          <label>Профиль подвижности</label>
          <div className="profile-selector">
            {MOBILITY_PROFILES.map((p) => (
              <button
                key={p.value}
                className={`profile-btn ${profile === p.value ? 'active' : ''}`}
                onClick={() => setProfile(p.value)}
              >
                <span>{p.icon}</span>
                <span>{p.label}</span>
              </button>
            ))}
          </div>
        </div>

        <button
          className="btn btn-primary btn-block"
          onClick={handleBuildRoute}
          disabled={!start || !finish || isBuilding}
        >
          {isBuilding ? 'Строим маршрут...' : 'Построить маршрут'}
        </button>

        {buildRouteMutation.error && (
          <div className="alert-error">
            {(buildRouteMutation.error as any)?.response?.data?.error || 'Не удалось построить маршрут'}
          </div>
        )}

        {buildRouteMutation.data && (
          <div className="route-result">
            <h4>Маршрут построен</h4>
            <p>Расстояние: {(buildRouteMutation.data.total_distance / 1000).toFixed(2)} км</p>
            <p>Макс. серьезность: {buildRouteMutation.data.max_severity}/5</p>
            <p>Точек: {buildRouteMutation.data.nodes.length}</p>
          </div>
        )}
      </div>
    </div>
  )
}