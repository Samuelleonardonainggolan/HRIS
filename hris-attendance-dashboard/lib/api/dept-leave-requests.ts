import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export type LeaveRequestStatus = "PENDING" | "APPROVED" | "REJECTED";

export interface DeptLeaveRequestApprovalResponse {
  pengajuan: {
    id: string;
    user_id: string;
    request_type_id: string;
    type_name: string;
    start_date: string;
    end_date: string;
    days_total: number;
    reason: string;
    document_url?: string;
    status_kepala_departemen: LeaveRequestStatus;
    kepala_departemen_id: string;
    status_manager_hr: LeaveRequestStatus;
    final_status: LeaveRequestStatus;
    rejection_reason_kepala_dept?: string;
    created_at: string;
    updated_at: string;
  };
  employee?: {
    id: string;
    payroll_number: string;
    full_name: string;
    department_name: string;
    position_name: string;
    avatar_url?: string;
    avatarUrl?: string;
    avatar?: string;
  };
}

class DeptLeaveRequestsApi {
  private async safeJson(res: Response): Promise<any> {
    const raw = await res.text();
    if (!raw) return {};
    try {
      return JSON.parse(raw);
    } catch {
      return { message: raw };
    }
  }

  async list(params?: { status?: LeaveRequestStatus | "ALL"; search?: string }): Promise<DeptLeaveRequestApprovalResponse[]> {
    const qs = new URLSearchParams();
    if (params?.status) qs.set("status", params.status);
    if (params?.search) qs.set("search", params.search);
    const query = qs.toString() ? `?${qs.toString()}` : "";

    const res = await fetch(`${API_BASE}/dept-leave-requests${query}`, {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await this.safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal memuat pengajuan");
    return (data.data || []) as DeptLeaveRequestApprovalResponse[];
  }

  async approve(id: string): Promise<DeptLeaveRequestApprovalResponse> {
    const res = await fetch(`${API_BASE}/dept-leave-requests/${id}/approve`, {
      method: "POST",
      headers: authService.getAuthHeaders(),
    });
    const data = await this.safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal menyetujui pengajuan");
    return data.data as DeptLeaveRequestApprovalResponse;
  }

  async reject(id: string, rejectionReason: string): Promise<DeptLeaveRequestApprovalResponse> {
    const res = await fetch(`${API_BASE}/dept-leave-requests/${id}/reject`, {
      method: "POST",
      headers: {
        ...authService.getAuthHeaders(),
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ rejection_reason: rejectionReason }),
    });
    const data = await this.safeJson(res);
    if (!res.ok) throw new Error(data.error || data.message || "Gagal menolak pengajuan");
    return data.data as DeptLeaveRequestApprovalResponse;
  }
}

export const deptLeaveRequestsApi = new DeptLeaveRequestsApi();
