"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { positionApi } from "@/lib/api/position";
import { departmentApi } from "@/lib/api/department";
import toast from "react-hot-toast";
import type { Department } from "@/types";

const POSITION_CODE_STOPWORDS = new Set(["dan", "and", "&", "of", "the"]);

function buildPositionCodePrefix(name: string) {
  const normalized = name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, " ")
    .trim()
    .split(/\s+/)
    .filter(Boolean);

  const words = normalized.filter((w) => !POSITION_CODE_STOPWORDS.has(w));
  const source = words.length > 0 ? words : normalized;

  if (source.length >= 2) {
    return `${source[0].slice(0, 1)}${source[1].slice(0, 1)}`.toUpperCase();
  }

  const w = source[0] ?? "";
  if (!w) return "";
  if (w.length >= 2) return w.slice(0, 2).toUpperCase();
  return `${w.slice(0, 1)}X`.toUpperCase();
}

function suggestNextPositionCode(name: string, existingCodes: string[]) {
  const prefix = buildPositionCodePrefix(name);
  if (!prefix) return "";

  const re = new RegExp(`^${prefix}[^0-9]*(\\d+)$`, "i");
  let max = 0;
  for (const raw of existingCodes) {
    const code = (raw ?? "").trim().toUpperCase();
    const m = code.match(re);
    if (!m) continue;
    const n = Number.parseInt(m[1], 10);
    if (Number.isFinite(n) && n > max) max = n;
  }

  const next = max + 1;
  return `${prefix}_${String(next).padStart(3, "0")}`;
}

// ─── Level options ─────────────────────────────────────────────────────────────

const LEVEL_OPTIONS = [
  { value: "1", label: "Level 1 — Staff" },
  { value: "2", label: "Level 2 — Supervisor" },
  { value: "3", label: "Level 3 — Manager" },
  { value: "4", label: "Level 4 — Director" },
  { value: "5", label: "Level 5 — C-Level" },
];

// ─── Form state type ──────────────────────────────────────────────────────────

interface FormData {
  name: string;
  code: string;
  department_id: string;
  level: string;
  description: string;
  is_active: string; // "true" | "false" — only shown in edit mode
}

const INITIAL_FORM: FormData = {
  name: "",
  code: "",
  department_id: "",
  level: "",
  description: "",
  is_active: "true",
};

// ─── Page component ────────────────────────────────────────────────────────────

