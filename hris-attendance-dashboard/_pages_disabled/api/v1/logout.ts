import type { NextApiRequest, NextApiResponse } from "next";

const BACKEND_BASE_URL = process.env.BACKEND_BASE_URL || "http://localhost:8080";

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method === "OPTIONS") {
    res.status(204).end();
    return;
  }

  if (req.method !== "POST") {
    res.setHeader("Allow", "POST, OPTIONS");
    res.status(405).json({ error: "Method Not Allowed" });
    return;
  }

  try {
    const upstream = await fetch(`${BACKEND_BASE_URL}/api/v1/logout`, {
      method: "POST",
      headers: {
        Authorization: req.headers.authorization || "",
        "Content-Type": "application/json",
      },
    });

    const text = await upstream.text();
    res.status(upstream.status);
    res.setHeader("Content-Type", upstream.headers.get("content-type") || "application/json");
    res.send(text);
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : "Failed to reach backend";
    res.status(502).json({ error: message, message: "Backend unavailable" });
  }
}

