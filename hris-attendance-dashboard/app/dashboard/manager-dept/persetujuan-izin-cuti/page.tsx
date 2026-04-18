"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Search, Loader2, MoreVertical } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { format } from "date-fns";
import { id as idLocale } from "date-fns/locale";
import { deptLeaveRequestsApi, type DeptLeaveRequestApprovalResponse } from "@/lib/api/dept-leave-requests";

type RequestStatus = "Pending" | "Disetujui" | "Ditolak";
type RequestType = "SAKIT" | "TAHUNAN" | "IZIN KHUSUS";

interface LeaveApprovalItem {
  id: string;
  employeeName: string;
  employeeId: string;
  department: string;
  position: string;

  type: RequestType;
  startAt: string;
  endAt: string;
  startDateLabel: string;
  startTimeLabel: string;
  endDateLabel: string;
  endTimeLabel: string;

  reason: string;
  attachmentName?: string;
  attachmentSize?: string;

  status: RequestStatus;

  avatarUrl?: string;
  avatarFallback: string;
}

function typeBadgeClass(t: RequestType) {
  switch (t) {
    case "SAKIT":
      return "bg-red-100 text-red-700 border border-red-200";
    case "TAHUNAN":
      return "bg-gray-100 text-gray-800 border border-gray-200";
    case "IZIN KHUSUS":
      return "bg-blue-100 text-blue-700 border border-blue-200";
    default:
      return "bg-gray-100 text-gray-800 border border-gray-200";
  }
}

function statusDotColor(s: RequestStatus) {
  switch (s) {
    case "Pending":
      return "bg-orange-500";
    case "Disetujui":
      return "bg-green-600";
    case "Ditolak":
      return "bg-red-600";
    default:
      return "bg-gray-400";
  }
}

