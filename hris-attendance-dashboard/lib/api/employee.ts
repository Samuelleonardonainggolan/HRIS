// lib/api/employee.ts
import { authService, User } from './auth';
import { Department, Position, CreateEmployeeRequest } from '@/types';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}

export interface CreateEmployeeResponse {
  employee: User;
  temporary_password?: string;
}

class EmployeeService {
  // ==================== EMPLOYEES ====================

  async getAllEmployees(): Promise<User[]> {
    try {
      const response = await fetch(`${API_URL}/employees`, {
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
      const response = await fetch(`${API_URL}/employees/${id}`, {
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
        payroll_number: employeeData.payrollNumber ?? employeeData.nik,
        nik: employeeData.nik,
        email: employeeData.email ?? employeeData.officeEmail,
        office_email: employeeData.officeEmail,
        full_name: employeeData.fullName,
        birth_date: employeeData.birthDate,
        religion: employeeData.religion,
        last_education: employeeData.lastEducation,
        year_enrolled: employeeData.yearEnrolled,
        employment_status: employeeData.employmentStatus,
        department_id: employeeData.departmentID ?? employeeData.department,
        position_id: employeeData.positionID ?? employeeData.position,
        phone: employeeData.phone ?? employeeData.phoneNumber,
        phone_number: employeeData.phoneNumber,
        address: employeeData.address,
        role: employeeData.role ?? 'staf',
      };

      const response = await fetch(`${API_URL}/employees`, {
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
      const response = await fetch(`${API_URL}/employees/${id}`, {
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
      const response = await fetch(`${API_URL}/employees/${id}`, {
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

      const response = await fetch(`${API_URL}/employees/import`, {
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
      const response = await fetch(`${API_URL}/employees/template`, {
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
      const response = await fetch(`${API_URL}/departments`, {
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
      let url = `${API_URL}/positions`;
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

      return data.data;
    } catch (error) {
      console.error('Fetch positions error:', error);
      throw error;
    }
  }
}

export const employeeService = new EmployeeService();
