"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Search } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { EmployeeDetailPanel } from "@/components/employee-detail-panel";
import { Employee } from "@/types";
import { employeeService } from "@/lib/api/employee";
import { User } from "@/lib/api/auth";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export default function PegawaiManagerDepartemenPage() {
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [selectedEmployee, setSelectedEmployee] = useState<Employee | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // filter
  const [searchQuery, setSearchQuery] = useState("");
  const [jabatanFilter, setJabatanFilter] = useState("all");

  const mapUserToEmployee = useCallback((user: User): Employee => {
    const yearEnrolled = user.year_enrolled ? parseInt(user.year_enrolled, 10) : NaN;
    const workYears = Number.isFinite(yearEnrolled)
      ? new Date().getFullYear() - yearEnrolled
      : undefined;

    return {
      id: user.id,
      name: user.full_name,
      nik: user.payroll_number || user.nik || "",
      // manager departemen: kita tetap simpan department utk keperluan lain,
      // tapi di tabel nanti kolomnya diganti menjadi "Jabatan"
      department: user.department_name || user.department || "",
      position: user.position_name || user.position || "", // ✅ jabatan
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

      const data = await employeeService.getAllEmployees();

      // NOTE: kalau backend Anda sudah punya endpoint "employees by department manager",
      // sebaiknya panggil endpoint itu di sini.
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

  const jabatanOptions = useMemo(() => {
    const uniq = Array.from(
      new Set(employees.map((e) => e.position).filter(Boolean))
    ).sort();
    return uniq;
  }, [employees]);

  const filteredEmployees = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    return employees.filter((e) => {
      const matchText =
        e.name.toLowerCase().includes(q) ||
        (e.nik || "").toLowerCase().includes(q) ||
        (e.email || "").toLowerCase().includes(q);

      const matchJabatan = jabatanFilter === "all" ? true : e.position === jabatanFilter;

      return matchText && matchJabatan;
    });
  }, [employees, searchQuery, jabatanFilter]);

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
      {/* Left: Table */}
      <div className="flex-1 overflow-y-auto">
        <Card className="rounded-2xl">
          <CardContent className="p-6">
            {/* Header + Filter */}
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div>
                <h2 className="text-lg font-semibold text-gray-900">Data Karyawan</h2>
                <p className="text-sm text-gray-600">
                  Daftar karyawan yang tergabung di departemen Anda
                </p>
              </div>

              <div className="flex items-center gap-3">
                <div className="relative w-[260px]">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                  <input
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Cari nama / NIK / email..."
                    className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm
                               focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </div>

                <Select value={jabatanFilter} onValueChange={setJabatanFilter}>
                  <SelectTrigger className="rounded-xl w-[220px]">
                    <SelectValue placeholder="Semua Jabatan" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">Semua Jabatan</SelectItem>
                    {jabatanOptions.map((j) => (
                      <SelectItem key={j} value={j}>
                        {j}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            {/* Table */}
            <div className="mt-5 overflow-hidden rounded-xl border border-gray-100">
              <table className="w-full">
                <thead>
                  <tr className="bg-gray-50 border-b border-gray-100">
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Karyawan
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      No Payroll
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Jabatan
                    </th>
                    <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Status
                    </th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-gray-100 bg-white">
                  {filteredEmployees.map((e) => {
                    const isSelected = e.id === selectedEmployee?.id;

                    return (
                      <tr
                        key={e.id}
                        onClick={() => handleSelectEmployee(e)}
                        className={[
                          "cursor-pointer hover:bg-gray-50 transition-colors",
                          isSelected ? "bg-blue-50" : "",
                        ].join(" ")}
                      >
                        {/* Karyawan */}
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <div className="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center text-sm font-semibold text-gray-700">
                              {e.name
                                .split(/\s+/)
                                .filter(Boolean)
                                .map((p) => p[0])
                                .join("")
                                .slice(0, 2)
                                .toUpperCase()}
                            </div>
                            <div>
                              <div className="font-semibold text-gray-900">{e.name}</div>
                              <div className="text-xs text-gray-500">{e.email}</div>
                            </div>
                          </div>
                        </td>

                        {/* NO PAYROLL */}
                        <td className="px-6 py-4">
                          <div className="text-sm font-semibold text-gray-900">{e.nik || "-"}</div>
                        </td>

                        {/* Jabatan */}
                        <td className="px-6 py-4">
                          <div className="text-sm font-semibold text-gray-900">{e.position || "-"}</div>
                        </td>

                        {/* Status */}
                        <td className="px-6 py-4">
                          <span
                            className={[
                              "inline-flex items-center rounded-full px-3 py-1 text-xs font-semibold",
                              e.status === "AKTIF"
                                ? "bg-green-50 text-green-700"
                                : "bg-gray-100 text-gray-600",
                            ].join(" ")}
                          >
                            {e.status}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>

              {filteredEmployees.length === 0 && (
                <div className="py-12 text-center text-sm text-gray-500">
                  Tidak ada karyawan ditemukan.
                </div>
              )}
            </div>

            <div className="mt-4 text-xs text-gray-500">
              Menampilkan {filteredEmployees.length} karyawan
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Right: Detail Panel */}
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