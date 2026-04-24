"use client";

import { useEffect, useMemo, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Search, Plus, ChevronLeft, ChevronRight, Ban, CheckCircle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import toast from "react-hot-toast";

import { departmentApi } from "@/lib/api/department";
import { positionApi } from "@/lib/api/position";
import { employeeService } from "@/lib/api/employee";

type Status = "Aktif" | "Nonaktif";

interface DepartmentDetail {
  id: string;
  name: string;
}

interface PositionRow {
  id: string;
  name: string;
  subTitle?: string; // contoh: "Backend Engineering"
  staffCount: number;
  status: Status;
  // untuk ikon kecil di kiri (opsional)
  icon?: React.ReactNode;
}

export default function DepartmentPositionsPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const departmentId = searchParams.get("departmentId") || "";

  const [searchQuery, setSearchQuery] = useState("");
  const [department, setDepartment] = useState<DepartmentDetail | null>(null);
  const [positions, setPositions] = useState<PositionRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // pagination simple (client-side)
  const [page, setPage] = useState(1);
  const pageSize = 4;

  useEffect(() => {
    if (!departmentId) {
      setError("Departemen belum dipilih. Kembali ke halaman Departemen dan klik salah satu departemen.");
      setLoading(false);
      return;
    }
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [departmentId]);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      // 1) Get department detail
      const dept = await departmentApi.getById(departmentId);
      setDepartment({
        id: dept.id,
        name: dept.name,
      });

      // 2) Get positions by department
      const allPositions = await positionApi.getAll(departmentId);

      // 3) (Opsional) hitung staf per posisi dari employees
      let staffByPosition: Record<string, number> = {};
      try {
        const employees = await employeeService.getAllEmployees();
        staffByPosition = employees.reduce<Record<string, number>>((acc, emp) => {
          const posId = (emp as { position_id?: string }).position_id;
          if (!posId) return acc;
          acc[posId] = (acc[posId] || 0) + 1;
          return acc;
        }, {});
      } catch (e) {
        // tidak fatal
        console.error("Failed to fetch employees for staff count:", e);
      }

      const mapped: PositionRow[] = allPositions.map((p) => ({
        id: p.id,
        name: p.name ?? "-",
        subTitle: p.description ?? "",
        staffCount: staffByPosition[p.id] ?? 0,
        status: p.is_active ? "Aktif" : "Nonaktif",
      }));

      setPositions(mapped);
      setPage(1);
    } catch (err) {
      console.error(err);
      setError("Gagal memuat data posisi");
    } finally {
      setLoading(false);
    }
  };

  const filteredPositions = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    return positions.filter((p) => p.name.toLowerCase().includes(q));
  }, [positions, searchQuery]);

  const totalItems = filteredPositions.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));

  const pagedPositions = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredPositions.slice(start, start + pageSize);
  }, [filteredPositions, page]);

  useEffect(() => {
    // kalau search berubah dan page kepotong
    if (page > totalPages) setPage(totalPages);
  }, [page, totalPages]);

  const handleAddPosition = () => {
    const qs = departmentId ? `?departmentId=${departmentId}` : "";
    router.push(`/dashboard/manager-hr/jabatan/tambah-jabatan${qs}`);
  };

  const handleEditPosition = (id: string) => {
    router.push(`/dashboard/manager-hr/jabatan/tambah-jabatan?edit=${id}`);
  };

  const handleDeletePosition = async (id: string) => {
    if (!confirm("Apakah Anda yakin ingin menghapus jabatan ini?")) return;
    try {
      await positionApi.delete(id);
      toast.success("Jabatan berhasil dihapus");
      await loadData();
    } catch (e) {
      console.error(e);
      toast.error(e instanceof Error ? e.message : "Gagal menghapus jabatan");
    }
  };

  const handleToggleActive = async (pos: PositionRow) => {
    const willActivate = pos.status === "Nonaktif";
    const confirmText = willActivate
      ? "Apakah Anda yakin ingin mengaktifkan jabatan ini?"
      : "Apakah Anda yakin ingin menonaktifkan jabatan ini?";
    if (!confirm(confirmText)) return;
    try {
      await positionApi.update(pos.id, { is_active: willActivate });
      toast.success(willActivate ? "Jabatan berhasil diaktifkan" : "Jabatan berhasil dinonaktifkan");
      await loadData();
    } catch (e) {
      console.error(e);
      toast.error(willActivate ? "Gagal mengaktifkan jabatan" : "Gagal menonaktifkan jabatan");
    }
  };

  const from = totalItems === 0 ? 0 : (page - 1) * pageSize + 1;
  const to = Math.min(page * pageSize, totalItems);

  return (
    <div className="p-6">
      {/* Breadcrumb */}
      <div className="mb-3 text-sm text-gray-500">
        <button
          onClick={() => router.push("/dashboard/manager-hr/departemen")}
          className="hover:text-blue-600 transition-colors"
        >
          Manajemen Departemen
        </button>
        <span className="mx-2">›</span>
        <span className="text-gray-700 font-medium">{department?.name ?? "-"}</span>
      </div>

      {/* Title + Description + Action */}
      <div className="mb-5 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            {department?.name ?? "Departemen"}
          </h1>
          <p className="mt-1 text-sm text-gray-600">
            Kelola struktur jabatan dan alokasi tim teknis
          </p>
        </div>

        <Button
          onClick={handleAddPosition}
          className="bg-blue-600 hover:bg-blue-700 text-white flex items-center gap-2"
        >
          <Plus className="h-4 w-4" />
          Tambah Posisi
        </Button>
      </div>

      <Card className="rounded-2xl">
        <CardContent className="p-6">
          {loading && <div className="mb-4 text-sm text-gray-600">Memuat posisi...</div>}
          {error && <div className="mb-4 text-sm text-red-600">{error}</div>}

          {/* Search */}
          <div className="mb-5">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Cari posisi..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm
                           focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* Table */}
          <div className="overflow-hidden rounded-xl border border-gray-100">
            <table className="w-full">
              <thead>
                <tr className="bg-gray-50 border-b border-gray-100">
                  <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Jabatan
                  </th>
                  <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Jumlah Staf
                  </th>
                  <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Status
                  </th>
                  <th className="px-5 py-4 text-right text-xs font-semibold uppercase tracking-wider text-gray-500">
                    Aksi
                  </th>
                </tr>
              </thead>

              <tbody className="divide-y divide-gray-100 bg-white">
                {pagedPositions.map((pos) => (
                  <tr key={pos.id} className="hover:bg-gray-50 transition-colors">
                    {/* Jabatan */}
                    <td className="px-5 py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-xl bg-blue-50 flex items-center justify-center text-blue-600">
                          {/* icon placeholder */}
                          <span className="text-sm font-semibold">🏷️</span>
                        </div>
                        <div>
                          <div className="font-semibold text-gray-900">{pos.name}</div>
                          {pos.subTitle ? (
                            <div className="text-xs text-gray-500">{pos.subTitle}</div>
                          ) : null}
                        </div>
                      </div>
                    </td>

                    {/* Staff */}
                    <td className="px-5 py-4">
                      <span className="text-sm font-semibold text-gray-900">
                        {pos.staffCount} Orang
                      </span>
                    </td>

                    {/* Status */}
                    <td className="px-5 py-4">
                      <Badge
                        variant={pos.status === "Aktif" ? "success" : "secondary"}
                      >
                        {pos.status.toUpperCase()}
                      </Badge>
                    </td>

                    {/* Aksi */}
                    <td className="px-5 py-4 text-right">
                      <div className="inline-flex items-center gap-2">
                        <button
                          onClick={() => handleEditPosition(pos.id)}
                          className="rounded-lg px-3 py-2 text-sm text-blue-600 hover:bg-blue-50 transition-colors"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => handleToggleActive(pos)}
                          className={[
                            "rounded-lg px-3 py-2 text-sm transition-colors",
                            pos.status === "Aktif"
                              ? "text-amber-700 hover:bg-amber-50"
                              : "text-emerald-700 hover:bg-emerald-50",
                          ].join(" ")}
                          title={pos.status === "Aktif" ? "Nonaktifkan" : "Aktifkan"}
                        >
                          <span className="inline-flex items-center gap-2">
                            {pos.status === "Aktif" ? (
                              <Ban className="h-4 w-4" />
                            ) : (
                              <CheckCircle className="h-4 w-4" />
                            )}
                            {pos.status === "Aktif" ? "Nonaktifkan" : "Aktifkan"}
                          </span>
                        </button>
                        <button
                          onClick={() => handleDeletePosition(pos.id)}
                          className="rounded-lg px-3 py-2 text-sm text-red-600 hover:bg-red-50 transition-colors"
                        >
                          Hapus
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>

            {/* Empty state */}
            {filteredPositions.length === 0 && !loading && (
              <div className="py-12 text-center">
                <div className="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-gray-100">
                  <Search className="h-6 w-6 text-gray-400" />
                </div>
                <div className="text-sm font-medium text-gray-900">
                  Tidak ada posisi ditemukan
                </div>
                <div className="mt-1 text-sm text-gray-500">
                  Coba ubah kata kunci pencarian Anda
                </div>
              </div>
            )}
          </div>

          {/* Footer: range + pagination */}
          <div className="mt-4 flex items-center justify-between text-sm text-gray-600">
            <div>
              Menampilkan {from}–{to} dari {totalItems} Posisi
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="h-9 w-9 rounded-lg border border-gray-200 bg-white flex items-center justify-center
                           disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
                title="Sebelumnya"
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
                title="Berikutnya"
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
