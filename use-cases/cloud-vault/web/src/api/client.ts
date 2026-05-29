export interface APIError {
  code: string
  message: string
}

export class APIResponseError extends Error {
  code: string
  status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.code = code
    this.status = status
  }
}

// Prevent multiple simultaneous redirects
let isRedirecting = false

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })

  const body = await res.json().catch(() => null)

  if (!res.ok) {
    // Handle 401 Unauthorized - redirect to login (except for auth endpoints)
    if (
      res.status === 401 &&
      !path.includes('/api/v1/auth/login') &&
      !path.includes('/api/v1/auth/setup') &&
      !isRedirecting
    ) {
      isRedirecting = true
      window.location.href = '/login'
      // Delay error throw to allow redirect to complete
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    const err = body?.error as APIError | undefined
    throw new APIResponseError(
      err?.code ?? 'UNKNOWN_ERROR',
      err?.message ?? `HTTP ${res.status}`,
      res.status,
    )
  }

  return (body?.data ?? body) as T
}

export const client = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, data: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(data) }),
  put: <T>(path: string, data: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (path: string) => request<void>(path, { method: 'DELETE' }),
}
