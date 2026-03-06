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

import { useState } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Download, Upload } from "lucide-react";
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

export default function AddEmployeePage() {
  const router = useRouter();
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [formData, setFormData] = useState({
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
  });

  const handleBack = () => {
    router.back();
  };

  const handleCancel = () => {
    router.back();
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log("Form data:", formData);
  };

  const handleDownloadTemplate = () => {
    console.log("Download template");
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
              onClick={() => router.push("/dashboard/manager-hr/employees")}
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

              {/* Form Content - Same as before */}
              <div className="p-6">
                <form onSubmit={handleSubmit} className="space-y-6">
{/* Row 1: NIK & Nama Lengkap */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="nik" className="text-sm font-medium text-gray-700">
                      NIK (NOMOR INDUK KARYAWAN)
                    </Label>
                    <Input
                      id="nik"
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
                      NAMA LENGKAP
                    </Label>
                    <Input
                      id="fullName"
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
                      TANGGAL LAHIR
                    </Label>
                    <Input
                      id="birthDate"
                      type="date"
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
                      AGAMA
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
                        <SelectItem value="islam">Islam</SelectItem>
                        <SelectItem value="kristen">Kristen</SelectItem>
                        <SelectItem value="katolik">Katolik</SelectItem>
                        <SelectItem value="hindu">Hindu</SelectItem>
                        <SelectItem value="buddha">Buddha</SelectItem>
                        <SelectItem value="konghucu">Konghucu</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                {/* Row 3: Pendidikan & Tahun Masuk */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="lastEducation" className="text-sm font-medium text-gray-700">
                      PENDIDIKAN TERAKHIR
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
                        <SelectItem value="sd">SD</SelectItem>
                        <SelectItem value="smp">SMP</SelectItem>
                        <SelectItem value="sma">SMA/SMK</SelectItem>
                        <SelectItem value="d3">D3</SelectItem>
                        <SelectItem value="s1">S1</SelectItem>
                        <SelectItem value="s2">S2</SelectItem>
                        <SelectItem value="s3">S3</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="yearEnrolled" className="text-sm font-medium text-gray-700">
                      TAHUN MASUK
                    </Label>
                    <Input
                      id="yearEnrolled"
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
                      STATUS KEPEGAWAIAN
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
                        <SelectItem value="tetap">Tetap</SelectItem>
                        <SelectItem value="kontrak">Kontrak</SelectItem>
                        <SelectItem value="magang">Magang</SelectItem>
                        <SelectItem value="outsourcing">Outsourcing</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="department" className="text-sm font-medium text-gray-700">
                      DEPARTEMEN
                    </Label>
                    <Select
                      value={formData.department}
                      onValueChange={(value) =>
                        setFormData({ ...formData, department: value })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Pilih Departemen" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="hr">Human Resources</SelectItem>
                        <SelectItem value="it">Information Technology</SelectItem>
                        <SelectItem value="finance">Finance & Accounting</SelectItem>
                        <SelectItem value="marketing">Sales & Marketing</SelectItem>
                        <SelectItem value="operations">Operations</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>

                {/* Row 5: Jabatan & Email Kantor */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="position" className="text-sm font-medium text-gray-700">
                      JABATAN
                    </Label>
                    <Select
                      value={formData.position}
                      onValueChange={(value) =>
                        setFormData({ ...formData, position: value })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="Pilih Jabatan" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="staff">Staff</SelectItem>
                        <SelectItem value="supervisor">Supervisor</SelectItem>
                        <SelectItem value="manager">Manager</SelectItem>
                        <SelectItem value="director">Director</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="officeEmail" className="text-sm font-medium text-gray-700">
                      EMAIL KANTOR
                    </Label>
                    <Input
                      id="officeEmail"
                      type="email"
                      placeholder="nama@perusahaan.com"
                      value={formData.officeEmail}
                      onChange={(e) =>
                        setFormData({ ...formData, officeEmail: e.target.value })
                      }
                      className="w-full"
                    />
                  </div>
                </div>

                {/* Row 6: Nomor Telepon */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="space-y-2">
                    <Label htmlFor="phoneNumber" className="text-sm font-medium text-gray-700">
                      NOMOR TELEPON
                    </Label>
                    <Input
                      id="phoneNumber"
                      placeholder="+62 812 3456 7890"
                      value={formData.phoneNumber}
                      onChange={(e) =>
                        setFormData({ ...formData, phoneNumber: e.target.value })
                      }
                      className="w-full"
                    />
                  </div>
                </div>

                {/* Row 7: Alamat Lengkap */}
                <div className="space-y-2">
                  <Label htmlFor="address" className="text-sm font-medium text-gray-700">
                    ALAMAT LENGKAP
                  </Label>
                  <textarea
                    id="address"
                    rows={3}
                    placeholder="Masukkan alamat lengkap sesuai KTP"
                    value={formData.address}
                    onChange={(e) =>
                      setFormData({ ...formData, address: e.target.value })
                    }
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-gray-900 placeholder:text-gray-400"
                  />
                </div>
                  <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={handleCancel}
                      className="px-6"
                    >
                      Batal
                    </Button>
                    <Button
                      type="submit"
                      className="px-6 bg-blue-600 hover:bg-blue-700 text-white"
                    >
                      Simpan Pegawai
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