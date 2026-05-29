import React from 'react'
import { Badge } from '@/components/ui/badge'

const variantMap = {
  healthy: 'success',
  unhealthy: 'destructive',
  degraded: 'warning',
  unknown: 'secondary',
}

const labelMap = {
  healthy: 'Healthy',
  unhealthy: 'Unhealthy',
  degraded: 'Degraded',
  unknown: 'Unknown',
}

const dotClassMap = {
  healthy: 'bg-green-400',
  unhealthy: 'bg-red-400',
  degraded: 'bg-yellow-400',
  unknown: 'bg-gray-400',
}

export default function StatusBadge({ status }) {
  const variant = variantMap[status] || 'secondary'
  const label = labelMap[status] || 'Unknown'
  const dotClass = dotClassMap[status] || 'bg-gray-400'

  return (
    <Badge variant={variant} className="flex items-center gap-1">
      <span className={`w-2 h-2 rounded-full ${dotClass}`}></span>
      {label}
    </Badge>
  )
}
