import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { BarrierFormModal } from '@features/barriers/BarrierFormModal'

export function BarrierFormPage() {
  const navigate = useNavigate()
  const { id } = useParams()
  const [saved, setSaved] = useState(false)

  const handleClose = () => navigate('/')

  return (
    <div className="barrier-form-page">
      <div className="page-header">
        <h1>{id ? 'Редактирование барьера' : 'Новый барьер'}</h1>
      </div>
      {!saved && (
        <BarrierFormModal
          barrier={null}
          onClose={handleClose}
          onSave={() => {
            setSaved(true)
            navigate('/')
          }}
        />
      )}
    </div>
  )
}