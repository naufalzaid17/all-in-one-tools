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

export interface DownloadResult {
  blob: Blob
  filename: string
}

/**
 * POST files + fields as multipart/form-data and return the binary
 * result. The backend answers either with a file download
 * (Content-Disposition: attachment) or a JSON error envelope — this
 * helper distinguishes the two by content type.
 */
export async function apiUpload(
  path: string,
  files: File[],
  fields: Record<string, string> = {},
): Promise<DownloadResult> {
  const form = new FormData()
  for (const [key, value] of Object.entries(fields)) form.append(key, value)
  for (const file of files) form.append('files', file, file.name)

  const res = await fetch(path, { method: 'POST', body: form })

  const contentType = res.headers.get('Content-Type') ?? ''
  if (contentType.includes('application/json')) {
    let envelope: Envelope<never> | undefined
    try {
      envelope = (await res.json()) as Envelope<never>
    } catch {
      throw new ApiError(res.status)
    }
    throw new ApiError(res.status, envelope.error)
  }
  if (!res.ok) throw new ApiError(res.status)

  return {
    blob: await res.blob(),
    filename: filenameFromDisposition(res.headers.get('Content-Disposition')) ?? 'download',
  }
}

/** Extract the file name from a Content-Disposition header. */
function filenameFromDisposition(header: string | null): string | undefined {
  if (!header) return undefined
  // RFC 5987 form: filename*=UTF-8''na%C3%AFve.pdf
  const star = /filename\*=(?:UTF-8'')?([^;]+)/i.exec(header)
  if (star?.[1]) {
    try {
      return decodeURIComponent(star[1].trim().replace(/^"|"$/g, ''))
    } catch {
      /* fall through to the plain form */
    }
  }
  const plain = /filename="?([^";]+)"?/i.exec(header)
  return plain?.[1]?.trim()
}

/** Trigger a browser download for a fetched blob. */
export function saveBlob({ blob, filename }: DownloadResult) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  // Give the browser a beat to start the download before revoking.
  setTimeout(() => URL.revokeObjectURL(url), 10_000)
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
