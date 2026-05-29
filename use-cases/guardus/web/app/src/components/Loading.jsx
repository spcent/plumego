import React from 'react'
import logoSvg from '@/logo.svg'

const sizeClasses = {
  xs: 'w-4 h-4',
  sm: 'w-6 h-6',
  md: 'w-8 h-8',
  lg: 'w-12 h-12',
  xl: 'w-16 h-16',
}

export default function Loading({ size = 'md' }) {
  const sizeClass = sizeClasses[size] || sizeClasses.md

  return (
    <div className="flex justify-center items-center">
      <img
        className={`animate-spin rounded-full opacity-60 grayscale ${sizeClass}`}
        src={logoSvg}
        alt="Gatus logo"
      />
    </div>
  )
}
