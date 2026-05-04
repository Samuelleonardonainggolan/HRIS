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
  const [errors, setErrors] = useState<Partial<Record<keyof CreateEmployeeRequest, string>>>({});
  const [loadingPayroll, setLoadingPayroll] = useState(true);

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
    fetchNextPayrollNumber();
  }, []);

  useEffect(() => {
    if (formData.department_id) {
      fetchPositions(formData.department_id);
    } else {
      setPositions([]);
    }
  }, [formData.department_id]);

  const fetchNextPayrollNumber = async () => {
    try {
      setLoadingPayroll(true);
      const nextNumber = await employeeService.getNextPayrollNumber();
      setFormData((prev) => ({ ...prev, payroll_number: nextNumber }));
    } catch (err) {
      console.error("Failed to fetch next payroll number:", err);
    } finally {
      setLoadingPayroll(false);
    }
  };

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

  const validateForm = () => {
    const newErrors: Partial<Record<keyof CreateEmployeeRequest, string>> = {};
    
    if (!formData.payroll_number) newErrors.payroll_number = "Kolom Nomor Payroll tidak boleh kosong";
    if (!formData.full_name) newErrors.full_name = "Kolom Nama Lengkap tidak boleh kosong";
    if (!formData.birth_date) newErrors.birth_date = "Kolom Tanggal Lahir tidak boleh kosong";
    if (!formData.religion) newErrors.religion = "Kolom Agama tidak boleh kosong";
    if (!formData.last_education) newErrors.last_education = "Kolom Pendidikan Terakhir tidak boleh kosong";
    if (!formData.year_enrolled) newErrors.year_enrolled = "Kolom Tanggal Masuk tidak boleh kosong";
    if (!formData.employment_status) newErrors.employment_status = "Kolom Status Kepegawaian tidak boleh kosong";
    if (!formData.department_id) newErrors.department_id = "Kolom Departemen tidak boleh kosong";
    if (!formData.position_id) newErrors.position_id = "Kolom Jabatan tidak boleh kosong";
    if (!formData.email) newErrors.email = "Kolom Email Kantor tidak boleh kosong";
    if (!formData.phone) newErrors.phone = "Kolom Nomor Telepon tidak boleh kosong";
    if (!formData.address) newErrors.address = "Kolom Alamat Lengkap tidak boleh kosong";
    if (!formData.role) newErrors.role = "Kolom Role Akses tidak boleh kosong";

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      toast.error("Mohon lengkdi semua field wajib");
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const response = await employeeService.createEmployee(formData);

      sessionStorage.setItem("flash_created_employee", JSON.stringify({
        email: formData.email,
        temporary_password: response?.temporary_password,
        full_name: formData.full_name,
        payroll_number: formData.payroll_number,
      }));

      router.push("/dashboard/manager-hr/karyawan?created=1");
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
                      <div className="relative">
                        <Input
                          id="payroll_number"
                          value={formData.payroll_number}
                          onChange={(e) => {
                            setFormData({ ...formData, payroll_number: e.target.value });
                            if (errors.payroll_number) {
                              setErrors((prev) => ({ ...prev, payroll_number: undefined }));
                            }
                          }}
                          className={`w-full ${loadingPayroll ? "bg-gray-50 text-gray-400" : ""} ${
                            errors.payroll_number ? "border-red-500 focus:ring-red-500" : ""
                          }`}
                          disabled={loadingPayroll}
                        />
                        {errors.payroll_number && (
                          <p className="text-red-500 text-xs mt-1">{errors.payroll_number}</p>
                        )}
                        {loadingPayroll && (
                          <div className="absolute right-3 top-1/2 -translate-y-1/2">
                            <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
                          </div>
                        )}
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="full_name" className="text-sm font-medium text-gray-700">
                        NAMA LENGKAP <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="full_name"
                        placeholder="Masukkan nama sesuai KTP"
                        value={formData.full_name}
                        onChange={(e) => {
                          setFormData({ ...formData, full_name: e.target.value });
                          if (errors.full_name) {
                            setErrors((prev) => ({ ...prev, full_name: undefined }));
                          }
                        }}
                        className={`w-full ${errors.full_name ? "border-red-500 focus:ring-red-500" : ""}`}
                      />
                      {errors.full_name && (
                        <p className="text-red-500 text-xs mt-1">{errors.full_name}</p>
                      )}
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
                        value={formData.birth_date}
                        onChange={(e) => {
                          setFormData({ ...formData, birth_date: e.target.value });
                          if (errors.birth_date) {
                            setErrors((prev) => ({ ...prev, birth_date: undefined }));
                          }
                        }}
                        className={`w-full ${errors.birth_date ? "border-red-500 focus:ring-red-500" : ""}`}
                      />
                      {errors.birth_date && (
                        <p className="text-red-500 text-xs mt-1">{errors.birth_date}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="religion" className="text-sm font-medium text-gray-700">
                        AGAMA <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.religion}
                        onValueChange={(value) => {
                          setFormData({ ...formData, religion: value });
                          if (errors.religion) {
                            setErrors((prev) => ({ ...prev, religion: undefined }));
                          }
                        }}
                      >
                        <SelectTrigger className={errors.religion ? "border-red-500" : ""}>
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
                      {errors.religion && (
                        <p className="text-red-500 text-xs mt-1">{errors.religion}</p>
                      )}
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
                        onValueChange={(value) => {
                          setFormData({ ...formData, last_education: value });
                          if (errors.last_education) {
                            setErrors((prev) => ({ ...prev, last_education: undefined }));
                          }
                        }}
                      >
                        <SelectTrigger className={errors.last_education ? "border-red-500" : ""}>
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
                      {errors.last_education && (
                        <p className="text-red-500 text-xs mt-1">{errors.last_education}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="year_enrolled" className="text-sm font-medium text-gray-700">
                        TANGGAL MASUK <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="year_enrolled"
                        type="date"
                        value={formData.year_enrolled}
                        onChange={(e) => {
                          setFormData({ ...formData, year_enrolled: e.target.value });
                          if (errors.year_enrolled) {
                            setErrors((prev) => ({ ...prev, year_enrolled: undefined }));
                          }
                        }}
                        className={`w-full ${errors.year_enrolled ? "border-red-500 focus:ring-red-500" : ""}`}
                      />
                      {errors.year_enrolled && (
                        <p className="text-red-500 text-xs mt-1">{errors.year_enrolled}</p>
                      )}
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
                        onValueChange={(value) => {
                          setFormData({ ...formData, employment_status: value });
                          if (errors.employment_status) {
                            setErrors((prev) => ({ ...prev, employment_status: undefined }));
                          }
                        }}
                      >
                        <SelectTrigger className={errors.employment_status ? "border-red-500" : ""}>
                          <SelectValue placeholder="Pilih Status" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="Permanent">Tetap</SelectItem>
                          <SelectItem value="Contract">Kontrak</SelectItem>
                          <SelectItem value="Internship">Magang</SelectItem>
                          <SelectItem value="Outsourcing">Outsourcing</SelectItem>
                        </SelectContent>
                      </Select>
                      {errors.employment_status && (
                        <p className="text-red-500 text-xs mt-1">{errors.employment_status}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="department_id" className="text-sm font-medium text-gray-700">
                        DEPARTEMEN <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.department_id}
                        onValueChange={(value) => {
                          setFormData({ ...formData, department_id: value, position_id: "" });
                          if (errors.department_id) {
                            setErrors((prev) => ({ ...prev, department_id: undefined }));
                          }
                        }}
                      >
                        <SelectTrigger className={errors.department_id ? "border-red-500" : ""}>
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
                      {errors.department_id && (
                        <p className="text-red-500 text-xs mt-1">{errors.department_id}</p>
                      )}
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
                        onValueChange={(value) => {
                          setFormData({ ...formData, position_id: value });
                          if (errors.position_id) {
                            setErrors((prev) => ({ ...prev, position_id: undefined }));
                          }
                        }}
                        disabled={!formData.department_id}
                      >
                        <SelectTrigger className={errors.position_id ? "border-red-500" : ""}>
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
                      {errors.position_id && (
                        <p className="text-red-500 text-xs mt-1">{errors.position_id}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="email" className="text-sm font-medium text-gray-700">
                        EMAIL KANTOR <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="email"
                        type="email"
                        placeholder="nama@perusahaan.com"
                        value={formData.email}
                        onChange={(e) => {
                          setFormData({ ...formData, email: e.target.value });
                          if (errors.email) {
                            setErrors((prev) => ({ ...prev, email: undefined }));
                          }
                        }}
                        className={`w-full ${errors.email ? "border-red-500 focus:ring-red-500" : ""}`}
                      />
                      {errors.email && (
                        <p className="text-red-500 text-xs mt-1">{errors.email}</p>
                      )}
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
                        placeholder="+62 812 3456 7890"
                        value={formData.phone}
                        onChange={(e) => {
                          setFormData({ ...formData, phone: e.target.value });
                          if (errors.phone) {
                            setErrors((prev) => ({ ...prev, phone: undefined }));
                          }
                        }}
                        className={`w-full ${errors.phone ? "border-red-500 focus:ring-red-500" : ""}`}
                      />
                      {errors.phone && (
                        <p className="text-red-500 text-xs mt-1">{errors.phone}</p>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="role" className="text-sm font-medium text-gray-700">
                        ROLE AKSES <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.role}
                        onValueChange={(value) => {
                          setFormData({ ...formData, role: value });
                          if (errors.role) {
                            setErrors((prev) => ({ ...prev, role: undefined }));
                          }
                        }}
                      >
                        <SelectTrigger className={errors.role ? "border-red-500" : ""}>
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
                      {errors.role && (
                        <p className="text-red-500 text-xs mt-1">{errors.role}</p>
                      )}
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
                      placeholder="Masukkan alamat lengkap sesuai KTP"
                      value={formData.address}
                      onChange={(e) => {
                        setFormData({ ...formData, address: e.target.value });
                        if (errors.address) {
                          setErrors((prev) => ({ ...prev, address: undefined }));
                        }
                      }}
                      className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-gray-900 placeholder:text-gray-400 ${
                        errors.address ? "border-red-500 focus:ring-red-500" : "border-gray-300"
                      }`}
                    />
                    {errors.address && (
                      <p className="text-red-500 text-xs mt-1">{errors.address}</p>
                    )}
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
