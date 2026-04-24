"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Download, Upload, Loader2 } from "lucide-react";
import { toast, Toaster } from "react-hot-toast";

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
import { ImportExcelModal } from "@/components/import-excel-modal";
import { employeeService } from "@/lib/api/employee";
import { Department, Position, CreateEmployeeRequest } from "@/types";

export default function AddEmployeePage() {
  const router = useRouter();
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [departments, setDepartments] = useState<Department[]>([]);
  const [positions, setPositions] = useState<Position[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [formData, setFormData] = useState<CreateEmployeeRequest>({
    payroll_number: "",
    full_name: "",
    birth_date: "",
    religion: "",
    last_education: "",
    year_enrolled: "",
    employment_status: "",
    department_id: "",
    position_id: "",
    email: "",
    phone: "",
    address: "",
    role: "staf",
  });

  useEffect(() => {
    fetchDepartments();
  }, []);

  useEffect(() => {
    if (formData.department_id) {
      fetchPositions(formData.department_id);
    } else {
      setPositions([]);
    }
  }, [formData.department_id]);

  const fetchDepartments = async () => {
    try {
      const data = await employeeService.getAllDepartments();
      setDepartments(data);
    } catch (err) {
      console.error("Failed to fetch departments:", err);
      toast.error("Gagal memuat data departemen");
    }
  };

  const fetchPositions = async (departmentId: string) => {
    try {
      const data = await employeeService.getAllPositions(departmentId);
      setPositions(data);
    } catch (err) {
      console.error("Failed to fetch positions:", err);
      toast.error("Gagal memuat data jabatan");
    }
  };

  const handleBack = () => router.back();
  const handleCancel = () => router.back();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      // Validate required fields
      const missing: string[] = [];
      if (!formData.payroll_number) missing.push("Nomor Payroll");
      if (!formData.full_name) missing.push("Nama Lengkap");
      if (!formData.birth_date) missing.push("Tanggal Lahir");
      if (!formData.religion) missing.push("Agama");
      if (!formData.last_education) missing.push("Pendidikan Terakhir");
      if (!formData.year_enrolled) missing.push("Tahun Masuk");
      if (!formData.employment_status) missing.push("Status Kepegawaian");
      if (!formData.department_id) missing.push("Departemen");
      if (!formData.position_id) missing.push("Jabatan");
      if (!formData.email) missing.push("Email Kantor");
      if (!formData.phone) missing.push("Nomor Telepon");
      if (!formData.address) missing.push("Alamat");

      if (missing.length > 0) {
        const msg = `Field wajib belum diisi: ${missing.join(", ")}`;
        setError(msg);
        toast.error(msg);
        return;
      }

      const response = await employeeService.createEmployee(formData);

      // ✅ toast success like geofencing
      toast.success("Pegawai berhasil dibuat");

      // ✅ show temporary password (if any) using custom toast (no alert)
      if (response?.temporary_password) {
        const tempPass = response.temporary_password;

        toast.custom(
          (t) => (
            <div className="rounded-xl border border-gray-200 bg-white shadow-lg p-4 w-[360px]">
              <div className="text-sm font-semibold text-gray-900 mb-1">
                Akun berhasil dibuat
              </div>

              <div className="text-xs text-gray-600">
                Email: <span className="font-medium">{formData.email}</span>
              </div>

              <div className="text-xs text-gray-600 mt-1">
                Password sementara:{" "}
                <span className="font-mono font-semibold text-gray-900">
                  {tempPass}
                </span>
              </div>

              <div className="text-xs text-gray-500 mt-2">
                Mohon catat password ini dan berikan kepada karyawan.
              </div>

              <div className="mt-3 flex justify-end gap-2">
                <button
                  type="button"
                  className="text-xs px-3 py-1.5 rounded-lg border border-gray-200 hover:bg-gray-50"
                  onClick={() => {
                    navigator.clipboard
                      .writeText(tempPass)
                      .then(() => toast.success("Password disalin"))
                      .catch(() => toast.error("Gagal menyalin password"));
                  }}
                >
                  Salin
                </button>

                <button
                  type="button"
                  className="text-xs px-3 py-1.5 rounded-lg bg-blue-600 text-white hover:bg-blue-700"
                  onClick={() => toast.dismiss(t.id)}
                >
                  OK
                </button>
              </div>
            </div>
          ),
          { duration: 9000 }
        );
      }

      router.push("/dashboard/manager-hr/karyawan");
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Gagal membuat pegawai. Silakan coba lagi.";
      console.error("Failed to create employee:", err);
      setError(message);
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDownloadTemplate = async () => {
    try {
      const blob = await employeeService.downloadTemplate();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "employee_template.xlsx";
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      toast.success("Template berhasil diunduh");
    } catch (err) {
      console.error("Failed to download template:", err);
      toast.error("Gagal mengunduh template");
    }
  };

  const handleImportExcel = () => setIsImportModalOpen(true);

  return (
    <>
      {/* ✅ Toast container */}
      <Toaster position="top-right" />

      <div className="min-h-screen bg-gray-50 p-6">
        <div className="max-w-5xl mx-auto">
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
              onClick={() => router.push("/dashboard/manager-hr/karyawan")}
              className="hover:text-blue-600 transition-colors"
            >
              Manajemen Pegawai
            </button>
            <span>/</span>
            <span className="text-gray-900 font-medium">Tambah Pegawai Baru</span>
          </div>

          {/* Main Card */}
          <Card>
            <CardContent className="p-0">
              {/* Header */}
              <div className="px-6 py-4 border-b border-gray-200">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <button
                      onClick={handleBack}
                      className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                    >
                      <ArrowLeft className="h-5 w-5 text-gray-600" />
                    </button>
                    <h1 className="text-xl font-semibold text-gray-900">
                      Tambah Pegawai Baru
                    </h1>
                  </div>

                  <div className="flex items-center gap-3">
                    <Button
                      variant="outline"
                      onClick={handleDownloadTemplate}
                      className="flex items-center gap-2"
                      disabled={isSubmitting}
                    >
                      <Download className="h-4 w-4" />
                      Unduh Template
                    </Button>
                    <Button
                      onClick={handleImportExcel}
                      className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white"
                      disabled={isSubmitting}
                    >
                      <Upload className="h-4 w-4" />
                      Import Excel
                    </Button>
                  </div>
                </div>
              </div>

              {/* Form Content */}
              <div className="p-6">
                {error && (
                  <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
                    {error}
                  </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                  {/* Row 1 */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="payroll_number" className="text-sm font-medium text-gray-700">
                        NOMOR PAYROLL <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="payroll_number"
                        required
                        placeholder="Contoh: PAY001"
                        value={formData.payroll_number}
                        onChange={(e) =>
                          setFormData({ ...formData, payroll_number: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="full_name" className="text-sm font-medium text-gray-700">
                        NAMA LENGKAP <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="full_name"
                        required
                        placeholder="Masukkan nama sesuai KTP"
                        value={formData.full_name}
                        onChange={(e) =>
                          setFormData({ ...formData, full_name: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>

                  {/* Row 2 */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="birth_date" className="text-sm font-medium text-gray-700">
                        TANGGAL LAHIR <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="birth_date"
                        type="date"
                        required
                        value={formData.birth_date}
                        onChange={(e) =>
                          setFormData({ ...formData, birth_date: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="religion" className="text-sm font-medium text-gray-700">
                        AGAMA <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.religion}
                        onValueChange={(value) =>
                          setFormData({ ...formData, religion: value })
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Pilih Agama" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="Islam">Islam</SelectItem>
                          <SelectItem value="Kristen">Kristen</SelectItem>
                          <SelectItem value="Katolik">Katolik</SelectItem>
                          <SelectItem value="Hindu">Hindu</SelectItem>
                          <SelectItem value="Buddha">Buddha</SelectItem>
                          <SelectItem value="Konghucu">Konghucu</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* Row 3 */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="last_education" className="text-sm font-medium text-gray-700">
                        PENDIDIKAN TERAKHIR <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.last_education}
                        onValueChange={(value) =>
                          setFormData({ ...formData, last_education: value })
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Pilih Pendidikan" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="SD">SD</SelectItem>
                          <SelectItem value="SMP">SMP</SelectItem>
                          <SelectItem value="SMA/SMK">SMA/SMK</SelectItem>
                          <SelectItem value="D3">D3</SelectItem>
                          <SelectItem value="S1">S1</SelectItem>
                          <SelectItem value="S2">S2</SelectItem>
                          <SelectItem value="S3">S3</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="year_enrolled" className="text-sm font-medium text-gray-700">
                        TANGGAL MASUK <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="year_enrolled"
                        type="date"
                        required
                        value={formData.year_enrolled}
                        onChange={(e) =>
                          setFormData({ ...formData, year_enrolled: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>

                  {/* Row 4 */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="employment_status" className="text-sm font-medium text-gray-700">
                        STATUS KEPEGAWAIAN <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.employment_status}
                        onValueChange={(value) =>
                          setFormData({ ...formData, employment_status: value })
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Pilih Status" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="Permanent">Tetap</SelectItem>
                          <SelectItem value="Contract">Kontrak</SelectItem>
                          <SelectItem value="Internship">Magang</SelectItem>
                          <SelectItem value="Outsourcing">Outsourcing</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="department_id" className="text-sm font-medium text-gray-700">
                        DEPARTEMEN <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.department_id}
                        onValueChange={(value) =>
                          setFormData({ ...formData, department_id: value, position_id: "" })
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Pilih Departemen" />
                        </SelectTrigger>
                        <SelectContent>
                          {departments.map((dept) => (
                            <SelectItem key={dept.id} value={dept.id}>
                              {dept.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* Row 5 */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="position_id" className="text-sm font-medium text-gray-700">
                        JABATAN <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.position_id}
                        onValueChange={(value) =>
                          setFormData({ ...formData, position_id: value })
                        }
                        disabled={!formData.department_id}
                      >
                        <SelectTrigger>
                          <SelectValue
                            placeholder={
                              formData.department_id
                                ? "Pilih Jabatan"
                                : "Pilih Departemen Terlebih Dahulu"
                            }
                          />
                        </SelectTrigger>
                        <SelectContent>
                          {positions.map((pos) => (
                            <SelectItem key={pos.id} value={pos.id}>
                              {pos.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="email" className="text-sm font-medium text-gray-700">
                        EMAIL KANTOR <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="email"
                        type="email"
                        required
                        placeholder="nama@perusahaan.com"
                        value={formData.email}
                        onChange={(e) =>
                          setFormData({ ...formData, email: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>

                  {/* Row 6 */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="phone" className="text-sm font-medium text-gray-700">
                        NOMOR TELEPON <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="phone"
                        required
                        placeholder="+62 812 3456 7890"
                        value={formData.phone}
                        onChange={(e) =>
                          setFormData({ ...formData, phone: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="role" className="text-sm font-medium text-gray-700">
                        ROLE AKSES <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.role}
                        onValueChange={(value) =>
                          setFormData({ ...formData, role: value })
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Pilih Role" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="staf">Staf</SelectItem>
                          <SelectItem value="manager_departemen">
                            Manager Departemen
                          </SelectItem>
                          <SelectItem value="admin_departemen">
                            Admin Departemen
                          </SelectItem>
                          <SelectItem value="accountant">Accountant</SelectItem>
                          <SelectItem value="manager_hr">Manager HR</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* Row 7 */}
                  <div className="space-y-2">
                    <Label htmlFor="address" className="text-sm font-medium text-gray-700">
                      ALAMAT LENGKAP <span className="text-red-500">*</span>
                    </Label>
                    <textarea
                      id="address"
                      rows={3}
                      required
                      placeholder="Masukkan alamat lengkap sesuai KTP"
                      value={formData.address}
                      onChange={(e) =>
                        setFormData({ ...formData, address: e.target.value })
                      }
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-gray-900 placeholder:text-gray-400"
                    />
                  </div>

                  {/* Action Buttons */}
                  <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={handleCancel}
                      className="px-6"
                      disabled={isSubmitting}
                    >
                      Batal
                    </Button>
                    <Button
                      type="submit"
                      className="px-6 bg-blue-600 hover:bg-blue-700 text-white"
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Menyimpan...
                        </>
                      ) : (
                        "Simpan Pegawai"
                      )}
                    </Button>
                  </div>
                </form>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      {/* Import Excel Modal */}
      <ImportExcelModal open={isImportModalOpen} onOpenChange={setIsImportModalOpen} />
    </>
  );
}