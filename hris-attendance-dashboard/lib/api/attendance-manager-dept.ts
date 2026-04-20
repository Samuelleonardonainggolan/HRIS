import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export type AttendanceStatusUI = "HADIR" | "TELAT" | "IZIN" | "ALFA";

export interface ManagerDeptAttendanceItem {
  id: string;
  user_id: string;
  full_name: string;
  email: string;
  payroll_number: string;
  department_name: string;
  position_name: string;
  date: string;
  clock_in_time: string;
  clock_out_time: string;
  status: AttendanceStatusUI;
  location: string;
}

export interface ManagerDeptAttendanceSummary {
  total_records: number;
  tepat_waktu: number;
  terlambat: number;
  izin_sakit: number;
  alfa: number;
  total_kehadiran_pct: number;
}

export interface ManagerDeptAttendanceListResponse {
  items: ManagerDeptAttendanceItem[];
  page: number;
  page_size: number;
  total: number;
  summary: ManagerDeptAttendanceSummary;
}

function toQuery(params: Record<string, string | number | undefined>): string {
  const qs = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined) continue;
    const s = String(v);
    if (s.trim() === "") continue;
    qs.set(k, s);
  }
  const out = qs.toString();
  return out ? `?${out}` : "";
}

class AttendanceManagerDeptApi {
  async list(params: {
    from: string;
    to: string;
    q?: string;
    page?: number;
    page_size?: number;
  }): Promise<ManagerDeptAttendanceListResponse> {
    const res = await fetch(
      `${API_BASE}/attendance/records/my-department${toQuery(params)}`,
      {
        method: "GET",
        headers: authService.getAuthHeaders(),
      }
    );

    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.message || "Failed to fetch attendance records");
    return data.data as ManagerDeptAttendanceListResponse;
  }

  async exportCsv(params: { from: string; to: string; q?: string }): Promise<Blob> {
    const res = await fetch(
      `${API_BASE}/attendance/records/my-department/export${toQuery(params)}`,
      {
        method: "GET",
        headers: {
          ...authService.getAuthHeaders(),
          Accept: "text/csv",
        },
      }
    );

    if (!res.ok) {
      const raw = await res.text();
      try {
        const j = JSON.parse(raw);
        throw new Error(j.error || j.message || "Failed to export CSV");
      } catch {
        throw new Error(raw || "Failed to export CSV");
      }
    }
    return await res.blob();
  }
}

export const attendanceManagerDeptApi = new AttendanceManagerDeptApi();
