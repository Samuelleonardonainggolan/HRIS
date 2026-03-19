import { Card, CardContent } from "@/components/ui/card";

type Props = {
  avgPresentPct: number;
  totalEmployees: number;
  totalLateIncidents: number;
  totalAbsentIncidents: number;
};

export function AttendanceKpiCards({
  avgPresentPct,
  totalEmployees,
  totalLateIncidents,
  totalAbsentIncidents,
}: Props) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      <Card className="border-blue-100 bg-blue-50/60">
        <CardContent className="p-6">
          <div className="text-xs font-semibold tracking-wide text-blue-700">RATA-RATA HADIR</div>
          <div className="mt-1 text-2xl font-bold text-blue-700">{avgPresentPct.toFixed(1)}%</div>
        </CardContent>
      </Card>
      <Card className="border-green-100 bg-green-50/60">
        <CardContent className="p-6">
          <div className="text-xs font-semibold tracking-wide text-green-700">TOTAL KARYAWAN</div>
          <div className="mt-1 text-2xl font-bold text-green-700">{totalEmployees} Orang</div>
        </CardContent>
      </Card>
      <Card className="border-orange-100 bg-orange-50/60">
        <CardContent className="p-6">
          <div className="text-xs font-semibold tracking-wide text-orange-700">TOTAL KETERLAMBATAN</div>
          <div className="mt-1 text-2xl font-bold text-orange-700">{totalLateIncidents} Insiden</div>
        </CardContent>
      </Card>
      <Card className="border-red-100 bg-red-50/60">
        <CardContent className="p-6">
          <div className="text-xs font-semibold tracking-wide text-red-700">TOTAL ALPA</div>
          <div className="mt-1 text-2xl font-bold text-red-700">{totalAbsentIncidents} Insiden</div>
        </CardContent>
      </Card>
    </div>
  );
}

