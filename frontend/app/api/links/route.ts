import { NextRequest, NextResponse } from "next/server"
import { getLinks, createLink } from "@/lib/api"

export async function GET(req: NextRequest) {
  try {
    const cursor = req.nextUrl.searchParams.get("cursor") ?? undefined
    const data = await getLinks(cursor)
    return NextResponse.json(data)
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Failed to fetch links"
    const status = (err as { status?: number }).status ?? 500
    return NextResponse.json({ error: message }, { status })
  }
}

export async function POST(req: NextRequest) {
  try {
    const body = await req.json()
    const link = await createLink(body)
    return NextResponse.json(link, { status: 201 })
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Failed to create link"
    const status = (err as { status?: number }).status ?? 500
    return NextResponse.json({ error: message }, { status })
  }
}
