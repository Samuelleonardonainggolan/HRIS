import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

// ─── Interfaces ────────────────────────────────────────────────────────────────

export interface SalaryRange {
  min: number;
  max: number;
  currency: string;
}

export interface PositionResponse {
  id: string;
  code: string;
  name: string;
  department_id: string;
  level: number;       // 1=Staff, 2=Supervisor, 3=Manager, 4=Director, 5=C-Level
  level_name: string;  // "Staff", "Supervisor", dll
  description: string;
  requirements: string;
  salary_range: SalaryRange;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreatePositionRequest {
  code: string;
  name: string;
  department_id: string;
  level: number;
  description?: string;
  requirements?: string;
  salary_range?: SalaryRange;
}

export interface UpdatePositionRequest {
  code?: string;
  name?: string;
  department_id?: string;
  level?: number;
  description?: string;
  requirements?: string;
  salary_range?: SalaryRange;
  is_active?: boolean;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

async function safeJson(res: Response): Promise<Record<string, unknown>> {
  const raw = await res.text();
  try {
    return raw ? (JSON.parse(raw) as Record<string, unknown>) : {};
  } catch {
    return { message: raw };
  }
}

function extractError(data: Record<string, unknown>): string {
  if (typeof data.error === "string" && data.error) return data.error;
  if (typeof data.message === "string" && data.message) return data.message;
  return "Unknown error";
}

// ─── PositionApi class ────────────────────────────────────────────────────────

class PositionApi {
  /**
   * GET /api/v1/positions?department_id=xxx
   * Returns all positions, optionally filtered by department.
   */
  async getAll(departmentId?: string): Promise<PositionResponse[]> {
    const url = departmentId
      ? `${API_BASE}/positions?department_id=${encodeURIComponent(departmentId)}`
      : `${API_BASE}/positions`;

    const res = await fetch(url, {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });

    const data = await safeJson(res);

    if (!res.ok) {
      throw new Error(extractError(data) || "Gagal mengambil daftar jabatan");
    }

    const raw = data.data;

    if (!Array.isArray(raw)) return [];

    return raw as PositionResponse[];
  }

  /**
   * GET /api/v1/positions/:id
   * Returns a single position by ID.
   */
  async getById(id: string): Promise<PositionResponse> {
    const res = await fetch(`${API_BASE}/positions/${encodeURIComponent(id)}`, {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });

    const data = await safeJson(res);

    if (!res.ok) {
      throw new Error(extractError(data) || "Gagal mengambil data jabatan");
    }

    return data.data as PositionResponse;
  }

  /**
   * POST /api/v1/positions
   * Creates a new position.
   */
  async create(req: CreatePositionRequest): Promise<PositionResponse> {
    const res = await fetch(`${API_BASE}/positions`, {
      method: "POST",
      headers: authService.getAuthHeaders(),
      body: JSON.stringify(req),
    });

    const data = await safeJson(res);

    if (!res.ok) {
      throw new Error(extractError(data) || "Gagal membuat jabatan baru");
    }

    return data.data as PositionResponse;
  }

  /**
   * PUT /api/v1/positions/:id
   * Updates an existing position.
   */
  async update(id: string, req: UpdatePositionRequest): Promise<PositionResponse> {
    const res = await fetch(`${API_BASE}/positions/${encodeURIComponent(id)}`, {
      method: "PUT",
      headers: authService.getAuthHeaders(),
      body: JSON.stringify(req),
    });

    const data = await safeJson(res);

    if (!res.ok) {
      throw new Error(extractError(data) || "Gagal memperbarui jabatan");
    }

    return data.data as PositionResponse;
  }

  /**
   * DELETE /api/v1/positions/:id
   * Deletes a position by ID.
   */
  async delete(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/positions/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: authService.getAuthHeaders(),
    });

    if (!res.ok) {
      const data = await safeJson(res);
      throw new Error(extractError(data) || "Gagal menghapus jabatan");
    }
  }
}

// ─── Singleton export ─────────────────────────────────────────────────────────

export const positionApi = new PositionApi();
