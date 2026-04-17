"use client";

import { useEffect, useMemo, useState } from "react";
import { Search, SlidersHorizontal, ChevronLeft, ChevronRight } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import type { BadgeProps } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { jamKerjaApi } from "@/lib/api/jam-kerja";

type WorkDays = string;

interface WorkScheduleRow {
  id: string;
  name: string;
  nik: string;
  jabatan: string;
  workDays: WorkDays;
  startTime: string;
  endTime: string;
  dayOfWeek: string[];
}

function workDaysBadgeVariant(v: WorkDays) {
  type BadgeVariant = BadgeProps["variant"];
  if (v === "Senin - Sabtu") return "success" as BadgeVariant;
  return "secondary" as BadgeVariant;
}

const hariOrder = ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"] as const;

function workDaysLabelFromHari(hari: string[]): WorkDays {
  const set = new Set(hari);
  const isMonFri = ["Senin", "Selasa", "Rabu", "Kamis", "Jumat"].every((h) => set.has(h));
  if (isMonFri && hari.length === 5) return "Senin - Jumat";
  const isMonSat = ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"].every((h) => set.has(h));
  if (isMonSat && hari.length === 6) return "Senin - Sabtu";
  return "Shift";
}

export default function ManajemenJamKerjaManagerDepartemenPage() {
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  const [searchQuery, setSearchQuery] = useState("");
  const [jabatanFilter, setJabatanFilter] = useState("all");

  const [rows, setRows] = useState<WorkScheduleRow[]>([]);

  const [page, setPage] = useState(1);
  const pageSize = 10;

  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [selectedUserName, setSelectedUserName] = useState<string | null>(null);
  const [formDays, setFormDays] = useState<string[]>([]);
  const [formStartTime, setFormStartTime] = useState("08:00");
  const [formEndTime, setFormEndTime] = useState("17:00");
  const [formActive, setFormActive] = useState(true);
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      try {
        setLoading(true);
        setLoadError(null);
        const data = await jamKerjaApi.listMyDepartment();
        if (cancelled) return;

        const mapped: WorkScheduleRow[] = (data || []).map((r) => {
          const hari = Array.isArray(r.day_of_week) ? r.day_of_week : [];
          return {
            id: r.id,
            name: r.name,
            nik: r.nik,
            jabatan: r.position,
            dayOfWeek: hari,
            workDays: workDaysLabelFromHari(hari),
            startTime: r.startTime,
            endTime: r.endTime,
          };
        });
        setRows(mapped);
      } catch (e) {
        const message = e instanceof Error ? e.message : "Gagal memuat data jam kerja";
        if (!cancelled) {
          setLoadError(message);
          setRows([]);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  const jabatanOptions = useMemo(() => {
    // list jabatan unik untuk dropdown
    const uniq = Array.from(new Set(rows.map((r) => r.jabatan))).sort();
    return uniq;
  }, [rows]);

  const filtered = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    return rows.filter((r) => {
      const matchText = r.name.toLowerCase().includes(q) || r.nik.toLowerCase().includes(q);
      const matchJabatan = jabatanFilter === "all" ? true : r.jabatan === jabatanFilter;
      return matchText && matchJabatan;
    });
  }, [rows, searchQuery, jabatanFilter]);

  const totalItems = filtered.length;
  const totalPages = Math.max(1, Math.ceil(totalItems / pageSize));

  const currentPage = Math.min(page, totalPages);

  const paged = useMemo(() => {
    const start = (currentPage - 1) * pageSize;
    return filtered.slice(start, start + pageSize);
  }, [filtered, currentPage]);

  const from = totalItems === 0 ? 0 : (currentPage - 1) * pageSize + 1;
  const to = Math.min(currentPage * pageSize, totalItems);

  const handleOpenSetSchedule = (employeeId: string) => {
    const row = rows.find((r) => r.id === employeeId);
    setSelectedUserId(employeeId);
    setSelectedUserName(row?.name ?? null);
    setDialogOpen(true);
    setFormError(null);
    setSaving(false);

    jamKerjaApi
      .getByUserId(employeeId)
      .then((detail) => {
        setFormDays(Array.isArray(detail.day_of_week) ? detail.day_of_week : []);
        setFormStartTime(detail.start_time);
        setFormEndTime(detail.end_time);
        setFormActive(!!detail.is_active);
      })
      .catch((e) => {
        const message = e instanceof Error ? e.message : "Gagal memuat detail jam kerja";
        setFormError(message);
      });
  };

  const toggleHari = (hari: string) => {
    setFormDays((prev) => {
      const set = new Set(prev);
      if (set.has(hari)) set.delete(hari);
      else set.add(hari);
      return Array.from(set);
    });
  };

  const handleSave = async () => {
    if (!selectedUserId) return;
    if (formDays.length === 0) {
      setFormError("Pilih minimal 1 hari kerja");
      return;
    }
    if (!formStartTime || !formEndTime) {
      setFormError("Jam mulai dan selesai wajib diisi");
      return;
    }

    setSaving(true);
    setFormError(null);
    try {
      const updated = await jamKerjaApi.updateByUserId(selectedUserId, {
        day_of_week: formDays,
        start_time: formStartTime,
        end_time: formEndTime,
        is_active: formActive,
      });

      setRows((prev) =>
        prev.map((r) =>
          r.id !== selectedUserId
            ? r
            : {
                ...r,
                dayOfWeek: updated.day_of_week,
                workDays: workDaysLabelFromHari(updated.day_of_week),
                startTime: updated.start_time,
                endTime: updated.end_time,
              }
        )
      );
      setDialogOpen(false);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal menyimpan jam kerja";
      setFormError(message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="p-6">
      {/* Header */}
      <div className="mb-5">
        <h1 className="text-xl font-semibold text-gray-900">Manajemen Jam Kerja</h1>
        <p className="mt-1 text-sm text-gray-600">
          Kelola jadwal operasional, shift, dan pengaturan waktu kerja untuk karyawan di departemen Anda.
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
                  placeholder="Cari karyawan berdasarkan nama atau NIK..."
                  className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Filter button + Jabatan select */}
            <div className="flex items-center gap-3">
              <Button variant="outline" className="rounded-xl gap-2">
                <SlidersHorizontal className="h-4 w-4" />
                Filter
              </Button>

              <Select
                value={jabatanFilter}
                onValueChange={(v) => {
                  setJabatanFilter(v);
                  setPage(1);
                }}
              >
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
                    Jabatan
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
                ) : loadError ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-10 text-center text-sm text-red-600">
                      {loadError}
                    </td>
                  </tr>
                ) : paged.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-6 py-10 text-center text-sm text-gray-500">
                      Tidak ada data karyawan ditemukan.
                    </td>
                  </tr>
                ) : (
                  paged.map((r) => (
                    <tr key={r.id} className="hover:bg-gray-50 transition-colors">
                      {/* Karyawan */}
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center text-sm font-semibold text-gray-700">
                            {r.name
                              .split(/\s+/)
                              .filter(Boolean)
                              .map((p) => p[0])
                              .join("")
                              .slice(0, 2)
                              .toUpperCase()}
                          </div>
                          <div>
                            <div className="font-semibold text-gray-900">{r.name}</div>
                            <div className="text-xs text-gray-500">NIK: {r.nik}</div>
                          </div>
                        </div>
                      </td>

                      {/* Jabatan */}
                      <td className="px-6 py-4">
                        <div className="text-sm font-semibold text-gray-900">{r.jabatan}</div>
                      </td>

                      {/* Hari kerja */}
                      <td className="px-6 py-4">
                        <Badge variant={workDaysBadgeVariant(r.workDays)}>{r.workDays}</Badge>
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
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Footer + pagination */}
          <div className="mt-4 flex items-center justify-between text-sm text-gray-600">
            <div>Menampilkan {from}-{to} dari {totalItems} karyawan</div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage(Math.max(1, currentPage - 1))}
                disabled={currentPage === 1}
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
                    p === currentPage
                      ? "border-blue-600 text-blue-600 bg-blue-50"
                      : "border-gray-200 text-gray-700 bg-white hover:bg-gray-50",
                  ].join(" ")}
                >
                  {p}
                </button>
              ))}

              <button
                onClick={() => setPage(Math.min(totalPages, currentPage + 1))}
                disabled={currentPage === totalPages}
                className="h-9 w-9 rounded-lg border border-gray-200 bg-white flex items-center justify-center
                           disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
              >
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Atur Jam Kerja</DialogTitle>
            <DialogDescription>
              {selectedUserName ? `Karyawan: ${selectedUserName}` : ""}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <div className="text-sm font-semibold text-gray-900">Hari Kerja</div>
              <div className="mt-2 grid grid-cols-2 gap-2">
                {hariOrder.map((h) => (
                  <label key={h} className="flex items-center gap-2 text-sm text-gray-700">
                    <input
                      type="checkbox"
                      checked={formDays.includes(h)}
                      onChange={() => toggleHari(h)}
                    />
                    {h}
                  </label>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <div className="text-sm font-semibold text-gray-900">Mulai</div>
                <input
                  type="time"
                  value={formStartTime}
                  onChange={(e) => setFormStartTime(e.target.value)}
                  className="mt-2 w-full rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm"
                />
              </div>

              <div>
                <div className="text-sm font-semibold text-gray-900">Selesai</div>
                <input
                  type="time"
                  value={formEndTime}
                  onChange={(e) => setFormEndTime(e.target.value)}
                  className="mt-2 w-full rounded-xl border border-gray-200 bg-white px-3 py-2 text-sm"
                />
              </div>
            </div>

            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={formActive} onChange={(e) => setFormActive(e.target.checked)} />
              Jadwal aktif
            </label>

            {formError && <div className="text-sm text-red-600">{formError}</div>}
          </div>

          <DialogFooter>
            <Button variant="outline" className="rounded-xl" onClick={() => setDialogOpen(false)} disabled={saving}>
              Batal
            </Button>
            <Button className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl" onClick={handleSave} disabled={saving}>
              {saving ? "Menyimpan..." : "Simpan"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
