"use client";

import { useState } from "react";
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

export default function AddDepartmentPage() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    code: "",
    name: "",
    description: "",
    icon: "🏢",
  });

  const iconOptions = [
    { value: "🏢", label: "🏢 Building" },
    { value: "🏨", label: "🏨 Hotel" },
    { value: "🍽️", label: "🍽️ Restaurant" },
    { value: "🛡️", label: "🛡️ Security" },
    { value: "💼", label: "💼 Business" },
    { value: "💰", label: "💰 Finance" },
    { value: "👥", label: "👥 Human Resources" },
    { value: "🔧", label: "🔧 Maintenance" },
    { value: "📊", label: "📊 Analytics" },
    { value: "🏭", label: "🏭 Operations" },
    { value: "🎯", label: "🎯 Target" },
    { value: "⚙️", label: "⚙️ Settings" },
  ];

  const handleBack = () => {
    router.back();
  };

  const handleCancel = () => {
    router.back();
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validation
    if (!formData.name.trim()) {
      alert("Nama departemen wajib diisi");
      return;
    }
    
    // TODO: Implement save logic (API call)
    console.log("Form data:", formData);
    
    // Show success message
    alert("Departemen berhasil ditambahkan!");
    
    // Redirect back to departments list
    router.push("/dashboard/manager-hr/departments");
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
            onClick={() => router.push("/dashboard/manager-hr/departments")}
            className="hover:text-blue-600 transition-colors"
          >
            Manajemen Departemen
          </button>
          <span>/</span>
          <span className="text-gray-900 font-medium">Tambah Departemen Baru</span>
        </div>

        {/* Main Card */}
        <Card>
          <CardContent className="p-6">
            {/* Header with Back Button */}
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-gray-200">
              <button
                onClick={handleBack}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <ArrowLeft className="h-5 w-5 text-gray-600" />
              </button>
              <h1 className="text-xl font-semibold text-gray-900">
                Tambah Departemen Baru
              </h1>
            </div>

            {/* Form */}
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Nama Departemen */}
              <div className="space-y-2">
                <Label htmlFor="name" className="text-sm font-medium text-gray-700">
                  NAMA DEPARTEMEN
                </Label>
                <Input
                  id="name"
                  placeholder="Contoh: Front Office"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value })
                  }
                  required
                  className="w-full"
                />
              </div>

              {/* Kode Departemen */}
              <div className="space-y-2">
                <Label htmlFor="code" className="text-sm font-medium text-gray-700">
                  KODE DEPARTEMEN
                </Label>
                <Input
                  id="code"
                  placeholder="Contoh: FO-01"
                  value={formData.code}
                  onChange={(e) =>
                    setFormData({ ...formData, code: e.target.value.toUpperCase() })
                  }
                  className="w-full"
                />
                <p className="text-xs text-gray-500">
                  Kode unik untuk departemen (opsional)
                </p>
              </div>

              {/* Deskripsi */}
              <div className="space-y-2">
                <Label htmlFor="description" className="text-sm font-medium text-gray-700">
                  DESKRIPSI
                </Label>
                <textarea
                  id="description"
                  rows={4}
                  placeholder="Jelaskan fungsi dan tanggung jawab departemen..."
                  value={formData.description}
                  onChange={(e) =>
                    setFormData({ ...formData, description: e.target.value })
                  }
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-gray-900 placeholder:text-gray-400"
                />
              </div>

              {/* Icon Departemen */}
              <div className="space-y-2">
                <Label htmlFor="icon" className="text-sm font-medium text-gray-700">
                  ICON DEPARTEMEN
                </Label>
                <Select
                  value={formData.icon}
                  onValueChange={(value) =>
                    setFormData({ ...formData, icon: value })
                  }
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Pilih Icon" />
                  </SelectTrigger>
                  <SelectContent>
                    {iconOptions.map((option) => (
                      <SelectItem key={option.value} value={option.value}>
                        <span className="text-base">{option.label}</span>
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-gray-500">
                  Icon akan ditampilkan di daftar departemen
                </p>
              </div>

              {/* Action Buttons */}
              <div className="flex justify-end gap-3 pt-6 border-t border-gray-200">
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
                  Simpan Departemen
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}