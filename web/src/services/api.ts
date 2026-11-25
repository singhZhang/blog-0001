export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export function apiBase() {
  return String(import.meta.env.VITE_API_BASE_URL || '/api').replace(/\/+$/, '')
}

export async function readError(response: Response) {
  let message = `请求失败（${response.status}）`
  try {
    const body = await response.json()
    if (body && typeof body === 'object' && 'error' in body) {
      const error = (body as { error?: { message?: unknown } }).error
      if (typeof error?.message === 'string' && error.message) message = error.message
    }
  } catch {
    // 非 JSON 响应保留状态码文案。
  }
  return new ApiError(message, response.status)
}
