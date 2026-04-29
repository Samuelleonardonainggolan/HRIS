"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Search,
  SlidersHorizontal,
  ChevronLeft,
  ChevronRight,
  Plus,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useRouter } from "next/navigation";

type WorkDays = string;

interface WorkScheduleRow {
  id: string; // user_id
  name: string;
  nik: string;
  department: string;
  position: string;

  // ✅ dari backend
  day_of_week?: string[];
  hari_kerja?: string[];

  // fallback (kalau masih ada)
  workDays?: WorkDays;

  startTime: string; // "08:00"
  endTime: string; // "17:00"
}

const hariOrder = ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"] as const;

function sortHariKerja(hari: string[] | undefined) {
  if (!hari) return [];
  const set = new Set(hari);
  return hariOrder.filter((h) => set.has(h));
}

function hariBadgeVariant(hari: string) {
  switch (hari) {
    case "Senin":
      return "bg-blue-50 text-blue-700 border-blue-200";
    case "Selasa":
      return "bg-indigo-50 text-indigo-700 border-indigo-200";
    case "Rabu":
      return "bg-violet-50 text-violet-700 border-violet-200";
    case "Kamis":
      return "bg-amber-50 text-amber-800 border-amber-200";
    case "Jumat":
      return "bg-emerald-50 text-emerald-700 border-emerald-200";
    case "Sabtu":
      return "bg-teal-50 text-teal-700 border-teal-200";
    case "Minggu":
      return "bg-rose-50 text-rose-700 border-rose-200";
    default:
      return "bg-gray-50 text-gray-700 border-gray-200";
  }
}

