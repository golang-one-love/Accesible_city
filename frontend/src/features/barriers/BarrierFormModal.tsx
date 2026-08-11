import { useState } from 'react'
import { Barrier, BarrierType, Severity, CreateBarrierRequest } from '@entities/barrier/types'
import { api } from '@shared/api/axios'
import { useBarrierStore } from '@features/barriers/store'

interface BarrierFormModalProps {
  barrier: Barrier | null
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

export function BarrierFormModal({ barrier, onClose, onSave }: BarrierFormModalProps) {
  const { addBarrier, updateBarrier } = useBarrierStore()
  const [formData, setFormData] = useState<CreateBarrierRequest>({
    type: 'high_curb',
    coordinates: { latitude: 55.7558, longitude: 37.6173 },
    description: '',
    severity: 2,
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isEditing = barrier !== null

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsSubmitting(true)

    try {
      if (isEditing) {
        // For editing, we'd need an update endpoint
        // This is a simplified version
        await api.patch(`/barriers/${barrier.id}`, formData)
        updateBarrier(barrier.id, formData)
      } else {
        const response = await api.post('/barriers', formData)
        addBarrier(response.data)
      }
      onSave()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Ошибка при сохранении')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal barrier-form-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h3>{isEditing ? 'Редактировать барьер' : 'Добавить барьер'}</h3>
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
                step="0.000001"
                value={formData.coordinates.latitude}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  coordinates: { ...formData.coordinates, latitude: parseFloat(e.target.value) } 
                })}
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
                step="0.000001"
                value={formData.coordinates.longitude}
                onChange={(e) => setFormData({ 
                  ...formData, 
                  coordinates: { ...formData.coordinates, longitude: parseFloat(e.target.value) } 
                })}
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

          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={onClose} disabled={isSubmitting}>
              Отмена
            </button>
            <button type="submit" className="btn btn-primary" disabled={isSubmitting}>
              {isSubmitting ? 'Сохранение...' : (isEditing ? 'Сохранить' : 'Добавить')}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}