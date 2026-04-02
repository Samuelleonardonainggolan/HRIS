import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export type LeaveRequestStatus = "PENDING" | "APPROVED" | "REJECTED";

export interface LeaveRequestApprovalResponse {
  pengajuan: {
    id: string;
    user_id: string;
    tipe_pengajuan_id: string;
    nama_tipe: string;
    tanggal_mulai: string;
    tanggal_selesai: string;
    total_hari: number;
    alasan: string;
    dokumen_url?: string;
    kuota_cuti_id?: string;
    status_kepala_departemen: LeaveRequestStatus;
    kepala_departemen_id: string;
    manager_hr_id: string;
    status_manager_hr: LeaveRequestStatus;
    status_final: LeaveRequestStatus;
    created_at: string;
    updated_at: string;
  };
  employee?: {
    id: string;
    payroll_number: string;
    full_name: string;
    department_name: string;
    position_name: string;
  };
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
    return raw ? (JSON.parse(raw) as any) : {};
  } catch {
    throw new Error(`Invalid JSON response (status ${res.status})`);
  }
}

class LeaveRequestsApi {
  async listForManagerHR(params?: { status?: LeaveRequestStatus | "ALL"; search?: string }): Promise<LeaveRequestApprovalResponse[]> {
    const res = await fetch(buildUrl("/leave-requests", { status: params?.status, search: params?.search }), {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to fetch leave requests");
    return data.data as LeaveRequestApprovalResponse[];
  }

  async getForManagerHR(id: string): Promise<LeaveRequestApprovalResponse> {
    const res = await fetch(buildUrl(`/leave-requests/${id}`), {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to fetch leave request");
    return data.data as LeaveRequestApprovalResponse;
  }

  async approve(id: string): Promise<LeaveRequestApprovalResponse> {
    const res = await fetch(buildUrl(`/leave-requests/${id}/approve`), {
      method: "POST",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to approve leave request");
    return data.data as LeaveRequestApprovalResponse;
  }

  async reject(id: string): Promise<LeaveRequestApprovalResponse> {
    const res = await fetch(buildUrl(`/leave-requests/${id}/reject`), {
      method: "POST",
      headers: authService.getAuthHeaders(),
    });
    const data = await readJsonSafely(res);
    if (!res.ok) throw new Error(data.error || data.message || "Failed to reject leave request");
    return data.data as LeaveRequestApprovalResponse;
  }
}

export const leaveRequestsApi = new LeaveRequestsApi();
