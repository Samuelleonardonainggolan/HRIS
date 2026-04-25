"use client";

import { useEffect, useMemo, useState } from "react";
import { Search, Eye, X } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

type ApiSuccess<T> = {
  success?: boolean;
  message?: string;
  data: T;
};

type FaceEmbeddingApprovalItem = {
  id: string;
  user_id: string;
  face_image_url?: string;

  full_name?: string;
  payroll_number?: string;
  department_name?: string;
  position_name?: string;
  email?: string;

  registered_at?: string; // ISO
  updated_at?: string; // ISO
};

type FaceRequestRow = {
  id: string;

  fullName: string;
  payrollNumber: string;

  departmentName: string;
  positionName: string;

  submittedAt: string;
  submittedAtTime: string;

  email: string;
  faceImageUrl: string;
};

function getToken() {
  return (
    localStorage.getItem("access_token") ||
    localStorage.getItem("token") ||
    localStorage.getItem("auth_token") ||
    ""
  );
}

function joinUrl(base: string, path: string) {
  const b = base.replace(/\/+$/, "");
  const p = path.replace(/^\/+/, "");
  return `${b}/${p}`;
}

// same helper style you used before
function apiUrl(base: string, endpoint: string) {
  const normalized = base.replace(/\/+$/, "");
  const hasV1 = /\/api\/v1$/.test(normalized);
  return hasV1 ? joinUrl(normalized, endpoint) : joinUrl(normalized, `api/v1/${endpoint}`);
}

async function readJsonSafe<T>(res: Response): Promise<T> {
  const text = await res.text();
  try {
    return (text ? JSON.parse(text) : null) as T;
  } catch {
    const snippet = text.replace(/\s+/g, " ").slice(0, 300);
    throw new Error(`Response bukan JSON valid. Status=${res.status}. Body="${snippet}"`);
  }
}

