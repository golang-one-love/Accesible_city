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

export function RoutePanel() {
  const [start, setStart] = useState<Coordinates | null>(null)
  const [finish, setFinish] = useState<Coordinates | null>(null)
  const [profile, setProfile] = useState<MobilityProfile>('default')
  const [isBuilding, setIsBuilding] = useState(false)

  const buildRouteMutation = useMutation({
    mutationFn: async (params: { start: Coordinates; finish: Coordinates; profile: MobilityProfile }) => {
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
    buildRouteMutation.mutate({ start, finish, profile })
  }

  return (
    <div className="route-panel">
      <div className="route-panel-header">
        <h3>🧭 Построить маршрут</h3>
      </div>
      
      <div className="route-form">
        <div className="form-group">
          <label>Точка начала</label>
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
          <label>Точка конца</label>
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
          disabled={!start || !finish || isBuilding}
        >
          {isBuilding ? 'Строим маршрут...' : 'Построить маршрут'}
        </button>

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