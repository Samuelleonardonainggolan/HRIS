"use client";

import { useEffect, useMemo, useState } from "react";
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

type ApiSuccess<T> = {
  success?: boolean;
  message?: string;
  data: T;
};

type SalaryListItem = {
  id: string;
  user_id: string;
  full_name: string;
  payroll_number: string;
  department: string;
  position: string;
  basic_salary: number;
  effective_from: string; // ISO string from backend
  is_active: boolean;
};

type Row = {
  id: string; // salary record id
  userId: string;
  initials: string;
  name: string;
  payrollNumber: string;

  department: string;
  position: string;

  basicSalary: number;
  effectiveFrom: string; // YYYY-MM-DD for display
  isActive: boolean;
};

function formatIDR(n: number) {
  return `Rp ${n.toLocaleString("id-ID")}`;
}

function toYYYYMMDD(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

function initialsOf(fullName: string) {
  const parts = fullName.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "U";
  if (parts.length === 1) return parts[0][0].toUpperCase();
  return (parts[0][0] + parts[1][0]).toUpperCase();
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

function getToken() {
  // ✅ sesuaikan dengan cara Anda menyimpan token
  return (
    localStorage.getItem("access_token") ||
    localStorage.getItem("token") ||
    localStorage.getItem("auth_token") ||
    ""
  );
}

/**
 * ✅ Parser aman:
 * - baca res.text()
 * - coba JSON.parse
 * - kalau gagal, tampilkan body untuk debug (biasanya HTML / text "Unauthorized")
 */
async function readJsonSafe<T>(res: Response): Promise<T> {
  const text = await res.text();

  try {
    return (text ? JSON.parse(text) : null) as T;
  } catch {
    const snippet = text.replace(/\s+/g, " ").slice(0, 300);
    throw new Error(
      `Response bukan JSON valid. Status=${res.status}. Body="${snippet}"`
    );
  }
}

export default function MasterGajiPage() {
  const router = useRouter();
  const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const [q, setQ] = useState("");
  const [dept, setDept] = useState<string>("Semua Departemen");

  const [rows, setRows] = useState<Row[]>([]);
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string>("");

  async function fetchRows() {
    setLoading(true);
    setErrorMsg("");

    try {
      const token = getToken();
      if (!token) throw new Error("Token tidak ditemukan. Silakan login ulang.");

      const params = new URLSearchParams();
      if (q.trim()) params.set("q", q.trim());
      if (dept && dept !== "Semua Departemen") params.set("department", dept);

      const url = `${API_URL}/employee-basic-salaries?${params.toString()}`;

      const res = await fetch(url, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        cache: "no-store",
      });

      const json = await readJsonSafe<ApiSuccess<SalaryListItem[]>>(res);

      if (!res.ok) {
        throw new Error(json?.message || `HTTP ${res.status}`);
      }

      const mapped: Row[] = (json.data || []).map((it) => ({
        id: it.id,
        userId: it.user_id,
        initials: initialsOf(it.full_name),
        name: it.full_name,
        payrollNumber: it.payroll_number,
        department: it.department,
        position: it.position,
        basicSalary: it.basic_salary,
        effectiveFrom: toYYYYMMDD(it.effective_from),
        isActive: it.is_active,
      }));

      setRows(mapped);
    } catch (e: any) {
      setErrorMsg(e?.message || "Gagal mengambil data gaji karyawan.");
      setRows([]);
    } finally {
      setLoading(false);
    }
  }

  // initial load
  useEffect(() => {
    fetchRows();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // dropdown departemen berasal dari rows (hasil backend)
  const departments = useMemo(() => {
    const uniq = Array.from(new Set(rows.map((r) => r.department))).sort();
    return ["Semua Departemen", ...uniq];
  }, [rows]);

  // Filter client-side untuk responsif (tetap), walau server-side juga dipakai
  const filtered = useMemo(() => {
    const qq = q.trim().toLowerCase();

    return rows.filter((r) => {
      const matchQ =
        !qq ||
        r.name.toLowerCase().includes(qq) ||
        r.payrollNumber.toLowerCase().includes(qq) ||
        r.department.toLowerCase().includes(qq) ||
        r.position.toLowerCase().includes(qq);

      const matchDept = dept === "Semua Departemen" || r.department === dept;

      return matchQ && matchDept;
    });
  }, [q, rows, dept]);

  // Debounce fetch when q/dept changes (server-side filter)
  useEffect(() => {
    const t = setTimeout(() => {
      fetchRows();
    }, 350);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [q, dept]);

  async function onDeactivate(row: Row) {
    if (!row.isActive) return;

    const ok = window.confirm(
      `Nonaktifkan gaji pokok aktif untuk ${row.name} (${row.payrollNumber})?`
    );
    if (!ok) return;

    try {
      const token = getToken();
      if (!token) throw new Error("Token tidak ditemukan. Silakan login ulang.");

      const url = `${API_URL}/employee-basic-salaries/users/${row.userId}/deactivate`;

      const res = await fetch(url, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      const json = await readJsonSafe<ApiSuccess<null>>(res);

      if (!res.ok) {
        throw new Error(json?.message || `HTTP ${res.status}`);
      }

      await fetchRows();
    } catch (e: any) {
      alert(e?.message || "Gagal menonaktifkan gaji pokok.");
    }
  }

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
                placeholder="Find by name, payroll number, dept, posisi..."
                className="pl-9 rounded-xl"
              />
            </div>

            <div className="flex flex-wrap items-center gap-2">
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

              <Button
                variant="outline"
                className="rounded-xl gap-2"
                onClick={fetchRows}
              >
                <Filter className="h-4 w-4" />
                Refresh
              </Button>
            </div>
          </div>

          {/* Error */}
          {errorMsg && (
            <div className="rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-700">
              {errorMsg}
            </div>
          )}

          {/* Loading */}
          {loading && <div className="text-sm text-gray-500">Memuat data...</div>}

          {/* Table */}
          <div className="overflow-hidden rounded-xl border border-gray-100">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-100">
                  <th className="px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Employee
                  </th>
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
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-9 w-9 rounded-full bg-gray-100 flex items-center justify-center text-xs font-semibold text-gray-700">
                          {r.initials}
                        </div>
                        <div>
                          <div className="text-sm font-semibold text-gray-900">
                            {r.name}
                          </div>
                          <div className="text-xs text-gray-500">
                            {r.payrollNumber}
                          </div>
                        </div>
                      </div>
                    </td>

                    <td className="px-6 py-4">
                      <div className="text-sm font-semibold text-gray-900">
                        {r.department}
                      </div>
                      <div className="text-xs text-gray-500">{r.position}</div>
                    </td>

                    <td className="px-6 py-4">
                      <div className="text-sm font-semibold text-gray-900">
                        {formatIDR(r.basicSalary)}
                      </div>
                      <div className="text-xs text-gray-500">/mo</div>
                    </td>

                    <td className="px-6 py-4 text-sm text-gray-700">
                      {r.effectiveFrom}
                    </td>

                    <td className="px-6 py-4">
                      <StatusBadge active={r.isActive} />
                    </td>

                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="outline"
                          className="rounded-xl"
                          onClick={() =>
                            router.push(
                              `/dashboard/manager-hr/gaji-karyawan/${r.userId}`
                            )
                          }
                        >
                          Edit
                        </Button>

                        <Button
                          variant="outline"
                          className="rounded-xl border-rose-200 text-rose-700 hover:bg-rose-50"
                          onClick={() => onDeactivate(r)}
                          disabled={!r.isActive}
                          title={
                            r.isActive
                              ? "Nonaktifkan gaji pokok aktif"
                              : "Sudah nonaktif"
                          }
                        >
                          Nonaktifkan
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}

                {!loading && filtered.length === 0 && (
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