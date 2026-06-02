import type { ReactNode } from 'react'
import { EmptyIcon } from './Icons'

interface EmptyStateProps {
  message: string
  hint?: string
  action?: ReactNode
}

export default function EmptyState({ message, hint, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      <div className="grid h-11 w-11 place-items-center rounded-xl mb-3" style={{ background: 'var(--bg-muted)', color: 'var(--text-subtle)' }}>
        <EmptyIcon className="h-5 w-5" />
      </div>
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