export default function TambahJabatanPage() {
  const router = useRouter();

  const [editingId, setEditingId]     = useState<string | null>(null);
  const [formData, setFormData]       = useState<FormData>(INITIAL_FORM);
  const [existingPositionCodes, setExistingPositionCodes] = useState<string[]>([]);
  const [isCodeManuallyEdited, setIsCodeManuallyEdited] = useState(false);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [loadingDepts, setLoadingDepts] = useState(true);
  const [loadingData, setLoadingData]   = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError]               = useState<string | null>(null);
  const [errors, setErrors]             = useState<Record<string, string>>({});

  // departmentId yang dikunci dari query (tidak bisa diubah user)
  const [lockedDepartmentId, setLockedDepartmentId] = useState<string | null>(null);

  // ── On mount: parse query params ──────────────────────────────────────────
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const editId       = params.get("edit");
    const departmentId = params.get("departmentId");

    // Kunci departemen dari query
    if (departmentId) {
      setLockedDepartmentId(departmentId);
      setFormData((prev) => ({ ...prev, department_id: departmentId }));
    }

    if (editId) {
      setEditingId(editId);
      loadPosition(editId);
    }

    loadDepartments();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    (async () => {
      try {
        const all = await positionApi.getAll();
        setExistingPositionCodes(
          all.map((p) => p.code).filter((c): c is string => Boolean(c?.trim()))
        );
      } catch {
        setExistingPositionCodes([]);
      }
    })();
  }, []);

  useEffect(() => {
    if (editingId) return;
    if (isCodeManuallyEdited) return;
    setFormData((prev) => ({
      ...prev,
      code: suggestNextPositionCode(prev.name, existingPositionCodes),
    }));
  }, [editingId, existingPositionCodes, isCodeManuallyEdited]);

  // ── Load departments for dropdown ─────────────────────────────────────────
  const loadDepartments = async () => {
    try {
      setLoadingDepts(true);
      const data = await departmentApi.getAll();
      setDepartments(data);
    } catch (err) {
      console.error("Failed to load departments:", err);
      toast.error("Gagal memuat daftar departemen");
    } finally {
      setLoadingDepts(false);
    }
  };

  // ── Load existing position when editing ───────────────────────────────────
  const loadPosition = async (id: string) => {
    try {
      setLoadingData(true);
      const pos = await positionApi.getById(id);
      setFormData((prev) => ({
        ...prev,
        name:          pos.name          ?? "",
        code:          pos.code          ?? "",
        department_id: pos.department_id ?? prev.department_id,
        level:         pos.level ? String(pos.level) : "",
        description:   pos.description   ?? "",
        is_active:     pos.is_active ? "true" : "false",
      }));
    } catch (err) {
      console.error("Failed to load position:", err);
      toast.error("Gagal memuat data jabatan");
    } finally {
      setLoadingData(false);
    }
  };

  // ── Field helpers ─────────────────────────────────────────────────────────
  const set = (field: keyof FormData, value: string) =>
    setFormData((prev) => ({ ...prev, [field]: value }));

  // ── Validation ────────────────────────────────────────────────────────────
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!formData.name.trim())      newErrors.name = "Kolom Nama Jabatan tidak boleh kosong";
    if (!formData.code.trim())      newErrors.code = "Kolom Kode Jabatan tidak boleh kosong";
    if (!formData.department_id)    newErrors.department_id = "Kolom Departemen tidak boleh kosong";
    if (!formData.level)            newErrors.level = "Kolom Level Jabatan tidak boleh kosong";
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // ── Submit ────────────────────────────────────────────────────────────────
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setErrors({});

    if (!validateForm()) {
      toast.error("Mohon lengkapi field wajib");
      return;
    }

    setIsSubmitting(true);

    try {
      if (editingId) {
        // ── UPDATE ──────────────────────────────────────────────────────
        await positionApi.update(editingId, {
          code:          formData.code.trim()        || undefined,
          name:          formData.name.trim(),
          department_id: formData.department_id,
          level:         Number(formData.level),
          description:   formData.description.trim() || undefined,
          is_active:     formData.is_active === "true",
        });
        toast.success("Jabatan berhasil diperbarui");
        router.back();
      } else {
        // ── CREATE ──────────────────────────────────────────────────────
        const codeToSend =
          formData.code.trim() ||
          suggestNextPositionCode(formData.name, existingPositionCodes);
        await positionApi.create({
          code:          codeToSend,
          name:          formData.name.trim(),
          department_id: formData.department_id,
          level:         Number(formData.level),
          description:   formData.description.trim() || undefined,
        });
        toast.success("Jabatan berhasil ditambahkan");
        router.push(
          `/dashboard/manager-hr/jabatan?departmentId=${formData.department_id}`
        );
      }
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Gagal menyimpan jabatan";
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  // ── Derived state ─────────────────────────────────────────────────────────
  const isEditMode  = Boolean(editingId);
  const pageTitle   = isEditMode ? "Edit Jabatan" : "Tambah Jabatan Baru";
  const isDeptLocked = Boolean(lockedDepartmentId);

  // Nama departemen yang terkunci (untuk ditampilkan sebagai teks readonly)
  const lockedDeptName =
    isDeptLocked
      ? departments.find((d) => d.id === lockedDepartmentId)?.name ?? ""
      : null;

  // ── Render ────────────────────────────────────────────────────────────────
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
            onClick={() =>
              router.push(
                lockedDepartmentId
                  ? `/dashboard/manager-hr/jabatan?departmentId=${lockedDepartmentId}`
                  : "/dashboard/manager-hr/jabatan"
              )
            }
            className="hover:text-blue-600 transition-colors"
          >
            Manajemen Jabatan
          </button>
          <span>/</span>
          <span className="text-gray-900 font-medium">{pageTitle}</span>
        </div>

        {/* Main Card */}
        <Card>
          <CardContent className="p-6">

            {/* Header with Back Button */}
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-gray-200">
              <button
                type="button"
                onClick={() => router.back()}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <ArrowLeft className="h-5 w-5 text-gray-600" />
              </button>
              <div>
                <h1 className="text-xl font-semibold text-gray-900">
                  {pageTitle}
                </h1>
                {isDeptLocked && lockedDeptName && (
                  <p className="text-sm text-gray-500 mt-0.5">
                    Departemen:{" "}
                    <span className="font-medium text-gray-700">
                      {lockedDeptName}
                    </span>
                  </p>
                )}
              </div>
            </div>

            {/* Loading indicator for edit data */}
            {loadingData && (
              <div className="mb-4 text-sm text-gray-500 animate-pulse">
                Memuat data jabatan...
              </div>
            )}

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">

              {/* Error Banner */}
              {error && (
                <div className="rounded-lg bg-red-50 border border-red-200 px-4 py-3 text-sm text-red-700">
                  {error}
                </div>
              )}

              {/* 1. Nama Jabatan */}
              <div className="space-y-2">
                <Label htmlFor="name" className="text-sm font-medium text-gray-700">
                  NAMA JABATAN <span className="text-red-500">*</span>
                </Label>
                <Input
                  id="name"
                  placeholder="Contoh: Software Engineer"
                  value={formData.name}
                  onChange={(e) => {
                    const nextName = e.target.value;
                    if (errors.name) {
                      setErrors((prev) => ({ ...prev, name: "" }));
                    }
                    setFormData((prev) => {
                      const next = { ...prev, name: nextName };
                      if (editingId) return next;
                      if (isCodeManuallyEdited) return next;
                      return {
                        ...next,
                        code: suggestNextPositionCode(
                          nextName,
                          existingPositionCodes
                        ),
                      };
                    });
                  }}
                  className={`w-full ${errors.name ? "border-red-500 focus:ring-red-500" : ""}`}
                />
                {errors.name && (
                  <p className="text-red-500 text-xs mt-1">{errors.name}</p>
                )}
              </div>

              {/* 2. Kode Jabatan */}
              <div className="space-y-2">
                <Label htmlFor="code" className="text-sm font-medium text-gray-700">
                  KODE JABATAN
                </Label>
                <Input
                  id="code"
                  placeholder="Otomatis: CH_001"
                  value={formData.code}
                  onChange={(e) => {
                    const nextCode = e.target.value.toUpperCase();
                    if (errors.code) {
                      setErrors((prev) => ({ ...prev, code: "" }));
                    }
                    setIsCodeManuallyEdited(Boolean(nextCode.trim()));
                    set("code", nextCode);
                  }}
                  className={`w-full ${errors.code ? "border-red-500 focus:ring-red-500" : ""}`}
                />
                {errors.code && (
                  <p className="text-red-500 text-xs mt-1">{errors.code}</p>
                )}
                <p className="text-xs text-gray-500">
                  Terisi otomatis sesuai nama jabatan, tetapi tetap bisa diedit
                </p>
              </div>

              {/* 3. Departemen */}
              <div className="space-y-2">
                <Label htmlFor="department_id" className="text-sm font-medium text-gray-700">
                  DEPARTEMEN <span className="text-red-500">*</span>
                </Label>

                {/* Jika departemen sudah dikunci dari query, tampilkan sebagai field readonly */}
                {isDeptLocked ? (
                  <div className="flex items-center gap-2 w-full rounded-lg border border-gray-200 bg-gray-50 px-3 py-2.5">
                    <span className="text-sm text-gray-700 font-medium">
                      {loadingDepts ? "Memuat..." : (lockedDeptName || lockedDepartmentId)}
                    </span>
                    <span className="ml-auto text-xs text-gray-400 bg-gray-200 rounded px-2 py-0.5">
                      Terkunci
                    </span>
                  </div>
                ) : (
                  <Select
                    value={formData.department_id}
                    onValueChange={(val) => {
                      set("department_id", val);
                      if (errors.department_id) {
                        setErrors((prev) => ({ ...prev, department_id: "" }));
                      }
                    }}
                    disabled={loadingDepts}
                  >
                    <SelectTrigger id="department_id" className={`w-full ${errors.department_id ? "border-red-500" : ""}`}>
                      <SelectValue
                        placeholder={
                          loadingDepts ? "Memuat departemen..." : "Pilih Departemen"
                        }
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {departments.map((dept) => (
                        <SelectItem key={dept.id} value={dept.id}>
                          {dept.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
                {errors.department_id && (
                  <p className="text-red-500 text-xs mt-1">{errors.department_id}</p>
                )}
              </div>

              {/* 4. Level Jabatan */}
              <div className="space-y-2">
                <Label htmlFor="level" className="text-sm font-medium text-gray-700">
                  LEVEL JABATAN <span className="text-red-500">*</span>
                </Label>
                <Select
                  value={formData.level}
                  onValueChange={(val) => {
                    set("level", val);
                    if (errors.level) {
                      setErrors((prev) => ({ ...prev, level: "" }));
                    }
                  }}
                >
                  <SelectTrigger id="level" className={`w-full ${errors.level ? "border-red-500" : ""}`}>
                    <SelectValue placeholder="Pilih Level Jabatan" />
                  </SelectTrigger>
                  <SelectContent>
                    {LEVEL_OPTIONS.map((opt) => (
                      <SelectItem key={opt.value} value={opt.value}>
                        {opt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {errors.level && (
                  <p className="text-red-500 text-xs mt-1">{errors.level}</p>
                )}
              </div>

              {/* 5. Deskripsi Jabatan */}
              <div className="space-y-2">
                <Label htmlFor="description" className="text-sm font-medium text-gray-700">
                  DESKRIPSI JABATAN
                </Label>
                <textarea
                  id="description"
                  rows={4}
                  placeholder="Jelaskan tugas dan tanggung jawab jabatan ini..."
                  value={formData.description}
                  onChange={(e) => set("description", e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-gray-900 placeholder:text-gray-400 text-sm"
                />
              </div>

              {/* 6. Status — hanya saat edit */}
              {isEditMode && (
                <div className="space-y-2">
                  <Label htmlFor="is_active" className="text-sm font-medium text-gray-700">
                    STATUS
                  </Label>
                  <Select
                    value={formData.is_active}
                    onValueChange={(val) => set("is_active", val)}
                  >
                    <SelectTrigger id="is_active" className="w-full">
                      <SelectValue placeholder="Pilih Status" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="true">Aktif</SelectItem>
                      <SelectItem value="false">Nonaktif</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex justify-end gap-3 pt-6 border-t border-gray-200">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => router.back()}
                  className="px-6"
                  disabled={isSubmitting}
                >
                  Batal
                </Button>
                <Button
                  type="submit"
                  className="px-6 bg-blue-600 hover:bg-blue-700 text-white"
                  disabled={isSubmitting || loadingData}
                >
                  {isSubmitting
                    ? "Menyimpan..."
                    : isEditMode
                    ? "Simpan Perubahan"
                    : "Simpan Jabatan"}
                </Button>
              </div>

            </form>
          </CardContent>
        </Card>

      </div>
    </div>
  );
}
