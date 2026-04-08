import { authService } from "./auth";
import { Department } from "@/types";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export interface CreateDepartmentRequest {
  code?: string;
  name: string;
  description?: string;
  icon?: string;
  manager_id?: string;
}

export interface UpdateDepartmentRequest {
  code?: string;
  name?: string;
  description?: string;
  icon?: string;
  manager_id?: string;
  is_active?: boolean;
}

class DepartmentApi {
  async getAll(): Promise<Department[]> {
    const res = await fetch(`${API_BASE}/departments`, {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.message || "Failed to fetch departments");
    return (data.data as Array<{
      id: string;
      code: string;
      name: string;
      description?: string;
      icon?: string;
      manager_id?: string;
      manager_name?: string;
      total_staff?: number;
      is_active: boolean;
      created_at: string;
      updated_at: string;
    }>).map((d) => ({
      id: d.id,
      code: d.code,
      name: d.name,
      description: d.description,
      icon: d.icon,
      managerId: d.manager_id,
      managerName: d.manager_name,
      totalEmployees: d.total_staff ?? 0,
      isActive: d.is_active,
      createdAt: d.created_at,
      updatedAt: d.updated_at,
    }));
  }

  async getById(id: string): Promise<Department> {
    const res = await fetch(`${API_BASE}/departments/${id}`, {
      method: "GET",
      headers: authService.getAuthHeaders(),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.message || "Failed to fetch department");
    const d = data.data;
    return {
      id: d.id,
      code: d.code,
      name: d.name,
      description: d.description,
      icon: d.icon,
      managerId: d.manager_id,
      managerName: d.manager_name,
      totalEmployees: d.total_staff,
      isActive: d.is_active,
      createdAt: d.created_at,
      updatedAt: d.updated_at,
    };
  }

  async create(req: CreateDepartmentRequest): Promise<Department> {
    const res = await fetch(`${API_BASE}/departments`, {
      method: "POST",
      headers: authService.getAuthHeaders(),
      body: JSON.stringify(req),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.message || "Failed to create department");
    const d = data.data;
    return {
      id: d.id,
      code: d.code,
      name: d.name,
      description: d.description,
      icon: d.icon,
      managerId: d.manager_id,
      managerName: d.manager_name,
      totalEmployees: d.total_staff,
      isActive: d.is_active,
      createdAt: d.created_at,
      updatedAt: d.updated_at,
    };
  }

  async update(id: string, req: UpdateDepartmentRequest): Promise<Department> {
    const res = await fetch(`${API_BASE}/departments/${id}`, {
      method: "PUT",
      headers: authService.getAuthHeaders(),
      body: JSON.stringify(req),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.message || "Failed to update department");
    const d = data.data;
    return {
      id: d.id,
      code: d.code,
      name: d.name,
      description: d.description,
      icon: d.icon,
      managerId: d.manager_id,
      managerName: d.manager_name,
      totalEmployees: d.total_staff,
      isActive: d.is_active,
      createdAt: d.created_at,
      updatedAt: d.updated_at,
    };
  }

  async delete(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/departments/${id}`, {
      method: "DELETE",
      headers: authService.getAuthHeaders(),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || data.message || "Failed to delete department");
  }
}

export const departmentApi = new DepartmentApi();
