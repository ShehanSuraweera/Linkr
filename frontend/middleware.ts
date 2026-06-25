import { NextRequest, NextResponse } from "next/server"

const COOKIE_NAME = process.env.JWT_COOKIE_NAME ?? "linkr_token"
const PUBLIC_PATHS = ["/login", "/api/auth/login", "/api/auth/register"]

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl
  const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p))
  const hasToken = req.cookies.has(COOKIE_NAME)

  if (!isPublic && !hasToken) {
    return NextResponse.redirect(new URL("/login", req.url))
  }
  if (pathname === "/login" && hasToken) {
    return NextResponse.redirect(new URL("/dashboard", req.url))
  }
  return NextResponse.next()
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico|swagger).*)"],
}
