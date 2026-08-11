import { ButtonHTMLAttributes, forwardRef } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'ghost'
  size?: 'sm' | 'md' | 'lg'
  isLoading?: boolean
  fullWidth?: boolean
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', isLoading, fullWidth, children, className = '', disabled, ...props }, ref) => {
    const baseClasses = 'btn'
    const variantClasses = `btn-${variant}`
    const sizeClasses = `btn-${size}`
    const widthClass = fullWidth ? 'btn-block' : ''
    const loadingClass = isLoading ? 'loading' : ''
    
    return (
      <button
        ref={ref}
        className={`${baseClasses} ${variantClasses} ${sizeClasses} ${widthClass} ${loadingClass} ${className}`}
        disabled={disabled || isLoading}
        {...props}
      >
        {isLoading && <span className="spinner-sm" />}
        {children}
      </button>
    )
  }
)

Button.displayName = 'Button'