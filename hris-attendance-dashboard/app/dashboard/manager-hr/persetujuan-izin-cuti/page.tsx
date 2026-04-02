"use client";

import { useEffect, useMemo, useState } from "react";
import { Search, Filter, Download, Loader2, MoreVertical } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge, type BadgeProps } from "@/components/ui/badge";
import { leaveRequestsApi, LeaveRequestStatus, LeaveRequestApprovalResponse } from "@/lib/api/leave-requests";
import toast from "react-hot-toast";

type RequestStatus = "Pending" | "Disetujui" | "Ditolak";

type RequestType = string;

interface LeaveApprovalItem {
  id: string;
  employeeName: string;
  employeeId: string;
  department: string;
  position: string;

  type: RequestType;
  startAt: string; // display string
  endAt: string;   // display string
  startDateLabel: string; // "22 Okt 2023"
  startTimeLabel: string; // "Pukul 08:00 WIB"
  endDateLabel: string;   // "24 Okt 2023"
  endTimeLabel: string;   // "Pukul 17:00 WIB"

  reason: string;
  attachmentName?: string;
  attachmentSize?: string;

  status: RequestStatus;

  avatarUrl?: string; // optional
  avatarFallback: string; // initials
}

function typeBadgeVariant(t: RequestType) {
  const fallback: BadgeProps["variant"] = "secondary";
  const key = t.toUpperCase();
  switch (key) {
    case "SAKIT":
      return "danger";
    case "TAHUNAN":
      return "secondary";
    case "IZIN KHUSUS":
    case "CUTI KHUSUS":
      return "warning";
    default:
      return fallback;
  }
}

function mapStatus(status: LeaveRequestStatus): RequestStatus {
  switch (status) {
    case "APPROVED":
      return "Disetujui";
    case "REJECTED":
      return "Ditolak";
    default:
      return "Pending";
  }
}

function getInitials(name: string) {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  const first = parts[0]?.[0] ?? "";
  const last = parts.length > 1 ? parts[parts.length - 1]?.[0] ?? "" : "";
  return (first + last).toUpperCase() || "?";
}

function formatDateLabel(d: Date) {
  return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
}

function formatTimeLabel(d: Date) {
  const t = d.toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" });
  return `Pukul ${t} WIB`;
}

function formatListDateTime(d: Date) {
  const date = formatDateLabel(d);
  const time = d.toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" });
  return `${date}, ${time}`;
}

function getFileNameFromUrl(url?: string) {
  if (!url) return undefined;
  try {
    const u = new URL(url);
    const pathname = u.pathname.split("/").filter(Boolean);
    return pathname[pathname.length - 1];
  } catch {
    const parts = url.split("/").filter(Boolean);
    return parts[parts.length - 1];
  }
}

