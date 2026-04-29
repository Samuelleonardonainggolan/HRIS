"use client";

import { useState, useEffect, useCallback } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { toast, Toaster } from "react-hot-toast";

import { EmployeeTable } from "@/components/employee-table";
import { EmployeeDetailPanel } from "@/components/employee-detail-panel";
import { Employee } from "@/types";
import { employeeService } from "@/lib/api/employee";
import { User } from "@/lib/api/auth";

const CREATED_EMPLOYEE_FLASH_KEY = "flash_created_employee";

type CreatedEmployeeFlashPayload = {
  email?: string;
  temporary_password?: string;
  full_name?: string;
  payroll_number?: string;
  created_at?: string;
};

export default function PegawaiPage() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

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
      setError(null);

      const data = await employeeService.getEmployeesByScope();
      const mappedEmployees = data.map(mapUserToEmployee);
      setEmployees(mappedEmployees);
    } catch (err) {
      setError("Gagal memuat data pegawai");
      console.error(err);
      toast.error("Gagal memuat data pegawai");
    } finally {
      setLoading(false);
    }
  }, [mapUserToEmployee]);

  useEffect(() => {
    fetchEmployees();
  }, [fetchEmployees]);

  // ✅ Flash toast setelah redirect dari page tambah
  useEffect(() => {
    const created = searchParams.get("created");
    if (created !== "1") return;

    let payload: CreatedEmployeeFlashPayload | null = null;

    try {
      const raw = sessionStorage.getItem(CREATED_EMPLOYEE_FLASH_KEY);
      if (raw) payload = JSON.parse(raw) as CreatedEmployeeFlashPayload;
    } catch {
      payload = null;
    }

    toast.custom(
      (t) => (
        <div className="rounded-xl border border-gray-200 bg-white shadow-lg p-4 w-[360px]">
          <div className="flex items-center gap-3 mb-3">
            <div className="h-8 w-8 rounded-full bg-green-100 flex items-center justify-center text-green-600 shrink-0">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <div>
              <div className="text-sm font-semibold text-gray-900">Pegawai berhasil dibuat</div>
              {payload?.full_name && (
                <div className="text-xs text-gray-500">
                  {payload.full_name} {payload?.payroll_number ? `(${payload.payroll_number})` : ""}
                </div>
              )}
            </div>
          </div>

          {payload?.temporary_password && (
            <div className="mt-3 p-3 bg-gray-50 rounded-lg border border-gray-100">
              <div className="text-xs text-gray-600 mb-1">
                Password sementara:{" "}
                <span className="font-mono font-semibold text-gray-900">{payload.temporary_password}</span>
              </div>
              <div className="text-[10px] text-gray-500 mb-2">
                Mohon catat password ini dan berikan kepada karyawan.
              </div>
              <button
                type="button"
                className="text-xs w-full px-3 py-1.5 rounded-md border border-gray-200 hover:bg-white bg-gray-50 text-gray-700 transition-colors"
                onClick={() => {
                  navigator.clipboard
                    .writeText(payload.temporary_password!)
                    .then(() => toast.success("Password disalin"))
                    .catch(() => toast.error("Gagal menyalin password"));
                }}
              >
                Salin Password
              </button>
            </div>
          )}

          <div className="mt-3 p-3 bg-amber-50 border border-amber-100 rounded-lg">
            <div className="text-xs text-amber-800 font-medium mb-1">Tindakan Diperlukan</div>
            <div className="text-xs text-amber-700 mb-2">
              Pegawai baru belum memiliki jam kerja. Segera atur jam kerja agar dapat melakukan presensi.
            </div>
            <button
              type="button"
              className="text-xs w-full px-3 py-1.5 rounded-md bg-amber-100 text-amber-800 hover:bg-amber-200 font-medium transition-colors"
              onClick={() => {
                toast.dismiss(t.id);
                router.push("/dashboard/manager-hr/jam-kerja");
              }}
            >
              Atur Jam Kerja Sekarang
            </button>
          </div>

          <div className="mt-3 flex justify-end">
            <button
              type="button"
              className="text-xs px-4 py-1.5 rounded-lg bg-gray-900 text-white hover:bg-gray-800 transition-colors"
              onClick={() => toast.dismiss(t.id)}
            >
              Tutup
            </button>
          </div>
        </div>
      ),
      { duration: Infinity }
    );

    // cleanup supaya tidak muncul lagi saat refresh
    try {
      sessionStorage.removeItem(CREATED_EMPLOYEE_FLASH_KEY);
    } catch {
      // ignore
    }

    // hapus query param (?created=1)
    router.replace(pathname);
  }, [searchParams, router, pathname]);

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
      <Toaster position="top-right" />

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
          <EmployeeDetailPanel employee={selectedEmployee} onClose={handleCloseDetail} />
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
