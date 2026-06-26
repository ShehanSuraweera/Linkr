import { NextRequest, NextResponse } from "next/server"

const COOKIE_NAME = process.env.JWT_COOKIE_NAME!

// POST — intentional logout (sidebar/header button)
export async function POST() {
  const res = NextResponse.json({ ok: true })
  res.cookies.set(COOKIE_NAME, "", { maxAge: 0, path: "/" })
  return res
}

// GET — session-expired auto-logout; clears cookie then redirects to /login
export async function GET(req: NextRequest) {
  const res = NextResponse.redirect(new URL("/login", req.url))
  res.cookies.set(COOKIE_NAME, "", { maxAge: 0, path: "/" })
  return res
}
