"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Search, MoreVertical, UserX, UserCheck } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Employee } from "@/types";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmationDialog } from "@/components/ui/confirmation-dialog";
import { employeeService } from "@/lib/api/employee";
import { authService } from "@/lib/api/auth";
import toast from "react-hot-toast";

interface EmployeeTableProps {
  employees: Employee[];
  onSelectEmployee: (employee: Employee) => void;
  selectedEmployeeId?: string;
  onEmployeeUpdated?: () => void;
}

export function EmployeeTable({
  employees,
  onSelectEmployee,
  selectedEmployeeId,
  onEmployeeUpdated,
}: EmployeeTableProps) {
  const router = useRouter();
  const [searchQuery, setSearchQuery] = useState("");

  // ── State untuk dialog konfirmasi
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [targetEmployee, setTargetEmployee] = useState<Employee | null>(null);
  const [isUpdating, setIsUpdating] = useState(false);

  const filteredEmployees = employees.filter((employee) =>
    employee.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleAddEmployee = () => {
    const role = authService.getUser()?.role;
    const path =
      role === "manager_hr"
        ? "/dashboard/manager-hr/karyawan/tambah-pegawai"
        : "/dashboard/manager-dept/karyawan/tambah-pegawai";
    router.push(path);
  };

  // Buka dialog konfirmasi — hanya untuk aksi Nonaktifkan
  // Untuk Aktifkan, langsung eksekusi tanpa konfirmasi
  const handleToggleClick = (e: React.MouseEvent, employee: Employee) => {
    e.stopPropagation();

    if (employee.status === "AKTIF") {
      // Nonaktifkan → tampilkan konfirmasi
      setTargetEmployee(employee);
      setConfirmOpen(true);
    } else {
      // Aktifkan kembali → langsung eksekusi
      void executeToggle(employee, true);
    }
  };

  // Eksekusi perubahan status
  const executeToggle = async (employee: Employee, newActive: boolean) => {
    setIsUpdating(true);
    try {
      await employeeService.updateEmployee(employee.id, { is_active: newActive });
      toast.success(
        newActive ? "Pegawai berhasil diaktifkan" : "Pegawai berhasil dinonaktifkan"
      );
      onEmployeeUpdated?.();
    } catch (error) {
      console.error("Failed to update status:", error);
      toast.error("Gagal mengubah status pegawai");
    } finally {
      setIsUpdating(false);
    }
  };

  // Dipanggil saat user klik "Ya, Nonaktifkan" di dialog
  const handleConfirmDeactivate = () => {
    if (!targetEmployee) return;
    void executeToggle(targetEmployee, false);
    setTargetEmployee(null);
  };

  return (
    <>
      <Card>
        <CardContent className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-gray-900">
              Manajemen Pegawai
            </h2>
            <Button
              onClick={handleAddEmployee}
              className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white"
            >
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
                    No Payroll
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
                        {employee.avatar && employee.avatar.startsWith("http") ? (
                          <AvatarImage src={employee.avatar} alt={employee.name} />
                        ) : (
                          <AvatarFallback
                            className={
                              employee.status === "AKTIF"
                                ? "bg-teal-500 text-white"
                                : employee.status === "NONAKTIF"
                                ? "bg-gray-400 text-white"
                                : "bg-blue-500 text-white"
                            }
                          >
                            {employee.name
                              ? employee.name
                                  .split(" ")
                                  .map((n) => n[0])
                                  .join("")
                                  .toUpperCase()
                              : "?"}
                          </AvatarFallback>
                        )}
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
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <button
                            className="text-gray-400 hover:text-gray-600 p-1 rounded-full hover:bg-gray-100 disabled:opacity-50"
                            onClick={(e) => e.stopPropagation()}
                            disabled={isUpdating}
                          >
                            <MoreVertical className="h-5 w-5" />
                          </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={(e) => handleToggleClick(e, employee)}
                            className={
                              employee.status === "AKTIF"
                                ? "text-red-600"
                                : "text-green-600"
                            }
                          >
                            {employee.status === "AKTIF" ? (
                              <>
                                <UserX className="mr-2 h-4 w-4" />
                                Nonaktifkan
                              </>
                            ) : (
                              <>
                                <UserCheck className="mr-2 h-4 w-4" />
                                Aktifkan
                              </>
                            )}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      {/* Dialog Konfirmasi Nonaktifkan */}
      <ConfirmationDialog
        open={confirmOpen}
        onOpenChange={(open) => {
          setConfirmOpen(open);
          if (!open) setTargetEmployee(null);
        }}
        icon="deactivate"
        variant="destructive"
        title="Nonaktifkan Pegawai"
        description={
          targetEmployee
            ? `Apakah Anda yakin ingin menonaktifkan ${targetEmployee.name} (${targetEmployee.nik})? Pegawai ini tidak akan dapat melakukan presensi atau mengakses sistem.`
            : "Apakah Anda yakin ingin menonaktifkan pegawai ini?"
        }
        confirmText="Ya, Nonaktifkan"
        cancelText="Batal"
        onConfirm={handleConfirmDeactivate}
      />
    </>
  );
}
