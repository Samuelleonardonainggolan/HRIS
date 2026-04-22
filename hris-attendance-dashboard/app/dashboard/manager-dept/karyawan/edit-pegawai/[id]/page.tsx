"use client";

import { useState, useEffect, useCallback } from "react";
import { useRouter, useParams } from "next/navigation";
import { ArrowLeft, Loader2 } from "lucide-react";
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
import { Card, CardContent } from "@/components/ui/card";
import { employeeService } from "@/lib/api/employee";
import { CreateEmployeeRequest, Position } from "@/types";
import toast from "react-hot-toast";

export default function EditEmployeePage() {
  const router = useRouter();
  const params = useParams<{ id: string | string[] }>();
  const rawId = params?.id;
  const id = Array.isArray(rawId) ? rawId[0] : rawId;

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isFetching, setIsFetching] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [positions, setPositions] = useState<Position[]>([]);
  const [departmentName, setDepartmentName] = useState("");

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

  const fetchInitialData = useCallback(async () => {
    if (!id) return;
    try {
      setIsFetching(true);
      setError(null);
      const empData = await employeeService.getEmployeeByID(id);

      const deptName =
        empData.department_name || empData.department || "";
      setDepartmentName(deptName);

      const birthDate = empData.birth_date
        ? new Date(empData.birth_date).toISOString().split("T")[0]
        : "";

      const yearEnrolledRaw = empData.year_enrolled || "";
      const yearEnrolled =
        /^\d{4}$/.test(yearEnrolledRaw) ? `${yearEnrolledRaw}-01-01` : yearEnrolledRaw;

      setFormData({
        payroll_number: empData.payroll_number || empData.nik || "",
        full_name: empData.full_name || "",
        birth_date: birthDate,
        religion: empData.religion || "",
        last_education: empData.last_education || "",
        year_enrolled: yearEnrolled,
        employment_status: empData.employment_status || "",
        department_id: empData.department_id || "",
        position_id: empData.position_id || "",
        email: empData.email || "",
        phone: empData.phone || "",
        address: empData.address || "",
        role: empData.role || "staf",
      });

      if (empData.department_id) {
        const posData = await employeeService.getAllPositions(empData.department_id);
        setPositions(posData);
      }

    } catch (error) {
      console.error("Failed to fetch data:", error);
      setError("Gagal memuat data pegawai");
      toast.error("Gagal memuat data pegawai");
      router.push("/dashboard/manager-dept/karyawan");
    } finally {
      setIsFetching(false);
    }
  }, [id, router]);

  useEffect(() => {
    fetchInitialData();
  }, [fetchInitialData]);

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const missing: string[] = [];
      if (!formData.payroll_number) missing.push("Nomor Payroll");
      if (!formData.full_name) missing.push("Nama Lengkap");
      if (!formData.birth_date) missing.push("Tanggal Lahir");
      if (!formData.religion) missing.push("Agama");
      if (!formData.last_education) missing.push("Pendidikan Terakhir");
      if (!formData.year_enrolled) missing.push("Tanggal Masuk");
      if (!formData.employment_status) missing.push("Status Kepegawaian");
      if (!formData.department_id) missing.push("Departemen");
      if (!formData.position_id) missing.push("Jabatan");
      if (!formData.email) missing.push("Email Kantor");
      if (!formData.phone) missing.push("Nomor Telepon");
      if (!formData.address) missing.push("Alamat");

      if (missing.length > 0) {
        setError(`Field wajib belum diisi: ${missing.join(", ")}`);
        return;
      }

      await employeeService.updateEmployee(id, {
        payroll_number: formData.payroll_number,
        full_name: formData.full_name,
        birth_date: formData.birth_date,
        religion: formData.religion,
        last_education: formData.last_education,
        year_enrolled: formData.year_enrolled,
        employment_status: formData.employment_status,
        department_id: formData.department_id,
        position_id: formData.position_id,
        phone: formData.phone,
        address: formData.address,
      });

      toast.success("Data pegawai berhasil diperbarui");
      router.push("/dashboard/manager-dept/karyawan");
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : "Gagal memperbarui data pegawai";
      console.error("Failed to update employee:", error);
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isFetching) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-5xl mx-auto">
        <div className="flex items-center gap-2 text-sm text-gray-600 mb-4">
          <button
            onClick={() => router.push("/dashboard/manager-dept")}
            className="hover:text-blue-600 transition-colors"
          >
            Dashboard
          </button>
          <span>/</span>
          <button
            onClick={() => router.push("/dashboard/manager-dept/karyawan")}
            className="hover:text-blue-600 transition-colors"
          >
            Manajemen Pegawai
          </button>
          <span>/</span>
          <span className="text-gray-900 font-medium">Edit Pegawai</span>
        </div>

        <Card>
          <CardContent className="p-0">
            <div className="px-6 py-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <button
                    onClick={() => router.back()}
                    className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
                  >
                    <ArrowLeft className="h-5 w-5 text-gray-600" />
                  </button>
                  <h1 className="text-xl font-semibold text-gray-900">Edit Pegawai</h1>
                </div>
                <div className="flex items-center gap-3" />
              </div>
            </div>

            <div className="p-6">
              {error && (
                <div className="mb-6 p-4 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
                  {error}
                </div>
              )}

              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="payroll_number" className="text-sm font-medium text-gray-700">
                      NOMOR PAYROLL <span className="text-red-500">*</span>
                    </Label>
                    <Input
                      id="payroll_number"
                      name="payroll_number"
                      required
                      placeholder="Contoh: PAY001"
                      value={formData.payroll_number}
                      onChange={handleChange}
                      className="w-full"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="full_name" className="text-sm font-medium text-gray-700">
                      NAMA LENGKAP <span className="text-red-500">*</span>
                    </Label>
                    <Input
                      id="full_name"
                      name="full_name"
                      required
                      placeholder="Masukkan nama sesuai KTP"
                      value={formData.full_name}
                      onChange={handleChange}
                      className="w-full"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="birth_date" className="text-sm font-medium text-gray-700">
                      TANGGAL LAHIR <span className="text-red-500">*</span>
                    </Label>
                    <Input
                      id="birth_date"
                      name="birth_date"
                      type="date"
                      required
                      value={formData.birth_date}
                      onChange={handleChange}
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
                        setFormData((prev) => ({ ...prev, religion: value }))
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

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="last_education" className="text-sm font-medium text-gray-700">
                      PENDIDIKAN TERAKHIR <span className="text-red-500">*</span>
                    </Label>
                    <Select
                      value={formData.last_education}
                      onValueChange={(value) =>
                        setFormData((prev) => ({ ...prev, last_education: value }))
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
                      name="year_enrolled"
                      type="date"
                      required
                      value={formData.year_enrolled}
                      onChange={handleChange}
                      className="w-full"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="employment_status" className="text-sm font-medium text-gray-700">
                      STATUS KEPEGAWAIAN <span className="text-red-500">*</span>
                    </Label>
                    <Select
                      value={formData.employment_status}
                      onValueChange={(value) =>
                        setFormData((prev) => ({ ...prev, employment_status: value }))
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
                    <Input
                      id="department_id"
                      value={departmentName || "-"}
                      disabled
                      className="w-full bg-gray-100"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="position_id" className="text-sm font-medium text-gray-700">
                      JABATAN <span className="text-red-500">*</span>
                    </Label>
                    <Select
                      value={formData.position_id}
                      onValueChange={(value) =>
                        setFormData((prev) => ({ ...prev, position_id: value }))
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
                      name="email"
                      type="email"
                      required
                      placeholder="nama@perusahaan.com"
                      value={formData.email}
                      disabled
                      className="w-full bg-gray-100"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="phone" className="text-sm font-medium text-gray-700">
                      NOMOR TELEPON <span className="text-red-500">*</span>
                    </Label>
                    <Input
                      id="phone"
                      name="phone"
                      required
                      placeholder="+62 812 3456 7890"
                      value={formData.phone}
                      onChange={handleChange}
                      className="w-full"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="role" className="text-sm font-medium text-gray-700">
                      ROLE AKSES <span className="text-red-500">*</span>
                    </Label>
                    <Select value={formData.role} disabled>
                      <SelectTrigger className="bg-gray-100">
                        <SelectValue placeholder="Pilih Role" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="staf">Staf</SelectItem>
                        <SelectItem value="manager_departemen">Manager Departemen</SelectItem>
                        <SelectItem value="admin_departemen">Admin Departemen</SelectItem>
                        <SelectItem value="accountant">Accountant</SelectItem>
                        <SelectItem value="manager_hr">Manager HR</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="address" className="text-sm font-medium text-gray-700">
                    ALAMAT LENGKAP <span className="text-red-500">*</span>
                  </Label>
                  <textarea
                    id="address"
                    name="address"
                    rows={3}
                    required
                    placeholder="Masukkan alamat lengkap sesuai KTP"
                    value={formData.address}
                    onChange={handleChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-gray-900 placeholder:text-gray-400"
                  />
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
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
                    disabled={isSubmitting}
                  >
                    {isSubmitting ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Menyimpan...
                      </>
                    ) : (
                      "Simpan Perubahan"
                    )}
                  </Button>
                </div>
              </form>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
