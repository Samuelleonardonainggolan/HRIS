// lib/api/employee.ts
import { authService, User } from './auth';
import { Department, Position, CreateEmployeeRequest } from '@/types';

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith('/')
    ? process.env.NEXT_PUBLIC_API_URL
    : '/api/v1';

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}

export interface CreateEmployeeResponse {
  employee: User;
  temporary_password?: string;
}

export type Employee = {
  id: string;
  full_name: string;
  payroll_number?: string;
  nik?: string;
  department_name?: string;
  position_name?: string;
  is_active?: boolean;
};

export const employeeApi = {
  async getAll(): Promise<Employee[]> {
    const res = await fetch("/api/v1/employees", {
      headers: authService.getAuthHeaders(),
    });
    const json = await res.json().catch(() => null);

    if (!res.ok) {
      throw new Error(json?.error || json?.message || "Gagal memuat data karyawan");
    }

    return (json?.data || []) as Employee[];
  },
};

class EmployeeService {
  // ==================== EMPLOYEES ====================

  async getAllEmployees(): Promise<User[]> {
    try {
      const response = await fetch(`${API_BASE}/employees`, {
        method: 'GET',
        headers: authService.getAuthHeaders(),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to fetch employees');
      }

      return data.data;
    } catch (error) {
      console.error('Fetch employees error:', error);
      throw error;
    }
  }

  async getEmployeeByID(id: string): Promise<User> {
    try {
      const response = await fetch(`${API_BASE}/employees/${id}`, {
        method: 'GET',
        headers: authService.getAuthHeaders(),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to fetch employee');
      }

      return data.data;
    } catch (error) {
      console.error('Fetch employee error:', error);
      throw error;
    }
  }

  async createEmployee(employeeData: CreateEmployeeRequest): Promise<CreateEmployeeResponse> {
    try {
      const payload = {
        ...employeeData,
        payroll_number: employeeData.payroll_number || employeeData.nik || "",
        office_email: employeeData.office_email ?? employeeData.email,
        phone_number: employeeData.phone_number ?? employeeData.phone,
        role: employeeData.role || "staf",
      };

      const response = await fetch(`${API_BASE}/employees`, {
        method: 'POST',
        headers: authService.getAuthHeaders(),
        body: JSON.stringify(payload),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to create employee');
      }

      // If temporary password is returned (wrapped in data object)
      if (data.data && data.data.temporary_password) {
        return data.data;
      }

      // If just employee data
      return { employee: data.data };
    } catch (error) {
      console.error('Create employee error:', error);
      throw error;
    }
  }

  async updateEmployee(id: string, employeeData: Partial<User>): Promise<User> {
    try {
      const response = await fetch(`${API_BASE}/employees/${id}`, {
        method: 'PUT',
        headers: authService.getAuthHeaders(),
        body: JSON.stringify(employeeData),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to update employee');
      }

      return data.data;
    } catch (error) {
      console.error('Update employee error:', error);
      throw error;
    }
  }

  async deleteEmployee(id: string): Promise<void> {
    try {
      const response = await fetch(`${API_BASE}/employees/${id}`, {
        method: 'DELETE',
        headers: authService.getAuthHeaders(),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to delete employee');
      }
    } catch (error) {
      console.error('Delete employee error:', error);
      throw error;
    }
  }

  async importEmployees(file: File): Promise<{ created: number; failed: number; errors: string[] }> {
    try {
      const formData = new FormData();
      formData.append('file', file);

      // Note: Do not set Content-Type header manually for FormData, 
      // let the browser set it with boundary
      const headers = authService.getAuthHeaders();
      delete headers['Content-Type'];

      const response = await fetch(`${API_BASE}/employees/import`, {
        method: 'POST',
        headers: headers,
        body: formData,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to import employees');
      }

      return data.data;
    } catch (error) {
      console.error('Import employees error:', error);
      throw error;
    }
  }

  async downloadTemplate(): Promise<Blob> {
    try {
      const response = await fetch(`${API_BASE}/employees/template`, {
        method: 'GET',
        headers: authService.getAuthHeaders(),
      });

      if (!response.ok) {
        throw new Error('Failed to download template');
      }

      return await response.blob();
    } catch (error) {
      console.error('Download template error:', error);
      throw error;
    }
  }

  // ==================== DEPARTMENTS ====================

  async getAllDepartments(): Promise<Department[]> {
    try {
      const response = await fetch(`${API_BASE}/departments`, {
        method: 'GET',
        headers: authService.getAuthHeaders(),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to fetch departments');
      }

      return data.data;
    } catch (error) {
      console.error('Fetch departments error:', error);
      throw error;
    }
  }

  // ==================== POSITIONS ====================

  async getAllPositions(departmentId?: string): Promise<Position[]> {
    try {
      let url = `${API_BASE}/positions`;
      if (departmentId) {
        url += `?department_id=${departmentId}`;
      }

      const response = await fetch(url, {
        method: 'GET',
        headers: authService.getAuthHeaders(),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to fetch positions');
      }

      const rawList = (data.data || []) as Array<Record<string, unknown>>;
      const toPosition = (p: Record<string, unknown>): Position => {
        const id = String(p.id ?? '');
        const departmentId = String(p.departmentId ?? p.department_id ?? '');
        const code = String(p.code ?? '');
        const name = String(p.name ?? '');
        const level = typeof p.level === 'number' ? p.level : 0;
        const grade = String(p.grade ?? '');
        const description = typeof p.description === 'string' ? p.description : undefined;

        const responsibilitiesRaw = p.responsibilities;
        const responsibilities = Array.isArray(responsibilitiesRaw)
          ? responsibilitiesRaw.filter((x): x is string => typeof x === 'string')
          : [];

        const requirementsRaw = p.requirements;
        const requirements = Array.isArray(requirementsRaw)
          ? requirementsRaw.filter((x): x is string => typeof x === 'string')
          : typeof requirementsRaw === 'string' && requirementsRaw.trim() !== ''
            ? [requirementsRaw]
            : [];

        const salaryRangeCandidate = (p.salaryRange ?? p.salary_range) as unknown;
        const salaryRangeObj =
          typeof salaryRangeCandidate === 'object' && salaryRangeCandidate !== null
            ? (salaryRangeCandidate as Record<string, unknown>)
            : {};

        const salaryMin =
          typeof salaryRangeObj.min === 'number' ? salaryRangeObj.min : 0;
        const salaryMax =
          typeof salaryRangeObj.max === 'number' ? salaryRangeObj.max : 0;
        const salaryCurrency =
          typeof salaryRangeObj.currency === 'string' ? salaryRangeObj.currency : 'IDR';

        const isActiveRaw = (p.isActive ?? p.is_active) as unknown;
        const isActive = typeof isActiveRaw === 'boolean' ? isActiveRaw : false;

        const createdAt = String(p.createdAt ?? p.created_at ?? '');
        const updatedAt = String(p.updatedAt ?? p.updated_at ?? '');

        return {
          id,
          departmentId,
          code,
          name,
          level,
          grade,
          description,
          responsibilities,
          requirements,
          salaryRange: {
            min: salaryMin,
            max: salaryMax,
            currency: salaryCurrency,
          },
          isActive,
          createdAt,
          updatedAt,
        };
      };

      return rawList.map(toPosition);
    } catch (error) {
      console.error('Fetch positions error:', error);
      throw error;
    }
  }

  async updatePosition(
    id: string,
    positionData: Partial<Position> & { is_active?: boolean; isActive?: boolean }
  ): Promise<Position> {
    try {
      let response = await fetch(`${API_BASE}/positions/${id}`, {
        method: 'PUT',
        headers: authService.getAuthHeaders(),
        body: JSON.stringify(positionData),
      });

      if (!response.ok && (response.status === 404 || response.status === 405)) {
        response = await fetch(`${API_BASE}/positions/${id}`, {
          method: 'PATCH',
          headers: authService.getAuthHeaders(),
          body: JSON.stringify(positionData),
        });
      }

      const dataText = await response.text();
      let data: unknown = {};
      try {
        data = dataText ? JSON.parse(dataText) : {};
      } catch {
        throw new Error('Invalid response from server');
      }

      const dataObj = (val: unknown): val is { error?: string; message?: string; data?: unknown } =>
        typeof val === 'object' && val !== null;

      if (!response.ok) {
        const msg =
          dataObj(data) && (typeof data.error === 'string' || typeof data.message === 'string')
            ? (data.error as string) || (data.message as string)
            : 'Failed to update position';
        throw new Error(msg);
      }

      const p =
        dataObj(data) && typeof data.data === 'object' && data.data !== null
          ? (data.data as Record<string, unknown>)
          : {};
      const toPosition = (p2: Record<string, unknown>): Position => {
        const id2 = String(p2.id ?? '');
        const departmentId2 = String(p2.departmentId ?? p2.department_id ?? '');
        const code2 = String(p2.code ?? '');
        const name2 = String(p2.name ?? '');
        const level2 = typeof p2.level === 'number' ? p2.level : 0;
        const grade2 = String(p2.grade ?? '');
        const description2 = typeof p2.description === 'string' ? p2.description : undefined;

        const responsibilitiesRaw2 = p2.responsibilities;
        const responsibilities2 = Array.isArray(responsibilitiesRaw2)
          ? responsibilitiesRaw2.filter((x): x is string => typeof x === 'string')
          : [];

        const requirementsRaw2 = p2.requirements;
        const requirements2 = Array.isArray(requirementsRaw2)
          ? requirementsRaw2.filter((x): x is string => typeof x === 'string')
          : typeof requirementsRaw2 === 'string' && requirementsRaw2.trim() !== ''
            ? [requirementsRaw2]
            : [];

        const salaryRangeCandidate2 = (p2.salaryRange ?? p2.salary_range) as unknown;
        const salaryRangeObj2 =
          typeof salaryRangeCandidate2 === 'object' && salaryRangeCandidate2 !== null
            ? (salaryRangeCandidate2 as Record<string, unknown>)
            : {};

        const salaryMin2 =
          typeof salaryRangeObj2.min === 'number' ? salaryRangeObj2.min : 0;
        const salaryMax2 =
          typeof salaryRangeObj2.max === 'number' ? salaryRangeObj2.max : 0;
        const salaryCurrency2 =
          typeof salaryRangeObj2.currency === 'string' ? salaryRangeObj2.currency : 'IDR';

        const isActiveRaw2 = (p2.isActive ?? p2.is_active) as unknown;
        const isActive2 = typeof isActiveRaw2 === 'boolean' ? isActiveRaw2 : false;

        const createdAt2 = String(p2.createdAt ?? p2.created_at ?? '');
        const updatedAt2 = String(p2.updatedAt ?? p2.updated_at ?? '');

        return {
          id: id2,
          departmentId: departmentId2,
          code: code2,
          name: name2,
          level: level2,
          grade: grade2,
          description: description2,
          responsibilities: responsibilities2,
          requirements: requirements2,
          salaryRange: {
            min: salaryMin2,
            max: salaryMax2,
            currency: salaryCurrency2,
          },
          isActive: isActive2,
          createdAt: createdAt2,
          updatedAt: updatedAt2,
        };
      };

      return toPosition(p);
    } catch (error) {
      console.error('Update position error:', error);
      throw error;
    }
  }
}

export const employeeService = new EmployeeService();
