import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import type { AttendanceEmployeeSummary } from "@/components/presensi/attendance-mock";

function initials(name: string) {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "?";
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[1][0]).toUpperCase();
}

function percentColor(pct: number) {
  if (pct >= 90) return "text-green-600";
  if (pct >= 85) return "text-orange-600";
  return "text-red-600";
}

export function AttendanceSummaryTable({ rows }: { rows: AttendanceEmployeeSummary[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full">
        <thead>
          <tr className="border-b border-gray-200 text-xs font-semibold uppercase tracking-wide text-gray-500">
            <th className="px-6 py-3 text-left">NIK</th>
            <th className="px-6 py-3 text-left">Nama Karyawan</th>
            <th className="px-6 py-3 text-left">Departemen</th>
            <th className="px-6 py-3 text-right">Hari Kerja</th>
            <th className="px-6 py-3 text-right">Hadir</th>
            <th className="px-6 py-3 text-right">Telat</th>
            <th className="px-6 py-3 text-right">Izin/Sakit</th>
            <th className="px-6 py-3 text-right">Alpa</th>
            <th className="px-6 py-3 text-right">Persentase</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100">
          {rows.map((r) => {
            const pct = r.workDays > 0 ? (r.present / r.workDays) * 100 : 0;
            return (
              <tr key={r.nik} className="hover:bg-gray-50 transition-colors">
                <td className="px-6 py-4 text-sm text-gray-700">{r.nik}</td>
                <td className="px-6 py-4">
                  <div className="flex items-center gap-3">
                    <Avatar className="h-8 w-8">
                      <AvatarFallback className="bg-gray-200 text-gray-700 text-xs">
                        {initials(r.name)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="text-sm font-semibold text-gray-900">{r.name}</div>
                  </div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600">{r.department}</td>
                <td className="px-6 py-4 text-sm font-semibold text-gray-900 text-right">
                  {r.workDays}
                </td>
                <td className="px-6 py-4 text-sm text-gray-900 text-right">{r.present}</td>
                <td className="px-6 py-4 text-right">
                  {r.late > 0 ? (
                    <Badge variant="warning" className="justify-center min-w-7">
                      {r.late}
                    </Badge>
                  ) : (
                    <span className="text-sm text-gray-900">0</span>
                  )}
                </td>
                <td className="px-6 py-4 text-sm text-gray-900 text-right">{r.sickLeave}</td>
                <td className="px-6 py-4 text-right">
                  {r.absent > 0 ? (
                    <Badge variant="danger" className="justify-center min-w-7">
                      {r.absent}
                    </Badge>
                  ) : (
                    <span className="text-sm text-gray-400">0</span>
                  )}
                </td>
                <td className={cn("px-6 py-4 text-sm font-semibold text-right", percentColor(pct))}>
                  {pct.toFixed(1).replace(/\.0$/, "")}%
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

