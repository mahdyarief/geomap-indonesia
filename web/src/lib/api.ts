import { useAuthStore } from '@/store/auth'

export class ApiError extends Error {
  code: string
  status: number

  constructor(message: string, status: number, code: string) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

interface Envelope<T> {
  success: boolean
  data: T
  error?: { code: string; message: string }
}

export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = useAuthStore.getState().token
  const res = await fetch(path, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })

  const body: Envelope<T> | null = await res.json().catch(() => null)

  if (!res.ok) {
    const message = body?.error?.message || body?.error?.code || `HTTP ${res.status}`
    const code = body?.error?.code || 'UNKNOWN'
    throw new ApiError(message, res.status, code)
  }

  return body?.data as T
}