"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type ApiSuccess<T> = {
  success?: boolean;
  message?: string;
  data: T;
};

type AvailableEmployee = {
  id: string;
  full_name: string;
  payroll_number: string;
  department_name: string;
  position_name: string;
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

export default function TambahGajiKaryawanPage() {
  const router = useRouter();
  const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // autocomplete state
  const [query, setQuery] = useState("");
  const [selectedEmployee, setSelectedEmployee] = useState<AvailableEmployee | null>(null);
  const [suggestions, setSuggestions] = useState<AvailableEmployee[]>([]);
  const [loadingSuggest, setLoadingSuggest] = useState(false);
  const [openSuggest, setOpenSuggest] = useState(false);

  const containerRef = useRef<HTMLDivElement | null>(null);

  const [basicSalary, setBasicSalary] = useState<string>("");
  const [effectiveFrom, setEffectiveFrom] = useState<string>(() => {
    const d = new Date();
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, "0");
    const dd = String(d.getDate()).padStart(2, "0");
    return `${yyyy}-${mm}-${dd}`;
  });
  const [isActive, setIsActive] = useState(true);

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

    // kalau user sudah pilih employee, jangan fetch lagi kecuali user mengetik beda
    if (selectedEmployee && q === selectedEmployee.full_name) return;

    if (q.length < 2) {
      setSuggestions([]);
      setOpenSuggest(false);
      return;
    }

    const t = setTimeout(async () => {
      try {
        setLoadingSuggest(true);
        const token = getToken();
        if (!token) throw new Error("Token tidak ditemukan. Silakan login ulang.");

        const url = apiUrl(
          BASE_URL,
          `employee-basic-salaries/available-employees?q=${encodeURIComponent(q)}`
        );

        const res = await fetch(url, {
          headers: { Authorization: `Bearer ${token}` },
          cache: "no-store",
        });

        const json = await readJsonSafe<ApiSuccess<AvailableEmployee[]>>(res);
        if (!res.ok) throw new Error(json?.message || "Gagal mencari karyawan");

        const data = (json?.data || []) as AvailableEmployee[];
        setSuggestions(Array.isArray(data) ? data : []);
        setOpenSuggest(true);
      } catch (e) {
        console.error(e);
        setSuggestions([]);
        setOpenSuggest(false);
      } finally {
        setLoadingSuggest(false);
      }
    }, 250);

    return () => clearTimeout(t);
  }, [query, selectedEmployee, BASE_URL]);

  const selectEmployee = (emp: AvailableEmployee) => {
    setSelectedEmployee(emp);
    setQuery(emp.full_name);
    setOpenSuggest(false);
  };

  const selectedLabel = useMemo(() => {
    if (!selectedEmployee) return "";
    return `${selectedEmployee.full_name} • ${selectedEmployee.payroll_number} • ${selectedEmployee.department_name} • ${selectedEmployee.position_name}`;
  }, [selectedEmployee]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setErrorMsg(null);

    if (!selectedEmployee) {
      const msg = "Karyawan wajib dipilih";
      setErrorMsg(msg);
      setIsSubmitting(false);
      return;
    }

    const salaryNumber = Number(basicSalary);
    if (!Number.isFinite(salaryNumber) || salaryNumber <= 0) {
      const msg = "Gaji pokok harus berupa angka dan lebih dari 0";
      setErrorMsg(msg);
      setIsSubmitting(false);
      return;
    }

    try {
      const token = getToken();
      if (!token) throw new Error("Token tidak ditemukan. Silakan login ulang.");

      const url = apiUrl(BASE_URL, "employee-basic-salaries");

      const res = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          user_id: selectedEmployee.id,
          basic_salary: salaryNumber,
          effective_from: effectiveFrom,
          is_active: isActive,
        }),
      });

      const json = await readJsonSafe<ApiSuccess<any>>(res);
      if (!res.ok) throw new Error(json?.message || "Gagal menambah gaji pokok");

      router.push("/dashboard/manager-hr/gaji-karyawan");
    } catch (err: any) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal menambah gaji pokok");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="p-6">
      <div className="max-w-3xl mx-auto">
        {/* Breadcrumb */}
        <div className="text-sm text-gray-500 mb-4">
          <button onClick={() => router.push("/dashboard/manager-hr")} className="hover:text-gray-700">
            Dashboard
          </button>{" "}
          /{" "}
          <button onClick={() => router.push("/dashboard/manager-hr/gaji-karyawan")} className="hover:text-gray-700">
            Gaji Karyawan
          </button>{" "}
          / <span className="text-gray-700">Tambah Gaji Pokok</span>
        </div>

        <Card className="rounded-2xl shadow-sm">
          <CardContent className="p-0">
            <div className="px-6 py-5 border-b border-gray-200">
              <div className="flex items-center gap-3">
                <button
                  onClick={() => router.back()}
                  className="h-9 w-9 rounded-lg hover:bg-gray-100 flex items-center justify-center"
                  aria-label="Kembali"
                >
                  <ArrowLeft className="h-5 w-5 text-gray-700" />
                </button>

                <div>
                  <h1 className="text-lg font-bold text-gray-900">Tambah Gaji Pokok</h1>
                  <p className="text-sm text-gray-600">Tambahkan basic salary untuk karyawan.</p>
                </div>
              </div>
            </div>

            <form onSubmit={onSubmit} className="px-6 py-6 space-y-5">
              {errorMsg && (
                <div className="rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-700">
                  {errorMsg}
                </div>
              )}

              {/* Cari Karyawan (Autocomplete) */}
              <div className="space-y-2" ref={containerRef}>
                <Label className="text-xs font-semibold tracking-wide text-gray-700 uppercase">
                  Karyawan
                </Label>

                <Input
                  placeholder="Ketik minimal 2 huruf (nama / payroll / departemen)..."
                  value={query}
                  onChange={(e) => {
                    setQuery(e.target.value);
                    setSelectedEmployee(null);
                  }}
                  onFocus={() => {
                    if (suggestions.length > 0) setOpenSuggest(true);
                  }}
                  className="rounded-xl"
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
                            ? "Tidak ada karyawan yang cocok / semua sudah punya gaji pokok."
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
                            <div className="text-sm font-semibold text-gray-900">{emp.full_name}</div>
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
                  Hanya menampilkan karyawan yang belum memiliki gaji pokok aktif.
                </p>
              </div>

              {/* Gaji pokok */}
              <div className="space-y-2">
                <Label className="text-xs font-semibold tracking-wide text-gray-700 uppercase">
                  Gaji Pokok (Rp)
                </Label>
                <Input
                  value={basicSalary}
                  onChange={(e) => setBasicSalary(e.target.value)}
                  placeholder="Contoh: 7500000"
                  className="rounded-xl"
                  inputMode="numeric"
                />
                <p className="text-xs text-gray-500">
                  Simpan sebagai angka tanpa titik/koma. (Mis. 7500000)
                </p>
              </div>

              {/* Effective From */}
              <div className="space-y-2">
                <Label className="text-xs font-semibold tracking-wide text-gray-700 uppercase">
                  Effective From
                </Label>
                <Input
                  type="date"
                  value={effectiveFrom}
                  onChange={(e) => setEffectiveFrom(e.target.value)}
                  className="rounded-xl"
                />
              </div>

              {/* Status aktif */}
              <div className="space-y-2">
                <Label className="text-xs font-semibold tracking-wide text-gray-700 uppercase">
                  Status Aktif
                </Label>

                <label className="flex items-center gap-3 rounded-xl border border-gray-200 p-3 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    className="h-4 w-4"
                    checked={isActive}
                    onChange={(e) => setIsActive(e.target.checked)}
                  />
                  <span className="text-sm text-gray-700">{isActive ? "Active" : "Inactive"}</span>
                </label>

                <p className="text-xs text-gray-500">
                  Catatan: 1 karyawan hanya boleh punya 1 basic salary yang aktif.
                </p>
              </div>

              {/* Footer */}
              <div className="pt-4 border-t border-gray-200 flex items-center justify-end gap-3">
                <Button
                  type="button"
                  variant="outline"
                  className="rounded-xl"
                  onClick={() => router.push("/dashboard/manager-hr/gaji-karyawan")}
                  disabled={isSubmitting}
                >
                  Batal
                </Button>

                <Button
                  type="submit"
                  className="rounded-xl bg-blue-600 hover:bg-blue-700 text-white"
                  disabled={isSubmitting || !selectedEmployee || !basicSalary}
                >
                  {isSubmitting ? "Menyimpan..." : "Simpan Gaji"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}