export interface ApiResponse<T> {
  success: boolean
  data?: T
  meta?: PaginationMeta
  error?: ApiError
}

export interface PaginationMeta {
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface ApiError {
  code: string
  message: string
  details?: ErrorDetail[]
}

export interface ErrorDetail {
  type?: string
  field?: string
  message: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface RefreshRequest {
  refresh_token: string
}
