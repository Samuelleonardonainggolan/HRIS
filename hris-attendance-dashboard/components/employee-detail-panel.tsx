"use client";

import { X, Mail, Phone, MapPin } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Employee } from "@/types";

interface EmployeeDetailPanelProps {
  employee: Employee | null;
  onClose: () => void;
}

export function EmployeeDetailPanel({
  employee,
  onClose,
}: EmployeeDetailPanelProps) {
  if (!employee) return null;

  return (
    <Card className="h-fit overflow-hidden shadow-lg border-gray-200 sticky top-0">
      <div className="overflow-y-auto max-h-[calc(100vh-8rem)]">
        {/* Profile Card */}
        <div className="relative bg-gradient-to-br from-gray-800 to-gray-900">
          <div className="p-6 text-center">
            {/* Close Button */}
            <button
              onClick={onClose}
              className="absolute top-4 right-4 text-white/60 hover:text-white transition-colors"
            >
              <X className="h-5 w-5" />
            </button>

            <div className="relative inline-block mb-4">
              <Avatar className="h-20 w-20">
                <AvatarFallback className="bg-teal-500 text-white text-2xl">
                  {employee.avatar}
                </AvatarFallback>
              </Avatar>
              <div className="absolute bottom-0 right-0 h-5 w-5 rounded-full bg-green-500 border-2 border-gray-900"></div>
            </div>
            <h2 className="text-xl font-bold text-white mb-1">
              {employee.name}
            </h2>
            <p className="text-sm text-gray-300">
              {employee.position} • {employee.department}
            </p>
          </div>
        </div>

        {/* Content Area */}
        <div className="p-6 space-y-8 bg-white">
          {/* Informasi Pribadi */}
          <div>
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-4">
              Informasi Pribadi
            </h3>
            <div className="space-y-4">
              {/* Email */}
              <div className="flex items-start gap-3">
                <Mail className="h-5 w-5 text-blue-500 shrink-0 mt-0.5" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">
                    Email
                  </p>
                  <p className="text-sm text-gray-900 font-normal break-all">
                    {employee.email || "Tidak tersedia"}
                  </p>
                </div>
              </div>

              {/* Telepon */}
              <div className="flex items-start gap-3">
                <Phone className="h-5 w-5 text-blue-500 shrink-0 mt-0.5" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">
                    Telepon
                  </p>
                  <p className="text-sm text-gray-900 font-normal">
                    {employee.phone || "Tidak tersedia"}
                  </p>
                </div>
              </div>

              {/* Alamat */}
              <div className="flex items-start gap-3">
                <MapPin className="h-5 w-5 text-blue-500 shrink-0 mt-0.5" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">
                    Alamat
                  </p>
                  <p className="text-sm text-gray-900 font-normal leading-relaxed">
                    {employee.address || "Tidak tersedia"}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Informasi Pekerjaan */}
          <div>
            <h3 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-4">
              Informasi Pekerjaan
            </h3>
            <div className="space-y-5">
              {/* Tanggal Masuk */}
              <div>
                <p className="text-xs text-gray-400 uppercase tracking-wide mb-1.5">
                  Tanggal Masuk
                </p>
                <p className="text-base text-gray-900 font-semibold mb-1">
                  {employee.joinDate || "Tidak tersedia"}
                </p>
                {employee.workYears && (
                  <p className="text-xs text-blue-600 font-normal">
                    {employee.workYears} Tahun • Masa Kerja
                  </p>
                )}
              </div>

              {/* Status Kepegawaian */}
              <div>
                <p className="text-xs text-gray-400 uppercase tracking-wide mb-1.5">
                  Status Kepegawaian
                </p>
                <p className="text-base text-gray-900 font-semibold mb-1">
                  {employee.employmentStatus || "Tidak tersedia"}
                </p>
                <p className="text-xs text-gray-400 font-normal">
                  Grade Senior II
                </p>
              </div>
            </div>
          </div>

          {/* Edit Button */}
          <Button
            variant="outline"
            className="w-full border-gray-300 text-gray-700 hover:bg-gray-50 font-medium"
          >
            Edit Data
          </Button>
        </div>
      </div>
    </Card>
  );
}