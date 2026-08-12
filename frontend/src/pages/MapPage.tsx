import { useEffect, useRef, useState } from 'react'
import { MapContainer, TileLayer, Popup, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { useQuery } from '@tanstack/react-query'
import { api } from '@shared/api/axios'
import { Barrier, BarrierType, BarrierStatus, Severity, Coordinates } from '@entities/barrier/types'
import { BarrierMarker } from '@widgets/BarrierMarker'
import { BarrierFormModal } from '@features/barriers/BarrierFormModal'
import { RoutePanel } from '@widgets/RoutePanel'
import { VolunteerPanel } from '@widgets/VolunteerPanel'
import { useBarrierStore } from '@features/barriers/store'

const iconDefault = L.icon({
  iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
  iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
  shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
})

L.Marker.prototype.options.icon = iconDefault

const MOSCOW_CENTER = [55.7558, 37.6173] as [number, number]

const roundCoord = (value: number) => Math.round(value * 1e6) / 1e6

function MapClickHandler({ onMapClick }: { onMapClick: (point: Coordinates) => void }) {
  useMapEvents({
    click: (e) => {
      const target = e.originalEvent.target as HTMLElement | null
      if (target?.closest('.leaflet-marker-icon')) return
      onMapClick({ latitude: roundCoord(e.latlng.lat), longitude: roundCoord(e.latlng.lng) })
    },
  })
  return null
}

interface MapViewProps {
  selectedBarrier: Barrier | null
  onSelectBarrier: (barrier: Barrier | null) => void
  onCloseModal: () => void
  onMapClick: (point: Coordinates) => void
}

function MapView({ selectedBarrier, onSelectBarrier, onCloseModal, onMapClick }: MapViewProps) {
  const { barriers, setBarriers, filters } = useBarrierStore()
  const mapRef = useRef<L.Map | null>(null)

  const { data: barriersData } = useQuery({
    queryKey: ['barriers', filters],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (filters.status) params.append('status', filters.status)
      if (filters.type) params.append('type', filters.type)
      if (filters.severity_min) params.append('severity_min', String(filters.severity_min))
      if (filters.severity_max) params.append('severity_max', String(filters.severity_max))
      if (filters.bounds) {
        params.append('sw_lat', String(filters.bounds[0].latitude))
        params.append('sw_lon', String(filters.bounds[0].longitude))
        params.append('ne_lat', String(filters.bounds[1].latitude))
        params.append('ne_lon', String(filters.bounds[1].longitude))
      }
      params.append('limit', '500')
      const response = await api.get(`/barriers?${params.toString()}`)
      return response.data.barriers
    },
    enabled: true,
  })

  useEffect(() => {
    if (barriersData) {
      setBarriers(barriersData)
    }
  }, [barriersData, setBarriers])

  return (
    <MapContainer
      center={MOSCOW_CENTER}
      zoom={13}
      scrollWheelZoom={true}
      style={{ height: '100%', width: '100%' }}
      ref={mapRef}
    >
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      
      <MapClickHandler onMapClick={onMapClick} />

      {barriers.map((barrier) => (
        <BarrierMarker
          key={barrier.id}
          barrier={barrier}
          isSelected={selectedBarrier?.id === barrier.id}
          onClick={() => onSelectBarrier(barrier)}
        />
      ))}

      {selectedBarrier && (
        <Popup
          position={[selectedBarrier.coordinates.latitude, selectedBarrier.coordinates.longitude]}
          eventHandlers={{ popupclose: onCloseModal }}
        >
          <div className="barrier-popup">
            <h4>{getBarrierTypeLabel(selectedBarrier.type)}</h4>
            <p className="severity-badge" style={{ backgroundColor: getSeverityColor(selectedBarrier.severity) }}>
              Severity: {selectedBarrier.severity}
            </p>
            <p>{selectedBarrier.description || 'Нет описания'}</p>
            <p className={`status-badge status-${selectedBarrier.status}`}>{getStatusLabel(selectedBarrier.status)}</p>
            <p className="popup-coords">
              Ш: {selectedBarrier.coordinates.latitude.toFixed(5)}, Д: {selectedBarrier.coordinates.longitude.toFixed(5)}
            </p>
            <button onClick={() => onCloseModal()} className="btn btn-sm">Закрыть</button>
          </div>
        </Popup>
      )}
    </MapContainer>
  )
}

