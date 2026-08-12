import { Link, useLocation } from 'react-router-dom'
import { useAuthStore } from '@features/auth/store'
import { useState } from 'react'

export function Header() {
  const { user, logout } = useAuthStore()
  const location = useLocation()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  const navItems = [
    { path: '/', label: 'Карта', icon: '🗺️' },
    { path: '/route', label: 'Маршрут', icon: '🧭' },
    { path: '/barrier/new', label: 'Добавить барьер', icon: '➕' },
    { path: '/profile', label: 'Профиль', icon: '👤' },
  ]

  const adminNavItems = [
    { path: '/moderation', label: 'Модерация', icon: '✅' },
  ]

  const isAdmin = user?.role === 'moderator' || user?.role === 'admin'

  return (
    <header className="header">
      <div className="header-container">
        <Link to="/" className="logo">
          <span className="logo-icon">♿</span>
          <span className="logo-text">Доступный путь</span>
        </Link>

        <nav className={`nav ${mobileMenuOpen ? 'open' : ''}`}>
          <ul className="nav-list">
            {navItems.map((item) => (
              <li key={item.path}>
                <Link
                  to={item.path}
                  className={`nav-link ${location.pathname === item.path ? 'active' : ''}`}
                  onClick={() => setMobileMenuOpen(false)}
                >
                  <span className="nav-icon">{item.icon}</span>
                  <span>{item.label}</span>
                </Link>
              </li>
            ))}
            {isAdmin && (
              <>
                {adminNavItems.map((item) => (
                  <li key={item.path}>
                    <Link
                      to={item.path}
                      className={`nav-link ${location.pathname === item.path ? 'active' : ''}`}
                      onClick={() => setMobileMenuOpen(false)}
                    >
                      <span className="nav-icon">{item.icon}</span>
                      <span>{item.label}</span>
                    </Link>
                  </li>
                ))}
              </>
            )}
            <li>
              <Link
                to="/notifications"
                className={`nav-link ${location.pathname === '/notifications' ? 'active' : ''}`}
                onClick={() => setMobileMenuOpen(false)}
              >
                <span className="nav-icon">🔔</span>
                <span>Уведомления</span>
              </Link>
            </li>
          </ul>
        </nav>

        <div className="header-actions">
          {user ? (
            <div className="user-menu">
              <span className="user-role">{user.role}</span>
              <span className="user-email">{user.nickname || user.email}</span>
              <button onClick={logout} className="btn btn-secondary btn-sm">
                Выйти
              </button>
            </div>
          ) : (
            <>
              <Link to="/login" className="btn btn-secondary btn-sm">Войти</Link>
              <Link to="/register" className="btn btn-primary btn-sm">Регистрация</Link>
            </>
          )}
          <button
            className="mobile-menu-btn"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            aria-label="Toggle menu"
          >
            <span className={`hamburger ${mobileMenuOpen ? 'open' : ''}`} />
          </button>
        </div>
      </div>
    </header>
  )
}