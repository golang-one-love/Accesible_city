import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { UserRole } from '@entities/user/types'
import { useAuthStore } from '@features/auth/store'
import { api } from '@shared/api/axios'

interface AdminUser {
  id: string
  email: string
  nickname: string
  role: UserRole
  is_active: boolean
  created_at: string
}

type Alert = { type: 'success' | 'error'; text: string } | null

const ROLE_OPTIONS: { value: UserRole; label: string }[] = [
  { value: 'user', label: 'Пользователь' },
  { value: 'volunteer', label: 'Волонтёр' },
  { value: 'business_owner', label: 'Владелец бизнеса' },
  { value: 'moderator', label: 'Модератор' },
]

function UserRow({ userEntry, canAssignAdmin }: { userEntry: AdminUser; canAssignAdmin: boolean }) {
  const queryClient = useQueryClient()
  const [role, setRole] = useState<UserRole>(userEntry.role)

  const roleMutation = useMutation({
    mutationFn: async () => {
      const response = await api.patch(`/auth/users/${userEntry.id}/role`, { role })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const options = canAssignAdmin
    ? [...ROLE_OPTIONS, { value: 'admin' as UserRole, label: 'Администратор' }]
    : ROLE_OPTIONS

  return (
    <div className="users-row">
      <div className="users-info">
        <span className="users-email">{userEntry.email}</span>
        <span className="users-nickname">{userEntry.nickname || '—'}</span>
      </div>
      <span className="role-badge">{userEntry.role}</span>
      <select
        className="select"
        value={role}
        onChange={(e) => setRole(e.target.value as UserRole)}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>{option.label}</option>
        ))}
      </select>
      <button
        className="btn btn-primary btn-sm"
        onClick={() => roleMutation.mutate()}
        disabled={roleMutation.isPending || role === userEntry.role}
      >
        {roleMutation.isPending ? 'Сохранение...' : 'Сохранить'}
      </button>
    </div>
  )
}

export function ProfilePage() {
  const user = useAuthStore((state) => state.user)
  const updateUser = useAuthStore((state) => state.updateUser)

  const [nickname, setNickname] = useState(user?.nickname ?? '')
  const [profileAlert, setProfileAlert] = useState<Alert>(null)

  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordAlert, setPasswordAlert] = useState<Alert>(null)

  const profileMutation = useMutation({
    mutationFn: async () => {
      const response = await api.patch('/auth/profile', { nickname })
      return response.data
    },
    onSuccess: (data) => {
      updateUser(data)
      setProfileAlert({ type: 'success', text: 'Профиль сохранён' })
    },
    onError: () => {
      setProfileAlert({ type: 'error', text: 'Не удалось сохранить профиль' })
    },
  })

  const passwordMutation = useMutation({
    mutationFn: async () => {
      await api.post('/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword,
      })
    },
    onSuccess: () => {
      setPasswordAlert({ type: 'success', text: 'Пароль успешно изменён' })
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    },
    onError: (err: any) => {
      setPasswordAlert({ type: 'error', text: err.response?.data?.error || 'Не удалось изменить пароль' })
    },
  })

  const usersQuery = useQuery({
    queryKey: ['users'],
    queryFn: async () => {
      const response = await api.get('/auth/users')
      return response.data.users as AdminUser[]
    },
    enabled: !!user && (user.role === 'moderator' || user.role === 'admin'),
  })

  if (!user) return null

  const canManageUsers = user.role === 'moderator' || user.role === 'admin'
  const canAssignAdmin = user.role === 'admin'

  const handleProfileSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setProfileAlert(null)
    profileMutation.mutate()
  }

  const handlePasswordSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setPasswordAlert(null)

    if (newPassword.length < 8) {
      setPasswordAlert({ type: 'error', text: 'Новый пароль должен содержать не менее 8 символов' })
      return
    }

    if (newPassword !== confirmPassword) {
      setPasswordAlert({ type: 'error', text: 'Пароли не совпадают' })
      return
    }

    passwordMutation.mutate()
  }

  return (
    <div className="profile-page">
      <div className="profile-card">
        <div className="profile-section">
          <div className="profile-header">
            <h2 className="profile-title">Профиль</h2>
            <span className="role-badge">{user.role}</span>
          </div>
          <form onSubmit={handleProfileSubmit}>
            <div className="form-group">
              <label className="form-label" htmlFor="profile-email">Email</label>
              <input
                id="profile-email"
                className="form-input"
                type="email"
                value={user.email}
                readOnly
              />
            </div>
            <div className="form-group">
              <label className="form-label" htmlFor="profile-nickname">Никнейм</label>
              <input
                id="profile-nickname"
                className="form-input"
                type="text"
                value={nickname}
                onChange={(e) => setNickname(e.target.value)}
              />
            </div>
            <button type="submit" className="btn btn-primary" disabled={profileMutation.isPending}>
              {profileMutation.isPending ? 'Сохранение...' : 'Сохранить'}
            </button>
          </form>
          {profileAlert && <div className={`alert-${profileAlert.type}`}>{profileAlert.text}</div>}
        </div>

        <div className="profile-section">
          <h2 className="profile-title">Смена пароля</h2>
          <form onSubmit={handlePasswordSubmit}>
            <div className="form-group">
              <label className="form-label" htmlFor="old-password">Текущий пароль</label>
              <input
                id="old-password"
                className="form-input"
                type="password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                required
                autoComplete="current-password"
              />
            </div>
            <div className="form-group">
              <label className="form-label" htmlFor="new-password">Новый пароль</label>
              <input
                id="new-password"
                className="form-input"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
                minLength={8}
                autoComplete="new-password"
              />
            </div>
            <div className="form-group">
              <label className="form-label" htmlFor="confirm-password">Подтвердите новый пароль</label>
              <input
                id="confirm-password"
                className="form-input"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
                autoComplete="new-password"
              />
            </div>
            <button type="submit" className="btn btn-primary" disabled={passwordMutation.isPending}>
              {passwordMutation.isPending ? 'Сохранение...' : 'Изменить пароль'}
            </button>
          </form>
          {passwordAlert && <div className={`alert-${passwordAlert.type}`}>{passwordAlert.text}</div>}
        </div>

        {canManageUsers && (
          <div className="profile-section">
            <h2 className="profile-title">Пользователи</h2>
            {usersQuery.isLoading ? (
              <div className="spinner" />
            ) : usersQuery.isError ? (
              <div className="alert-error">Ошибка загрузки пользователей</div>
            ) : usersQuery.data?.length === 0 ? (
              <div className="empty-state">
                <p>Пользователей нет</p>
              </div>
            ) : (
              <div className="users-table">
                {usersQuery.data?.map((userEntry) => (
                  <UserRow key={userEntry.id} userEntry={userEntry} canAssignAdmin={canAssignAdmin} />
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
