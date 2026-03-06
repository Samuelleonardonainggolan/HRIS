"use client";

import { useState } from "react";
import { EmployeeTable } from "@/components/employee-table";
import { EmployeeDetailPanel } from "@/components/employee-detail-panel";
import { allEmployeesData } from "@/lib/mock-data";
import { Employee } from "@/types";

export default function PegawaiPage() {
  const [selectedEmployee, setSelectedEmployee] = useState<Employee | null>(
    allEmployeesData[1] // Default: David Miller
  );

  const handleSelectEmployee = (employee: Employee) => {
    setSelectedEmployee(employee);
  };

  const handleCloseDetail = () => {
    setSelectedEmployee(null);
  };

  return (
    <div className="flex h-full gap-6 p-6">
      {/* Main Content - Table Area */}
      <div className="flex-1 overflow-y-auto">
        <EmployeeTable
          employees={allEmployeesData}
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
        <div className="w-80 shrink-0 flex items-center justify-center">
          <div className="text-center text-gray-400">
            <p className="text-sm">Pilih pegawai untuk melihat detail</p>
          </div>
        </div>
      )}
    </div>
  );
}