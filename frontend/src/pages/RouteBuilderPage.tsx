import { useState } from 'react'
import { MapContainer, TileLayer, Marker, Polyline } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { useMutation } from '@tanstack/react-query'
import { api } from '@shared/api/axios'
import { Coordinates, MobilityProfile, RouteNode } from '@entities/route/types'
import { MapClickHandler } from '@shared/ui/MapClickHandler'

const MOBILITY_PROFILES: { value: MobilityProfile; label: string; icon: string }[] = [
  { value: 'wheelchair', label: 'Инвалидное кресло', icon: '♿' },
  { value: 'stroller', label: 'Коляска', icon: '👶' },
  { value: 'elderly', label: 'Пожилой человек', icon: '👴' },
  { value: 'default', label: 'Обычный пешеход', icon: '🚶' },
]

const MOSCOW_CENTER = [55.7558, 37.6173] as [number, number]

const startIcon = L.divIcon({
  className: 'route-point-marker',
  html: '<span class="route-point-icon route-point-start">A</span>',
  iconSize: [30, 30],
  iconAnchor: [15, 15],
})

const finishIcon = L.divIcon({
  className: 'route-point-marker',
  html: '<span class="route-point-icon route-point-finish">B</span>',
  iconSize: [30, 30],
  iconAnchor: [15, 15],
})

function pointToLatLng(p: Coordinates): [number, number] {
  return [p.latitude, p.longitude]
}

export function RouteBuilderPage() {
  const [start, setStart] = useState<Coordinates | null>(null)
  const [finish, setFinish] = useState<Coordinates | null>(null)
  const [clickPoint, setClickPoint] = useState<Coordinates | null>(null)
  const [profile, setProfile] = useState<MobilityProfile>('default')
  const [routeResult, setRouteResult] = useState<any>(null)
  const [error, setError] = useState<string | null>(null)

  const buildRouteMutation = useMutation({
    mutationFn: async (params: { start: Coordinates; finish: Coordinates; profile: MobilityProfile }) => {
      const response = await api.post('/routes/build', params)
      return response.data
    },
    onSuccess: (data) => {
      setRouteResult(data)
      setError(null)
    },
    onError: (error: any) => {
      setError(error.response?.data?.error || 'Не удалось построить маршрут')
      setRouteResult(null)
    },
  })

  const handleMapClick = (point: Coordinates) => {
    setClickPoint(point)
  }

  const handleSetStart = (point: Coordinates) => {
    setStart(point)
    setClickPoint(null)
    setRouteResult(null)
  }

  const handleSetFinish = (point: Coordinates) => {
    setFinish(point)
    setClickPoint(null)
    setRouteResult(null)
  }

  const routeNodes: [number, number][] = (routeResult?.nodes ?? []).map((n: RouteNode) => [n.latitude, n.longitude])

  return (
    <div className="route-page">
      <div className="page-header">
        <h1>Построение маршрута</h1>
      </div>

      <div className="route-builder">
        <div className="route-controls-panel">
          <div className="form-group">
            <label>Отсюда (A)</label>
            {start ? (
              <div className="route-point-set">
                <span className="route-point-coords">{start.latitude}, {start.longitude}</span>
                <button className="btn btn-sm btn-secondary" onClick={() => { setStart(null); setRouteResult(null) }}>
                  Очистить
                </button>
              </div>
            ) : (
              <p className="route-hint">Кликните по карте и выберите «Отсюда»</p>
            )}
          </div>

          <div className="form-group">
            <label>Сюда (B)</label>
            {finish ? (
              <div className="route-point-set">
                <span className="route-point-coords">{finish.latitude}, {finish.longitude}</span>
                <button className="btn btn-sm btn-secondary" onClick={() => { setFinish(null); setRouteResult(null) }}>
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
            onClick={() => buildRouteMutation.mutate({ start: start!, finish: finish!, profile })}
            disabled={!start || !finish || buildRouteMutation.isPending}
          >
            {buildRouteMutation.isPending ? 'Строим маршрут...' : 'Построить маршрут'}
          </button>

          {error && <div className="alert-error">{error}</div>}

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
          <MapContainer
            center={MOSCOW_CENTER}
            zoom={13}
            scrollWheelZoom={true}
            style={{ height: '100%', width: '100%' }}
          >
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />

            <MapClickHandler onMapClick={handleMapClick} />

            {start && <Marker position={pointToLatLng(start)} icon={startIcon} />}
            {finish && <Marker position={pointToLatLng(finish)} icon={finishIcon} />}

            {routeNodes.length > 1 && (
              <Polyline
                positions={routeNodes}
                pathOptions={{ color: '#2563eb', weight: 5, opacity: 0.85 }}
              />
            )}
          </MapContainer>

          {clickPoint && (
            <div className="click-point-panel">
              <div className="click-point-coords">
                <span>Широта: <b>{clickPoint.latitude}</b></span>
                <span>Долгота: <b>{clickPoint.longitude}</b></span>
              </div>
              <div className="click-point-actions">
                <button className="btn btn-primary btn-sm" onClick={() => handleSetStart(clickPoint)}>
                  Отсюда
                </button>
                <button className="btn btn-success btn-sm" onClick={() => handleSetFinish(clickPoint)}>
                  Сюда
                </button>
                <button className="btn btn-secondary btn-sm" onClick={() => setClickPoint(null)}>
                  Скрыть
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}