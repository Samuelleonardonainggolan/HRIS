// "use client";

// import { useState } from "react";
// import { useRouter } from "next/navigation";
// import { ArrowLeft, Download, Upload } from "lucide-react";
// import { Card, CardContent } from "@/components/ui/card";
// import { Button } from "@/components/ui/button";
// import { Input } from "@/components/ui/input";
// import { Label } from "@/components/ui/label";
// import {
//   Select,
//   SelectContent,
//   SelectItem,
//   SelectTrigger,
//   SelectValue,
// } from "@/components/ui/select";
// import { ImportExcelModal } from "@/components/import-excel-modal";

// export default function AddEmployeePage() {
//   const router = useRouter();
//   const [formData, setFormData] = useState({
//     nik: "",
//     fullName: "",
//     birthDate: "",
//     religion: "",
//     lastEducation: "",
//     yearEnrolled: "",
//     employmentStatus: "",
//     department: "",
//     position: "",
//     officeEmail: "",
//     phoneNumber: "",
//     address: "",
//   });

//   const handleBack = () => {
//     router.back();
//   };

//   const handleCancel = () => {
//     router.back();
//   };

//   const handleSubmit = (e: React.FormEvent) => {
//     e.preventDefault();
//     // TODO: Implement save employee logic
//     console.log("Form data:", formData);
//   };

//   const handleDownloadTemplate = () => {
//     // TODO: Implement download template
//     console.log("Download template");
//   };

//   const handleImportExcel = () => {
//     setIsImportModalOpen(true);
//   };

//   return (
//     <>
//     <div className="min-h-screen bg-gray-50 p-6">
//       <div className="max-w-5xl mx-auto">
//         {/* Breadcrumb - Outside Card */}
//         <div className="flex items-center gap-2 text-sm text-gray-600 mb-4">
//           <button
//             onClick={() => router.push("/dashboard/manager-hr")}
//             className="hover:text-blue-600 transition-colors"
//           >
//             Dashboard
//           </button>
//           <span>/</span>
//           <button
//             onClick={() => router.push("/dashboard/manager-hr/employees")}
//             className="hover:text-blue-600 transition-colors"
//           >
//             Manajemen Pegawai
//           </button>
//           <span>/</span>
//           <span className="text-gray-900 font-medium">Tambah Pegawai Baru</span>
//         </div>

//         {/* Main Card */}
//         <Card>
//           <CardContent className="p-0">
//             {/* Header Inside Card */}
//             <div className="px-6 py-4 border-b border-gray-200">
//               <div className="flex items-center justify-between">
//                 {/* Title with Back Button */}
//                 <div className="flex items-center gap-3">
//                   <button
//                     onClick={handleBack}
//                     className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
//                   >
//                     <ArrowLeft className="h-5 w-5 text-gray-600" />
//                   </button>
//                   <h1 className="text-xl font-semibold text-gray-900">
//                     Tambah Pegawai Baru
//                   </h1>
//                 </div>

//                 {/* Action Buttons */}
//                 <div className="flex items-center gap-3">
//                   <Button
//                     variant="outline"
//                     onClick={handleDownloadTemplate}
//                     className="flex items-center gap-2"
//                   >
//                     <Download className="h-4 w-4" />
//                     Unduh Template
//                   </Button>
//                   <Button
//                     onClick={handleImportExcel}
//                     className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white"
//                   >
//                     <Upload className="h-4 w-4" />
//                     Import Excel
//                   </Button>
//                 </div>
//               </div>
//             </div>

//             {/* Form Content */}
//             <div className="p-6">
//               <form onSubmit={handleSubmit} className="space-y-6">
//                 {/* Row 1: NIK & Nama Lengkap */}
//                 <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
//                   <div className="space-y-2">
//                     <Label htmlFor="nik" className="text-sm font-medium text-gray-700">
//                       NIK (NOMOR INDUK KARYAWAN)
//                     </Label>
//                     <Input
//                       id="nik"
//                       placeholder="Contoh: 3210001234"
//                       value={formData.nik}
//                       onChange={(e) =>
//                         setFormData({ ...formData, nik: e.target.value })
//                       }
//                       className="w-full"
//                     />
//                   </div>

