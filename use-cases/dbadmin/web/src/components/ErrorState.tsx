interface ErrorStateProps {
  title?: string
  message: string
  detail?: string
  onRetry?: () => void
}

export default function ErrorState({ title = 'Error', message, detail, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      <div className="text-3xl mb-3 select-none">⚠</div>
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
          className="mt-4 text-xs px-3 py-1.5 rounded border"
          style={{
            borderColor: 'var(--border-strong)',
            color: 'var(--text-muted)',
          }}
        >
          Retry
        </button>
      )}
    </div>
  )
}
