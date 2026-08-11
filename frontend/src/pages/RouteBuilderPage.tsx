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

export function RouteBuilderPage() {
  const [start, setStart] = useState<Coordinates | null>(null)
  const [finish, setFinish] = useState<Coordinates | null>(null)
  const [profile, setProfile] = useState<MobilityProfile>('default')
  const [routeResult, setRouteResult] = useState<any>(null)

  const buildRouteMutation = useMutation({
    mutationFn: async (params: { start: Coordinates; finish: Coordinates; profile: MobilityProfile }) => {
      const response = await api.post('/routes/build', params)
      return response.data
    },
    onSuccess: (data) => {
      setRouteResult(data)
    },
    onError: (error) => {
      console.error('Route error:', error)
      alert('Не удалось построить маршрут')
    },
  })

  const handleBuildRoute = () => {
    if (!start || !finish) return
    buildRouteMutation.mutate({ start, finish, profile })
  }

  return (
    <div className="route-page">
      <div className="page-header">
        <h1>Построение маршрута</h1>
      </div>

      <div className="route-builder">
        <div className="route-form-panel">
          <div className="form-group">
            <label>Откуда</label>
            <div className="coords-input">
              <input
                type="number"
                step="0.000001"
                placeholder="Широта"
                value={start?.latitude || ''}
                onChange={(e) => setStart({ ...(start || { latitude: 0, longitude: 0 }), latitude: parseFloat(e.target.value) })}
              />
              <input
                type="number"
                step="0.000001"
                placeholder="Долгота"
                value={start?.longitude || ''}
                onChange={(e) => setStart({ ...(start || { latitude: 0, longitude: 0 }), longitude: parseFloat(e.target.value) })}
              />
            </div>
            <button className="btn btn-sm btn-secondary" onClick={() => navigator.geolocation.getCurrentPosition((pos) => setStart({ latitude: pos.coords.latitude, longitude: pos.coords.longitude }))}>
              Моя позиция
            </button>
          </div>

          <div className="form-group">
            <label>Куда</label>
            <div className="coords-input">
              <input
                type="number"
                step="0.000001"
                placeholder="Широта"
                value={finish?.latitude || ''}
                onChange={(e) => setFinish({ ...(finish || { latitude: 0, longitude: 0 }), latitude: parseFloat(e.target.value) })}
              />
              <input
                type="number"
                step="0.000001"
                placeholder="Долгота"
                value={finish?.longitude || ''}
                onChange={(e) => setFinish({ ...(finish || { latitude: 0, longitude: 0 }), longitude: parseFloat(e.target.value) })}
              />
            </div>
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
            disabled={!start || !finish || buildRouteMutation.isPending}
          >
            {buildRouteMutation.isPending ? 'Строим маршрут...' : 'Построить маршрут'}
          </button>

          {routeResult && (
            <div className="route-result">
              <h4>Маршрут построен</h4>
              <p>Расстояние: {(routeResult.total_distance / 1000).toFixed(2)} км</p>
              <p>Макс. серьезность препятствий: {routeResult.max_severity}/5</p>
              <p>Количество точек: {routeResult.nodes.length}</p>
            </div>
          )}
        </div>

        <div className="route-map-panel">
          <div className="map-placeholder">
            <p>Карта маршрута будет здесь</p>
            <p className="hint">Кликните по карте для выбора точек</p>
          </div>
        </div>
      </div>
    </div>
  )
}