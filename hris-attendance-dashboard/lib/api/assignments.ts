import { apiClient as api } from "./client";

export interface AssignmentEmployee {
  user_id: string;
  assigned_start_time?: string;
  assigned_end_time?: string;
}

export interface CreateAssignmentRequest {
  department_id: string;
  date: string; // YYYY-MM-DD
  reason: string;
  status: string;
  notes: string;
  start_time: string;
  end_time: string;
  employees: AssignmentEmployee[];
}

export interface UpdateAssignmentRequest {
  date?: string;
  reason?: string;
  status?: string;
  notes?: string;
  start_time?: string;
  end_time?: string;
  employees?: AssignmentEmployee[];
}

export const assignmentsApi = {
  list: async (departmentId?: string) => {
    const params = departmentId ? { department_id: departmentId } : {};
    const res = await api.get("/dept-assignments", { params });
    return res.data.data;
  },
  getById: async (id: string) => {
    const res = await api.get(`/dept-assignments/${id}`);
    return res.data.data;
  },
  create: async (req: CreateAssignmentRequest) => {
    const res = await api.post("/dept-assignments", req);
    return res.data.data;
  },
  update: async (id: string, req: UpdateAssignmentRequest) => {
    const res = await api.put(`/dept-assignments/${id}`, req);
    return res.data.data;
  },
  delete: async (id: string) => {
    const res = await api.delete(`/dept-assignments/${id}`);
    return res.data.data;
  },
  previewSchedule: async (userId: string, date: string) => {
    const res = await api.get("/dept-assignments/preview-schedule", {
      params: { user_id: userId, date },
    });
    return res.data.data;
  },
};
