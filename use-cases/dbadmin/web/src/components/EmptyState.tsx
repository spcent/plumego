import type { ReactNode } from 'react'

interface EmptyStateProps {
  message: string
  hint?: string
  action?: ReactNode
}

export default function EmptyState({ message, hint, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      <div className="text-4xl mb-3 select-none opacity-30">∅</div>
      <p className="text-sm font-medium" style={{ color: 'var(--text-muted)' }}>
        {message}
      </p>
      {hint && (
        <p className="text-xs mt-1" style={{ color: 'var(--text-subtle)' }}>
          {hint}
        </p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
