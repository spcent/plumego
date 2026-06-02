import type { SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement>

function Icon({ children, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      {...props}
    >
      {children}
    </svg>
  )
}

export function DatabaseIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <ellipse cx="12" cy="5" rx="7" ry="3" />
      <path d="M5 5v7c0 1.7 3.1 3 7 3s7-1.3 7-3V5" />
      <path d="M5 12v7c0 1.7 3.1 3 7 3s7-1.3 7-3v-7" />
    </Icon>
  )
}

export function MySQLIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <ellipse cx="12" cy="5.5" rx="6.5" ry="2.8" />
      <path d="M5.5 5.5v5.2c0 1.5 2.9 2.8 6.5 2.8s6.5-1.3 6.5-2.8V5.5" />
      <path d="M7.5 16.4c1.2.8 2.7 1.1 4.5 1.1 3.6 0 6.5-1.3 6.5-2.8" />
      <path d="M15.5 13.6c1.6 1.2 3.1 2.7 4.2 4.5" />
      <path d="M17.2 17.5h3.4v-3.4" />
    </Icon>
  )
}

export function SQLiteIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M6 3.8h8.1L18 7.7v12.5H6z" />
      <path d="M14 3.8V8h4" />
      <path d="M8.7 15.8c2.8-1.8 5.4-4.6 7.4-8.2" />
      <path d="M9.4 11.7c.9.4 2.3.5 3.7.2" />
    </Icon>
  )
}

export function RedisIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="m12 4 7 3.2-7 3.2-7-3.2Z" />
      <path d="m5 11 7 3.2 7-3.2" />
      <path d="m5 15 7 3.2 7-3.2" />
    </Icon>
  )
}

export function MongoDBIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12 21s5-5.2 5-11.1C17 5.2 12 2 12 2s-5 3.2-5 7.9C7 15.8 12 21 12 21Z" />
      <path d="M12 8v13" />
    </Icon>
  )
}

export function ElasticsearchIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M8.5 4.5h6.9l4.1 7.5-4.1 7.5H8.5L4.5 12Z" />
      <path d="M8.5 4.5 12 12l-3.5 7.5" />
      <path d="M12 12h7.5" />
    </Icon>
  )
}

export function TableIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <rect x="4" y="5" width="16" height="14" rx="2" />
      <path d="M4 10h16M9 5v14" />
    </Icon>
  )
}

export function SettingsIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12 15.2A3.2 3.2 0 1 0 12 8.8a3.2 3.2 0 0 0 0 6.4Z" />
      <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.9l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.9-.3 1.7 1.7 0 0 0-1 1.6V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1-1.6 1.7 1.7 0 0 0-1.9.3l-.1.1A2 2 0 1 1 4.2 17l.1-.1a1.7 1.7 0 0 0 .3-1.9 1.7 1.7 0 0 0-1.6-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.6-1 1.7 1.7 0 0 0-.3-1.9l-.1-.1A2 2 0 1 1 7.1 4.2l.1.1a1.7 1.7 0 0 0 1.9.3h.1a1.7 1.7 0 0 0 .9-1.6V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.6 1.7 1.7 0 0 0 1.9-.3l.1-.1A2 2 0 1 1 19.8 7l-.1.1a1.7 1.7 0 0 0-.3 1.9v.1a1.7 1.7 0 0 0 1.6.9h.1a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.6 1Z" />
    </Icon>
  )
}

export function RefreshIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M20 6v5h-5" />
      <path d="M4 18v-5h5" />
      <path d="M6.1 9A7 7 0 0 1 18.5 6.5L20 11" />
      <path d="M17.9 15A7 7 0 0 1 5.5 17.5L4 13" />
    </Icon>
  )
}

export function MoonIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M20 14.5A8 8 0 0 1 9.5 4a7 7 0 1 0 10.5 10.5Z" />
    </Icon>
  )
}

export function SunIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <circle cx="12" cy="12" r="4" />
      <path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" />
    </Icon>
  )
}

export function LogoutIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M10 17l5-5-5-5" />
      <path d="M15 12H3" />
      <path d="M14 4h4a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-4" />
    </Icon>
  )
}

export function XIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M18 6 6 18M6 6l12 12" />
    </Icon>
  )
}

export function AlertIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M12 9v4" />
      <path d="M12 17h.01" />
      <path d="M10.3 3.9 2.5 18a2 2 0 0 0 1.8 3h15.4a2 2 0 0 0 1.8-3L13.7 3.9a2 2 0 0 0-3.4 0Z" />
    </Icon>
  )
}

export function EmptyIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="M4 7h16M7 7V5a2 2 0 0 1 2-2h6a2 2 0 0 1 2 2v2" />
      <path d="M6 7l1 13h10l1-13" />
      <path d="M10 11v5M14 11v5" />
    </Icon>
  )
}

export function ChevronRightIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="m9 18 6-6-6-6" />
    </Icon>
  )
}

export function ChevronDownIcon(props: IconProps) {
  return (
    <Icon {...props}>
      <path d="m6 9 6 6 6-6" />
    </Icon>
  )
}