//                   <div className="space-y-2">
//                     <Label htmlFor="fullName" className="text-sm font-medium text-gray-700">
//                       NAMA LENGKAP
//                     </Label>
//                     <Input
//                       id="fullName"
//                       placeholder="Masukkan nama sesuai KTP"
//                       value={formData.fullName}
//                       onChange={(e) =>
//                         setFormData({ ...formData, fullName: e.target.value })
//                       }
//                       className="w-full"
//                     />
//                   </div>
//                 </div>

//                 {/* Row 2: Tanggal Lahir & Agama */}
//                 <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
//                   <div className="space-y-2">
//                     <Label htmlFor="birthDate" className="text-sm font-medium text-gray-700">
//                       TANGGAL LAHIR
//                     </Label>
//                     <Input
//                       id="birthDate"
//                       type="date"
//                       placeholder="mm/dd/yyyy"
//                       value={formData.birthDate}
//                       onChange={(e) =>
//                         setFormData({ ...formData, birthDate: e.target.value })
//                       }
//                       className="w-full"
//                     />
//                   </div>

//                   <div className="space-y-2">
//                     <Label htmlFor="religion" className="text-sm font-medium text-gray-700">
//                       AGAMA
//                     </Label>
//                     <Select
//                       value={formData.religion}
//                       onValueChange={(value) =>
//                         setFormData({ ...formData, religion: value })
//                       }
//                     >
//                       <SelectTrigger>
//                         <SelectValue placeholder="Pilih Agama" />
//                       </SelectTrigger>
//                       <SelectContent>
//                         <SelectItem value="islam">Islam</SelectItem>
//                         <SelectItem value="kristen">Kristen</SelectItem>
//                         <SelectItem value="katolik">Katolik</SelectItem>
//                         <SelectItem value="hindu">Hindu</SelectItem>
//                         <SelectItem value="buddha">Buddha</SelectItem>
//                         <SelectItem value="konghucu">Konghucu</SelectItem>
//                       </SelectContent>
//                     </Select>
//                   </div>
//                 </div>

//                 {/* Row 3: Pendidikan & Tahun Masuk */}
//                 <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
//                   <div className="space-y-2">
//                     <Label htmlFor="lastEducation" className="text-sm font-medium text-gray-700">
//                       PENDIDIKAN TERAKHIR
//                     </Label>
//                     <Select
//                       value={formData.lastEducation}
//                       onValueChange={(value) =>
//                         setFormData({ ...formData, lastEducation: value })
//                       }
//                     >
//                       <SelectTrigger>
//                         <SelectValue placeholder="Pilih Pendidikan" />
//                       </SelectTrigger>
//                       <SelectContent>
//                         <SelectItem value="sd">SD</SelectItem>
//                         <SelectItem value="smp">SMP</SelectItem>
//                         <SelectItem value="sma">SMA/SMK</SelectItem>
//                         <SelectItem value="d3">D3</SelectItem>
//                         <SelectItem value="s1">S1</SelectItem>
//                         <SelectItem value="s2">S2</SelectItem>
//                         <SelectItem value="s3">S3</SelectItem>
//                       </SelectContent>
//                     </Select>
//                   </div>

//                   <div className="space-y-2">
//                     <Label htmlFor="yearEnrolled" className="text-sm font-medium text-gray-700">
//                       TAHUN MASUK
//                     </Label>
//                     <Input
//                       id="yearEnrolled"
//                       placeholder="Contoh: 2023"
//                       value={formData.yearEnrolled}
//                       onChange={(e) =>
//                         setFormData({ ...formData, yearEnrolled: e.target.value })
//                       }
//                       className="w-full"
//                     />
//                   </div>
//                 </div>

