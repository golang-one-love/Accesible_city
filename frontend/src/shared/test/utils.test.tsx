import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Button } from '@shared/ui/Button'
import { Input } from '@shared/ui/Input'
import { haversineDistance, formatDistance } from '@shared/lib/geo'

describe('Shared UI Components', () => {
  it('renders button with correct variant', () => {
    render(<Button variant="primary">Click me</Button>)
    const button = screen.getByRole('button', { name: /click me/i })
    expect(button).toHaveClass('btn-primary')
  })

  it('shows loading state', () => {
    render(<Button isLoading>Loading</Button>)
    const button = screen.getByRole('button', { name: /loading/i })
    expect(button).toHaveClass('loading')
    expect(button).toBeDisabled()
  })

  it('renders input with label', () => {
    render(<Input label="Email" placeholder="Enter email" />)
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/enter email/i)).toBeInTheDocument()
  })

  it('shows error message', () => {
    render(<Input label="Email" error="Invalid email" />)
    expect(screen.getByText(/invalid email/i)).toBeInTheDocument()
  })
})

describe('Geo utilities', () => {
  it('calculates haversine distance correctly', () => {
    // Distance between Moscow center and Red Square (~0.4km)
    const dist = haversineDistance(55.7558, 37.6173, 55.7522, 37.6156)
    expect(dist).toBeGreaterThan(300)
    expect(dist).toBeLessThan(600)
  })

  it('formats distance in meters', () => {
    expect(formatDistance(500)).toBe('500 м')
    expect(formatDistance(1500)).toBe('1.50 км')
  })
})