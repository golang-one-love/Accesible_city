import { Marker } from 'react-leaflet'
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

  return (
    <Marker
      position={[barrier.coordinates.latitude, barrier.coordinates.longitude]}
      eventHandlers={{ click: onClick }}
    >
      <div
        className={`barrier-marker ${isSelected ? 'selected' : ''}`}
        style={{
          width: size,
          height: size,
          borderColor: color,
          backgroundColor: color,
        } as React.CSSProperties}
        role="button"
        tabIndex={0}
        onKeyDown={(e) => e.key === 'Enter' && onClick()}
      >
        <span style={{ fontSize: size * 0.5 }}>{icon}</span>
        {isSelected && <div className="marker-pulse" style={{ borderColor: color }} />}
      </div>
    </Marker>
  )
}