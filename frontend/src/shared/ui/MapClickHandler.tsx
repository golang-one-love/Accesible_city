import { useMapEvents } from 'react-leaflet'
import { Coordinates } from '@entities/barrier/types'

const roundCoord = (value: number) => Math.round(value * 1e6) / 1e6

export function MapClickHandler({ onMapClick }: { onMapClick: (point: Coordinates) => void }) {
  useMapEvents({
    click: (e) => {
      const target = e.originalEvent.target as HTMLElement | null
      if (target?.closest('.leaflet-marker-icon')) return
      onMapClick({ latitude: roundCoord(e.latlng.lat), longitude: roundCoord(e.latlng.lng) })
    },
  })
  return null
}