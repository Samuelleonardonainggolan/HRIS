import { Card } from "@/components/ui/card";
import type { PresensiKpi } from "@/components/presensi/types";

type Props = {
  kpi: PresensiKpi;
};

export function PresensiKpiCards({ kpi }: Props) {
  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-4">
      <Card className="border-blue-100 bg-blue-50 p-4 shadow-sm">
        <div className="text-[11px] font-semibold tracking-wide text-blue-700">
          RATA-RATA HADIR
        </div>
        <div className="mt-1 text-2xl font-extrabold text-blue-800">
          {kpi.averageAttendancePercent.toFixed(1)}%
        </div>
      </Card>

      <Card className="border-green-100 bg-green-50 p-4 shadow-sm">
        <div className="text-[11px] font-semibold tracking-wide text-green-700">
          TOTAL KARYAWAN
        </div>
        <div className="mt-1 text-2xl font-extrabold text-green-800">
          {kpi.totalEmployees} Orang
        </div>
      </Card>

      <Card className="border-orange-100 bg-orange-50 p-4 shadow-sm">
        <div className="text-[11px] font-semibold tracking-wide text-orange-700">
          TOTAL KETERLAMBATAN
        </div>
        <div className="mt-1 text-2xl font-extrabold text-orange-800">
          {kpi.totalLateIncidents} Insiden
        </div>
      </Card>

      <Card className="border-red-100 bg-red-50 p-4 shadow-sm">
        <div className="text-[11px] font-semibold tracking-wide text-red-700">
          TOTAL ALPA
        </div>
        <div className="mt-1 text-2xl font-extrabold text-red-800">
          {kpi.totalAbsentIncidents} Insiden
        </div>
      </Card>
    </div>
  );
}

