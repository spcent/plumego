import { forwardRef } from 'react'
import type { ButtonHTMLAttributes, InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import { Icon, type IconName } from './icons'

export function cn(...inputs: Parameters<typeof clsx>) {
  return twMerge(clsx(inputs))
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md'
  icon?: IconName
}

export function Button({
  variant = 'secondary',
  size = 'md',
  icon,
  className,
  children,
  type = 'button',
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        'inline-flex shrink-0 items-center justify-center gap-2 rounded-md font-medium transition-[background-color,border-color,color,transform,opacity] duration-150 active:translate-y-px disabled:pointer-events-none disabled:opacity-45',
        size === 'sm' ? 'h-8 px-2.5 text-xs' : 'h-9 px-3 text-sm',
        variant === 'primary' && 'border border-primary bg-primary text-primary-foreground hover:bg-primary/90',
        variant === 'secondary' && 'border border-border bg-surface text-foreground hover:bg-accent',
        variant === 'ghost' && 'border border-transparent text-muted-foreground hover:bg-accent hover:text-foreground',
        variant === 'danger' && 'border border-destructive/25 bg-destructive/10 text-destructive hover:bg-destructive/15',
        className,
      )}
      {...props}
    >
      {icon && <Icon name={icon} className="h-4 w-4" />}
      {children}
    </button>
  )
}

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: IconName
  label: string
  active?: boolean
  tone?: 'default' | 'danger' | 'favorite'
}

export function IconButton({ icon, label, active, tone = 'default', className, type = 'button', ...props }: IconButtonProps) {
  return (
    <button
      type={type}
      aria-label={label}
      title={label}
      className={cn(
        'inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-transparent text-muted-foreground transition-[background-color,border-color,color,transform] duration-150 hover:bg-accent hover:text-foreground active:translate-y-px disabled:pointer-events-none disabled:opacity-40',
        active && tone === 'default' && 'border-primary/30 bg-primary/10 text-primary',
        active && tone === 'favorite' && 'border-amber-300/50 bg-amber-100/70 text-amber-700 dark:bg-amber-400/10 dark:text-amber-300',
        tone === 'danger' && 'hover:bg-destructive/10 hover:text-destructive',
        className,
      )}
      {...props}
    >
      <Icon name={icon} className="h-4 w-4" />
    </button>
  )
}

interface BadgeProps {
  children: ReactNode
  tone?: 'neutral' | 'accent' | 'success' | 'warning' | 'danger'
  className?: string
}

export function Badge({ children, tone = 'neutral', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] font-medium leading-5',
        tone === 'neutral' && 'border-border bg-muted text-muted-foreground',
        tone === 'accent' && 'border-primary/25 bg-primary/10 text-primary',
        tone === 'success' && 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
        tone === 'warning' && 'border-amber-500/25 bg-amber-500/10 text-amber-700 dark:text-amber-300',
        tone === 'danger' && 'border-destructive/25 bg-destructive/10 text-destructive',
        className,
      )}
    >
      {children}
    </span>
  )
}

export const TextInput = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(function TextInput({ className, ...props }, ref) {
  return (
    <input
      ref={ref}
      className={cn(
        'h-9 w-full rounded-md border border-input bg-surface px-3 text-sm text-foreground outline-none transition-[border-color,box-shadow] placeholder:text-muted-foreground focus:border-primary/60 focus:ring-2 focus:ring-primary/10',
        className,
      )}
      {...props}
    />
  )
})

export function SelectInput({ className, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className={cn(
        'h-9 rounded-md border border-input bg-surface px-3 text-sm text-foreground outline-none transition-[border-color,box-shadow] focus:border-primary/60 focus:ring-2 focus:ring-primary/10',
        className,
      )}
      {...props}
    />
  )
}

interface EmptyStateProps {
  icon?: IconName
  title: string
  description?: string
  action?: ReactNode
  compact?: boolean
}

export function EmptyState({ icon = 'file', title, description, action, compact }: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center text-center', compact ? 'px-4 py-8' : 'px-8 py-14')}>
      <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-lg border border-border bg-surface text-muted-foreground">
        <Icon name={icon} className="h-5 w-5" />
      </div>
      <div className="text-sm font-medium text-foreground">{title}</div>
      {description && <div className="mt-1 max-w-[28ch] text-xs leading-5 text-muted-foreground">{description}</div>}
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}

export function SkeletonRows({ count = 5 }: { count?: number }) {
  return (
    <div className="divide-y divide-border">
      {Array.from({ length: count }).map((_, idx) => (
        <div key={idx} className="space-y-2 px-3 py-3">
          <div className="h-3 w-4/5 animate-pulse rounded bg-muted" />
          <div className="h-2.5 w-3/5 animate-pulse rounded bg-muted" />
        </div>
      ))}
    </div>
  )
}

interface PanelHeaderProps {
  title: string
  description?: string
  action?: ReactNode
  className?: string
}

export function PanelHeader({ title, description, action, className }: PanelHeaderProps) {
  return (
    <div className={cn('flex min-h-12 items-center gap-3 border-b border-border px-4', className)}>
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-semibold text-foreground">{title}</div>
        {description && <div className="truncate text-xs text-muted-foreground">{description}</div>}
      </div>
      {action}
    </div>
  )
}
