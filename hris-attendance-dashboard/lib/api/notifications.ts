import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export interface NotificationResponse {
  id: string;
  user_id: string;
  title: string;
  message: string;
  type: string;
  reference_id?: string;
  is_read: boolean;
  created_at: string;
  updated_at: string;
}

function buildUrl(path: string, params?: Record<string, string | undefined>) {
  const base =
    API_BASE.startsWith("http")
      ? API_BASE
      : `${typeof window !== "undefined" ? window.location.origin : "http://localhost:3000"}${API_BASE}`;
  const url = new URL(`${base}${path}`);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value) url.searchParams.set(key, value);
    }
  }
  return url.toString();
}

async function readJsonSafely(res: Response) {
  const raw = await res.text();
  try {
    return raw ? JSON.parse(raw) : {};
  } catch {
    throw new Error(`Invalid JSON response (status ${res.status})`);
  }
}

class NotificationsApi {
  async getNotifications(limit: number = 10): Promise<NotificationResponse[]> {
    const res = await fetch(buildUrl("/notifications", { limit: limit.toString() }), {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to fetch notifications");
    return data.data as NotificationResponse[];
  }

  async getUnreadCount(): Promise<number> {
    const res = await fetch(buildUrl("/notifications/unread-count"), {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to fetch unread count");
    return data.data?.unread_count || 0;
  }

  async markAsRead(id: string): Promise<void> {
    const res = await fetch(buildUrl(`/notifications/${id}/read`), {
      method: "PATCH",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to mark notification as read");
  }

  async markAllAsRead(): Promise<void> {
    const res = await fetch(buildUrl("/notifications/read-all"), {
      method: "POST",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to mark all as read");
  }
}

export const notificationsApi = new NotificationsApi();
