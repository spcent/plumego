import type { ReactNode } from 'react'
import { AlertIcon, EmptyIcon, XIcon } from './Icons'

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

export function LoadingState({
  title,
  detail,
  compact = false,
}: {
  title: ReactNode
  detail?: ReactNode
  compact?: boolean
}) {
  return (
    <div className={`flex flex-col items-center justify-center px-6 text-center ${compact ? 'py-6' : 'py-14'}`}>
      <div className="mb-3 grid h-10 w-10 place-items-center rounded-xl" style={{ background: 'var(--bg-muted)', color: 'var(--accent)' }}>
        <span className="h-4 w-4 animate-spin rounded-full border-2 border-current border-r-transparent" />
      </div>
      <div className="text-sm font-medium" style={{ color: 'var(--text-default)' }}>{title}</div>
      {detail && <div className="mt-1 max-w-md text-xs leading-relaxed" style={{ color: 'var(--text-muted)' }}>{detail}</div>}
    </div>
  )
}

export function EmptyStatePanel({
  title,
  detail,
  action,
  compact = false,
}: {
  title: ReactNode
  detail?: ReactNode
  action?: ReactNode
  compact?: boolean
}) {
  return (
    <div className={`flex flex-col items-center justify-center px-6 text-center ${compact ? 'py-6' : 'py-14'}`}>
      <div className="mb-3 grid h-10 w-10 place-items-center rounded-xl" style={{ background: 'var(--bg-muted)', color: 'var(--text-subtle)' }}>
        <EmptyIcon className="h-5 w-5" />
      </div>
      <div className="text-sm font-medium" style={{ color: 'var(--text-default)' }}>{title}</div>
      {detail && <div className="mt-1 max-w-md text-xs leading-relaxed" style={{ color: 'var(--text-muted)' }}>{detail}</div>}
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}

export function ErrorStatePanel({
  title,
  message,
  action,
  compact = false,
}: {
  title?: ReactNode
  message: ReactNode
  action?: ReactNode
  compact?: boolean
}) {
  return (
    <div className={`flex flex-col items-center justify-center px-6 text-center ${compact ? 'py-6' : 'py-14'}`}>
      <div className="mb-3 grid h-10 w-10 place-items-center rounded-xl" style={{ background: 'var(--danger-soft)', color: 'var(--danger)' }}>
        <AlertIcon className="h-5 w-5" />
      </div>
      <div className="text-sm font-semibold" style={{ color: 'var(--danger)' }}>{title}</div>
      <div className="mt-1 max-w-md text-xs leading-relaxed" style={{ color: 'var(--text-muted)' }}>{message}</div>
      {action && <div className="mt-4">{action}</div>}
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
