/**
 * Minimal typed client for the Go backend. All endpoints share the same
 * response envelope:
 *
 *   success: { "success": true,  "data": {...} }
 *   failure: { "success": false, "error": { "code", "message" } }
 */

export interface ApiErrorBody {
  code: string
  message: string
}

interface Envelope<T> {
  success: boolean
  data?: T
  error?: ApiErrorBody
}

export class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(status: number, body?: ApiErrorBody) {
    super(body?.message ?? `Request failed with status ${status}`)
    this.name = 'ApiError'
    this.code = body?.code ?? 'unknown_error'
    this.status = status
  }
}

/** POST a JSON body to an API path and unwrap the response envelope. */
export async function apiPost<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  let envelope: Envelope<T> | undefined
  try {
    envelope = (await res.json()) as Envelope<T>
  } catch {
    throw new ApiError(res.status)
  }

  if (!res.ok || !envelope.success || envelope.data === undefined) {
    throw new ApiError(res.status, envelope.error)
  }
  return envelope.data
}
