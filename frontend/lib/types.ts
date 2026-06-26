export interface Link {
  id: number
  short_code: string
  original_url: string
  created_at: string
  expires_at?: string
  is_active: boolean
  total_clicks: number
}

export interface ListLinksResponse {
  items: Link[]
  has_more: boolean
  next_cursor?: string
}

export interface DailyStat {
  day: string
  count: number
}

export interface DeviceStat {
  device: string
  count: number
}

export interface BrowserStat {
  browser: string
  count: number
}

export interface RefererStat {
  domain: string
  count: number
}

export interface LinkStats {
  total_clicks: number
  daily: DailyStat[]
  devices: DeviceStat[]
  browsers: BrowserStat[]
  referers: RefererStat[]
}

export interface LinkClickStat {
  short_code: string
  total_clicks: number
}

export interface OverviewStats {
  total_links: number
  active_links: number
  total_clicks: number
  daily: DailyStat[]
  devices: DeviceStat[]
  browsers: BrowserStat[]
  referers: RefererStat[]
  top_links: LinkClickStat[]
}

export interface TokenResponse {
  token: string
  user_id: number
}

export interface ApiError {
  error: string
}
