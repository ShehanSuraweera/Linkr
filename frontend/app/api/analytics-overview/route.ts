import { NextResponse } from "next/server"
import { getOverview, ApiError } from "@/lib/api"

export async function GET() {
  try {
    const data = await getOverview()
    return NextResponse.json(data)
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Failed to fetch overview"
    const status = err instanceof ApiError ? err.status : 500
    return NextResponse.json({ error: message }, { status })
  }
}
