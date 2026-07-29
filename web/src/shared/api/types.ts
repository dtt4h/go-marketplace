export type APIErrorResponse = {
  error: {
    code: string
    message: string
    details?: Record<string, unknown>
  }
}