//                 {/* Row 4: Status Kepegawaian & Departemen */}
//                 <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
//                   <div className="space-y-2">
//                     <Label htmlFor="employmentStatus" className="text-sm font-medium text-gray-700">
//                       STATUS KEPEGAWAIAN
//                     </Label>
//                     <Select
//                       value={formData.employmentStatus}
//                       onValueChange={(value) =>
//                         setFormData({ ...formData, employmentStatus: value })
//                       }
//                     >
//                       <SelectTrigger>
//                         <SelectValue placeholder="Pilih Status" />
//                       </SelectTrigger>
//                       <SelectContent>
//                         <SelectItem value="tetap">Tetap</SelectItem>
//                         <SelectItem value="kontrak">Kontrak</SelectItem>
//                         <SelectItem value="magang">Magang</SelectItem>
//                         <SelectItem value="outsourcing">Outsourcing</SelectItem>
//                       </SelectContent>
//                     </Select>
//                   </div>

//                   <div className="space-y-2">
//                     <Label htmlFor="department" className="text-sm font-medium text-gray-700">
//                       DEPARTEMEN
//                     </Label>
//                     <Select
//                       value={formData.department}
//                       onValueChange={(value) =>
//                         setFormData({ ...formData, department: value })
//                       }
//                     >
//                       <SelectTrigger>
//                         <SelectValue placeholder="Pilih Departemen" />
//                       </SelectTrigger>
//                       <SelectContent>
//                         <SelectItem value="hr">Human Resources</SelectItem>
//                         <SelectItem value="it">Information Technology</SelectItem>
//                         <SelectItem value="finance">Finance & Accounting</SelectItem>
//                         <SelectItem value="marketing">Sales & Marketing</SelectItem>
//                         <SelectItem value="operations">Operations</SelectItem>
//                       </SelectContent>
//                     </Select>
//                   </div>
//                 </div>

//                 {/* Row 5: Jabatan & Email Kantor */}
//                 <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
//                   <div className="space-y-2">
//                     <Label htmlFor="position" className="text-sm font-medium text-gray-700">
//                       JABATAN
//                     </Label>
//                     <Select
//                       value={formData.position}
//                       onValueChange={(value) =>
//                         setFormData({ ...formData, position: value })
//                       }
//                     >
//                       <SelectTrigger>
//                         <SelectValue placeholder="Pilih Jabatan" />
//                       </SelectTrigger>
//                       <SelectContent>
//                         <SelectItem value="staff">Staff</SelectItem>
//                         <SelectItem value="supervisor">Supervisor</SelectItem>
//                         <SelectItem value="manager">Manager</SelectItem>
//                         <SelectItem value="director">Director</SelectItem>
//                       </SelectContent>
//                     </Select>
//                   </div>

//                   <div className="space-y-2">
//                     <Label htmlFor="officeEmail" className="text-sm font-medium text-gray-700">
//                       EMAIL KANTOR
//                     </Label>
//                     <Input
//                       id="officeEmail"
//                       type="email"
//                       placeholder="nama@perusahaan.com"
//                       value={formData.officeEmail}
//                       onChange={(e) =>
//                         setFormData({ ...formData, officeEmail: e.target.value })
//                       }
//                       className="w-full"
//                     />
//                   </div>
//                 </div>

//                 {/* Row 6: Nomor Telepon */}
//                 <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
//                   <div className="space-y-2">
//                     <Label htmlFor="phoneNumber" className="text-sm font-medium text-gray-700">
//                       NOMOR TELEPON
//                     </Label>
//                     <Input
//                       id="phoneNumber"
//                       placeholder="+62 812 3456 7890"
//                       value={formData.phoneNumber}
//                       onChange={(e) =>
//                         setFormData({ ...formData, phoneNumber: e.target.value })
//                       }
//                       className="w-full"
//                     />
//                   </div>
//                 </div>

//                 {/* Row 7: Alamat Lengkap */}
//                 <div className="space-y-2">
//                   <Label htmlFor="address" className="text-sm font-medium text-gray-700">
//                     ALAMAT LENGKAP
//                   </Label>
//                   <textarea
//                     id="address"
//                     rows={3}
//                     placeholder="Masukkan alamat lengkap sesuai KTP"
//                     value={formData.address}
//                     onChange={(e) =>
//                       setFormData({ ...formData, address: e.target.value })
//                     }
//                     className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-gray-900 placeholder:text-gray-400"
//                   />
//                 </div>

