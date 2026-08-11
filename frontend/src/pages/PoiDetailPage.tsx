import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '@shared/api/axios'

interface POI {
  id: string
  name: string
  category: string
  latitude: number
  longitude: number
  address: string
  phone: string
  website: string
  opening_hours: string
  accessibility: {
    features: string[]
    entrance_step_height_cm: number | null
    door_width_cm: number | null
    has_accessible_toilet: boolean
    notes: string
  }
  owner_id: string | null
  is_verified: boolean
  created_at: string
  updated_at: string
}

const categoryLabels: Record<string, string> = {
  restaurant: 'Ресторан',
  cafe: 'Кафе',
  shop: 'Магазин',
  pharmacy: 'Аптека',
  hospital: 'Больница',
  clinic: 'Клиника',
  bank: 'Банк',
  post_office: 'Почта',
  government: 'Госучреждение',
  park: 'Парк',
  museum: 'Музей',
  theater: 'Театр',
  library: 'Библиотека',
  school: 'Школа',
  university: 'Университет',
  hotel: 'Отель',
  transport: 'Транспорт',
  other: 'Другое',
}

const featureLabels: Record<string, string> = {
  ramp: 'Пандус',
  elevator: 'Лифт',
  wide_door: 'Широкая дверь',
  accessible_toilet: 'Доступный туалет',
  tactile_paving: 'Тактильное покрытие',
  braille_signs: 'Знаки Брайля',
  audio_guide: 'Аудиогид',
  low_counter: 'Низкая стойка',
  parking: 'Парковка',
  induction_loop: 'Индукционная петля',
}

export function PoiDetailPage() {
  const { id } = useParams<{ id: string }>()

  const { data: poi, isLoading, error } = useQuery<POI>({
    queryKey: ['poi', id],
    queryFn: async () => {
      const response = await api.get(`/poi/${id}`)
      return response.data
    },
    enabled: !!id,
  })

  if (isLoading) return <div className="loading">Загрузка...</div>
  if (error || !poi) return <div className="error">Точка не найдена</div>

  return (
    <div className="poi-page">
      <div className="page-header">
        <h1>{poi.name}</h1>
        <span className="category-badge">{categoryLabels[poi.category] || poi.category}</span>
      </div>

      <div className="poi-content">
        <div className="poi-main">
          <div className="poi-map-placeholder">
            <p>Карта: {poi.latitude}, {poi.longitude}</p>
          </div>

          <div className="poi-info">
            <div className="info-section">
              <h3>Адрес</h3>
              <p>{poi.address}</p>
            </div>

            {poi.phone && (
              <div className="info-section">
                <h3>Телефон</h3>
                <p>{poi.phone}</p>
              </div>
            )}

            {poi.website && (
              <div className="info-section">
                <h3>Сайт</h3>
                <p><a href={poi.website} target="_blank" rel="noopener">{poi.website}</a></p>
              </div>
            )}

            {poi.opening_hours && (
              <div className="info-section">
                <h3>Часы работы</h3>
                <p>{poi.opening_hours}</p>
              </div>
            )}

            <div className="info-section">
              <h3>Доступность</h3>
              <div className="accessibility-features">
                {poi.accessibility.features.map((feature: string) => (
                  <span key={feature} className="feature-tag">
                    {featureLabels[feature] || feature}
                  </span>
                ))}
              </div>
              
              {poi.accessibility.entrance_step_height_cm && (
                <p>Высота порога: {poi.accessibility.entrance_step_height_cm} см</p>
              )}
              {poi.accessibility.door_width_cm && (
                <p>Ширина двери: {poi.accessibility.door_width_cm} см</p>
              )}
              {poi.accessibility.has_accessible_toilet && (
                <p className="feature-available">✓ Есть доступный туалет</p>
              )}
              {poi.accessibility.notes && (
                <p className="accessibility-notes">{poi.accessibility.notes}</p>
              )}
            </div>

            <div className="info-section">
              <h3>Статус</h3>
              <p>{poi.is_verified ? '✓ Проверено' : '⚠ Не проверено'}</p>
              <p>Создано: {new Date(poi.created_at).toLocaleString('ru-RU')}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}