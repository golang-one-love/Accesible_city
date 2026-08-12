import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { BarrierFormModal } from '@features/barriers/BarrierFormModal'

export function BarrierFormPage() {
  const navigate = useNavigate()
  const [saved, setSaved] = useState(false)

  const handleClose = () => navigate('/')

  return (
    <div className="barrier-form-page">
      <div className="page-header">
        <h1>Новый барьер</h1>
      </div>
      {!saved && (
        <BarrierFormModal
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