//                 {/* Action Buttons */}
//                 <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
//                   <Button
//                     type="button"
//                     variant="outline"
//                     onClick={handleCancel}
//                     className="px-6"
//                   >
//                     Batal
//                   </Button>
//                   <Button
//                     type="submit"
//                     className="px-6 bg-blue-600 hover:bg-blue-700 text-white"
//                   >
//                     Simpan Pegawai
//                   </Button>
//                 </div>
//               </form>
//             </div>
//           </CardContent>
//         </Card>
//       </div>
//     </div>
//           {/* Import Excel Modal */}
//       <ImportExcelModal
//         open={isImportModalOpen}
//         onOpenChange={setIsImportModalOpen}
//       />
//     </>
//   );
// }


"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Download, Upload, Loader2 } from "lucide-react";
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
    nik: "",
    fullName: "",
    birthDate: "",
    religion: "",
    lastEducation: "",
    yearEnrolled: "",
    employmentStatus: "",
    department: "",
    position: "",
    officeEmail: "",
    phoneNumber: "",
    address: "",
    role: "staf",
  });

  useEffect(() => {
    fetchDepartments();
  }, []);

  useEffect(() => {
    if (formData.department) {
      fetchPositions(formData.department);
    } else {
      setPositions([]);
    }
  }, [formData.department]);

  const fetchDepartments = async () => {
    try {
      const data = await employeeService.getAllDepartments();
      setDepartments(data);
    } catch (err) {
      console.error("Failed to fetch departments:", err);
    }
  };

  const fetchPositions = async (departmentId: string) => {
    try {
      const data = await employeeService.getAllPositions(departmentId);
      setPositions(data);
    } catch (err) {
      console.error("Failed to fetch positions:", err);
    }
  };

  const handleBack = () => {
    router.back();
  };

  const handleCancel = () => {
    router.back();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const missing: string[] = [];
      if (!formData.nik) missing.push("NIK");
      if (!formData.fullName) missing.push("Nama Lengkap");
      if (!formData.birthDate) missing.push("Tanggal Lahir");
      if (!formData.religion) missing.push("Agama");
      if (!formData.lastEducation) missing.push("Pendidikan Terakhir");
      if (!formData.yearEnrolled) missing.push("Tahun Masuk");
      if (!formData.employmentStatus) missing.push("Status Kepegawaian");
      if (!formData.department) missing.push("Departemen");
      if (!formData.position) missing.push("Jabatan");
      if (!formData.officeEmail) missing.push("Email Kantor");
      if (!formData.phoneNumber) missing.push("Nomor Telepon");
      if (!formData.address) missing.push("Alamat");
      if (missing.length > 0) {
        setError(`Field wajib belum diisi: ${missing.join(", ")}`);
        return;
      }

      const response = await employeeService.createEmployee(formData);
      
      if (response.temporary_password) {
        alert(`Pegawai berhasil dibuat!\nPassword Sementara: ${response.temporary_password}`);
      } else {
        alert("Pegawai berhasil dibuat!");
      }
      
      router.push("/dashboard/manager-hr/karyawan");
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Gagal membuat pegawai. Silakan coba lagi.";
      console.error("Failed to create employee:", err);
      setError(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDownloadTemplate = async () => {
    try {
      const blob = await employeeService.downloadTemplate();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'employee_template.xlsx';
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      console.error("Failed to download template:", err);
      alert("Gagal mengunduh template");
    }
  };

  const handleImportExcel = () => {
    setIsImportModalOpen(true);
  };

  return (
    <>
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
                    >
                      <Download className="h-4 w-4" />
                      Unduh Template
                    </Button>
                    <Button
                      onClick={handleImportExcel}
                      className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white"
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
                  {/* Row 1: NIK & Nama Lengkap */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="nik" className="text-sm font-medium text-gray-700">
                        NIK (NOMOR INDUK KARYAWAN) <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="nik"
                        required
                        placeholder="Contoh: 3210001234"
                        value={formData.nik}
                        onChange={(e) =>
                          setFormData({ ...formData, nik: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="fullName" className="text-sm font-medium text-gray-700">
                        NAMA LENGKAP <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="fullName"
                        required
                        placeholder="Masukkan nama sesuai KTP"
                        value={formData.fullName}
                        onChange={(e) =>
                          setFormData({ ...formData, fullName: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>

                  {/* Row 2: Tanggal Lahir & Agama */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="birthDate" className="text-sm font-medium text-gray-700">
                        TANGGAL LAHIR <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="birthDate"
                        type="date"
                        required
                        placeholder="mm/dd/yyyy"
                        value={formData.birthDate}
                        onChange={(e) =>
                          setFormData({ ...formData, birthDate: e.target.value })
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

                  {/* Row 3: Pendidikan & Tahun Masuk */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="lastEducation" className="text-sm font-medium text-gray-700">
                        PENDIDIKAN TERAKHIR <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.lastEducation}
                        onValueChange={(value) =>
                          setFormData({ ...formData, lastEducation: value })
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
                      <Label htmlFor="yearEnrolled" className="text-sm font-medium text-gray-700">
                        TAHUN MASUK <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="yearEnrolled"
                        required
                        placeholder="Contoh: 2023"
                        value={formData.yearEnrolled}
                        onChange={(e) =>
                          setFormData({ ...formData, yearEnrolled: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>

                  {/* Row 4: Status Kepegawaian & Departemen */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="employmentStatus" className="text-sm font-medium text-gray-700">
                        STATUS KEPEGAWAIAN <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.employmentStatus}
                        onValueChange={(value) =>
                          setFormData({ ...formData, employmentStatus: value })
                        }
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="Pilih Status" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="Tetap">Tetap</SelectItem>
                          <SelectItem value="Kontrak">Kontrak</SelectItem>
                          <SelectItem value="Magang">Magang</SelectItem>
                          <SelectItem value="Outsourcing">Outsourcing</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="department" className="text-sm font-medium text-gray-700">
                        DEPARTEMEN <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.department}
                        onValueChange={(value) =>
                          setFormData({ ...formData, department: value, position: "" })
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

                  {/* Row 5: Jabatan & Email Kantor */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="position" className="text-sm font-medium text-gray-700">
                        JABATAN <span className="text-red-500">*</span>
                      </Label>
                      <Select
                        value={formData.position}
                        onValueChange={(value) =>
                          setFormData({ ...formData, position: value })
                        }
                        disabled={!formData.department}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder={formData.department ? "Pilih Jabatan" : "Pilih Departemen Terlebih Dahulu"} />
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
                      <Label htmlFor="officeEmail" className="text-sm font-medium text-gray-700">
                        EMAIL KANTOR <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="officeEmail"
                        type="email"
                        required
                        placeholder="nama@perusahaan.com"
                        value={formData.officeEmail}
                        onChange={(e) =>
                          setFormData({ ...formData, officeEmail: e.target.value })
                        }
                        className="w-full"
                      />
                    </div>
                  </div>

                  {/* Row 6: Nomor Telepon & Role */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="space-y-2">
                      <Label htmlFor="phoneNumber" className="text-sm font-medium text-gray-700">
                        NOMOR TELEPON <span className="text-red-500">*</span>
                      </Label>
                      <Input
                        id="phoneNumber"
                        required
                        placeholder="+62 812 3456 7890"
                        value={formData.phoneNumber}
                        onChange={(e) =>
                          setFormData({ ...formData, phoneNumber: e.target.value })
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
                          <SelectItem value="manager_departemen">Manager Departemen</SelectItem>
                          <SelectItem value="admin_departemen">Admin Departemen</SelectItem>
                          <SelectItem value="manager_hr">Manager HR</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>

                  {/* Row 7: Alamat Lengkap */}
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
      <ImportExcelModal
        open={isImportModalOpen}
        onOpenChange={setIsImportModalOpen}
      />
    </>
  );
}
