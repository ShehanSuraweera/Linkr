export interface Link {
  id: number
  short_code: string
  original_url: string
  created_at: string
  expires_at?: string
  is_active: boolean
}

export interface ListLinksResponse {
  items: Link[]
  has_more: boolean
  next_cursor?: string
}

export interface DailyStat {
  Day: string
  Count: number
}

export interface LinkStats {
  total_clicks: number
  daily: DailyStat[]
}

export interface TokenResponse {
  token: string
  user_id: number
}

export interface ApiError {
  error: string
}
