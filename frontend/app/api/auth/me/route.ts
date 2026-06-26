import { NextResponse } from "next/server"
import { getMe } from "@/lib/api"

export async function GET() {
  try {
    const user = await getMe()
    return NextResponse.json(user)
  } catch {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }
}
