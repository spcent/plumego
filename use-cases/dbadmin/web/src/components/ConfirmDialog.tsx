import { useEffect, useRef } from 'react'
import { AlertIcon } from './Icons'

interface ConfirmDialogProps {
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  dangerous?: boolean
  onConfirm: () => void
  onCancel: () => void
}

export default function ConfirmDialog({
  open, title, message, confirmLabel = 'Confirm', dangerous = false, onConfirm, onCancel,
}: ConfirmDialogProps) {
  const cancelRef = useRef<HTMLButtonElement>(null)

  useEffect(() => {
    if (open) cancelRef.current?.focus()
  }, [open])

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel()
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
  }, [open, onCancel])

  if (!open) return null

  return (
    <div
      className="fixed inset-0 bg-slate-950/55 flex items-center justify-center z-50"
      onClick={onCancel}
    >
      <div
        className="panel p-5 w-full max-w-sm mx-4"
        onClick={e => e.stopPropagation()}
      >
        <div className="flex items-start gap-3 mb-4">
          <span
            className="grid h-9 w-9 place-items-center rounded-lg shrink-0"
            style={{
              background: dangerous ? 'color-mix(in srgb, var(--danger) 12%, transparent)' : 'var(--bg-muted)',
              color: dangerous ? 'var(--danger)' : 'var(--accent)',
            }}
          >
            <AlertIcon className="h-5 w-5" />
          </span>
          <div>
            <h2 className="text-base font-semibold mb-1" style={{ color: 'var(--text-strong)' }}>
              {title}
            </h2>
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>{message}</p>
          </div>
        </div>
        <div className="flex justify-end gap-2">
          <button
            ref={cancelRef}
            onClick={onCancel}
            className="btn"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className={dangerous ? 'btn btn-danger' : 'btn btn-primary'}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  )
}
