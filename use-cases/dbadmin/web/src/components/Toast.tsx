import { useCallback, useRef, useState, type ReactNode } from 'react'
import { AlertIcon, XIcon } from './Icons'
import { ToastContext, type ToastType } from './toastContext'

interface ToastItem {
  id: number
  message: string
  type: ToastType
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([])
  const nextID = useRef(1)

  const showToast = useCallback((message: string, type: ToastType = 'error') => {
    const id = nextID.current++
    setToasts(p => [...p, { id, message, type }])
    setTimeout(() => setToasts(p => p.filter(t => t.id !== id)), 4500)
  }, [])

  const dismiss = (id: number) => setToasts(p => p.filter(t => t.id !== id))

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      <div className="fixed bottom-4 right-4 z-[200] flex flex-col gap-2 w-full max-w-sm pointer-events-none">
        {toasts.map(t => (
          <div
            key={t.id}
            role="status"
            className="pointer-events-auto panel flex items-start gap-3 px-3.5 py-3 text-sm"
            style={{
              borderColor: t.type === 'error'
                ? 'color-mix(in srgb, var(--danger) 45%, var(--border-subtle))'
                : t.type === 'success'
                ? 'color-mix(in srgb, var(--success) 45%, var(--border-subtle))'
                : 'var(--border-subtle)',
            }}
          >
            <span
              className="mt-0.5 grid h-5 w-5 place-items-center rounded-full shrink-0"
              style={{
                color: t.type === 'error' ? 'var(--danger)' : t.type === 'success' ? 'var(--success)' : 'var(--accent)',
              }}
            >
              <AlertIcon className="h-4 w-4" />
            </span>
            <span className="flex-1 leading-snug" style={{ color: 'var(--text-default)' }}>{t.message}</span>
            <button
              onClick={() => dismiss(t.id)}
              className="shrink-0 opacity-60 hover:opacity-100 leading-none"
              style={{ color: 'var(--text-muted)' }}
            >
              <XIcon className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}
