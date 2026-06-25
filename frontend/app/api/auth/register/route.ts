import { NextRequest, NextResponse } from "next/server"
import { register } from "@/lib/api"

const COOKIE_NAME = process.env.JWT_COOKIE_NAME!
const IS_PROD = process.env.NODE_ENV === "production"

export async function POST(req: NextRequest) {
  try {
    const { email, password } = await req.json()
    const data = await register(email, password)

    const res = NextResponse.json({ user_id: data.user_id }, { status: 201 })
    res.cookies.set(COOKIE_NAME, data.token, {
      httpOnly: true,
      secure: IS_PROD,
      sameSite: "lax",
      maxAge: 60 * 60 * 24,
      path: "/",
    })
    return res
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Registration failed"
    const status = (err as { status?: number }).status ?? 500
    return NextResponse.json({ error: message }, { status })
  }
}
