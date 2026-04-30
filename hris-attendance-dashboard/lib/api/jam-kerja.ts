import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export interface JamKerjaListRow {
  id: string;
  name: string;
  nik: string;
  department: string;
  position: string;
  day_of_week: string[];
  workDays: string;
  startTime: string;
  endTime: string;
  // API kadang mengembalikan snake_case
  start_time?: string;
  end_time?: string;
  avatar_url?: string;
  avatarUrl?: string;
  avatar?: string;
}

export interface JamKerjaDetail {
  user_id: string;
  name: string;
  nik: string;
  department: string;
  position: string;
  day_of_week: string[];
  start_time: string;
  end_time: string;
  is_active: boolean;
}

export interface UpdateJamKerjaRequest {
  day_of_week: string[];
  start_time: string;
  end_time: string;
  is_active?: boolean;
}

class JamKerjaApi {
  private async safeJson(res: Response): Promise<any> {
    const raw = await res.text();
    if (!raw) return {};
    try {
      return JSON.parse(raw);
    } catch {
      return { message: raw };
    }
  }

  async listMyDepartment(params?: { q?: string; position?: string }): Promise<JamKerjaListRow[]> {
    const qs = new URLSearchParams();
    if (params?.q) qs.set("q", params.q);
    if (params?.position) qs.set("position", params.position);
    const query = qs.toString() ? `?${qs.toString()}` : "";

    const res = await fetch(`${API_BASE}/jam-kerja/my-department${query}`, {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await this.safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal memuat jam kerja");
    return (data.data || []) as JamKerjaListRow[];
  }

  async getByUserId(userId: string): Promise<JamKerjaDetail> {
    const res = await fetch(`${API_BASE}/jam-kerja/user/${userId}`, {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await this.safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal memuat detail jam kerja");
    return data.data as JamKerjaDetail;
  }

  async updateByUserId(userId: string, req: UpdateJamKerjaRequest): Promise<JamKerjaDetail> {
    const res = await fetch(`${API_BASE}/jam-kerja/user/${userId}`, {
      method: "PUT",
      headers: authService.getAuthHeaders(),
      body: JSON.stringify(req),
    });
    const data = await this.safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal memperbarui jam kerja");
    return data.data as JamKerjaDetail;
  }
}

export const jamKerjaApi = new JamKerjaApi();
