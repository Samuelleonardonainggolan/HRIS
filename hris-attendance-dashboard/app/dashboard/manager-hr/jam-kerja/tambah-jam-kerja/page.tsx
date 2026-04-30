"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import toast from "react-hot-toast";
import { authService } from "@/lib/api/auth";

const HARI = ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"] as const;

type AvailableEmployee = {
  id: string;
  full_name: string;
  payroll_number: string;
  department_name: string;
  position_name: string;
};

export default function AddJamKerjaPage() {
  const router = useRouter();

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // autocomplete state
  const [query, setQuery] = useState("");
  const [selectedEmployee, setSelectedEmployee] = useState<AvailableEmployee | null>(null);
  const [suggestions, setSuggestions] = useState<AvailableEmployee[]>([]);
  const [loadingSuggest, setLoadingSuggest] = useState(false);
  const [openSuggest, setOpenSuggest] = useState(false);

  const containerRef = useRef<HTMLDivElement | null>(null);

  const [formData, setFormData] = useState({
    user_id: "",
    hari_kerja: ["Senin", "Selasa", "Rabu", "Kamis", "Jumat"] as string[],
    waktu_mulai: "09:00",
    waktu_selesai: "18:00",
    aktif: true,
  });

  const handleBack = () => router.back();
  const handleCancel = () => router.back();

  const toggleHari = (h: string) => {
    setFormData((prev) => {
      const exists = prev.hari_kerja.includes(h);
      const next = exists ? prev.hari_kerja.filter((x) => x !== h) : [...prev.hari_kerja, h];
      return { ...prev, hari_kerja: next };
    });
  };

  // close suggestions if click outside
  useEffect(() => {
    const onDown = (e: MouseEvent) => {
      const el = containerRef.current;
      if (!el) return;
      if (!el.contains(e.target as Node)) setOpenSuggest(false);
    };
    window.addEventListener("mousedown", onDown);
    return () => window.removeEventListener("mousedown", onDown);
  }, []);

  // fetch suggestions with debounce
  useEffect(() => {
    const q = query.trim();

    // kalau user sudah pilih employee, jangan auto-fetch kecuali user mulai mengetik lagi
    if (selectedEmployee && q === selectedEmployee.full_name) {
      return;
    }

    if (q.length < 2) {
      setSuggestions([]);
      setOpenSuggest(false);
      return;
    }

    const t = setTimeout(async () => {
      try {
        setLoadingSuggest(true);

        const res = await fetch(
          `/api/v1/jam-kerja/available-employees?q=${encodeURIComponent(q)}`,
          { headers: authService.getAuthHeaders() }
        );
        const json = await res.json().catch(() => null);

        if (!res.ok) {
          throw new Error(json?.error || json?.message || "Gagal mencari karyawan");
        }

        const data = (json?.data || []) as AvailableEmployee[];
        setSuggestions(Array.isArray(data) ? data : []);
        setOpenSuggest(true);
      } catch (e: any) {
        console.error(e);
        setSuggestions([]);
        setOpenSuggest(false);
      } finally {
        setLoadingSuggest(false);
      }
    }, 250);

    return () => clearTimeout(t);
  }, [query, selectedEmployee]);

  const selectEmployee = (emp: AvailableEmployee) => {
    setSelectedEmployee(emp);
    setQuery(emp.full_name);
    setFormData((prev) => ({ ...prev, user_id: emp.id }));
    setOpenSuggest(false);
  };

  const selectedLabel = useMemo(() => {
    if (!selectedEmployee) return "";
    return `${selectedEmployee.full_name} • ${selectedEmployee.payroll_number} • ${selectedEmployee.department_name} • ${selectedEmployee.position_name}`;
  }, [selectedEmployee]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    if (!formData.user_id) {
      const msg = "Karyawan wajib dipilih";
      setError(msg);
      toast.error(msg);
      setIsSubmitting(false);
      return;
    }

    if (formData.hari_kerja.length === 0) {
      const msg = "Hari kerja wajib dipilih minimal 1";
      setError(msg);
      toast.error(msg);
      setIsSubmitting(false);
      return;
    }

    if (formData.waktu_selesai <= formData.waktu_mulai) {
      const msg = "Waktu selesai harus lebih dari waktu mulai";
      setError(msg);
      toast.error(msg);
      setIsSubmitting(false);
      return;
    }

    try {
      const res = await fetch("/api/v1/jam-kerja", {
        method: "POST",
        headers: authService.getAuthHeaders(),
        body: JSON.stringify({
          user_id: formData.user_id,
          hari_kerja: formData.hari_kerja,
          waktu_mulai: formData.waktu_mulai,
          waktu_selesai: formData.waktu_selesai,
          aktif: formData.aktif,
        }),
      });

      const json = await res.json().catch(() => null);
      if (!res.ok) throw new Error(json?.error || json?.message || "Gagal menambah jam kerja");

      toast.success("Jam kerja berhasil ditambahkan");
      router.push("/dashboard/manager-hr/jam-kerja");
    } catch (err: any) {
      const message = err instanceof Error ? err.message : "Gagal menambah jam kerja";
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-3xl mx-auto">
        {/* Breadcrumb */}
        <div className="flex items-center gap-2 text-sm text-gray-600 mb-4">
          <button
            onClick={() => router.push("/dashboard/manager-hr")}
            className="hover:text-blue-600 transition-colors"
          >
            Dashboard
          </button>
          <span>/</span>
          <button
            onClick={() => router.push("/dashboard/manager-hr/jam-kerja")}
            className="hover:text-blue-600 transition-colors"
          >
            Manajemen Jam Kerja
          </button>
          <span>/</span>
          <span className="text-gray-900 font-medium">Tambah Jam Kerja</span>
        </div>

        <Card>
          <CardContent className="p-6">
            {/* Header */}
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-gray-200">
              <button
                onClick={handleBack}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <ArrowLeft className="h-5 w-5 text-gray-600" />
              </button>
              <h1 className="text-xl font-semibold text-gray-900">Tambah Jam Kerja</h1>
            </div>

            <form onSubmit={handleSubmit} className="space-y-6">
              {error && <div className="mb-2 text-sm text-red-600">{error}</div>}

              {/* Cari Karyawan (Autocomplete) */}
              <div className="space-y-2" ref={containerRef}>
                <Label className="text-sm font-medium text-gray-700">
                  CARI KARYAWAN
                </Label>

                <Input
                  placeholder="Ketik minimal 2 huruf (nama / NIK / departemen)..."
                  value={query}
                  onChange={(e) => {
                    setQuery(e.target.value);
                    setSelectedEmployee(null);
                    setFormData((p) => ({ ...p, user_id: "" }));
                  }}
                  onFocus={() => {
                    if (suggestions.length > 0) setOpenSuggest(true);
                  }}
                />

                {selectedEmployee && (
                  <div className="text-xs text-gray-600">
                    Dipilih: <span className="font-medium">{selectedLabel}</span>
                  </div>
                )}

                {openSuggest && (
                  <div className="relative">
                    <div className="absolute z-20 mt-2 w-full rounded-xl border border-gray-200 bg-white shadow-lg overflow-hidden">
                      <div className="px-3 py-2 text-xs text-gray-500 border-b bg-gray-50">
                        {loadingSuggest
                          ? "Mencari..."
                          : suggestions.length === 0
                            ? "Tidak ada karyawan yang cocok / semua sudah punya jam kerja."
                            : "Pilih karyawan"}
                      </div>

                      <div className="max-h-64 overflow-auto">
                        {suggestions.map((emp) => (
                          <button
                            key={emp.id}
                            type="button"
                            onClick={() => selectEmployee(emp)}
                            className="w-full text-left px-3 py-2 hover:bg-gray-50 transition-colors"
                          >
                            <div className="text-sm font-semibold text-gray-900">
                              {emp.full_name}
                            </div>
                            <div className="text-xs text-gray-600">
                              {emp.payroll_number} • {emp.department_name} • {emp.position_name}
                            </div>
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                )}

                <p className="text-xs text-gray-500">
                  Hanya menampilkan karyawan yang belum memiliki jam kerja.
                </p>
              </div>

              {/* Hari Kerja */}
              <div className="space-y-2">
                <Label className="text-sm font-medium text-gray-700">HARI KERJA</Label>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
                  {HARI.map((h) => {
                    const active = formData.hari_kerja.includes(h);
                    return (
                      <button
                        key={h}
                        type="button"
                        onClick={() => toggleHari(h)}
                        className={[
                          "rounded-xl border px-3 py-2 text-sm text-left transition-colors",
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
                <p className="text-xs text-gray-500">Pilih minimal 1 hari kerja.</p>
              </div>

              {/* Waktu */}
              <div className="space-y-2">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-sm font-medium text-gray-700">WAKTU MULAI</Label>
                    <Input
                      type="time"
                      value={formData.waktu_mulai}
                      onChange={(e) => setFormData({ ...formData, waktu_mulai: e.target.value })}
                      required
                    />
                  </div>

                  <div className="space-y-2">
                    <Label className="text-sm font-medium text-gray-700">WAKTU SELESAI</Label>
                    <Input
                      type="time"
                      value={formData.waktu_selesai}
                      onChange={(e) => setFormData({ ...formData, waktu_selesai: e.target.value })}
                      required
                    />
                  </div>
                </div>

                {formData.waktu_selesai && formData.waktu_mulai && formData.waktu_selesai <= formData.waktu_mulai && (
                  <p className="text-xs text-red-600">
                    ⚠ Waktu selesai harus lebih dari waktu mulai.
                  </p>
                )}
              </div>

              {/* Aktif */}
              <div className="flex items-center gap-2">
                <input
                  id="aktif"
                  type="checkbox"
                  checked={formData.aktif}
                  onChange={(e) => setFormData({ ...formData, aktif: e.target.checked })}
                />
                <Label htmlFor="aktif" className="text-sm font-medium text-gray-700">
                  Aktif
                </Label>
              </div>

              {/* Action Buttons */}
              <div className="flex justify-end gap-3 pt-6 border-t border-gray-200">
                <Button type="button" variant="outline" onClick={handleCancel} className="px-6">
                  Batal
                </Button>
                <Button
                  type="submit"
                  className="px-6 bg-blue-600 hover:bg-blue-700 text-white"
                  disabled={isSubmitting}
                >
                  {isSubmitting ? "Menyimpan..." : "Simpan Jam Kerja"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}