function mapResponseToItem(x: LeaveRequestApprovalResponse): LeaveApprovalItem {
  const employeeName = x.employee?.full_name ?? "(Karyawan)";
  const employeeId = x.employee?.payroll_number ?? x.pengajuan.user_id;
  const department = x.employee?.department_name ?? "-";
  const position = x.employee?.position_name ?? "-";
  const start = new Date(x.pengajuan.tanggal_mulai);
  const end = new Date(x.pengajuan.tanggal_selesai);
  const attachmentName = getFileNameFromUrl(x.pengajuan.dokumen_url);

  return {
    id: x.pengajuan.id,
    employeeName,
    employeeId,
    department,
    position,
    type: (x.pengajuan.nama_tipe || "IZIN KHUSUS").toUpperCase(),
    startAt: formatListDateTime(start),
    endAt: formatListDateTime(end),
    startDateLabel: formatDateLabel(start),
    startTimeLabel: formatTimeLabel(start),
    endDateLabel: formatDateLabel(end),
    endTimeLabel: formatTimeLabel(end),
    reason: x.pengajuan.alasan,
    attachmentName,
    attachmentSize: undefined,
    status: mapStatus(x.pengajuan.status_manager_hr),
    avatarUrl: "",
    avatarFallback: getInitials(employeeName),
  };
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

export default function PersetujuanIzinCutiPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isActing, setIsActing] = useState(false);
  const [searchEmployee, setSearchEmployee] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const [items, setItems] = useState<LeaveApprovalItem[]>([]);

  const activeStatus = "PENDING" as const;

  useEffect(() => {
    let cancelled = false;
    const t = setTimeout(() => {
      (async () => {
        try {
          setLoadError(null);
          setIsLoading(true);
          const res = await leaveRequestsApi.listForManagerHR({
            status: activeStatus,
            search: searchEmployee.trim() || undefined,
          });
          const mapped = res.map(mapResponseToItem);
          if (cancelled) return;
          setItems(mapped);
          setSelectedId((prev) => {
            if (prev && mapped.some((x) => x.id === prev)) return prev;
            return mapped[0]?.id ?? null;
          });
        } catch (e) {
          if (cancelled) return;
          setLoadError(e instanceof Error ? e.message : "Gagal memuat pengajuan");
        } finally {
          if (!cancelled) setIsLoading(false);
        }
      })();
    }, 250);

    return () => {
      cancelled = true;
      clearTimeout(t);
    };
  }, [activeStatus, searchEmployee]);

  const filtered = useMemo(() => {
    const q = searchEmployee.toLowerCase().trim();
    if (!q) return items;
    return items.filter((x) => x.employeeName.toLowerCase().includes(q));
  }, [items, searchEmployee]);

  const selected = useMemo(
    () => filtered.find((x) => x.id === selectedId) ?? filtered[0] ?? null,
    [filtered, selectedId]
  );

  useEffect(() => {
    if (!selected) return;
    setSelectedId(selected.id);
  }, [selected]);

  const handleApprove = async () => {
    if (!selected || isActing) return;
    try {
      setIsActing(true);
      await leaveRequestsApi.approve(selected.id);
      toast.success("Pengajuan disetujui");
      const res = await leaveRequestsApi.listForManagerHR({
        status: activeStatus,
        search: searchEmployee.trim() || undefined,
      });
      const mapped = res.map(mapResponseToItem);
      setItems(mapped);
      setSelectedId(mapped[0]?.id ?? null);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Gagal menyetujui pengajuan");
    } finally {
      setIsActing(false);
    }
  };

  const handleReject = async () => {
    if (!selected || isActing) return;
    try {
      setIsActing(true);
      await leaveRequestsApi.reject(selected.id);
      toast.success("Pengajuan ditolak");
      const res = await leaveRequestsApi.listForManagerHR({
        status: activeStatus,
        search: searchEmployee.trim() || undefined,
      });
      const mapped = res.map(mapResponseToItem);
      setItems(mapped);
      setSelectedId(mapped[0]?.id ?? null);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Gagal menolak pengajuan");
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

  if (loadError) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <div className="text-center">
          <p className="text-sm font-semibold text-gray-900">Gagal memuat data</p>
          <p className="mt-1 text-sm text-gray-500">{loadError}</p>
          <Button className="mt-4" onClick={() => window.location.reload()}>
            Muat Ulang
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full gap-6 p-6">
      {/* LEFT: list */}
      <div className="flex-1 min-w-0">
        <Card className="h-full">
          <CardContent className="p-6 h-full flex flex-col">
            {/* Header */}
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold text-gray-900">
                  Manajemen Izin &amp; Cuti
                </h2>
              </div>
              <div className="flex items-center gap-2">
                <Button variant="outline" className="h-9 gap-2">
                  <Filter className="h-4 w-4" />
                  Filter
                </Button>
                <Button variant="outline" className="h-9 gap-2">
                  <Download className="h-4 w-4" />
                  Export
                </Button>
              </div>
            </div>

            {/* Search bar (wajib di bawah judul) */}
            <div className="mt-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
                <input
                  value={searchEmployee}
                  onChange={(e) => setSearchEmployee(e.target.value)}
                  placeholder="Cari Karyawan"
                  className="w-full rounded-xl border border-gray-200 bg-white py-2.5 pl-10 pr-4 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Table */}
            <div className="mt-5 flex-1 overflow-hidden rounded-xl border border-gray-100">
              <div className="overflow-x-auto h-full">
                <table className="w-full">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-100">
                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Nama Karyawan
                      </th>
                      <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                        Departemen
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
                                  <img src={x.avatarUrl} alt={x.employeeName} className="h-full w-full object-cover" />
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

                          <td className="px-5 py-4 text-sm text-gray-700">
                            {x.department}
                          </td>

                          <td className="px-5 py-4">
                            <Badge variant={typeBadgeVariant(x.type)}>
                              {x.type}
                            </Badge>
                          </td>

                          <td className="px-5 py-4 text-sm text-gray-700">
                            {x.startAt}
                          </td>

                          <td className="px-5 py-4 text-sm text-gray-700">
                            {x.endAt}
                          </td>

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

            {/* Footer (simple info) */}
            <div className="mt-4 text-xs text-gray-500">
              Menampilkan {filtered.length} pengajuan
            </div>
          </CardContent>
        </Card>
      </div>

      {/* RIGHT: detail */}
      <div className="w-[360px] shrink-0">
        <Card className="h-full">
          <CardContent className="p-6 h-full flex flex-col">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-gray-900">
                Detail Pengajuan
              </h3>
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
                {/* Employee card */}
                <div className="mt-4 rounded-xl border border-gray-100 p-4">
                  <div className="flex items-center gap-3">
                    <div className="h-12 w-12 rounded-xl bg-gray-100 overflow-hidden flex items-center justify-center">
                      {selected.avatarUrl ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={selected.avatarUrl} alt={selected.employeeName} className="h-full w-full object-cover" />
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
                      <div className="text-xs text-gray-500 truncate">
                        {selected.department} • {selected.position}
                      </div>
                      <div className="text-xs text-gray-400">
                        {selected.employeeId}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Dates */}
                <div className="mt-5 grid grid-cols-2 gap-4">
                  <div>
                    <div className="text-[11px] font-semibold text-gray-500 uppercase">
                      Mulai
                    </div>
                    <div className="mt-2 text-sm font-semibold text-gray-900">
                      {selected.startDateLabel}
                    </div>
                    <div className="text-xs text-gray-500">
                      {selected.startTimeLabel}
                    </div>
                  </div>
                  <div>
                    <div className="text-[11px] font-semibold text-gray-500 uppercase">
                      Selesai
                    </div>
                    <div className="mt-2 text-sm font-semibold text-gray-900">
                      {selected.endDateLabel}
                    </div>
                    <div className="text-xs text-gray-500">
                      {selected.endTimeLabel}
                    </div>
                  </div>
                </div>

                {/* Reason */}
                <div className="mt-5">
                  <div className="text-[11px] font-semibold text-gray-500 uppercase">
                    Alasan Pengajuan
                  </div>
                  <div className="mt-2 rounded-xl border border-gray-100 bg-gray-50 p-4 text-sm text-gray-700 leading-relaxed">
                    {selected.reason}
                  </div>
                </div>

                {/* Attachment */}
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

                {/* Actions */}
                <div className="mt-auto pt-6 grid grid-cols-2 gap-3">
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
