import { createContext, useContext } from 'react'
import type { Connection } from '../api'

export const ConnectionsCtx = createContext<Connection[]>([])

export function useCurrentConn(connId: string | undefined): Connection | undefined {
  const conns = useContext(ConnectionsCtx)
  return conns.find(c => c.id === connId)
}
