"use client";

import { useState, useEffect } from "react";
import { EmployeeTable } from "@/components/employee-table";
import { EmployeeDetailPanel } from "@/components/employee-detail-panel";
import { Employee } from "@/types";
import { employeeService } from "@/lib/api/employee";
import { User } from "@/lib/api/auth";
import { Loader2 } from "lucide-react";

export default function PegawaiPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [selectedEmployee, setSelectedEmployee] = useState<Employee | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchEmployees = async () => {
      try {
        setIsLoading(true);
        const data = await employeeService.getAllEmployees();

        const mappedEmployees: Employee[] = data.map((user: User) => {
          const name = user.full_name?.trim() || user.email || "-";
          const initialsSource = user.full_name?.trim() || user.email || "";
          const avatar = initialsSource
            ? initialsSource
                .split(/\s+/)
                .filter(Boolean)
                .map((part) => part[0])
                .join("")
                .toUpperCase()
                .slice(0, 2)
            : "??";

          const joinDateCandidate = user.join_date || user.created_at;
          const parsedJoinDate = joinDateCandidate ? new Date(joinDateCandidate) : null;
          const joinDate =
            parsedJoinDate && !Number.isNaN(parsedJoinDate.getTime())
              ? parsedJoinDate.toLocaleDateString("id-ID", {
                  day: "numeric",
                  month: "long",
                  year: "numeric",
                })
              : undefined;

          return {
            id: user.id,
            name,
            nik: user.nik || "-",
            department: user.department || "-",
            position: user.position || "-",
            status: user.is_active ? "AKTIF" : "NONAKTIF",
            email: user.email,
            phone: user.phone,
            joinDate,
            avatar,
          };
        });

        setEmployees(mappedEmployees);
        if (mappedEmployees.length > 0) {
          setSelectedEmployee(mappedEmployees[0]);
        }
      } catch (err: unknown) {
        console.error("Failed to fetch employees:", err);
        setError(err instanceof Error ? err.message : "Gagal memuat data pegawai");
      } finally {
        setIsLoading(false);
      }
    };

    fetchEmployees();
  }, []);

  const handleSelectEmployee = (employee: Employee) => {
    setSelectedEmployee(employee);
  };

  const handleCloseDetail = () => {
    setSelectedEmployee(null);
  };

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <Loader2 className="mx-auto h-8 w-8 animate-spin text-indigo-600" />
          <p className="mt-2 text-sm text-gray-500">Memuat data pegawai...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center text-red-600">
          <p className="text-lg font-semibold">Error</p>
          <p className="text-sm">{error}</p>
          <button 
            onClick={() => window.location.reload()}
            className="mt-4 px-4 py-2 bg-indigo-600 text-white rounded-lg text-sm"
          >
            Coba Lagi
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full gap-6 p-6">
      {/* Main Content - Table Area */}
      <div className="flex-1 overflow-y-auto">
        <EmployeeTable
          employees={employees}
          onSelectEmployee={handleSelectEmployee}
          selectedEmployeeId={selectedEmployee?.id}
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
        <div className="w-80 shrink-0 flex items-center justify-center border-l border-gray-100">
          <div className="text-center text-gray-400">
            <p className="text-sm">Pilih pegawai untuk melihat detail</p>
          </div>
        </div>
      )}
    </div>
  );
}
