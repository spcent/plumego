import { AlertIcon } from './Icons'

interface ErrorStateProps {
  title?: string
  message: string
  detail?: string
  onRetry?: () => void
}

export default function ErrorState({ title = 'Error', message, detail, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      <div className="grid h-11 w-11 place-items-center rounded-xl mb-3" style={{ background: 'color-mix(in srgb, var(--danger) 12%, transparent)', color: 'var(--danger)' }}>
        <AlertIcon className="h-5 w-5" />
      </div>
      <p className="text-sm font-semibold mb-1" style={{ color: 'var(--danger)' }}>
        {title}
      </p>
      <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
        {message}
      </p>
      {detail && (
        <pre className="mt-2 text-xs font-mono px-3 py-2 rounded max-w-lg text-left overflow-auto"
          style={{
            background: 'var(--bg-muted)',
            color: 'var(--text-subtle)',
            border: '1px solid var(--border-subtle)',
          }}>
          {detail}
        </pre>
      )}
      {onRetry && (
        <button
          onClick={onRetry}
          className="btn mt-4"
        >
          Retry
        </button>
      )}
    </div>
  )
}
