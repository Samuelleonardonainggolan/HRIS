import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export type OvertimeRequestStatus = "PENDING" | "APPROVED" | "REJECTED";

export interface OvertimeApprovalResponse {
  overtime: {
    id: string;
    user_id: string;
    date: string;
    start_time: string;
    end_time: string;
    reason: string;
    total: string;
    status_kepala_departemen: OvertimeRequestStatus;
    kepala_departemen_id?: string;
    rejection_reason_kepala_dept?: string;
    status_manager_hr: OvertimeRequestStatus;
    manager_hr_id?: string;
    rejection_reason_manager_hr?: string;
    final_status: OvertimeRequestStatus;
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
  const base = API_BASE.startsWith("http")
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

async function safeJson(res: Response): Promise<any> {
  const raw = await res.text();
  if (!raw) return {};
  try {
    return JSON.parse(raw);
  } catch {
    return { message: raw };
  }
}

// ─── Manager HR API ──────────────────────────────────────────────────────────

class OvertimeRequestsApi {
  async listForManagerHR(params?: {
    status?: OvertimeRequestStatus | "ALL";
    search?: string;
  }): Promise<OvertimeApprovalResponse[]> {
    const res = await fetch(
      buildUrl("/overtime-requests", {
        status: params?.status,
        search: params?.search,
      }),
      {
        method: "GET",
        headers: authService.getAuthHeaders(),
      }
    );
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal mengambil data pengajuan lembur");
    return (data.data ?? []) as OvertimeApprovalResponse[];
  }

  async getForManagerHR(id: string): Promise<OvertimeApprovalResponse> {
    const res = await fetch(buildUrl(`/overtime-requests/${id}`), {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal mengambil pengajuan lembur");
    return data.data as OvertimeApprovalResponse;
  }

  async approve(id: string): Promise<OvertimeApprovalResponse> {
    const res = await fetch(buildUrl(`/overtime-requests/${id}/approve`), {
      method: "POST",
      headers: authService.getAuthHeaders(),
    });
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal menyetujui pengajuan lembur");
    return data.data as OvertimeApprovalResponse;
  }

  async reject(id: string, rejectionReason: string): Promise<OvertimeApprovalResponse> {
    const res = await fetch(buildUrl(`/overtime-requests/${id}/reject`), {
      method: "POST",
      headers: {
        ...authService.getAuthHeaders(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ rejection_reason: rejectionReason }),
    });
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal menolak pengajuan lembur");
    return data.data as OvertimeApprovalResponse;
  }
}

// ─── Kepala Departemen API ────────────────────────────────────────────────────

class DeptOvertimeRequestsApi {
  async list(params?: {
    status?: OvertimeRequestStatus | "ALL";
    search?: string;
  }): Promise<OvertimeApprovalResponse[]> {
    const res = await fetch(
      buildUrl("/dept-overtime-requests", {
        status: params?.status,
        search: params?.search,
      }),
      {
        method: "GET",
        headers: authService.getAuthHeaders(),
      }
    );
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal mengambil data pengajuan lembur");
    return (data.data ?? []) as OvertimeApprovalResponse[];
  }

  async get(id: string): Promise<OvertimeApprovalResponse> {
    const res = await fetch(buildUrl(`/dept-overtime-requests/${id}`), {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal mengambil pengajuan lembur");
    return data.data as OvertimeApprovalResponse;
  }

  async approve(id: string): Promise<OvertimeApprovalResponse> {
    const res = await fetch(buildUrl(`/dept-overtime-requests/${id}/approve`), {
      method: "POST",
      headers: authService.getAuthHeaders(),
    });
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal menyetujui pengajuan lembur");
    return data.data as OvertimeApprovalResponse;
  }

  async reject(id: string, rejectionReason: string): Promise<OvertimeApprovalResponse> {
    const res = await fetch(buildUrl(`/dept-overtime-requests/${id}/reject`), {
      method: "POST",
      headers: {
        ...authService.getAuthHeaders(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ rejection_reason: rejectionReason }),
    });
    const data = await safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal menolak pengajuan lembur");
    return data.data as OvertimeApprovalResponse;
  }
}

export const overtimeRequestsApi = new OvertimeRequestsApi();
export const deptOvertimeRequestsApi = new DeptOvertimeRequestsApi();
