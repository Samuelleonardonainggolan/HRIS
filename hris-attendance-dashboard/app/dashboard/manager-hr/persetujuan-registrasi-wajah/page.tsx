"use client";

import { useMemo, useState } from "react";
import { Search, Eye, Check, X } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

type Status = "Menunggu" | "Disetujui" | "Ditolak";

type FaceRequestRow = {
  id: string;

  fullName: string;
  payrollNumber: string;

  departmentName: string;
  positionName: string;

  submittedAt: string; // display "12 Okt 2023"
  submittedAtTime: string; // "09:41"

  status: Status;

  email: string;
  faceImageUrl: string; // preview image
};

function StatusBadge({ status }: { status: Status }) {
  if (status === "Menunggu") {
    return (
      <Badge className="rounded-lg bg-amber-50 text-amber-700 border border-amber-200">
        Menunggu
      </Badge>
    );
  }
  if (status === "Disetujui") {
    return (
      <Badge className="rounded-lg bg-emerald-50 text-emerald-700 border border-emerald-200">
        Disetujui
      </Badge>
    );
  }
  return (
    <Badge className="rounded-lg bg-rose-50 text-rose-700 border border-rose-200">
      Ditolak
    </Badge>
  );
}

function Avatar({ name }: { name: string }) {
  const initials = useMemo(() => {
    const parts = name.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return "U";
    if (parts.length === 1) return parts[0][0].toUpperCase();
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }, [name]);

  return (
    <div className="h-9 w-9 rounded-full bg-gray-100 flex items-center justify-center text-xs font-semibold text-gray-700">
      {initials}
    </div>
  );
}

