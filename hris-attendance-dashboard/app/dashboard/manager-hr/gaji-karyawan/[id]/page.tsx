"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
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

type SalaryResponse = {
  id: string; // salary record id
  user_id: string;
  basic_salary: number;
  effective_from: string; // ISO
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

type EmployeeDetail = {
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

function toYYYYMMDD(iso: string) {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export default function EditGajiKaryawanPage() {
  const router = useRouter();
  const params = useParams();

  // ✅ karena folder dynamic Anda [id]
  const userId = typeof params?.id === "string" ? params.id : "";

  const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const [loading, setLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const [employee, setEmployee] = useState<EmployeeDetail | null>(null);

  // ✅ untuk submit PATCH /employee-basic-salaries/:salaryId
  const [salaryId, setSalaryId] = useState<string>("");

  const [basicSalary, setBasicSalary] = useState<string>("");
  const [effectiveFrom, setEffectiveFrom] = useState<string>("");
  const [isActive, setIsActive] = useState<boolean>(true);

  const employeeLabel = useMemo(() => {
    if (!employee) return userId;
    return `${employee.full_name} • ${employee.payroll_number} • ${employee.department_name} • ${employee.position_name}`;
  }, [employee, userId]);

  async function fetchInitial() {
    setLoading(true);
    setErrorMsg(null);

    try {
      const token = getToken();
      if (!token) throw new Error("Token tidak ditemukan. Silakan login ulang.");
      if (!userId) throw new Error("User ID tidak valid.");

      // ✅ salary terbaru (aktif / nonaktif)
      const salaryUrl = apiUrl(BASE_URL, `employee-basic-salaries/users/${userId}/latest`);
      const salaryRes = await fetch(salaryUrl, {
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      });

      const salaryJson = await readJsonSafe<ApiSuccess<SalaryResponse>>(salaryRes);
      if (!salaryRes.ok) throw new Error(salaryJson?.message || "Gagal mengambil data gaji");

      setSalaryId(salaryJson.data.id);
      setBasicSalary(String(salaryJson.data.basic_salary ?? ""));
      setEffectiveFrom(toYYYYMMDD(salaryJson.data.effective_from));
      setIsActive(Boolean(salaryJson.data.is_active));

      // Optional: employee detail (mungkin AdminOnly -> ignore kalau gagal)
      try {
        const empUrl = apiUrl(BASE_URL, `employees/${userId}`);
        const empRes = await fetch(empUrl, {
          headers: { Authorization: `Bearer ${token}` },
          cache: "no-store",
        });
        const empJson = await readJsonSafe<ApiSuccess<any>>(empRes);
        if (empRes.ok && empJson?.data) {
          const u = empJson.data;
          setEmployee({
            id: u.id,
            full_name: u.full_name,
            payroll_number: u.payroll_number,
            department_name: u.department_name,
            position_name: u.position_name,
          });
        }
      } catch {
        // ignore
      }
    } catch (e: any) {
      setErrorMsg(e?.message || "Gagal memuat data.");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    if (!userId) return;
    fetchInitial();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [userId]);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setErrorMsg(null);

    try {
      const token = getToken();
      if (!token) throw new Error("Token tidak ditemukan. Silakan login ulang.");
      if (!salaryId) throw new Error("Salary ID tidak ditemukan (data belum termuat).");

      const salaryNumber = Number(basicSalary);
      if (!Number.isFinite(salaryNumber) || salaryNumber <= 0) {
        throw new Error("Gaji pokok harus berupa angka dan lebih dari 0");
      }
      if (!effectiveFrom) {
        throw new Error("Effective From wajib diisi");
      }

      const url = apiUrl(BASE_URL, `employee-basic-salaries/${salaryId}`);

      const res = await fetch(url, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          basic_salary: salaryNumber,
          effective_from: effectiveFrom,
          is_active: isActive,
        }),
      });

      const json = await readJsonSafe<ApiSuccess<any>>(res);
      if (!res.ok) throw new Error(json?.message || "Gagal memperbarui gaji pokok");

      router.push("/dashboard/manager-hr/gaji-karyawan");
    } catch (err: any) {
      setErrorMsg(err instanceof Error ? err.message : "Gagal memperbarui gaji pokok");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="p-6">
      <div className="max-w-3xl mx-auto">
        <div className="text-sm text-gray-500 mb-4">
          <button onClick={() => router.push("/dashboard/manager-hr")} className="hover:text-gray-700">
            Dashboard
          </button>{" "}
          /{" "}
          <button onClick={() => router.push("/dashboard/manager-hr/gaji-karyawan")} className="hover:text-gray-700">
            Gaji Karyawan
          </button>{" "}
          / <span className="text-gray-700">Edit Gaji Pokok</span>
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
                  <h1 className="text-lg font-bold text-gray-900">Edit Gaji Pokok</h1>
                  <p className="text-sm text-gray-600">Perbarui basic salary untuk karyawan.</p>
                </div>
              </div>
            </div>

            <div className="px-6 py-6">
              {loading ? (
                <div className="text-sm text-gray-500">Memuat data...</div>
              ) : (
                <form onSubmit={onSubmit} className="space-y-5">
                  {errorMsg && (
                    <div className="rounded-xl border border-rose-200 bg-rose-50 p-3 text-sm text-rose-700">
                      {errorMsg}
                    </div>
                  )}

                  <div className="space-y-2">
                    <Label className="text-xs font-semibold tracking-wide text-gray-700 uppercase">
                      Karyawan
                    </Label>
                    <Input value={employeeLabel} readOnly className="rounded-xl bg-gray-50" />
                  </div>

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
                  </div>

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
                  </div>

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
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? "Menyimpan..." : "Simpan Perubahan"}
                    </Button>
                  </div>
                </form>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}