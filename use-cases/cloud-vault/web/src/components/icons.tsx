import type { ReactNode, SVGProps } from 'react'

export type IconName =
  | 'archive'
  | 'bolt'
  | 'book'
  | 'box'
  | 'check'
  | 'chevronDown'
  | 'chevronLeft'
  | 'chevronRight'
  | 'circle'
  | 'close'
  | 'collapse'
  | 'copy'
  | 'database'
  | 'file'
  | 'filter'
  | 'folder'
  | 'grid'
  | 'import'
  | 'layout'
  | 'lock'
  | 'logout'
  | 'more'
  | 'plus'
  | 'refresh'
  | 'save'
  | 'search'
  | 'settings'
  | 'spark'
  | 'star'
  | 'tag'
  | 'trash'
  | 'user'

interface IconProps extends SVGProps<SVGSVGElement> {
  name: IconName
}

export function Icon({ name, className = 'h-4 w-4', ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      aria-hidden="true"
      {...props}
    >
      {paths[name]}
    </svg>
  )
}

const paths: Record<IconName, ReactNode> = {
  archive: <><path d="M4 7.5h16" /><path d="M6 7.5v10A2.5 2.5 0 0 0 8.5 20h7a2.5 2.5 0 0 0 2.5-2.5v-10" /><path d="M8 4h8l2 3.5H6L8 4Z" /><path d="M10 12h4" /></>,
  bolt: <path d="m13 2-8 12h6l-1 8 8-12h-6l1-8Z" />,
  book: <><path d="M5 5.5A2.5 2.5 0 0 1 7.5 3H19v16H7.5A2.5 2.5 0 0 1 5 16.5v-11Z" /><path d="M5 16.5A2.5 2.5 0 0 1 7.5 14H19" /></>,
  box: <><path d="m4 8 8-4 8 4-8 4-8-4Z" /><path d="M4 8v8l8 4 8-4V8" /><path d="M12 12v8" /></>,
  check: <path d="m5 12.5 4.2 4.2L19 6.8" />,
  chevronDown: <path d="m7 9.5 5 5 5-5" />,
  chevronLeft: <path d="m15 18-6-6 6-6" />,
  chevronRight: <path d="m9 18 6-6-6-6" />,
  circle: <circle cx="12" cy="12" r="7.5" />,
  close: <><path d="m7 7 10 10" /><path d="M17 7 7 17" /></>,
  collapse: <><path d="M4 5h16" /><path d="M4 19h16" /><path d="M8 9l-3 3 3 3" /><path d="M20 12H5" /></>,
  copy: <><path d="M8 8h9.5A1.5 1.5 0 0 1 19 9.5v9A1.5 1.5 0 0 1 17.5 20h-9A1.5 1.5 0 0 1 7 18.5V9.5A1.5 1.5 0 0 1 8.5 8Z" /><path d="M5 16H4.5A1.5 1.5 0 0 1 3 14.5v-9A1.5 1.5 0 0 1 4.5 4h9A1.5 1.5 0 0 1 15 5.5V6" /></>,
  database: <><ellipse cx="12" cy="5.5" rx="7" ry="3" /><path d="M5 5.5v6c0 1.7 3.1 3 7 3s7-1.3 7-3v-6" /><path d="M5 11.5v6c0 1.7 3.1 3 7 3s7-1.3 7-3v-6" /></>,
  file: <><path d="M7 3h6l4 4v14H7a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2Z" /><path d="M13 3v5h5" /><path d="M8.5 13h7" /><path d="M8.5 17h5" /></>,
  filter: <><path d="M4 6h16" /><path d="M7 12h10" /><path d="M10 18h4" /></>,
  folder: <><path d="M3.5 7.5A2.5 2.5 0 0 1 6 5h4l2 2h6a2.5 2.5 0 0 1 2.5 2.5v7A2.5 2.5 0 0 1 18 19H6a2.5 2.5 0 0 1-2.5-2.5v-9Z" /></>,
  grid: <><rect x="4" y="4" width="6" height="6" rx="1.5" /><rect x="14" y="4" width="6" height="6" rx="1.5" /><rect x="4" y="14" width="6" height="6" rx="1.5" /><rect x="14" y="14" width="6" height="6" rx="1.5" /></>,
  import: <><path d="M12 3v11" /><path d="m8 10 4 4 4-4" /><path d="M4 17v1.5A2.5 2.5 0 0 0 6.5 21h11a2.5 2.5 0 0 0 2.5-2.5V17" /></>,
  layout: <><rect x="4" y="4" width="16" height="16" rx="2" /><path d="M9 4v16" /><path d="M15 4v16" /></>,
  lock: <><rect x="5" y="10" width="14" height="10" rx="2" /><path d="M8 10V7a4 4 0 0 1 8 0v3" /></>,
  logout: <><path d="M10 5H6.5A1.5 1.5 0 0 0 5 6.5v11A1.5 1.5 0 0 0 6.5 19H10" /><path d="M14 8l4 4-4 4" /><path d="M18 12H9" /></>,
  more: <><circle cx="6" cy="12" r="1" /><circle cx="12" cy="12" r="1" /><circle cx="18" cy="12" r="1" /></>,
  plus: <><path d="M12 5v14" /><path d="M5 12h14" /></>,
  refresh: <><path d="M20 12a8 8 0 0 1-13.7 5.6" /><path d="M4 12A8 8 0 0 1 17.7 6.4" /><path d="M17.5 3.5v3h-3" /><path d="M6.5 20.5v-3h3" /></>,
  save: <><path d="M5 4h12l2 2v14H5V4Z" /><path d="M8 4v6h8V4" /><path d="M8 20v-6h8v6" /></>,
  search: <><circle cx="10.5" cy="10.5" r="6.5" /><path d="m16 16 4 4" /></>,
  settings: <><circle cx="12" cy="12" r="3" /><path d="M19 12a6.8 6.8 0 0 0-.1-1l2-1.5-2-3.4-2.4 1a7.8 7.8 0 0 0-1.7-1L14.5 3h-5l-.3 3.1a7.8 7.8 0 0 0-1.7 1l-2.4-1-2 3.4 2 1.5a6.8 6.8 0 0 0 0 2l-2 1.5 2 3.4 2.4-1a7.8 7.8 0 0 0 1.7 1l.3 3.1h5l.3-3.1a7.8 7.8 0 0 0 1.7-1l2.4 1 2-3.4-2-1.5c.1-.3.1-.7.1-1Z" /></>,
  spark: <><path d="m12 3 1.6 5.4L19 10l-5.4 1.6L12 17l-1.6-5.4L5 10l5.4-1.6L12 3Z" /><path d="m18 15 .8 2.7 2.7.8-2.7.8L18 21l-.8-2.7-2.7-.8 2.7-.8L18 15Z" /></>,
  star: <path d="m12 4 2.3 4.8 5.2.7-3.8 3.7.9 5.2L12 16l-4.6 2.4.9-5.2-3.8-3.7 5.2-.7L12 4Z" />,
  tag: <><path d="M4 11V5h6l9.2 9.2a2.1 2.1 0 0 1 0 3L17.2 19a2.1 2.1 0 0 1-3 0L4 11Z" /><circle cx="8" cy="8" r="1" /></>,
  trash: <><path d="M5 7h14" /><path d="M9 7V5h6v2" /><path d="M7 7l1 13h8l1-13" /><path d="M10 11v5" /><path d="M14 11v5" /></>,
  user: <><circle cx="12" cy="8" r="4" /><path d="M4.5 20a7.5 7.5 0 0 1 15 0" /></>,
}
