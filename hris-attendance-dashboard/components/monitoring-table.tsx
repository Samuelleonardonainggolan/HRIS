import { Check, MapPin } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Employee } from "@/types";

interface MonitoringTableProps {
  employees: Employee[];
  loading?: boolean;
  emptyMessage?: string;
}

export function MonitoringTable({
  employees,
  loading = false,
  emptyMessage = "Belum ada data presensi hari ini",
}: MonitoringTableProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Monitoring Real-time</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Pegawai
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Departemen
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Status
                </th>
                <th className="pb-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  Verifikasi
                </th>
              </tr>
            </thead>
            {loading ? (
              <tbody>
                {[1, 2, 3, 4].map((i) => (
                  <tr key={i} className="border-b border-gray-100">
                    <td className="py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-gray-200 animate-pulse" />
                        <div className="space-y-1.5">
                          <div className="h-3.5 w-28 bg-gray-200 animate-pulse rounded" />
                          <div className="h-3 w-20 bg-gray-100 animate-pulse rounded" />
                        </div>
                      </div>
                    </td>
                    <td className="py-4">
                      <div className="h-3.5 w-20 bg-gray-200 animate-pulse rounded" />
                    </td>
                    <td className="py-4">
                      <div className="h-6 w-16 bg-gray-200 animate-pulse rounded-full" />
                    </td>
                    <td className="py-4">
                      <div className="h-6 w-12 bg-gray-200 animate-pulse rounded" />
                    </td>
                  </tr>
                ))}
              </tbody>
            ) : employees.length === 0 ? (
              <tbody>
                <tr>
                  <td
                    colSpan={4}
                    className="py-12 text-center text-sm text-gray-400"
                  >
                    {emptyMessage}
                  </td>
                </tr>
              </tbody>
            ) : (
              <tbody className="divide-y divide-gray-200">
                {employees.map((employee) => (
                  <tr key={employee.id} className="hover:bg-gray-50">
                    <td className="py-4">
                      <div className="flex items-center gap-3">
                        <Avatar>
                          <AvatarFallback
                            className={
                              employee.status === "HADIR"
                                ? "bg-teal-500 text-white"
                                : employee.status === "TELAT"
                                ? "bg-orange-500 text-white"
                                : employee.status === "TELAMBAT"
                                ? "bg-orange-500 text-white"
                                : "bg-blue-500 text-white"
                            }
                          >
                            {employee.avatar}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium text-gray-900">
                            {employee.name}
                          </div>
                          {employee.checkInTime && (
                            <div className="text-sm text-gray-500">
                              absen pukul {employee.checkInTime}
                            </div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="py-4">
                      <span className="text-sm text-gray-900">
                        {employee.department}
                      </span>
                    </td>
                    <td className="py-4">
                      <Badge
                        variant={
                          employee.status === "HADIR"
                            ? "success"
                            : employee.status === "TELAT"
                            ? "warning"
                            : "warning"
                        }
                      >
                        {employee.status}
                      </Badge>
                    </td>
                    <td className="py-4">
                      <div className="flex items-center gap-2">
                        {employee.verified?.biometric && (
                          <div className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-100">
                            <Check className="h-4 w-4 text-blue-600" />
                          </div>
                        )}
                        {employee.verified?.geofencing && (
                          <div className="flex h-6 w-6 items-center justify-center rounded-full bg-blue-100">
                            <MapPin className="h-4 w-4 text-blue-600" />
                          </div>
                        )}
                        {!employee.verified && (
                          <span className="text-xs text-gray-400">-</span>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            )}
          </table>
        </div>
      </CardContent>
    </Card>
  );
}
