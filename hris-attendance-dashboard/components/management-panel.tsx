"use client";

import { useRouter } from "next/navigation";
import {
  MapPin,
  Calendar,
  UserPlus,
  Download,
  ChevronDown,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

interface ManagementPanelProps {
  pendingLeaveCount?: number;
  loadingLeave?: boolean;
  onExportCsv?: () => void;
  onExportPdf?: () => void;
}

export function ManagementPanel({
  pendingLeaveCount = 0,
  loadingLeave = false,
  onExportCsv,
  onExportPdf,
}: ManagementPanelProps) {
  const router = useRouter();

  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold text-gray-900">Manajemen</h2>

      {/* Konfigurasi Geofencing */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
              <MapPin className="h-5 w-5 text-blue-600" />
            </div>
            <div>
              <CardTitle className="text-base">
                Konfigurasi Geofencing
              </CardTitle>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <CardDescription className="mb-4">
            Atur lokasi yang sah untuk absensi dengan radius lokasi wilayah
          </CardDescription>
          <Button
            variant="primary"
            className="w-full"
            onClick={() => router.push("/dashboard/geofencing")}
          >
            Konfigurasi Zona
          </Button>
        </CardContent>
      </Card>

      {/* Persetujuan Cuti */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100">
              <Calendar className="h-5 w-5 text-green-600" />
            </div>
            <div>
              <CardTitle className="text-base">Persetujuan Cuti</CardTitle>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <CardDescription className="mb-4">
            {loadingLeave ? (
              <span className="inline-block h-4 w-40 bg-gray-200 animate-pulse rounded" />
            ) : pendingLeaveCount > 0
              ? `${pendingLeaveCount} pengajuan menunggu yang membutuhkan persetujuan segera Anda`
              : "Tidak ada pengajuan yang menunggu saat ini"}
          </CardDescription>
          <Button
            variant="success"
            className="w-full"
            onClick={() =>
              router.push("/dashboard/manager-hr/persetujuan-izin-cuti")
            }
          >
            Tinjau Pengajuan
          </Button>
        </CardContent>
      </Card>

      {/* Pintasan Cepat */}
      <Card className="bg-gray-900">
        <CardHeader>
          <CardTitle className="text-white">Pintasan Cepat</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-3">
            <Button
              variant="dark"
              className="flex flex-col items-center gap-2 h-auto py-4 bg-gray-800 hover:bg-gray-700"
              onClick={() =>
                router.push("/dashboard/manager-hr/karyawan/tambah-pegawai")
              }
            >
              <UserPlus className="h-5 w-5" />
              <span className="text-xs">Tambah User</span>
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="dark"
                  className="flex flex-col items-center gap-2 h-auto py-4 bg-gray-800 hover:bg-gray-700 w-full"
                >
                  <Download className="h-5 w-5" />
                  <div className="flex items-center gap-1">
                    <span className="text-xs">Ekspor Data</span>
                    <ChevronDown className="h-3 w-3 opacity-50" />
                  </div>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="center" className="w-40">
                <DropdownMenuItem onClick={onExportCsv} disabled={!onExportCsv}>
                  Export CSV
                </DropdownMenuItem>
                <DropdownMenuItem onClick={onExportPdf} disabled={!onExportPdf}>
                  Export PDF
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
