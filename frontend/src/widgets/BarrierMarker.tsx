import { Marker } from 'react-leaflet'
import L from 'leaflet'
import { useMemo } from 'react'
import { Barrier, BarrierType, Severity } from '@entities/barrier/types'

interface BarrierMarkerProps {
  barrier: Barrier
  isSelected: boolean
  onClick: () => void
}

const barrierTypeIcons: Record<BarrierType, string> = {
  high_curb: '🚧',
  broken_elevator: '🛗',
  closed_sidewalk: '🚫',
  stairs: '🪜',
  pothole: '🕳️',
  uneven_surface: '📏',
  parked_car: '🚗',
}

const severityColors: Record<Severity, string> = {
  1: '#22c55e',
  2: '#84cc16',
  3: '#eab308',
  4: '#f97316',
  5: '#ef4444',
}

export function BarrierMarker({ barrier, isSelected, onClick }: BarrierMarkerProps) {
  const icon = barrierTypeIcons[barrier.type] || '📍'
  const color = severityColors[barrier.severity] || '#64748b'
  const size = 24 + barrier.severity * 4

  const divIcon = useMemo(() => {
    return L.divIcon({
      className: `barrier-marker ${isSelected ? 'selected' : ''}`,
      html: `<span class="marker-icon" style="width:${size}px;height:${size}px;border:3px solid ${color};background:var(--color-card);font-size:${Math.round(size * 0.5)}px">${icon}</span>${isSelected ? `<span class="marker-pulse" style="border-color:${color}"></span>` : ''}`,
      iconSize: [size, size],
      iconAnchor: [size / 2, size / 2],
      popupAnchor: [0, -size / 2],
    })
  }, [icon, color, size, isSelected])

  return (
    <Marker
      position={[barrier.coordinates.latitude, barrier.coordinates.longitude]}
      icon={divIcon}
      eventHandlers={{ click: onClick }}
    />
  )
}