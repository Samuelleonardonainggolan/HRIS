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
import { departmentApi } from "@/lib/api/department";
import toast from "react-hot-toast";

export default function AddDepartmentPage() {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
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

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const editId = params.get("edit");
    if (editId) {
      setEditingId(editId);
      loadDepartment(editId);
    }
  }, []);

  const loadDepartment = async (id: string) => {
    try {
      const d = await departmentApi.getById(id);
      setFormData({
        code: d.code || "",
        name: d.name || "",
        description: d.description || "",
        icon: d.icon || "🏢",
      });
    } catch (err) {
      console.error(err);
      toast.error("Gagal memuat data departemen");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    if (!formData.name.trim()) {
      setError("Nama departemen wajib diisi");
      setIsSubmitting(false);
      return;
    }

    try {
      if (editingId) {
        await departmentApi.update(editingId, {
          code: formData.code || undefined,
          name: formData.name,
          description: formData.description || undefined,
          icon: formData.icon || undefined,
        });
        toast.success("Departemen berhasil diperbarui");
      } else {
        await departmentApi.create({
          code: formData.code || undefined,
          name: formData.name,
          description: formData.description || undefined,
          icon: formData.icon || undefined,
        });
        toast.success("Departemen berhasil ditambahkan");
      }
      router.push("/dashboard/manager-hr/departemen");
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Gagal menyimpan departemen";
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
                {error && (
                  <div className="mb-2 text-sm text-red-600">{error}</div>
                )}
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
                  disabled={isSubmitting}
                >
                  {isSubmitting ? "Menyimpan..." : editingId ? "Simpan Perubahan" : "Simpan Departemen"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
