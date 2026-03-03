import {
  MapPin,
  Calendar,
  UserPlus,
  Download,
  Settings,
  HelpCircle,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export function ManagementPanel() {
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
          <Button variant="primary" className="w-full">
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
            6 pengajuan menunggu yang membutuhkan persetujuan segera Anda
          </CardDescription>
          <Button variant="success" className="w-full">
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
            >
              <UserPlus className="h-5 w-5" />
              <span className="text-xs">Tambah User</span>
            </Button>
            <Button
              variant="dark"
              className="flex flex-col items-center gap-2 h-auto py-4 bg-gray-800 hover:bg-gray-700"
            >
              <Download className="h-5 w-5" />
              <span className="text-xs">Ekspor Data</span>
            </Button>
            <Button
              variant="dark"
              className="flex flex-col items-center gap-2 h-auto py-4 bg-gray-800 hover:bg-gray-700"
            >
              <Settings className="h-5 w-5" />
              <span className="text-xs">Pengaturan</span>
            </Button>
            <Button
              variant="dark"
              className="flex flex-col items-center gap-2 h-auto py-4 bg-gray-800 hover:bg-gray-700"
            >
              <HelpCircle className="h-5 w-5" />
              <span className="text-xs">Bantuan</span>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}