function getBarrierTypeLabel(type: BarrierType): string {
  const labels: Record<BarrierType, string> = {
    high_curb: 'Высокий бордюр',
    broken_elevator: 'Сломанный лифт',
    closed_sidewalk: 'Перекрытый тротуар',
    stairs: 'Лестница',
    pothole: 'Яма',
    uneven_surface: 'Неровное покрытие',
    parked_car: 'Припаркованная машина',
  }
  return labels[type] || type
}

function getStatusLabel(status: BarrierStatus): string {
  const labels: Record<BarrierStatus, string> = {
    pending: 'На модерации',
    approved: 'Подтвержден',
    rejected: 'Отклонен',
    resolved: 'Устранен',
  }
  return labels[status] || status
}

function getSeverityColor(severity: Severity): string {
  if (severity <= 2) return '#22c55e'
  if (severity === 3) return '#eab308'
  if (severity === 4) return '#f97316'
  return '#ef4444'
}

function MapControls({ onAddBarrier }: { onAddBarrier: () => void }) {
  return (
    <div className="map-controls">
      <button className="btn btn-primary map-btn" onClick={onAddBarrier}>
        + Добавить барьер
      </button>
    </div>
  )
}

function ClickPointPanel({ point, onPlaceBarrier, onHide }: {
  point: Coordinates
  onPlaceBarrier: () => void
  onHide: () => void
}) {
  return (
    <div className="click-point-panel">
      <div className="click-point-coords">
        <span>Широта: <b>{point.latitude.toFixed(6)}</b></span>
        <span>Долгота: <b>{point.longitude.toFixed(6)}</b></span>
      </div>
      <div className="click-point-actions">
        <button className="btn btn-primary btn-sm" onClick={onPlaceBarrier}>Поставить барьер</button>
        <button className="btn btn-secondary btn-sm" onClick={onHide}>Скрыть</button>
      </div>
    </div>
  )
}

export function MapPage() {
  const [selectedBarrier, setSelectedBarrier] = useState<Barrier | null>(null)
  const [clickPoint, setClickPoint] = useState<Coordinates | null>(null)
  const [formOpen, setFormOpen] = useState(false)

  const handleMapClick = (point: Coordinates) => {
    setSelectedBarrier(null)
    setClickPoint(point)
  }

  const handleAddBarrier = () => {
    setSelectedBarrier(null)
    setClickPoint(null)
    setFormOpen(true)
  }

  const handlePlaceBarrier = () => {
    setFormOpen(true)
  }

  const closeForm = () => {
    setFormOpen(false)
    setClickPoint(null)
  }

  return (
    <div className="map-page">
      <div className="map-container">
        <MapView
          selectedBarrier={selectedBarrier}
          onSelectBarrier={setSelectedBarrier}
          onCloseModal={() => setSelectedBarrier(null)}
          onMapClick={handleMapClick}
        />
        <MapControls onAddBarrier={handleAddBarrier} />
        {clickPoint && !formOpen && (
          <ClickPointPanel
            point={clickPoint}
            onPlaceBarrier={handlePlaceBarrier}
            onHide={() => setClickPoint(null)}
          />
        )}
      </div>

      {formOpen && (
        <BarrierFormModal
          initialCoordinates={clickPoint ?? undefined}
          onClose={closeForm}
          onSave={() => {
            closeForm()
            setSelectedBarrier(null)
          }}
        />
      )}
      
      <VolunteerPanel />
      <RoutePanel />
    </div>
  )
}