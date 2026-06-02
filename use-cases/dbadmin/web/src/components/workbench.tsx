import type { ReactNode } from 'react'
import { AlertIcon, XIcon } from './Icons'

export function PageShell({ children }: { children: ReactNode }) {
  return <div className="page-shell">{children}</div>
}

export function PageBody({ children, scroll = false }: { children: ReactNode; scroll?: boolean }) {
  return <div className={scroll ? 'page-scroll' : 'page-body'}>{children}</div>
}

export function PageToolbar({
  children,
  leading,
  trailing,
}: {
  children?: ReactNode
  leading?: ReactNode
  trailing?: ReactNode
}) {
  return (
    <div className="toolbar flex-wrap justify-between">
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        {leading}
        {children}
      </div>
      {trailing && <div className="flex shrink-0 flex-wrap items-center justify-end gap-2">{trailing}</div>}
    </div>
  )
}

export function PageStatusBar({
  left,
  center,
  right,
}: {
  left?: ReactNode
  center?: ReactNode
  right?: ReactNode
}) {
  return (
    <div className="page-statusbar">
      <div className="min-w-0 flex-1 truncate">{left}</div>
      {center && <div className="hidden min-w-0 flex-1 justify-center truncate md:flex">{center}</div>}
      <div className="min-w-0 flex-1 truncate text-right">{right}</div>
    </div>
  )
}

export function CodePanel({
  children,
  height,
}: {
  children: ReactNode
  height?: number | string
}) {
  return <div className="code-panel shrink-0" style={{ height }}>{children}</div>
}

export function StatusBanner({
  tone,
  title,
  children,
  action,
}: {
  tone: 'success' | 'warning' | 'danger'
  title?: ReactNode
  children: ReactNode
  action?: ReactNode
}) {
  return (
    <div className="status-banner" data-tone={tone}>
      <AlertIcon className="mt-0.5 h-4 w-4 shrink-0" />
      <div className="min-w-0 flex-1">
        {title && <div className="mb-0.5 font-medium">{title}</div>}
        <div style={{ color: 'var(--text-default)' }}>{children}</div>
      </div>
      {action && <div className="shrink-0">{action}</div>}
    </div>
  )
}

export function ModalShell({
  title,
  children,
  footer,
  onClose,
  widthClass = 'max-w-lg',
}: {
  title: ReactNode
  children: ReactNode
  footer?: ReactNode
  onClose: () => void
  widthClass?: string
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div className={`panel flex max-h-[90vh] w-full flex-col overflow-hidden ${widthClass}`} onClick={e => e.stopPropagation()}>
        <div className="flex shrink-0 items-center justify-between border-b px-5 py-4" style={{ borderColor: 'var(--border-subtle)' }}>
          <h2 className="text-sm font-semibold" style={{ color: 'var(--text-strong)' }}>{title}</h2>
          <button onClick={onClose} className="icon-btn" aria-label="Close">
            <XIcon className="h-4 w-4" />
          </button>
        </div>
        <div className="min-h-0 flex-1 overflow-auto p-5">{children}</div>
        {footer && (
          <div className="flex shrink-0 justify-end gap-2 border-t px-5 py-4" style={{ borderColor: 'var(--border-subtle)' }}>
            {footer}
          </div>
        )}
      </div>
    </div>
  )
}
