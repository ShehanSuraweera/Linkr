import { NextRequest, NextResponse } from "next/server"
import { login } from "@/lib/api"

const COOKIE_NAME = process.env.JWT_COOKIE_NAME!
const IS_PROD = process.env.NODE_ENV === "production"

export async function POST(req: NextRequest) {
  try {
    const { email, password } = await req.json()
    const data = await login(email, password)

    const res = NextResponse.json({ user_id: data.user_id })
    res.cookies.set(COOKIE_NAME, data.token, {
      httpOnly: true,
      secure: IS_PROD,
      sameSite: "lax",
      maxAge: 60 * 60 * 24, // 24 h — matches Go token TTL
      path: "/",
    })
    return res
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Login failed"
    const status = (err as { status?: number }).status ?? 500
    return NextResponse.json({ error: message }, { status })
  }
}
