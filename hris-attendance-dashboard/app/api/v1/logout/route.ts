import { NextResponse } from "next/server";

const BACKEND_BASE_URL = process.env.BACKEND_BASE_URL || "http://localhost:8080";

export async function POST(req: Request) {
  try {
    const upstream = await fetch(`${BACKEND_BASE_URL}/api/v1/logout`, {
      method: "POST",
      headers: {
        Authorization: req.headers.get("authorization") || "",
        "Content-Type": req.headers.get("content-type") || "application/json",
      },
    });

    const text = await upstream.text();

    return new NextResponse(text, {
      status: upstream.status,
      headers: {
        "Content-Type": upstream.headers.get("content-type") || "application/json",
      },
    });
  } catch (error: unknown) {
    const message =
      error instanceof Error ? error.message : "Failed to reach backend";
    return NextResponse.json(
      { error: message, message: "Backend unavailable" },
      { status: 502 }
    );
  }
}

export async function OPTIONS() {
  return new NextResponse(null, { status: 204 });
}
