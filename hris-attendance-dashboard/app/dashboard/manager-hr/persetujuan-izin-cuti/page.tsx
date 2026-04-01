"use client";

import { useEffect, useMemo, useState } from "react";
import { Search, Filter, Download, Loader2, MoreVertical } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

type RequestStatus = "Pending" | "Disetujui" | "Ditolak";

type RequestType = "SAKIT" | "TAHUNAN" | "IZIN KHUSUS";

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

export default function PersetujuanIzinCutiPage() {
  const [isLoading, setIsLoading] = useState(true);
  const [searchEmployee, setSearchEmployee] = useState("");
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // Dummy data (ganti nanti dengan API)
  const [items, setItems] = useState<LeaveApprovalItem[]>([]);

  useEffect(() => {
    // Simulasi fetch
    const t = setTimeout(() => {
      const mock: LeaveApprovalItem[] = [
        {
          id: "req-1",
          employeeName: "Adinda Larasati",
          employeeId: "SI-2024-089",
          department: "Marketing",
          position: "Marketing Coordinator",
          type: "SAKIT",
          startAt: "22 Okt 2023, 08:00",
          endAt: "24 Okt 2023, 17:00",
          startDateLabel: "22 Okt 2023",
          startTimeLabel: "Pukul 08:00 WIB",
          endDateLabel: "24 Okt 2023",
          endTimeLabel: "Pukul 17:00 WIB",
          reason:
            "Saya merasa kurang enak badan sejak semalam. Berdasarkan pemeriksaan dokter, saya memerlukan istirahat total selama 3 hari karena gejala flu berat dan demam tinggi.",
          attachmentName: "Surat_Dokter_Adinda_220ct.pdf",
          attachmentSize: "1.2 MB",
          status: "Pending",
          avatarUrl: "",
          avatarFallback: "AL",
        },
        {
          id: "req-2",
          employeeName: "Bagus Pranogo",
          employeeId: "SI-2024-112",
          department: "Engineering",
          position: "Backend Engineer",
          type: "TAHUNAN",
          startAt: "25 Okt 2023, 09:00",
          endAt: "27 Okt 2023, 17:00",
          startDateLabel: "25 Okt 2023",
          startTimeLabel: "Pukul 09:00 WIB",
          endDateLabel: "27 Okt 2023",
          endTimeLabel: "Pukul 17:00 WIB",
          reason: "Mengambil cuti tahunan untuk keperluan keluarga.",
          status: "Pending",
          avatarFallback: "BP",
        },
        {
          id: "req-3",
          employeeName: "Citra Kirana",
          employeeId: "SI-2024-045",
          department: "Human Resources",
          position: "HR Staff",
          type: "IZIN KHUSUS",
          startAt: "23 Okt 2023, 13:00",
          endAt: "23 Okt 2023, 17:00",
          startDateLabel: "23 Okt 2023",
          startTimeLabel: "Pukul 13:00 WIB",
          endDateLabel: "23 Okt 2023",
          endTimeLabel: "Pukul 17:00 WIB",
          reason: "Izin khusus untuk keperluan administrasi pribadi.",
          status: "Pending",
          avatarFallback: "CK",
        },
      ];

      setItems(mock);
      setSelectedId(mock[0]?.id ?? null);
      setIsLoading(false);
    }, 650);

    return () => clearTimeout(t);
  }, []);

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

  const handleApprove = () => {
    if (!selected) return;
    setItems((prev) =>
      prev.map((x) =>
        x.id === selected.id ? { ...x, status: "Disetujui" } : x
      )
    );
  };

  const handleReject = () => {
    if (!selected) return;
    setItems((prev) =>
      prev.map((x) => (x.id === selected.id ? { ...x, status: "Ditolak" } : x))
    );
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
                    disabled={selected.status !== "Pending"}
                  >
                    Tolak
                  </Button>
                  <Button
                    className="bg-blue-600 hover:bg-blue-700 text-white"
                    onClick={handleApprove}
                    disabled={selected.status !== "Pending"}
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