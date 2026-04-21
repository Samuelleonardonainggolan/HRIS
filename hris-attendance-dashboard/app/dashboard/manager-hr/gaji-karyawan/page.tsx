"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Plus, Search, Filter, ChevronDown } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

type Row = {
  id: string; // salary record id
  userId: string;
  initials: string;
  name: string;
  nik: string;

  // ✅ new
  department: string;
  position: string;

  basicSalary: number;
  effectiveFrom: string; // YYYY-MM-DD for display
  isActive: boolean;
};

function formatIDR(n: number) {
  return `Rp ${n.toLocaleString("id-ID")}`;
}

function StatusBadge({ active }: { active: boolean }) {
  if (active) {
    return (
      <Badge className="rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
        • Active
      </Badge>
    );
  }
  return (
    <Badge className="rounded-full bg-gray-100 text-gray-700 border border-gray-200">
      • Inactive
    </Badge>
  );
}

export default function MasterGajiPage() {
  const router = useRouter();
  const [q, setQ] = useState("");

  // ✅ department filter state
  const [dept, setDept] = useState<string>("Semua Departemen");

  // Mock: nanti dari API join (users + employee_basic_salaries aktif)
  const rows: Row[] = useMemo(
    () => [
      {
        id: "sal1",
        userId: "u1",
        initials: "AW",
        name: "Arya Wijaya",
        nik: "EMP-2023-041",
        department: "IT",
        position: "IT Developer",
        basicSalary: 15_500_000,
        effectiveFrom: "2024-06-01",
        isActive: true,
      },
      {
        id: "sal2",
        userId: "u2",
        initials: "BN",
        name: "Budi Nugroho",
        nik: "EMP-2023-089",
        department: "Housekeeping",
        position: "Supervisor",
        basicSalary: 8_450_000,
        effectiveFrom: "2024-05-01",
        isActive: true,
      },
      {
        id: "sal3",
        userId: "u3",
        initials: "CD",
        name: "Citra Dewi",
        nik: "EMP-2022-112",
        department: "Front Office",
        position: "Receptionist",
        basicSalary: 7_200_000,
        effectiveFrom: "2024-04-01",
        isActive: false,
      },
    ],
    []
  );

  // ✅ build department list from rows
  const departments = useMemo(() => {
    const uniq = Array.from(new Set(rows.map((r) => r.department))).sort();
    return ["Semua Departemen", ...uniq];
  }, [rows]);

  const filtered = useMemo(() => {
    const qq = q.trim().toLowerCase();

    return rows.filter((r) => {
      const matchQ =
        !qq ||
        r.name.toLowerCase().includes(qq) ||
        r.nik.toLowerCase().includes(qq) ||
        r.department.toLowerCase().includes(qq) ||
        r.position.toLowerCase().includes(qq);

      const matchDept =
        dept === "Semua Departemen" || r.department === dept;

      return matchQ && matchDept;
    });
  }, [q, rows, dept]);

  return (
    <div className="p-6 space-y-6">
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Master Data Gaji</h1>
          <p className="text-gray-600">
            Kelola gaji pokok karyawan (basic salary) dan histori perubahannya.
          </p>
        </div>

        <Button
          className="rounded-xl gap-2 bg-blue-600 hover:bg-blue-700 text-white"
          onClick={() => router.push("/dashboard/manager-hr/gaji-karyawan/tambah")}
        >
          <Plus className="h-4 w-4" />
          Tambah
        </Button>
      </div>

      <Card className="rounded-2xl">
        <CardContent className="p-5 space-y-4">
          {/* Toolbar */}
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="relative w-full md:max-w-lg">
              <Search className="h-4 w-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
              <Input
                value={q}
                onChange={(e) => setQ(e.target.value)}
                placeholder="Find by name, NIK, dept, posisi..."
                className="pl-9 rounded-xl"
              />
            </div>

            <div className="flex flex-wrap items-center gap-2">
              {/* ✅ Department Filter */}
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="rounded-xl gap-2">
                    {dept}
                    <ChevronDown className="h-4 w-4 text-gray-500" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {departments.map((d) => (
                    <DropdownMenuItem key={d} onClick={() => setDept(d)}>
                      {d}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              <Button variant="outline" className="rounded-xl gap-2">
                <Filter className="h-4 w-4" />
                Filter
              </Button>
            </div>
          </div>

          {/* Table */}
          <div className="overflow-hidden rounded-xl border border-gray-100">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-100">
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Employee
                  </th>

                  {/* ✅ new column */}
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Departemen
                  </th>

                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Basic Salary
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Effective From
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Actions
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y divide-gray-100 bg-white">
                {filtered.map((r) => (
                  <tr key={r.id} className="hover:bg-gray-50 transition-colors">
                    {/* Employee */}
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-9 w-9 rounded-full bg-gray-100 flex items-center justify-center text-xs font-semibold text-gray-700">
                          {r.initials}
                        </div>
                        <div>
                          <div className="text-sm font-semibold text-gray-900">
                            {r.name}
                          </div>
                          <div className="text-xs text-gray-500">{r.nik}</div>
                        </div>
                      </div>
                    </td>

                    {/* ✅ Department + position */}
                    <td className="px-6 py-4">
                      <div className="text-sm font-semibold text-gray-900">
                        {r.department}
                      </div>
                      <div className="text-xs text-gray-500">{r.position}</div>
                    </td>

                    {/* Basic salary */}
                    <td className="px-6 py-4">
                      <div className="text-sm font-semibold text-gray-900">
                        {formatIDR(r.basicSalary)}
                      </div>
                      <div className="text-xs text-gray-500">/mo</div>
                    </td>

                    {/* Effective from */}
                    <td className="px-6 py-4 text-sm text-gray-700">
                      {r.effectiveFrom}
                    </td>

                    {/* Status */}
                    <td className="px-6 py-4">
                      <StatusBadge active={r.isActive} />
                    </td>

                    {/* Actions */}
                    <td className="px-6 py-4 text-right">
                      <Button
                        variant="outline"
                        className="rounded-xl"
                        onClick={() =>
                          router.push(
                            `/dashboard/manager-hr/gaji-karyawan/${r.userId}/edit`
                          )
                        }
                      >
                        Edit
                      </Button>
                    </td>
                  </tr>
                ))}

                {filtered.length === 0 && (
                  <tr>
                    <td
                      colSpan={6}
                      className="px-6 py-10 text-center text-sm text-gray-500"
                    >
                      Tidak ada data.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between text-sm text-gray-500">
            <div>
              Showing 1 to {Math.min(filtered.length, 10)} of {filtered.length} results
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" className="rounded-xl">
                Previous
              </Button>
              <Button variant="outline" size="sm" className="rounded-xl">
                Next
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}