function ImagePreviewModal({
  open,
  title,
  subtitle,
  imageUrl,
  onClose,
}: {
  open: boolean;
  title: string;
  subtitle?: string;
  imageUrl: string;
  onClose: () => void;
}) {
  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50"
      aria-modal="true"
      role="dialog"
      onMouseDown={(e) => {
        // close on backdrop click
        if (e.target === e.currentTarget) onClose();
      }}
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/40" />

      {/* Modal */}
      <div className="relative h-full w-full flex items-center justify-center p-4">
        <Card className="w-full max-w-2xl rounded-2xl shadow-xl">
          <CardContent className="p-0">
            <div className="px-5 py-4 border-b border-gray-200 flex items-start justify-between gap-4">
              <div>
                <div className="text-base font-semibold text-gray-900">{title}</div>
                {subtitle && <div className="text-sm text-gray-600">{subtitle}</div>}
              </div>
              <Button
                variant="outline"
                className="rounded-xl"
                onClick={onClose}
                aria-label="Tutup"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="p-5">
              <div className="overflow-hidden rounded-xl border border-gray-200 bg-gray-50">
                <img
                  src={imageUrl}
                  alt="Registrasi wajah"
                  className="w-full max-h-[70vh] object-contain"
                />
              </div>

              <div className="mt-4 flex justify-end">
                <Button className="rounded-xl" variant="outline" onClick={onClose}>
                  Tutup
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function PersetujuanRegistrasiWajahPage() {
  const [q, setQ] = useState("");
  const [dept, setDept] = useState("Semua Departemen");
  const [status, setStatus] = useState<Status | "Semua Status">("Semua Status");

  // ✅ state modal
  const [openModal, setOpenModal] = useState(false);
  const [modalRowId, setModalRowId] = useState<string>("");

  // mock data (ganti ke fetch API nanti)
  const rows: FaceRequestRow[] = useMemo(
    () => [
      {
        id: "req1",
        fullName: "Siti Rahmawati",
        payrollNumber: "EMP-2023-089",
        departmentName: "Engineering",
        positionName: "Senior Developer",
        submittedAt: "12 Okt 2023",
        submittedAtTime: "09:41",
        status: "Menunggu",
        email: "siti.rahmawati@sapphire.id",
        faceImageUrl: "https://picsum.photos/seed/face-1/800/800",
      },
      {
        id: "req2",
        fullName: "Andi Pratama",
        payrollNumber: "EMP-2022-142",
        departmentName: "Marketing",
        positionName: "Marketing Specialist",
        submittedAt: "11 Okt 2023",
        submittedAtTime: "10:05",
        status: "Disetujui",
        email: "andi.pratama@sapphire.id",
        faceImageUrl: "https://picsum.photos/seed/face-2/800/800",
      },
      {
        id: "req3",
        fullName: "Budi Wijaya",
        payrollNumber: "EMP-2021-085",
        departmentName: "Sales",
        positionName: "Sales Manager",
        submittedAt: "10 Okt 2023",
        submittedAtTime: "15:20",
        status: "Ditolak",
        email: "budi.wijaya@sapphire.id",
        faceImageUrl: "https://picsum.photos/seed/face-3/800/800",
      },
    ],
    []
  );

  const departments = useMemo(() => {
    const uniq = Array.from(new Set(rows.map((r) => r.departmentName))).sort();
    return ["Semua Departemen", ...uniq];
  }, [rows]);

  const filtered = useMemo(() => {
    const qq = q.trim().toLowerCase();

    return rows.filter((r) => {
      const matchQ =
        !qq ||
        r.fullName.toLowerCase().includes(qq) ||
        r.payrollNumber.toLowerCase().includes(qq) ||
        r.departmentName.toLowerCase().includes(qq) ||
        r.positionName.toLowerCase().includes(qq);

      const matchDept = dept === "Semua Departemen" || r.departmentName === dept;
      const matchStatus = status === "Semua Status" || r.status === status;

      return matchQ && matchDept && matchStatus;
    });
  }, [rows, q, dept, status]);

  const [selectedId, setSelectedId] = useState<string>(() => rows[0]?.id ?? "");
  const selected = useMemo(
    () => rows.find((r) => r.id === selectedId) ?? null,
    [rows, selectedId]
  );

  const modalRow = useMemo(
    () => rows.find((r) => r.id === modalRowId) ?? null,
    [rows, modalRowId]
  );

  function onApprove() {
    alert("Setujui (mock)");
  }

  function onReject() {
    alert("Tolak (mock)");
  }

  function openDetailModal(rowId: string) {
    setModalRowId(rowId);
    setOpenModal(true);
  }

  return (
    <div className="p-6">
      {/* ✅ Modal preview gambar registrasi wajah */}
      {modalRow && (
        <ImagePreviewModal
          open={openModal}
          onClose={() => setOpenModal(false)}
          title="Detail Registrasi Wajah"
          subtitle={`${modalRow.fullName} • ${modalRow.payrollNumber}`}
          imageUrl={modalRow.faceImageUrl}
        />
      )}

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* LEFT */}
        <Card className="lg:col-span-8 rounded-2xl">
          <CardContent className="p-5 space-y-4">
            <div>
              <h1 className="text-lg font-bold text-gray-900">
                Persetujuan Registrasi Wajah
              </h1>
              <p className="text-sm text-gray-600">
                Kelola permintaan registrasi wajah karyawan
              </p>
            </div>

            {/* toolbar */}
            <div className="flex flex-col md:flex-row md:items-center gap-3">
              <div className="relative w-full md:max-w-md">
                <Search className="h-4 w-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
                <Input
                  value={q}
                  onChange={(e) => setQ(e.target.value)}
                  placeholder="Cari nama, ID, atau jabatan..."
                  className="pl-9 rounded-xl"
                />
              </div>

              <select
                value={dept}
                onChange={(e) => setDept(e.target.value)}
                className="h-10 rounded-xl border border-gray-200 bg-white px-3 text-sm text-gray-700"
              >
                {departments.map((d) => (
                  <option key={d} value={d}>
                    {d}
                  </option>
                ))}
              </select>

              <select
                value={status}
                onChange={(e) => setStatus(e.target.value as any)}
                className="h-10 rounded-xl border border-gray-200 bg-white px-3 text-sm text-gray-700"
              >
                <option value="Semua Status">Semua Status</option>
                <option value="Menunggu">Menunggu</option>
                <option value="Disetujui">Disetujui</option>
                <option value="Ditolak">Ditolak</option>
              </select>
            </div>

            {/* table */}
            <div className="overflow-hidden rounded-xl border border-gray-100">
              <table className="w-full">
                <thead>
                  <tr className="bg-gray-50 border-b border-gray-100">
                    <th className="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Karyawan
                    </th>
                    <th className="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Departemen
                    </th>
                    <th className="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Tanggal Pengajuan
                    </th>
                    <th className="px-5 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Status
                    </th>
                    <th className="px-5 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Aksi
                    </th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-gray-100 bg-white">
                  {filtered.map((r) => {
                    const active = r.id === selectedId;
                    return (
                      <tr
                        key={r.id}
                        className={[
                          "cursor-pointer transition-colors",
                          active ? "bg-blue-50/40" : "hover:bg-gray-50",
                        ].join(" ")}
                        onClick={() => setSelectedId(r.id)}
                      >
                        {/* ✅ Nama + payroll di bawah */}
                        <td className="px-5 py-4">
                          <div className="flex items-center gap-3">
                            <Avatar name={r.fullName} />
                            <div>
                              <div className="text-sm font-semibold text-gray-900">
                                {r.fullName}
                              </div>
                              <div className="text-xs text-gray-500">
                                {r.payrollNumber}
                              </div>
                            </div>
                          </div>
                        </td>

                        {/* ✅ Departemen + jabatan di bawah */}
                        <td className="px-5 py-4">
                          <div className="text-sm font-semibold text-gray-900">
                            {r.departmentName}
                          </div>
                          <div className="text-xs text-gray-500">
                            {r.positionName}
                          </div>
                        </td>

                        <td className="px-5 py-4 text-sm text-gray-700">
                          {r.submittedAt}
                        </td>

                        <td className="px-5 py-4">
                          <StatusBadge status={r.status} />
                        </td>

                        <td className="px-5 py-4">
                          <div className="flex items-center justify-end gap-2">
                            <Button
                              variant="outline"
                              className="rounded-xl"
                              onClick={(e) => {
                                e.stopPropagation();
                                setSelectedId(r.id);
                                openDetailModal(r.id); // ✅ buka modal
                              }}
                              title="Lihat detail"
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}

                  {filtered.length === 0 && (
                    <tr>
                      <td
                        colSpan={5}
                        className="px-6 py-10 text-center text-sm text-gray-500"
                      >
                        Tidak ada data.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>

        {/* RIGHT */}
        <Card className="lg:col-span-4 rounded-2xl">
          <CardContent className="p-5">
            <div className="text-xs font-semibold tracking-wide text-gray-500 uppercase">
              Detail Karyawan
            </div>

            {!selected ? (
              <div className="mt-4 text-sm text-gray-500">Pilih data di tabel.</div>
            ) : (
              <div className="mt-4 space-y-4">
                <div className="flex items-start gap-3">
                  <div className="h-12 w-12 rounded-xl overflow-hidden bg-gray-100">
                    <img
                      src={selected.faceImageUrl}
                      alt="preview"
                      className="h-12 w-12 object-cover"
                    />
                  </div>
                  <div>
                    <div className="font-semibold text-gray-900">
                      {selected.fullName}
                    </div>
                    <div className="text-sm text-gray-600">
                      {selected.positionName}
                    </div>
                    <div className="text-xs text-gray-500">
                      {selected.departmentName}
                    </div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <div className="text-xs text-gray-500">No Payroll</div>
                    <div className="font-medium text-gray-900">
                      {selected.payrollNumber}
                    </div>
                  </div>
                  <div>
                    <div className="text-xs text-gray-500">Tanggal Pengajuan</div>
                    <div className="font-medium text-gray-900">
                      {selected.submittedAt}, {selected.submittedAtTime}
                    </div>
                  </div>
                  <div className="col-span-2">
                    <div className="text-xs text-gray-500">Email</div>
                    <div className="font-medium text-gray-900">{selected.email}</div>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="text-sm font-semibold text-gray-900">
                      Verifikasi Wajah
                    </div>
                    <Badge className="rounded-lg bg-amber-50 text-amber-700 border border-amber-200">
                      Menunggu Audit
                    </Badge>
                  </div>

                  <div className="overflow-hidden rounded-xl border border-gray-200 bg-gray-50">
                    <img
                      src={selected.faceImageUrl}
                      alt="face"
                      className="w-full aspect-square object-cover"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-2">
                    <Button
                      className="rounded-xl bg-emerald-600 hover:bg-emerald-700 text-white gap-2"
                      onClick={onApprove}
                      disabled={selected.status !== "Menunggu"}
                    >
                      <Check className="h-4 w-4" />
                      Setujui
                    </Button>
                    <Button
                      variant="outline"
                      className="rounded-xl border-rose-200 text-rose-700 hover:bg-rose-50 gap-2"
                      onClick={onReject}
                      disabled={selected.status !== "Menunggu"}
                    >
                      <X className="h-4 w-4" />
                      Tolak
                    </Button>
                  </div>

                  <Button variant="ghost" className="w-full rounded-xl text-gray-600">
                    Minta Unggah Ulang
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}