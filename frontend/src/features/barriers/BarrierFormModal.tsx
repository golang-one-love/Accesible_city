import { useState } from 'react'
import { BarrierType, Severity, Coordinates } from '@entities/barrier/types'
import { api } from '@shared/api/axios'
import { useBarrierStore } from '@features/barriers/store'

interface BarrierFormModalProps {
  initialCoordinates?: Coordinates
  onClose: () => void
  onSave: () => void
}

const BARRIER_TYPES: { value: BarrierType; label: string }[] = [
  { value: 'high_curb', label: 'Высокий бордюр' },
  { value: 'broken_elevator', label: 'Сломанный лифт' },
  { value: 'closed_sidewalk', label: 'Перекрытый тротуар' },
  { value: 'stairs', label: 'Лестница' },
  { value: 'pothole', label: 'Яма' },
  { value: 'uneven_surface', label: 'Неровное покрытие' },
  { value: 'parked_car', label: 'Припаркованная машина' },
]

const SEVERITY_OPTIONS: { value: Severity; label: string }[] = [
  { value: 1, label: '1 - Низкая' },
  { value: 2, label: '2 - Средняя' },
  { value: 3, label: '3 - Высокая' },
  { value: 4, label: '4 - Критическая' },
  { value: 5, label: '5 - Блокирующая' },
]

export function BarrierFormModal({ initialCoordinates, onClose, onSave }: BarrierFormModalProps) {
  const { addBarrier } = useBarrierStore()
  const [formData, setFormData] = useState<{
    type: BarrierType
    latitude: string
    longitude: string
    description: string
    severity: Severity
  }>({
    type: 'high_curb',
    latitude: initialCoordinates ? String(initialCoordinates.latitude) : '',
    longitude: initialCoordinates ? String(initialCoordinates.longitude) : '',
    description: '',
    severity: 2,
  })
  const [photos, setPhotos] = useState<File[]>([])
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)

    const latitude = parseFloat(formData.latitude)
    const longitude = parseFloat(formData.longitude)
    if (isNaN(latitude) || isNaN(longitude)) {
      setError('Укажите широту и долготу')
      setIsSubmitting(false)
      return
    }

    try {
      const response = await api.post('/barriers', {
        type: formData.type,
        coordinates: { latitude, longitude },
        description: formData.description,
        severity: formData.severity,
      })
      const created = response.data
      addBarrier(created)

      for (const file of photos) {
        const formDataBody = new FormData()
        formDataBody.append('file', file)
        await api.post(`/barriers/${created.id}/photos`, formDataBody, {
          headers: { 'Content-Type': 'multipart/form-data' },
        })
      }

      onSave()
    } catch (err: any) {
      setError(err.response?.data?.error || err.response?.data?.detail || 'Ошибка при сохранении')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal barrier-form-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>Добавить барьер</h3>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>
        
        <form onSubmit={handleSubmit} className="modal-body">
          {error && <div className="error-message">{error}</div>}
          
          <div className="form-group">
            <label htmlFor="type">Тип барьера *</label>
            <select
              id="type"
              value={formData.type}
              onChange={(e) => setFormData({ ...formData, type: e.target.value as BarrierType })}
              required
            >
              {BARRIER_TYPES.map((t) => (
                <option key={t.value} value={t.value}>{t.label}</option>
              ))}
            </select>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="latitude">Широта *</label>
              <input
                id="latitude"
                type="number"
                step="any"
                value={formData.latitude}
                onChange={(e) => setFormData({ ...formData, latitude: e.target.value })}
                required
                min="-90"
                max="90"
              />
            </div>
            <div className="form-group">
              <label htmlFor="longitude">Долгота *</label>
              <input
                id="longitude"
                type="number"
                step="any"
                value={formData.longitude}
                onChange={(e) => setFormData({ ...formData, longitude: e.target.value })}
                required
                min="-180"
                max="180"
              />
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="severity">Серьезность *</label>
            <select
              id="severity"
              value={formData.severity}
              onChange={(e) => setFormData({ ...formData, severity: parseInt(e.target.value) as Severity })}
              required
            >
              {SEVERITY_OPTIONS.map((s) => (
                <option key={s.value} value={s.value}>{s.label}</option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="description">Описание</label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              rows={3}
              placeholder="Опишите проблему..."
            />
          </div>

          <div className="form-group">
            <label htmlFor="photos">Фотографии</label>
            <input
              id="photos"
              type="file"
              accept="image/*"
              multiple
              onChange={(e) => setPhotos(Array.from(e.target.files || []))}
            />
            {photos.length > 0 && (
              <div className="photo-previews">
                {photos.map((file, index) => (
                  <div key={`${file.name}-${index}`} className="photo-preview-item">
                    <img src={URL.createObjectURL(file)} alt={file.name} className="photo-preview-img" />
                    <span className="photo-preview-name">{file.name}</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={isSubmitting}>
              Отмена
            </button>
            <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
              {isSubmitting ? 'Сохранение...' : 'Добавить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}