function resolveImageUrl(baseApiUrl: string, faceImageUrl?: string) {
  if (!faceImageUrl) return "";
  // already absolute
  if (/^https?:\/\//i.test(faceImageUrl)) return faceImageUrl;

  // if backend returns "/uploads/xxx.jpg", join with BASE_URL
  return joinUrl(baseApiUrl, faceImageUrl);
}

function formatDateTime(iso?: string): { date: string; time: string } {
  if (!iso) return { date: "-", time: "" };
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return { date: "-", time: "" };

  // simple formatting (ID style)
  const date = d.toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
  const time = d.toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" });
  return { date, time };
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
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="absolute inset-0 bg-black/40" />

      <div className="relative h-full w-full flex items-center justify-center p-4">
        <Card className="w-full max-w-2xl rounded-2xl shadow-xl">
          <CardContent className="p-0">
            <div className="px-5 py-4 border-b border-gray-200 flex items-start justify-between gap-4">
              <div>
                <div className="text-base font-semibold text-gray-900">{title}</div>
                {subtitle && <div className="text-sm text-gray-600">{subtitle}</div>}
              </div>
              <Button variant="outline" className="rounded-xl" onClick={onClose} aria-label="Tutup">
                <X className="h-4 w-4" />
              </Button>
            </div>

            <div className="p-5">
              <div className="overflow-hidden rounded-xl border border-gray-200 bg-gray-50">
                {imageUrl ? (
                  imageUrl.toLowerCase().endsWith('.pdf') ? (
                    <iframe
                      src={imageUrl}
                      title="Registrasi wajah PDF"
                      className="w-full h-[70vh]"
                    />
                  ) : (
                    <img
                      src={imageUrl}
                      alt="Registrasi wajah"
                      className="w-full max-h-[70vh] object-contain"
                    />
                  )
                ) : (
                  <div className="p-10 text-center text-sm text-gray-500">
                    Gambar tidak tersedia
                  </div>
                )}
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
  const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const [q, setQ] = useState("");
  const [dept, setDept] = useState("Semua Departemen");

  const [loading, setLoading] = useState(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // ✅ state modal
  const [openModal, setOpenModal] = useState(false);
  const [modalRowId, setModalRowId] = useState<string>("");

  // ✅ fetched rows
  const [rows, setRows] = useState<FaceRequestRow[]>([]);
  const [selectedId, setSelectedId] = useState<string>("");

  async function fetchRows() {
    setLoading(true);
    setErrorMsg(null);

    try {
      const token = getToken();
      if (!token) throw new Error("Token tidak ditemukan. Silakan login ulang.");

      const endpoint = `face-embeddings/detail?q=${encodeURIComponent(q)}&department=${encodeURIComponent(
        dept
      )}`;

      const url = apiUrl(BASE_URL, endpoint);
      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      });

      const json = await readJsonSafe<ApiSuccess<FaceEmbeddingApprovalItem[]>>(res);
      if (!res.ok) throw new Error(json?.message || "Gagal memuat data registrasi wajah");

      const mapped: FaceRequestRow[] = (json.data || []).map((it) => {
        const dt = formatDateTime(it.registered_at || it.updated_at);
        return {
          id: it.id,
          fullName: it.full_name || "-",
          payrollNumber: it.payroll_number || "-",
          departmentName: it.department_name || "-",
          positionName: it.position_name || "-",
          submittedAt: dt.date,
          submittedAtTime: dt.time,
          email: it.email || "-",
          faceImageUrl: resolveImageUrl(BASE_URL, it.face_image_url),
        };
      });

      setRows(mapped);

      // auto select first row if current selected missing
      setSelectedId((prev) => {
        if (prev && mapped.some((r) => r.id === prev)) return prev;
        return mapped[0]?.id || "";
      });
    } catch (e: any) {
      setErrorMsg(e?.message || "Gagal memuat data.");
      setRows([]);
      setSelectedId("");
    } finally {
      setLoading(false);
    }
  }

  // fetch on mount + when dept/q changes (debounce untuk q)
  useEffect(() => {
    fetchRows();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [dept]);

  useEffect(() => {
    const t = setTimeout(() => {
      fetchRows();
    }, 350);
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [q]);

  const departments = useMemo(() => {
    const uniq = Array.from(new Set(rows.map((r) => r.departmentName)))
      .filter((d) => d && d !== "-")
      .sort();
    return ["Semua Departemen", ...uniq];
  }, [rows]);

  const selected = useMemo(
    () => rows.find((r) => r.id === selectedId) ?? null,
    [rows, selectedId]
  );

  const modalRow = useMemo(
    () => rows.find((r) => r.id === modalRowId) ?? null,
    [rows, modalRowId]
  );

  function openDetailModal(rowId: string) {
    setModalRowId(rowId);
    setOpenModal(true);
  }

  return (
    <div className="p-6">
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
              <h1 className="text-lg font-bold text-gray-900">Persetujuan Registrasi Wajah</h1>
              <p className="text-sm text-gray-600">Kelola permintaan registrasi wajah karyawan</p>
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

              <Button
                variant="outline"
                className="rounded-xl"
                onClick={() => fetchRows()}
                disabled={loading}
              >
                Refresh
              </Button>
            </div>

            {errorMsg && (
              <div className="rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-700">
                {errorMsg}
              </div>
            )}

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
                    <th className="px-5 py-3 text-right text-xs font-semibold uppercase tracking-wider text-gray-500">
                      Aksi
                    </th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-gray-100 bg-white">
                  {loading ? (
                    <tr>
                      <td colSpan={4} className="px-6 py-10 text-center text-sm text-gray-500">
                        Memuat data...
                      </td>
                    </tr>
                  ) : rows.length === 0 ? (
                    <tr>
                      <td colSpan={4} className="px-6 py-10 text-center text-sm text-gray-500">
                        Tidak ada data.
                      </td>
                    </tr>
                  ) : (
                    rows.map((r) => {
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
                          <td className="px-5 py-4">
                            <div className="flex items-center gap-3">
                              <Avatar name={r.fullName} />
                              <div>
                                <div className="text-sm font-semibold text-gray-900">{r.fullName}</div>
                                <div className="text-xs text-gray-500">{r.payrollNumber}</div>
                              </div>
                            </div>
                          </td>

                          <td className="px-5 py-4">
                            <div className="text-sm font-semibold text-gray-900">{r.departmentName}</div>
                            <div className="text-xs text-gray-500">{r.positionName}</div>
                          </td>

                          <td className="px-5 py-4 text-sm text-gray-700">{r.submittedAt}</td>

                          <td className="px-5 py-4">
                            <div className="flex items-center justify-end gap-2">
                              <Button
                                variant="outline"
                                className="rounded-xl"
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setSelectedId(r.id);
                                  openDetailModal(r.id);
                                }}
                                title="Lihat detail"
                              >
                                <Eye className="h-4 w-4" />
                              </Button>
                            </div>
                          </td>
                        </tr>
                      );
                    })
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
              <div className="mt-4 text-sm text-gray-500">
                {loading ? "Memuat..." : "Pilih data di tabel."}
              </div>
            ) : (
              <div className="mt-4 space-y-4">
                <div className="flex items-start gap-3">
                  <div className="h-12 w-12 rounded-xl overflow-hidden bg-gray-100 flex items-center justify-center">
                    {selected.faceImageUrl ? (
                      selected.faceImageUrl.toLowerCase().endsWith('.pdf') ? (
                        <div className="text-xs font-bold text-gray-500">PDF</div>
                      ) : (
                        <img
                          src={selected.faceImageUrl}
                          alt="preview"
                          className="h-12 w-12 object-cover"
                        />
                      )
                    ) : (
                      <div className="text-xs text-gray-400">
                        N/A
                      </div>
                    )}
                  </div>
                  <div>
                    <div className="font-semibold text-gray-900">{selected.fullName}</div>
                    <div className="text-sm text-gray-600">{selected.positionName}</div>
                    <div className="text-xs text-gray-500">{selected.departmentName}</div>
                  </div>
                </div>

                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <div className="text-xs text-gray-500">No Payroll</div>
                    <div className="font-medium text-gray-900">{selected.payrollNumber}</div>
                  </div>
                  <div>
                    <div className="text-xs text-gray-500">Tanggal Pengajuan</div>
                    <div className="font-medium text-gray-900">
                      {selected.submittedAt}
                      {selected.submittedAtTime ? `, ${selected.submittedAtTime}` : ""}
                    </div>
                  </div>
                  <div className="col-span-2">
                    <div className="text-xs text-gray-500">Email</div>
                    <div className="font-medium text-gray-900">{selected.email}</div>
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="text-sm font-semibold text-gray-900">Foto Registrasi Wajah</div>

                  <div className="overflow-hidden rounded-xl border border-gray-200 bg-gray-50">
                    {selected.faceImageUrl ? (
                      selected.faceImageUrl.toLowerCase().endsWith('.pdf') ? (
                        <iframe
                          src={selected.faceImageUrl}
                          title="face"
                          className="w-full aspect-square"
                        />
                      ) : (
                        <img
                          src={selected.faceImageUrl}
                          alt="face"
                          className="w-full aspect-square object-cover"
                        />
                      )
                    ) : (
                      <div className="p-10 text-center text-sm text-gray-500">
                        Gambar tidak tersedia
                      </div>
                    )}
                  </div>

                  <Button
                    variant="outline"
                    className="w-full rounded-xl"
                    onClick={() => openDetailModal(selected.id)}
                    disabled={!selected.faceImageUrl}
                  >
                    Lihat Detail Gambar
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