// Server-side API client — attaches the JWT from the httpOnly cookie.
// Only used in Server Components and Route Handlers (never runs in the browser).
import { cache } from "react"
import { cookies } from "next/headers"
import type { Link, ListLinksResponse, LinkStats, OverviewStats, TokenResponse, User } from "./types"

const API_URL = process.env.API_URL!
const COOKIE_NAME = process.env.JWT_COOKIE_NAME!

async function getToken(): Promise<string | undefined> {
  const store = await cookies()
  return store.get(COOKIE_NAME)?.value
}

async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
  token?: string
): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }))
    throw new ApiError(body.error ?? "Unknown error", res.status)
  }

  return res.json() as Promise<T>
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number
  ) {
    super(message)
  }
}

// Auth
export async function login(email: string, password: string): Promise<TokenResponse> {
  return apiFetch<TokenResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  })
}

export async function register(email: string, password: string): Promise<TokenResponse> {
  return apiFetch<TokenResponse>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  })
}

// Links — require auth token from cookie
export async function getLinks(cursor?: string, limit?: number, q?: string): Promise<ListLinksResponse> {
  const token = await getToken()
  const params = new URLSearchParams()
  if (cursor) params.set("cursor", cursor)
  if (limit) params.set("limit", String(limit))
  if (q) params.set("q", q)
  const qs = params.size ? `?${params}` : ""
  return apiFetch<ListLinksResponse>(`/api/links${qs}`, {}, token)
}

export async function createLink(payload: {
  url: string
  alias?: string
  expires_at?: string
}): Promise<Link> {
  const token = await getToken()
  return apiFetch<Link>("/api/links", { method: "POST", body: JSON.stringify(payload) }, token)
}

export async function updateLink(
  code: string,
  payload: { is_active?: boolean; expires_at?: string | null }
): Promise<Link> {
  const token = await getToken()
  return apiFetch<Link>(`/api/links/${code}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  }, token)
}

export async function deleteLink(code: string): Promise<void> {
  const token = await getToken()
  const res = await fetch(`${API_URL}/api/links/${code}`, {
    method: "DELETE",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "Unknown error" }))
    throw new ApiError(body.error ?? "Unknown error", res.status)
  }
}

export async function getLinkStats(code: string): Promise<LinkStats> {
  const token = await getToken()
  return apiFetch<LinkStats>(`/api/links/${code}/stats`, {}, token)
}

export async function getOverview(): Promise<OverviewStats> {
  const token = await getToken()
  return apiFetch<OverviewStats>("/api/analytics/overview", {}, token)
}

export const getMe = cache(async (): Promise<User> => {
  const token = await getToken()
  return apiFetch<User>("/api/auth/me", {}, token)
})
