import { NextResponse } from "next/server";

const BACKEND_ORIGIN = (() => {
  try {
    return new URL(process.env.BACKEND_BASE_URL || "http://localhost:8080").origin;
  } catch {
    return "http://localhost:8080";
  }
})();

export async function POST(req: Request) {
  const body = await req.text();

  try {
    const upstream = await fetch(`${BACKEND_ORIGIN}/api/v1/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": req.headers.get("content-type") || "application/json",
      },
      body,
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