export default function PersetujuanIzinCutiManagerDepartemenPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [searchEmployee, setSearchEmployee] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isActing, setIsActing] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  // ✅ filter jabatan
  const [positionFilter, setPositionFilter] = useState("all");

  const [items, setItems] = useState<LeaveApprovalItem[]>([]);

  const mapApiToUi = useCallback((rows: DeptLeaveRequestApprovalResponse[]): LeaveApprovalItem[] => {
    const toRequestType = (typeName: string): RequestType => {
      const t = (typeName || "").toUpperCase();
      if (t.includes("SAKIT")) return "SAKIT";
      if (t.includes("TAHUN")) return "TAHUNAN";
      if (t.includes("IZIN")) return "IZIN KHUSUS";
      return "IZIN KHUSUS";
    };

    const toStatus = (s: string): RequestStatus => {
      const x = (s || "").toUpperCase();
      if (x === "APPROVED") return "Disetujui";
      if (x === "REJECTED") return "Ditolak";
      return "Pending";
    };

    const fileNameFromUrl = (url?: string) => {
      if (!url) return undefined;
      try {
        const u = new URL(url);
        const path = u.pathname;
        const base = path.split("/").filter(Boolean).pop();
        return base || url;
      } catch {
        const parts = url.split("/").filter(Boolean);
        return parts[parts.length - 1] || url;
      }
    };

    return (rows || []).map((r) => {
      const empName = r.employee?.full_name || "Karyawan";
      const empId = r.employee?.payroll_number || r.pengajuan.user_id;
      const dept = r.employee?.department_name || "";
      const pos = r.employee?.position_name || "";

      const start = new Date(r.pengajuan.start_date);
      const end = new Date(r.pengajuan.end_date);

      const startDateLabel = isNaN(start.getTime()) ? "-" : format(start, "dd MMM yyyy", { locale: idLocale });
      const endDateLabel = isNaN(end.getTime()) ? "-" : format(end, "dd MMM yyyy", { locale: idLocale });
      const startTime = isNaN(start.getTime()) ? "--:--" : format(start, "HH:mm");
      const endTime = isNaN(end.getTime()) ? "--:--" : format(end, "HH:mm");

      return {
        id: r.pengajuan.id,
        employeeName: empName,
        employeeId: empId,
        department: dept,
        position: pos,
        type: toRequestType(r.pengajuan.type_name),
        startAt: isNaN(start.getTime()) ? "-" : format(start, "dd MMM yyyy, HH:mm", { locale: idLocale }),
        endAt: isNaN(end.getTime()) ? "-" : format(end, "dd MMM yyyy, HH:mm", { locale: idLocale }),
        startDateLabel,
        startTimeLabel: `Pukul ${startTime} WIB`,
        endDateLabel,
        endTimeLabel: `Pukul ${endTime} WIB`,
        reason: r.pengajuan.reason || "-",
        attachmentName: fileNameFromUrl(r.pengajuan.document_url),
        status: toStatus(r.pengajuan.status_kepala_departemen),
        avatarUrl: "",
        avatarFallback: empName
          .split(/\s+/)
          .filter(Boolean)
          .map((p) => p[0])
          .join("")
          .slice(0, 2)
          .toUpperCase(),
      };
    });
  }, []);

  const load = useCallback(async () => {
    try {
      setIsLoading(true);
      setLoadError(null);
      const data = await deptLeaveRequestsApi.list({ status: "PENDING", search: searchEmployee || undefined });
      const mapped = mapApiToUi(data);
      setItems(mapped);
      setSelectedId((prev) => {
        if (prev && mapped.some((x) => x.id === prev)) return prev;
        return mapped[0]?.id ?? null;
      });
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal memuat pengajuan";
      setLoadError(message);
      setItems([]);
      setSelectedId(null);
    } finally {
      setIsLoading(false);
    }
  }, [mapApiToUi, searchEmployee]);

  useEffect(() => {
    load();
  }, [load]);

  const positionOptions = useMemo(() => {
    const uniq = Array.from(new Set(items.map((x) => x.position).filter(Boolean))).sort();
    return uniq;
  }, [items]);

  const filtered = useMemo(() => {
    const q = searchEmployee.toLowerCase().trim();

    return items.filter((x) => {
      const matchText =
        !q ? true : x.employeeName.toLowerCase().includes(q) || x.employeeId.toLowerCase().includes(q);

      const matchPosition = positionFilter === "all" ? true : x.position === positionFilter;

      return matchText && matchPosition;
    });
  }, [items, searchEmployee, positionFilter]);

  const selected = useMemo(
    () => filtered.find((x) => x.id === selectedId) ?? filtered[0] ?? null,
    [filtered, selectedId]
  );

  const handleApprove = async () => {
    if (!selected) return;
    setIsActing(true);
    setActionError(null);
    try {
      await deptLeaveRequestsApi.approve(selected.id);
      setItems((prev) => prev.map((x) => (x.id === selected.id ? { ...x, status: "Disetujui" } : x)));
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal menyetujui pengajuan";
      setActionError(message);
    } finally {
      setIsActing(false);
    }
  };

  const handleReject = async () => {
    if (!selected) return;
    setIsActing(true);
    setActionError(null);
    try {
      await deptLeaveRequestsApi.reject(selected.id);
      setItems((prev) => prev.map((x) => (x.id === selected.id ? { ...x, status: "Ditolak" } : x)));
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal menolak pengajuan";
      setActionError(message);
    } finally {
      setIsActing(false);
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <div className="text-center">
          <Loader2 className="mx-auto h-8 w-8 animate-spin text-indigo-600" />
          <p className="mt-2 text-sm text-gray-500">Memuat pengajuan...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full gap-6 p-6">
      {/* LEFT */}
      <div className="flex-1 min-w-0">
        <Card className="h-full">
          <CardContent className="p-6 h-full flex flex-col">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold text-gray-900">
                  Manajemen Izin &amp; Cuti
                </h2>
                <p className="mt-1 text-sm text-gray-600">
                  Persetujuan izin &amp; cuti karyawan dalam departemen Anda
                </p>
              </div>
            </div>

            {/* ✅ Search + Filter Jabatan */}
            <div className="mt-4 flex flex-col gap-3 md:flex-row md:items-center">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input
                  value={searchEmployee}
                  onChange={(e) => setSearchEmployee(e.target.value)}
                  placeholder="Cari karyawan / ID..."
                  className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>

              <Select value={positionFilter} onValueChange={setPositionFilter}>
                <SelectTrigger className="rounded-xl w-full md:w-[240px]">
                  <SelectValue placeholder="Semua Jabatan" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">Semua Jabatan</SelectItem>
                  {positionOptions.map((p) => (
                    <SelectItem key={p} value={p}>
                      {p}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {loadError && (
              <div className="mt-3 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">
                {loadError}
              </div>
            )}

            {/* Table */}
            <div className="mt-5 flex-1 overflow-hidden rounded-xl border border-gray-100">
              <div className="overflow-x-auto h-full">
                <table className="w-full">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-100">
                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Nama Karyawan
                      </th>

                      {/* ✅ Departemen -> Jabatan */}
                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Jabatan
                      </th>

                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Tipe Pengajuan
                      </th>
                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Waktu Mulai
                      </th>
                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Waktu Selesai
                      </th>
                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Status
                      </th>
                    </tr>
                  </thead>

                  <tbody className="divide-y divide-gray-100 bg-white">
                    {filtered.map((x) => {
                      const isSelected = x.id === selected?.id;
                      return (
                        <tr
                          key={x.id}
                          onClick={() => setSelectedId(x.id)}
                          className={[
                            "cursor-pointer hover:bg-gray-50 transition-colors",
                            isSelected ? "bg-blue-50" : "",
                          ].join(" ")}
                        >
                          <td className="px-5 py-4">
                            <div className="flex items-center gap-3">
                              <div className="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center overflow-hidden">
                                {x.avatarUrl ? (
                                  // eslint-disable-next-line @next/next/no-img-element
                                  <img
                                    src={x.avatarUrl}
                                    alt={x.employeeName}
                                    className="h-full w-full object-cover"
                                  />
                                ) : (
                                  <span className="text-xs font-semibold text-gray-700">
                                    {x.avatarFallback}
                                  </span>
                                )}
                              </div>
                              <div>
                                <div className="font-semibold text-gray-900">
                                  {x.employeeName}
                                </div>
                                <div className="text-xs text-gray-500">
                                  ID: {x.employeeId}
                                </div>
                              </div>
                            </div>
                          </td>

                          {/* ✅ Jabatan */}
                          <td className="px-5 py-4 text-sm text-gray-700">
                            {x.position}
                          </td>

                          <td className="px-5 py-4">
                            <Badge variant="secondary" className={typeBadgeClass(x.type)}>
                              {x.type}
                            </Badge>
                          </td>

                          <td className="px-5 py-4 text-sm text-gray-700">{x.startAt}</td>
                          <td className="px-5 py-4 text-sm text-gray-700">{x.endAt}</td>

                          <td className="px-5 py-4">
                            <div className="flex items-center gap-2 text-sm text-gray-700">
                              <span className={`h-2 w-2 rounded-full ${statusDotColor(x.status)}`} />
                              {x.status}
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>

                {filtered.length === 0 && (
                  <div className="p-10 text-center text-sm text-gray-500">
                    Tidak ada pengajuan ditemukan.
                  </div>
                )}
              </div>
            </div>

            <div className="mt-4 text-xs text-gray-500">
              Menampilkan {filtered.length} pengajuan
            </div>
          </CardContent>
        </Card>
      </div>

      {/* RIGHT */}
      <div className="w-[360px] shrink-0">
        <Card className="h-full">
          <CardContent className="p-6 h-full flex flex-col">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-gray-900">Detail Pengajuan</h3>
              <button className="p-2 rounded-lg hover:bg-gray-100">
                <MoreVertical className="h-4 w-4 text-gray-500" />
              </button>
            </div>

            {!selected ? (
              <div className="flex-1 flex items-center justify-center text-sm text-gray-400">
                Pilih pengajuan untuk melihat detail
              </div>
            ) : (
              <>
                <div className="mt-4 rounded-xl border border-gray-100 p-4">
                  <div className="flex items-center gap-3">
                    <div className="h-12 w-12 rounded-xl bg-gray-100 overflow-hidden flex items-center justify-center">
                      {selected.avatarUrl ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img
                          src={selected.avatarUrl}
                          alt={selected.employeeName}
                          className="h-full w-full object-cover"
                        />
                      ) : (
                        <span className="text-sm font-semibold text-gray-700">
                          {selected.avatarFallback}
                        </span>
                      )}
                    </div>
                    <div className="min-w-0">
                      <div className="font-semibold text-gray-900 truncate">
                        {selected.employeeName}
                      </div>
                      {/* Anda boleh tetap tampilkan department di detail, atau hilangkan */}
                      <div className="text-xs text-gray-500 truncate">
                        {selected.department} • {selected.position}
                      </div>
                      <div className="text-xs text-gray-400">{selected.employeeId}</div>
                    </div>
                  </div>
                </div>

                <div className="mt-5 grid grid-cols-2 gap-4">
                  <div>
                    <div className="text-[11px] font-semibold text-gray-500 uppercase">
                      Mulai
                    </div>
                    <div className="mt-2 text-sm font-semibold text-gray-900">
                      {selected.startDateLabel}
                    </div>
                    <div className="text-xs text-gray-500">{selected.startTimeLabel}</div>
                  </div>
                  <div>
                    <div className="text-[11px] font-semibold text-gray-500 uppercase">
                      Selesai
                    </div>
                    <div className="mt-2 text-sm font-semibold text-gray-900">
                      {selected.endDateLabel}
                    </div>
                    <div className="text-xs text-gray-500">{selected.endTimeLabel}</div>
                  </div>
                </div>

                <div className="mt-5">
                  <div className="text-[11px] font-semibold text-gray-500 uppercase">
                    Alasan Pengajuan
                  </div>
                  <div className="mt-2 rounded-xl border border-gray-100 bg-gray-50 p-4 text-sm text-gray-700 leading-relaxed">
                    {selected.reason}
                  </div>
                </div>

                <div className="mt-5">
                  <div className="text-[11px] font-semibold text-gray-500 uppercase">
                    Dokumen Pendukung
                  </div>
                  <div className="mt-2 rounded-xl border border-gray-100 overflow-hidden">
                    <div className="h-28 bg-gradient-to-br from-slate-700 to-slate-900" />
                    <div className="p-3 text-xs text-gray-600">
                      {selected.attachmentName ? (
                        <div className="flex items-center justify-between">
                          <span className="truncate">{selected.attachmentName}</span>
                          <span className="text-gray-400">{selected.attachmentSize}</span>
                        </div>
                      ) : (
                        <div className="text-gray-400">Tidak ada dokumen</div>
                      )}
                    </div>
                  </div>
                </div>

                <div className="mt-auto pt-6 grid grid-cols-2 gap-3">
                  {actionError && (
                    <div className="col-span-2 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700">
                      {actionError}
                    </div>
                  )}
                  <Button
                    variant="outline"
                    className="border-red-200 text-red-600 hover:bg-red-50"
                    onClick={handleReject}
                    disabled={selected.status !== "Pending" || isActing}
                  >
                    Tolak
                  </Button>
                  <Button
                    className="bg-blue-600 hover:bg-blue-700 text-white"
                    onClick={handleApprove}
                    disabled={selected.status !== "Pending" || isActing}
                  >
                    Setuju
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
