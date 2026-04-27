"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { Search, Loader2, MoreVertical, X, AlertTriangle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import toast from "react-hot-toast";

type RequestStatus = "Pending" | "Disetujui" | "Ditolak";

type OvertimeApprovalItem = {
  id: string;

  employeeName: string;
  employeeId: string; // payroll number
  department: string;
  position: string;

  dateLabel: string; // still used in detail panel
  startAt: string;
  endAt: string;
  startTimeLabel: string;
  endTimeLabel: string;

  reason: string;
  total: string;

  status: RequestStatus;

  avatarUrl?: string;
  avatarFallback: string;

  statusKepalaDepartemen?: string;
  statusManagerHR?: string;
};

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

function statusBadgeClass(s: RequestStatus) {
  switch (s) {
    case "Pending":
      return "bg-orange-100 text-orange-700 border border-orange-200";
    case "Disetujui":
      return "bg-green-100 text-green-700 border border-green-200";
    case "Ditolak":
      return "bg-red-100 text-red-700 border border-red-200";
    default:
      return "bg-gray-100 text-gray-800 border border-gray-200";
  }
}

export default function PersetujuanLemburPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [isActing, setIsActing] = useState(false);

  const [searchEmployee, setSearchEmployee] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // Modal penolakan
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [rejectionReason, setRejectionReason] = useState("");
  const [rejectError, setRejectError] = useState<string | null>(null);
  const rejectTextareaRef = useRef<HTMLTextAreaElement>(null);

  const [items, setItems] = useState<OvertimeApprovalItem[]>([]);

  useEffect(() => {
    let cancelled = false;
    const t = setTimeout(() => {
      (async () => {
        try {
          setLoadError(null);
          setIsLoading(true);

          // TODO: ganti ini dengan API list overtime requests for manager-hr
          const mock: OvertimeApprovalItem[] = [
            (() => {
              const start = new Date("2026-04-22T18:00:00+07:00");
              const end = new Date("2026-04-22T21:00:00+07:00");
              return {
                id: "ot-1",
                employeeName: "Siti Rahmawati",
                employeeId: "EMP-2023-089",
                department: "Engineering",
                position: "Senior Developer",
                dateLabel: formatDateLabel(start),
                startAt: formatListDateTime(start),
                endAt: formatListDateTime(end),
                startTimeLabel: formatTimeLabel(start),
                endTimeLabel: formatTimeLabel(end),
                reason: "Deployment & bugfix urgent untuk release malam ini.",
                total: "3 jam",
                status: "Pending",
                avatarUrl: "",
                avatarFallback: getInitials("Siti Rahmawati"),
                statusKepalaDepartemen: "APPROVED",
                statusManagerHR: "PENDING",
              };
            })(),
            (() => {
              const start = new Date("2026-04-21T19:30:00+07:00");
              const end = new Date("2026-04-21T22:00:00+07:00");
              return {
                id: "ot-2",
                employeeName: "Andi Pratama",
                employeeId: "EMP-2022-142",
                department: "Marketing",
                position: "Marketing Specialist",
                dateLabel: formatDateLabel(start),
                startAt: formatListDateTime(start),
                endAt: formatListDateTime(end),
                startTimeLabel: formatTimeLabel(start),
                endTimeLabel: formatTimeLabel(end),
                reason: "Persiapan materi campaign besok pagi.",
                total: "2.5 jam",
                status: "Disetujui",
                avatarUrl: "",
                avatarFallback: getInitials("Andi Pratama"),
                statusKepalaDepartemen: "APPROVED",
                statusManagerHR: "APPROVED",
              };
            })(),
            (() => {
              const start = new Date("2026-04-20T18:00:00+07:00");
              const end = new Date("2026-04-20T20:00:00+07:00");
              return {
                id: "ot-3",
                employeeName: "Budi Wijaya",
                employeeId: "EMP-2021-085",
                department: "Sales",
                position: "Sales Manager",
                dateLabel: formatDateLabel(start),
                startAt: formatListDateTime(start),
                endAt: formatListDateTime(end),
                startTimeLabel: formatTimeLabel(start),
                endTimeLabel: formatTimeLabel(end),
                reason: "Follow up klien untuk closing deal.",
                total: "2 jam",
                status: "Ditolak",
                avatarUrl: "",
                avatarFallback: getInitials("Budi Wijaya"),
                statusKepalaDepartemen: "REJECTED",
                statusManagerHR: "PENDING",
              };
            })(),
          ];

          const q = searchEmployee.trim().toLowerCase();
          const filtered = q ? mock.filter((x) => x.employeeName.toLowerCase().includes(q)) : mock;

          if (cancelled) return;
          setItems(filtered);
          setSelectedId((prev) => {
            if (prev && filtered.some((x) => x.id === prev)) return prev;
            return filtered[0]?.id ?? null;
          });
        } catch (e) {
          if (cancelled) return;
          setLoadError(e instanceof Error ? e.message : "Gagal memuat pengajuan lembur");
        } finally {
          if (!cancelled) setIsLoading(false);
        }
      })();
    }, 250);

    return () => {
      cancelled = true;
      clearTimeout(t);
    };
  }, [searchEmployee]);

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
      // TODO: call API approve
      await new Promise((r) => setTimeout(r, 600));
      toast.success("Pengajuan lembur disetujui");
      setItems((prev) => prev.map((x) => (x.id === selected.id ? { ...x, status: "Disetujui" } : x)));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Gagal menyetujui pengajuan lembur");
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
      // TODO: call API reject + reason
      await new Promise((r) => setTimeout(r, 600));
      toast.success("Pengajuan lembur ditolak");
      closeRejectModal();
      setItems((prev) => prev.map((x) => (x.id === selected.id ? { ...x, status: "Ditolak" } : x)));
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Gagal menolak pengajuan lembur");
    } finally {
      setIsActing(false);
    }
  };

  const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) closeRejectModal();
  };

  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <div className="text-center">
          <Loader2 className="mx-auto h-8 w-8 animate-spin text-indigo-600" />
          <p className="mt-2 text-sm text-gray-500">Memuat pengajuan lembur...</p>
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
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">Persetujuan Lembur</h2>
                </div>
              </div>

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

              <div className="mt-5 flex-1 overflow-hidden rounded-xl border border-gray-100">
                <div className="overflow-x-auto h-full">
                  <table className="w-full">
                    <thead>
                      <tr className="bg-gray-50 border-b border-gray-100">
                        <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                          Nama Karyawan
                        </th>
                        <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                          Jabatan
                        </th>
                        <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                          Mulai
                        </th>
                        <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                          Selesai
                        </th>
                        <th className="px-5 py-4 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                          Total
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
                                    <span className="text-xs font-semibold text-gray-700">{x.avatarFallback}</span>
                                  )}
                                </div>
                                <div>
                                  <div className="font-semibold text-gray-900">{x.employeeName}</div>
                                  <div className="text-xs text-gray-500">ID: {x.employeeId}</div>
                                </div>
                              </div>
                            </td>

                            {/* ✅ Departemen (bold) + Jabatan (di bawah) */}
                            <td className="px-5 py-4">
                              <div className="text-sm font-semibold text-gray-900">{x.department}</div>
                              <div className="text-xs text-gray-500">{x.position}</div>
                            </td>

                            <td className="px-5 py-4 text-sm text-gray-700">{x.startAt}</td>
                            <td className="px-5 py-4 text-sm text-gray-700">{x.endAt}</td>

                            <td className="px-5 py-4">
                              <Badge variant="secondary" className="bg-gray-100 text-gray-800 border border-gray-200">
                                {x.total}
                              </Badge>
                            </td>

                            <td className="px-5 py-4">
                              <div className="flex items-center gap-2 text-sm text-gray-700">
                                <span className={`h-2 w-2 rounded-full ${statusDotColor(x.status)}`} />
                                <Badge variant="secondary" className={statusBadgeClass(x.status)}>
                                  {x.status}
                                </Badge>
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>

                  {filtered.length === 0 && (
                    <div className="p-10 text-center text-sm text-gray-500">
                      Tidak ada pengajuan lembur ditemukan.
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

        {/* RIGHT: detail */}
        <div className="w-[360px] shrink-0">
          <Card className="h-full">
            <CardContent className="p-6 h-full flex flex-col">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-gray-900">Detail Pengajuan Lembur</h3>
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
                          <img src={selected.avatarUrl} alt={selected.employeeName} className="h-full w-full object-cover" />
                        ) : (
                          <span className="text-sm font-semibold text-gray-700">{selected.avatarFallback}</span>
                        )}
                      </div>
                      <div className="min-w-0">
                        <div className="font-semibold text-gray-900 truncate">{selected.employeeName}</div>
                        <div className="text-xs text-gray-500 truncate">
                          {selected.department} • {selected.position}
                        </div>
                        <div className="text-xs text-gray-400">{selected.employeeId}</div>
                      </div>
                    </div>
                  </div>

                  <div className="mt-5 grid grid-cols-2 gap-4">
                    <div>
                      <div className="text-[11px] font-semibold text-gray-500 uppercase">Mulai</div>
                      <div className="mt-2 text-sm font-semibold text-gray-900">{selected.dateLabel}</div>
                      <div className="text-xs text-gray-500">{selected.startTimeLabel}</div>
                    </div>
                    <div>
                      <div className="text-[11px] font-semibold text-gray-500 uppercase">Selesai</div>
                      <div className="mt-2 text-sm font-semibold text-gray-900">{selected.dateLabel}</div>
                      <div className="text-xs text-gray-500">{selected.endTimeLabel}</div>
                    </div>
                  </div>

                  <div className="mt-5">
                    <div className="text-[11px] font-semibold text-gray-500 uppercase">Total Lembur</div>
                    <div className="mt-2">
                      <Badge variant="secondary" className="bg-gray-100 text-gray-800 border border-gray-200">
                        {selected.total}
                      </Badge>
                    </div>
                  </div>

                  <div className="mt-5">
                    <div className="text-[11px] font-semibold text-gray-500 uppercase">Alasan Lembur</div>
                    <div className="mt-2 rounded-xl border border-gray-100 bg-gray-50 p-4 text-sm text-gray-700 leading-relaxed">
                      {selected.reason}
                    </div>
                  </div>

                  <div className="mt-5">
                    <div className="text-[11px] font-semibold text-gray-500 uppercase">Status</div>
                    <div className="mt-2 flex items-center gap-2">
                      <span className={`h-2 w-2 rounded-full ${statusDotColor(selected.status)}`} />
                      <Badge variant="secondary" className={statusBadgeClass(selected.status)}>
                        {selected.status}
                      </Badge>
                    </div>
                  </div>

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

      {showRejectModal && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
          onClick={handleOverlayClick}
        >
          <div className="relative w-full max-w-md mx-4 bg-white rounded-2xl shadow-xl p-6 animate-in fade-in zoom-in-95 duration-200">
            <button
              onClick={closeRejectModal}
              className="absolute top-4 right-4 p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
            >
              <X className="h-4 w-4" />
            </button>

            <div className="flex items-center gap-3 mb-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-red-100 shrink-0">
                <AlertTriangle className="h-5 w-5 text-red-600" />
              </div>
              <div>
                <h3 className="text-base font-semibold text-gray-900">Tolak Pengajuan Lembur</h3>
                <p className="text-xs text-gray-500 mt-0.5">
                  Pengajuan atas nama{" "}
                  <span className="font-medium text-gray-700">{selected?.employeeName}</span>
                </p>
              </div>
            </div>

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
                placeholder="Tuliskan alasan penolakan pengajuan lembur ini..."
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