"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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

interface DepartmentModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  department?: {
    id: string;
    name: string;
    icon: string;
    managerName: string;
    totalStaff: number;
    status: string;
  };
}

export function DepartmentModal({
  open,
  onOpenChange,
  department,
}: DepartmentModalProps) {
  const isEdit = !!department;
  const [formData, setFormData] = useState({
    name: department?.name || "",
    icon: department?.icon || "🏢",
    managerName: department?.managerName || "",
    status: department?.status || "Aktif",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // TODO: Implement save logic
    console.log("Form data:", formData);
    onOpenChange(false);
  };

  const handleCancel = () => {
    onOpenChange(false);
  };

  const iconOptions = [
    { value: "🏢", label: "🏢 Building" },
    { value: "🏨", label: "🏨 Hotel" },
    { value: "🍽️", label: "🍽️ Restaurant" },
    { value: "🛡️", label: "🛡️ Security" },
    { value: "💼", label: "💼 Business" },
    { value: "🏭", label: "🏭 Factory" },
    { value: "🏪", label: "🏪 Store" },
    { value: "🏥", label: "🏥 Medical" },
  ];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? "Edit Departemen" : "Tambah Departemen Baru"}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6 py-4">
          {/* Department Name */}
          <div className="space-y-2">
            <Label htmlFor="name" className="text-sm font-medium text-gray-700">
              Nama Departemen
            </Label>
            <Input
              id="name"
              placeholder="Contoh: Human Resources"
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
              required
            />
          </div>

          {/* Icon */}
          <div className="space-y-2">
            <Label htmlFor="icon" className="text-sm font-medium text-gray-700">
              Icon Departemen
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
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Manager Name */}
          <div className="space-y-2">
            <Label
              htmlFor="managerName"
              className="text-sm font-medium text-gray-700"
            >
              Nama Kepala Departemen
            </Label>
            <Input
              id="managerName"
              placeholder="Contoh: Budi Santoso"
              value={formData.managerName}
              onChange={(e) =>
                setFormData({ ...formData, managerName: e.target.value })
              }
              required
            />
          </div>

          {/* Status */}
          <div className="space-y-2">
            <Label htmlFor="status" className="text-sm font-medium text-gray-700">
              Status
            </Label>
            <Select
              value={formData.status}
              onValueChange={(value) =>
                setFormData({ ...formData, status: value })
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Pilih Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Aktif">Aktif</SelectItem>
                <SelectItem value="Nonaktif">Nonaktif</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3 pt-4 border-t">
            <Button type="button" variant="outline" onClick={handleCancel}>
              Batal
            </Button>
            <Button
              type="submit"
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              {isEdit ? "Simpan Perubahan" : "Tambah Departemen"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}