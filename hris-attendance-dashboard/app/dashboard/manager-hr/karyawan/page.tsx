"use client";

import { useState, useEffect, useCallback } from "react";
import { EmployeeTable } from "@/components/employee-table";
import { EmployeeDetailPanel } from "@/components/employee-detail-panel";
import { Employee } from "@/types";
import { employeeService } from "@/lib/api/employee";
import { User } from "@/lib/api/auth";

export default function PegawaiPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [selectedEmployee, setSelectedEmployee] = useState<Employee | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const mapUserToEmployee = useCallback((user: User): Employee => {
    const yearEnrolled = user.year_enrolled ? parseInt(user.year_enrolled, 10) : NaN;
    const workYears = Number.isFinite(yearEnrolled)
      ? new Date().getFullYear() - yearEnrolled
      : undefined;

    return {
      id: user.id,
      name: user.full_name,
      nik: user.payroll_number || user.nik || "",
      department: user.department_name || user.department || "",
      position: user.position_name || user.position || "",
      email: user.email,
      phone: user.phone,
      avatar: user.avatar,
      status: user.is_active ? "AKTIF" : "NONAKTIF",
      joinDate: user.created_at,
      address: user.address,
      birthDate: user.birth_date,
      religion: user.religion,
      education: user.last_education,
      yearEnrolled: user.year_enrolled,
      employmentStatus: user.employment_status,
      workYears,
      checkInTime: undefined,
      verified: undefined,
    };
  }, []);

  const fetchEmployees = useCallback(async () => {
    try {
      setLoading(true);
      const data = await employeeService.getEmployeesByScope();
      const mappedEmployees = data.map(mapUserToEmployee);
      setEmployees(mappedEmployees);
    } catch (err) {
      setError("Gagal memuat data pegawai");
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [mapUserToEmployee]);

  useEffect(() => {
    fetchEmployees();
  }, [fetchEmployees]);

  const handleSelectEmployee = (employee: Employee) => {
    setSelectedEmployee(employee);
  };

  const handleCloseDetail = () => {
    setSelectedEmployee(null);
  };

  if (loading) {
    return <div className="flex h-full items-center justify-center">Loading...</div>;
  }

  if (error) {
    return <div className="flex h-full items-center justify-center text-red-500">{error}</div>;
  }

  return (
    <div className="flex h-full gap-6 p-6">
      {/* Main Content - Table Area */}
      <div className="flex-1 overflow-y-auto">
        <EmployeeTable
          employees={employees}
          onSelectEmployee={handleSelectEmployee}
          selectedEmployeeId={selectedEmployee?.id}
          onEmployeeUpdated={fetchEmployees}
        />
      </div>

      {/* Detail Panel - Always Visible */}
      {selectedEmployee ? (
        <div className="w-80 shrink-0">
          <EmployeeDetailPanel
            employee={selectedEmployee}
            onClose={handleCloseDetail}
          />
        </div>
      ) : (
        <div className="w-80 shrink-0 flex items-center justify-center">
          <div className="text-center text-gray-400">
            <p className="text-sm">Pilih pegawai untuk melihat detail</p>
          </div>
        </div>
      )}
    </div>
  );
}
