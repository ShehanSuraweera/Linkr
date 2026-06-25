import { NextRequest, NextResponse } from "next/server"
import { getLinkStats } from "@/lib/api"

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ code: string }> }
) {
  try {
    const { code } = await params
    const data = await getLinkStats(code)
    return NextResponse.json(data)
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Failed to fetch stats"
    const status = (err as { status?: number }).status ?? 500
    return NextResponse.json({ error: message }, { status })
  }
}