export default function ManajemenJamKerjaPage() {
  const router = useRouter();

  const [loading, setLoading] = useState(true);

  const [searchQuery, setSearchQuery] = useState("");
  const [departmentFilter, setDepartmentFilter] = useState("all");

  const [rows, setRows] = useState<WorkScheduleRow[]>([]);

  // pagination
  const [page, setPage] = useState(1);
  const pageSize = 4;

  const load = async () => {
    try {
      setLoading(true);

      const token = localStorage.getItem("access_token");

      const res = await fetch("/api/v1/jam-kerja", {
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
      });

      const json = await res.json().catch(() => null);

      if (!res.ok) {
        console.error("API error:", res.status, json);
        throw new Error(json?.error || json?.message || "Gagal memuat data jam kerja");
      }

      const data = (json?.data || []) as WorkScheduleRow[];
      setRows(Array.isArray(data) ? data : []);
    } catch (e) {
      console.error(e);
      setRows([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      if (cancelled) return;
      await load();
    };

    run();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const departmentOptions = useMemo(() => {
    const uniq = Array.from(new Set(rows.map((r) => r.department).filter(Boolean)));
    uniq.sort();
    return uniq;
  }, [rows]);

  const filtered = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    return rows.filter((r) => {
      const matchText = r.name.toLowerCase().includes(q) || r.nik.toLowerCase().includes(q);
      const matchDept = departmentFilter === "all" ? true : r.department === departmentFilter;
      return matchText && matchDept;
    });
  }, [rows, searchQuery, departmentFilter]);

  const totalItems = filtered.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));

  const paged = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filtered.slice(start, start + pageSize);
  }, [filtered, page]);

  useEffect(() => {
    if (page > totalPages) setPage(totalPages);
  }, [page, totalPages]);

  const from = totalItems === 0 ? 0 : (page - 1) * pageSize + 1;
  const to = Math.min(page * pageSize, totalItems);

  const handleOpenSetSchedule = (employeeId: string) => {
    router.push(`/dashboard/manager-hr/jam-kerja/${employeeId}`);
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-5">
        <h1 className="text-xl font-semibold text-gray-900">Manajemen Jam Kerja</h1>
        <p className="mt-1 text-sm text-gray-600">
          Kelola jadwal operasional, shift, dan pengaturan waktu kerja untuk seluruh departemen.
        </p>
      </div>

      <Card className="rounded-2xl">
        <CardContent className="p-6">
          {/* Top controls */}
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            {/* Search */}
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setPage(1);
                  }}
                  placeholder="Cari karyawan berdasarkan nama atau No Payroll..."
                  className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Filter + tambah + departemen */}
            <div className="flex items-center gap-3">
              <Button variant="outline" className="rounded-xl gap-2">
                <SlidersHorizontal className="h-4 w-4" />
                Filter
              </Button>

              <Button
                className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl gap-2"
                onClick={() => router.push("/dashboard/manager-hr/jam-kerja/tambah-jam-kerja")}
              >
                <Plus className="h-4 w-4" />
                Tambah
              </Button>

              <Select
                value={departmentFilter}
                onValueChange={(v) => {
                  setDepartmentFilter(v);
                  setPage(1);
                }}
              >
                <SelectTrigger className="rounded-xl w-[200px]">
                  <SelectValue placeholder="Semua Departemen" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Departemen</SelectItem>
                  {departmentOptions.map((d) => (
                    <SelectItem key={d} value={d}>
                      {d}
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
                    Departemen &amp; Posisi
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Hari Kerja
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Jam Kerja
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Aksi
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y divide-gray-100 bg-white">
                {loading ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-10 text-center text-sm text-gray-500">
                      Memuat data jam kerja...
                    </td>
                  </tr>
                ) : paged.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-10 text-center text-sm text-gray-500">
                      Tidak ada data karyawan ditemukan.
                    </td>
                  </tr>
                ) : (
                  paged.map((r) => {
                    const hari = sortHariKerja(r.day_of_week || r.hari_kerja);

                    return (
                      <tr key={r.id} className="hover:bg-gray-50 transition-colors">
                        {/* Karyawan */}
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-3">
                            <Avatar className="h-10 w-10">
                              {r.avatar && r.avatar.startsWith("http") ? (
                                <AvatarImage src={r.avatar} alt={r.name} />
                              ) : (
                                <AvatarFallback>
                                  {r.name
                                    ? r.name
                                        .split(" ")
                                        .map((n) => n[0])
                                        .join("")
                                        .toUpperCase()
                                    : "?"}
                                </AvatarFallback>
                              )}
                            </Avatar>
                            <div>
                              <div className="font-semibold text-gray-900">{r.name}</div>
                              <div className="text-xs text-gray-500">NO PAY: {r.nik}</div>
                            </div>
                          </div>
                        </td>

                        {/* Dept & Posisi */}
                        <td className="px-6 py-4">
                          <div className="text-sm font-semibold text-gray-900">{r.department}</div>
                          <div className="text-xs text-gray-500">{r.position}</div>
                        </td>

                        {/* Hari kerja (badge per hari) */}
                        <td className="px-6 py-4">
                          {hari.length === 0 ? (
                            <span className="inline-flex items-center rounded-full border border-gray-200 bg-gray-50 px-2.5 py-0.5 text-xs font-semibold text-gray-700">
                              -
                            </span>
                          ) : (
                            <div className="flex flex-wrap gap-2">
                              {hari.map((h) => (
                                <span
                                  key={h}
                                  className={[
                                    "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold",
                                    hariBadgeVariant(h),
                                  ].join(" ")}
                                >
                                  {h}
                                </span>
                              ))}
                            </div>
                          )}
                        </td>

                        {/* Jam kerja */}
                        <td className="px-6 py-4">
                          <div className="flex items-center gap-10">
                            <div>
                              <div className="text-sm font-semibold text-gray-900">{r.startTime}</div>
                              <div className="text-[11px] font-semibold text-gray-400 uppercase">
                                Mulai
                              </div>
                            </div>
                            <div>
                              <div className="text-sm font-semibold text-gray-900">{r.endTime}</div>
                              <div className="text-[11px] font-semibold text-gray-400 uppercase">
                                Selesai
                              </div>
                            </div>
                          </div>
                        </td>

                        {/* Aksi */}
                        <td className="px-6 py-4 text-right">
                          <Button
                            className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl"
                            onClick={() => handleOpenSetSchedule(r.id)}
                          >
                            Atur Jam Kerja
                          </Button>
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>

          {/* Footer + pagination */}
          <div className="mt-4 flex items-center justify-between text-sm text-gray-600">
            <div>
              Menampilkan {from}-{to} dari {totalItems} karyawan
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="h-9 w-9 rounded-lg border border-gray-200 bg-white flex items-center justify-center
                           disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                <ChevronLeft className="h-4 w-4" />
              </button>

              {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                <button
                  key={p}
                  onClick={() => setPage(p)}
                  className={[
                    "h-9 w-9 rounded-lg border text-sm font-medium",
                    p === page
                      ? "border-blue-600 text-blue-600 bg-blue-50"
                      : "border-gray-200 text-gray-700 bg-white hover:bg-gray-50",
                  ].join(" ")}
                >
                  {p}
                </button>
              ))}

              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="h-9 w-9 rounded-lg border border-gray-200 bg-white flex items-center justify-center
                           disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}