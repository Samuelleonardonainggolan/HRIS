"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Search, Loader2, MoreVertical, X, AlertTriangle, FileText, ExternalLink, ZoomIn } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
  rejectionReason?: string;
  attachmentName?: string;
  attachmentSize?: string;
  attachmentUrl?: string;

  status: RequestStatus;

  avatarUrl?: string; // optional
  avatarFallback: string; // initials
}

function typeBadgeClass(t: RequestType) {
  switch (t.toUpperCase()) {
    case "SAKIT":
      return "bg-red-100 text-red-700 border border-red-200";
    case "TAHUNAN":
      return "bg-gray-100 text-gray-800 border border-gray-200";
    case "IZIN KHUSUS":
      return "bg-blue-100 text-blue-700 border border-blue-200";
    case "CUTI KHUSUS":
      return "bg-orange-100 text-orange-800 border border-orange-200";
    default:
      return "bg-gray-100 text-gray-800 border border-gray-200";
  }
}

function mapStatus(finalStatus: LeaveRequestStatus | undefined, stageStatus: LeaveRequestStatus | undefined): RequestStatus {
  switch (finalStatus) {
    case "APPROVED":
      return "Disetujui";
    case "REJECTED":
      return "Ditolak";
  }
  switch (stageStatus) {
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
  const start = new Date(x.pengajuan.start_date ?? x.pengajuan.tanggal_mulai ?? "");
  const end = new Date(x.pengajuan.end_date ?? x.pengajuan.tanggal_selesai ?? "");
  const attachmentName = getFileNameFromUrl(x.pengajuan.document_url ?? x.pengajuan.dokumen_url);

  return {
    id: x.pengajuan.id,
    employeeName,
    employeeId,
    department,
    position,
    type: (x.pengajuan.type_name || x.pengajuan.nama_tipe || "IZIN KHUSUS").toUpperCase(),
    startAt: formatListDateTime(start),
    endAt: formatListDateTime(end),
    startDateLabel: formatDateLabel(start),
    startTimeLabel: formatTimeLabel(start),
    endDateLabel: formatDateLabel(end),
    endTimeLabel: formatTimeLabel(end),
    reason: x.pengajuan.reason || x.pengajuan.alasan || "-",
    rejectionReason: x.pengajuan.rejection_reason_manager_hr || undefined,
    attachmentName,
    attachmentSize: undefined,
    attachmentUrl: x.pengajuan.document_url ?? x.pengajuan.dokumen_url ?? undefined,
    status: mapStatus(
      x.pengajuan.final_status ?? x.pengajuan.status_final,
      x.pengajuan.status_manager_hr
    ),
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

  // Modal penolakan
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [rejectionReason, setRejectionReason] = useState("");
  const [rejectError, setRejectError] = useState<string | null>(null);
  const rejectTextareaRef = useRef<HTMLTextAreaElement>(null);

  // Modal preview dokumen
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  const activeStatus = "ALL" as const;

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

  const openRejectModal = () => {
    if (!selected || isActing) return;
    setRejectionReason("");
    setRejectError(null);
    setShowRejectModal(true);
    setTimeout(() => rejectTextareaRef.current?.focus(), 100);
  };

  const closeRejectModal = () => {
    setShowRejectModal(false);
    setRejectionReason("");
    setRejectError(null);
  };

  const handleReject = async () => {
    if (!selected || isActing) return;
    const trimmed = rejectionReason.trim();
    if (!trimmed) {
      setRejectError("Alasan penolakan wajib diisi.");
      rejectTextareaRef.current?.focus();
      return;
    }
    try {
      setIsActing(true);
      await leaveRequestsApi.reject(selected.id, trimmed);
      toast.success("Pengajuan ditolak");
      closeRejectModal();
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

  // Tutup modal jika klik di luar overlay
  const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) closeRejectModal();
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
    <>
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
                            <Badge variant="secondary" className={typeBadgeClass(x.type)}>
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

                {/* Rejection Reason - tampil hanya jika ditolak */}
                {selected.status === "Ditolak" && selected.rejectionReason && (
                  <div className="mt-5">
                    <div className="text-[11px] font-semibold text-red-500 uppercase flex items-center gap-1">
                      <span>Alasan Penolakan</span>
                    </div>
                    <div className="mt-2 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 leading-relaxed">
                      {selected.rejectionReason}
                    </div>
                  </div>
                )}

                {/* Dokumen Pendukung */}
                <div className="mt-5">
                  <div className="text-[11px] font-semibold text-gray-500 uppercase">
                    Dokumen Pendukung
                  </div>
                  {selected.attachmentUrl ? (
                    <div
                      className="mt-2 rounded-xl border border-gray-100 overflow-hidden cursor-pointer group"
                      onClick={() => setPreviewUrl(selected.attachmentUrl!)}
                    >
                      {/* Thumbnail preview */}
                      <div className="relative h-36 bg-gray-50 flex items-center justify-center overflow-hidden">
                        {/\.(png|jpe?g|gif|webp|bmp|svg)($|\?)/i.test(selected.attachmentUrl) ? (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img
                            src={selected.attachmentUrl}
                            alt={selected.attachmentName ?? "Dokumen"}
                            className="h-full w-full object-cover transition-transform duration-200 group-hover:scale-105"
                          />
                        ) : /\.pdf($|\?)/i.test(selected.attachmentUrl) ? (
                          <div className="flex flex-col items-center gap-2 text-gray-400">
                            <div className="flex items-center justify-center bg-red-100 text-red-600 font-bold text-xs px-2 py-1 rounded">PDF</div>
                            <span className="text-xs">Klik untuk pratinjau</span>
                          </div>
                        ) : (
                          <div className="flex flex-col items-center gap-2 text-gray-400">
                            <FileText className="h-10 w-10" />
                            <span className="text-xs">Klik untuk pratinjau</span>
                          </div>
                        )}
                        {/* Overlay hover */}
                        <div className="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors duration-200 flex items-center justify-center">
                          <ZoomIn className="h-7 w-7 text-white opacity-0 group-hover:opacity-100 transition-opacity duration-200" />
                        </div>
                      </div>
                      {/* Footer info */}
                      <div className="p-3 flex items-center justify-between bg-white border-t border-gray-100">
                        <div className="flex items-center gap-2 min-w-0">
                          <FileText className="h-3.5 w-3.5 text-gray-400 shrink-0" />
                          <span className="text-xs text-gray-600 truncate">{selected.attachmentName}</span>
                        </div>
                        <ExternalLink className="h-3.5 w-3.5 text-gray-400 shrink-0 ml-2" />
                      </div>
                    </div>
                  ) : (
                    <div className="mt-2 rounded-xl border border-gray-100 bg-gray-50 p-6 flex flex-col items-center gap-2 text-gray-400">
                      <FileText className="h-8 w-8" />
                      <span className="text-xs">Tidak ada dokumen pendukung</span>
                    </div>
                  )}
                </div>

                {/* Actions */}
                <div className="mt-auto pt-6 grid grid-cols-2 gap-3">
                  <Button
                    variant="outline"
                    className="border-red-200 text-red-600 hover:bg-red-50"
                    onClick={openRejectModal}
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

    {/* ===== MODAL ALASAN PENOLAKAN ===== */}
    {/* ===== MODAL PREVIEW DOKUMEN ===== */}
    {previewUrl && (
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4"
        onClick={() => setPreviewUrl(null)}
      >
        <div
          className="relative max-w-4xl w-full max-h-[90vh] flex flex-col bg-white rounded-2xl overflow-hidden shadow-2xl"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-3 border-b border-gray-100 bg-white shrink-0">
            <div className="flex items-center gap-2 min-w-0">
              <FileText className="h-4 w-4 text-gray-500 shrink-0" />
              <span className="text-sm font-medium text-gray-700 truncate">
                {selected?.attachmentName ?? "Dokumen Pendukung"}
              </span>
            </div>
            <div className="flex items-center gap-2 shrink-0 ml-3">
              <a
                href={previewUrl}
                target="_blank"
                rel="noopener noreferrer"
                onClick={(e) => e.stopPropagation()}
                className="flex items-center gap-1.5 text-xs text-blue-600 hover:text-blue-700 border border-blue-200 hover:bg-blue-50 rounded-lg px-3 py-1.5 transition-colors"
              >
                <ExternalLink className="h-3.5 w-3.5" />
                Buka di tab baru
              </a>
              <button
                onClick={() => setPreviewUrl(null)}
                className="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-auto bg-gray-50 flex items-center justify-center min-h-[400px]">
            {/\.(png|jpe?g|gif|webp|bmp|svg)($|\?)/i.test(previewUrl) ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={previewUrl}
                alt="Dokumen Pendukung"
                className="max-w-full max-h-[75vh] object-contain p-4"
              />
            ) : /\.pdf($|\?)/i.test(previewUrl) ? (
              <iframe
                src={previewUrl}
                className="w-full h-[75vh]"
                title="Dokumen Pendukung"
              />
            ) : (
              <div className="flex flex-col items-center gap-4 p-10 text-center">
                <FileText className="h-16 w-16 text-gray-300" />
                <p className="text-sm text-gray-500">
                  Pratinjau tidak tersedia untuk format file ini.
                </p>
                <a
                  href={previewUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-2 text-sm text-blue-600 hover:text-blue-700 underline"
                >
                  <ExternalLink className="h-4 w-4" />
                  Unduh / Buka file
                </a>
              </div>
            )}
          </div>
        </div>
      </div>
    )}

    {showRejectModal && (
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
        onClick={handleOverlayClick}
      >
        <div className="relative w-full max-w-md mx-4 bg-white rounded-2xl shadow-xl p-6 animate-in fade-in zoom-in-95 duration-200">
          {/* Close button */}
          <button
            onClick={closeRejectModal}
            className="absolute top-4 right-4 p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
          >
            <X className="h-4 w-4" />
          </button>

          {/* Icon + Judul */}
          <div className="flex items-center gap-3 mb-4">
            <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-100 shrink-0">
              <AlertTriangle className="h-5 w-5 text-red-600" />
            </div>
            <div>
              <h3 className="text-base font-semibold text-gray-900">Tolak Pengajuan</h3>
              <p className="text-xs text-gray-500 mt-0.5">
                Pengajuan atas nama <span className="font-medium text-gray-700">{selected?.employeeName}</span>
              </p>
            </div>
          </div>

          {/* Form alasan */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1.5">
              Alasan Penolakan <span className="text-red-500">*</span>
            </label>
            <textarea
              ref={rejectTextareaRef}
              value={rejectionReason}
              onChange={(e) => {
                setRejectionReason(e.target.value);
                if (rejectError) setRejectError(null);
              }}
              placeholder="Tuliskan alasan penolakan pengajuan izin/cuti ini..."
              rows={4}
              className={[
                "w-full rounded-xl border px-4 py-3 text-sm text-gray-700 leading-relaxed resize-none",
                "focus:outline-none focus:ring-2 focus:ring-red-400 focus:border-transparent",
                "placeholder:text-gray-400 transition-colors",
                rejectError ? "border-red-400 bg-red-50" : "border-gray-200 bg-white",
              ].join(" ")}
            />
            {rejectError && (
              <p className="mt-1.5 text-xs text-red-600 flex items-center gap-1">
                <AlertTriangle className="h-3 w-3" />
                {rejectError}
              </p>
            )}
            <p className="mt-1.5 text-xs text-gray-400">
              {rejectionReason.trim().length} karakter (minimal 5 karakter)
            </p>
          </div>

          {/* Tombol aksi */}
          <div className="flex gap-3 justify-end">
            <Button
              variant="outline"
              onClick={closeRejectModal}
              disabled={isActing}
              className="border-gray-200 text-gray-700 hover:bg-gray-50"
            >
              Batal
            </Button>
            <Button
              onClick={handleReject}
              disabled={isActing || rejectionReason.trim().length < 5}
              className="bg-red-600 hover:bg-red-700 text-white disabled:opacity-50"
            >
              {isActing ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Memproses...
                </span>
              ) : (
                "Tolak Pengajuan"
              )}
            </Button>
          </div>
        </div>
      </div>
    )}
    </>
  );
}
