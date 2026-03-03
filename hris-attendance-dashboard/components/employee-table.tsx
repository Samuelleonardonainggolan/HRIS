"use client";

import { useState } from "react";
import { Search, MoreVertical } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Employee } from "@/types";

interface EmployeeTableProps {
  employees: Employee[];
  onSelectEmployee: (employee: Employee) => void;
  selectedEmployeeId?: string;
}

export function EmployeeTable({
  employees,
  onSelectEmployee,
  selectedEmployeeId,
}: EmployeeTableProps) {
  const [searchQuery, setSearchQuery] = useState("");

  const filteredEmployees = employees.filter((employee) =>
    employee.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <Card>
      <CardContent className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-semibold text-gray-900">
            Manajemen Pegawai
          </h2>
          <Button variant="primary" className="flex items-center gap-2">
            <span className="text-lg">+</span>
            Tambah Pegawai
          </Button>
        </div>

        {/* Search Bar */}
        <div className="mb-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Cari pegawai..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full rounded-lg border border-gray-300 bg-white py-2 pl-10 pr-4 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Foto
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Nama Lengkap
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  NIK
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Departemen
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Jabatan
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Status
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Aksi
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filteredEmployees.map((employee) => (
                <tr
                  key={employee.id}
                  onClick={() => onSelectEmployee(employee)}
                  className={`hover:bg-gray-50 cursor-pointer transition-colors ${
                    selectedEmployeeId === employee.id ? "bg-blue-50" : ""
                  }`}
                >
                  <td className="py-4">
                    <Avatar className="h-10 w-10">
                      <AvatarFallback
                        className={
                          employee.status === "AKTIF"
                            ? "bg-teal-500 text-white"
                            : employee.status === "NONAKTIF"
                            ? "bg-gray-400 text-white"
                            : "bg-blue-500 text-white"
                        }
                      >
                        {employee.avatar}
                      </AvatarFallback>
                    </Avatar>
                  </td>
                  <td className="py-4">
                    <span className="font-medium text-gray-900">
                      {employee.name}
                    </span>
                  </td>
                  <td className="py-4">
                    <span className="text-sm text-gray-600">{employee.nik}</span>
                  </td>
                  <td className="py-4">
                    <span className="text-sm text-gray-900">
                      {employee.department}
                    </span>
                  </td>
                  <td className="py-4">
                    <span className="text-sm text-gray-900">
                      {employee.position}
                    </span>
                  </td>
                  <td className="py-4">
                    <Badge
                      variant={
                        employee.status === "AKTIF" ? "success" : "default"
                      }
                    >
                      {employee.status}
                    </Badge>
                  </td>
                  <td className="py-4">
                    <button className="text-gray-400 hover:text-gray-600">
                      <MoreVertical className="h-5 w-5" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}