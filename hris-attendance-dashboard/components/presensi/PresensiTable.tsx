import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import type { PresensiEmployeeSummary } from "@/components/presensi/types";

type Props = {
  rows: PresensiEmployeeSummary[];
};

function getInitials(name: string) {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  const first = parts[0]?.[0] || "U";
  const second = parts[1]?.[0] || "";
  return (first + second).toUpperCase();
}

function calcPercent(row: PresensiEmployeeSummary) {
  if (row.workDays <= 0) return 0;
  return (row.present / row.workDays) * 100;
}

function percentColor(percent: number) {
  if (percent >= 90) return "text-green-600";
  if (percent >= 80) return "text-orange-600";
  return "text-red-600";
}

export function PresensiTable({ rows }: Props) {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full">
        <thead>
          <tr className="border-b border-gray-200">
            <th className="px-6 py-3 text-left text-[11px] font-semibold tracking-wide text-gray-500">
              NIK
            </th>
            <th className="px-6 py-3 text-left text-[11px] font-semibold tracking-wide text-gray-500">
              NAMA KARYAWAN
            </th>
            <th className="px-6 py-3 text-left text-[11px] font-semibold tracking-wide text-gray-500">
              DEPARTEMEN
            </th>
            <th className="px-6 py-3 text-right text-[11px] font-semibold tracking-wide text-gray-500">
              HARI KERJA
            </th>
            <th className="px-6 py-3 text-right text-[11px] font-semibold tracking-wide text-gray-500">
              HADIR
            </th>
            <th className="px-6 py-3 text-right text-[11px] font-semibold tracking-wide text-gray-500">
              TELAT
            </th>
            <th className="px-6 py-3 text-right text-[11px] font-semibold tracking-wide text-gray-500">
              IZIN/SAKIT
            </th>
            <th className="px-6 py-3 text-right text-[11px] font-semibold tracking-wide text-gray-500">
              ALPA
            </th>
            <th className="px-6 py-3 text-right text-[11px] font-semibold tracking-wide text-gray-500">
              PERSENTASE
            </th>
          </tr>
        </thead>

        <tbody className="divide-y divide-gray-100">
          {rows.map((row) => {
            const percent = calcPercent(row);
            return (
              <tr key={row.nik} className="hover:bg-gray-50">
                <td className="px-6 py-4 text-sm text-gray-700">{row.nik}</td>
                <td className="px-6 py-4">
                  <div className="flex items-center gap-3">
                    <Avatar className="h-8 w-8">
                      <AvatarFallback className="text-xs">
                        {getInitials(row.name)}
                      </AvatarFallback>
                    </Avatar>
                    <div className="text-sm font-semibold text-gray-900">
                      {row.name}
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-600">
                  {row.department}
                </td>
                <td className="px-6 py-4 text-right text-sm font-semibold text-gray-900">
                  {row.workDays}
                </td>
                <td className="px-6 py-4 text-right text-sm text-gray-900">
                  {row.present}
                </td>
                <td className="px-6 py-4 text-right text-sm text-gray-900">
                  {row.late > 0 ? (
                    <Badge variant="warning" className="ml-auto w-fit rounded-full">
                      {row.late}
                    </Badge>
                  ) : (
                    "0"
                  )}
                </td>
                <td className="px-6 py-4 text-right text-sm text-gray-900">
                  {row.leaveSick}
                </td>
                <td className="px-6 py-4 text-right text-sm text-gray-900">
                  {row.absent > 0 ? (
                    <Badge variant="danger" className="ml-auto w-fit rounded-full">
                      {row.absent}
                    </Badge>
                  ) : (
                    "0"
                  )}
                </td>
                <td className={"px-6 py-4 text-right text-sm font-semibold " + percentColor(percent)}>
                  {percent.toFixed(1)}%
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

