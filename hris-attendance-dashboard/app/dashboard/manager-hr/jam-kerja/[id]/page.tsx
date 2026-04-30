"use client";

import { useEffect, useState } from "react";
import { ChevronLeft, Save } from "lucide-react";
import { useRouter, useParams } from "next/navigation";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authService } from "@/lib/api/auth";

const HARI = ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"] as const;

type Detail = {
  user_id: string;
  name: string;
  nik: string;
  department: string;
  position: string;
  hari_kerja: string[];
  waktu_mulai: string;   // "HH:mm"
  waktu_selesai: string; // "HH:mm"
  aktif: boolean;
};

export default function AturJamKerjaPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const userId = params?.id;

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [detail, setDetail] = useState<Detail | null>(null);

  useEffect(() => {
  const fetchDetail = async () => {
        try {
        setLoading(true);
        const res = await fetch(`/api/v1/jam-kerja/user/${userId}`, {
            headers: authService.getAuthHeaders(),
        });
        const json = await res.json();
        if (!res.ok) throw new Error(json?.error || json?.message || "Gagal memuat jam kerja");
        const d = json?.data ?? {};
        setDetail({
          user_id: d.user_id ?? d.userId ?? "",
          name: d.name ?? "",
          nik: d.nik ?? "",
          department: d.department ?? "",
          position: d.position ?? "",
          hari_kerja: Array.isArray(d.hari_kerja)
            ? d.hari_kerja
            : Array.isArray(d.day_of_week)
              ? d.day_of_week
              : [],
          waktu_mulai: d.waktu_mulai ?? d.start_time ?? "09:00",
          waktu_selesai: d.waktu_selesai ?? d.end_time ?? "18:00",
          aktif: typeof d.aktif === "boolean" ? d.aktif : (typeof d.is_active === "boolean" ? d.is_active : true),
        });
        } catch (e: any) {
        console.error(e);
        setDetail(null);
        } finally {
        setLoading(false);
        }
    };

    if (userId) fetchDetail();
    }, [userId]);

  const toggleHari = (h: string) => {
    if (!detail) return;
    const exists = detail.hari_kerja.includes(h);
    const next = exists
      ? detail.hari_kerja.filter((x) => x !== h)
      : [...detail.hari_kerja, h];
    setDetail({ ...detail, hari_kerja: next });
  };

  const handleSave = async () => {
    if (!detail) return;

    if (detail.waktu_selesai <= detail.waktu_mulai) {
      alert("Waktu selesai harus lebih dari waktu mulai");
      return;
    }

    setSaving(true);
    try {
        const res = await fetch(`/api/v1/jam-kerja/user/${userId}`, {
        method: "PUT",
        headers: authService.getAuthHeaders(),
        body: JSON.stringify({
            day_of_week: detail.hari_kerja,      // ✅
            start_time: detail.waktu_mulai,      // ✅ "HH:mm"
            end_time: detail.waktu_selesai,      // ✅ "HH:mm"
            is_active: detail.aktif,             // ✅ optional
        }),
        });

      if (!res.ok) {
        const j = await res.json().catch(() => null);
        throw new Error(j?.error || j?.message || "Gagal menyimpan jam kerja");
      }

      router.back();
    } catch (e: any) {
      alert(e?.message || "Gagal menyimpan jam kerja");
    } finally {
      setSaving(false);
    }
  };

  if (loading) return <div className="p-6">Memuat...</div>;
  if (!detail) return <div className="p-6">Data tidak ditemukan</div>;

  return (
    <div className="p-6 space-y-6">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <Button variant="outline" className="rounded-xl" onClick={() => router.back()}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <div>
            <h1 className="text-xl font-semibold text-gray-900">Atur Jam Kerja</h1>
            <p className="text-sm text-gray-600">
              {detail.name} • {detail.department} • {detail.position}
            </p>
          </div>
        </div>

        <Button
          className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl gap-2"
          onClick={handleSave}
          disabled={saving || detail.hari_kerja.length === 0}
        >
          <Save className="h-4 w-4" />
          {saving ? "Menyimpan..." : "Simpan"}
        </Button>
      </div>

      <Card className="rounded-2xl">
        <CardContent className="p-6 space-y-6">
          {/* Hari kerja */}
          <div>
            <div className="text-[11px] font-semibold text-gray-500 uppercase mb-3">
              Hari Kerja
            </div>

            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              {HARI.map((h) => {
                const active = detail.hari_kerja.includes(h);
                return (
                  <button
                    key={h}
                    type="button"
                    onClick={() => toggleHari(h)}
                    className={[
                      "rounded-xl border px-4 py-3 text-sm text-left transition-colors",
                      active
                        ? "border-blue-600 bg-blue-50 text-blue-700"
                        : "border-gray-200 bg-white text-gray-700 hover:bg-gray-50",
                    ].join(" ")}
                  >
                    {h}
                  </button>
                );
              })}
            </div>

            {detail.hari_kerja.length === 0 && (
              <div className="mt-2 text-sm text-red-600">
                Pilih minimal 1 hari kerja.
              </div>
            )}
          </div>

          {/* Jam kerja */}
          <div className="space-y-2">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
              <div>
                <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">
                  Waktu Mulai
                </div>
                <input
                  type="time"
                  value={detail.waktu_mulai}
                  onChange={(e) => setDetail({ ...detail, waktu_mulai: e.target.value })}
                  className="w-full rounded-xl border border-gray-200 bg-white px-3 py-2.5 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>

              <div>
                <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">
                  Waktu Selesai
                </div>
                <input
                  type="time"
                  value={detail.waktu_selesai}
                  onChange={(e) => setDetail({ ...detail, waktu_selesai: e.target.value })}
                  className="w-full rounded-xl border border-gray-200 bg-white px-3 py-2.5 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>

              <div className="flex items-center gap-3">
                <label className="flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={detail.aktif}
                    onChange={(e) => setDetail({ ...detail, aktif: e.target.checked })}
                  />
                  Aktif
                </label>
              </div>
            </div>

            {detail.waktu_selesai && detail.waktu_mulai && detail.waktu_selesai <= detail.waktu_mulai && (
              <p className="text-xs text-red-600">
                ⚠ Waktu selesai harus lebih dari waktu mulai.
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}