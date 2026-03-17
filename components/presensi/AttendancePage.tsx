"use client";

import * as React from "react";
import { CalendarDays, FileSpreadsheet, FileText, Filter } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  attendanceKpiMock,
  attendanceSummaryMock,
  departmentsMock,
} from "@/components/presensi/attendance-mock";
import { AttendanceSummaryTable } from "@/components/presensi/AttendanceSummaryTable";
import { Pagination } from "@/components/presensi/pagination";
import { AttendanceKpiCards } from "@/components/presensi/AttendanceKpiCards";

type Props = {
  title?: string;
};

export function AttendancePage({ title = "Presensi" }: Props) {
  const [monthLabel, setMonthLabel] = React.useState("Mei 2024 (01/05 - 31/05)");
  const [department, setDepartment] = React.useState("Semua Departemen");
  const [appliedMonthLabel, setAppliedMonthLabel] = React.useState(monthLabel);
  const [appliedDepartment, setAppliedDepartment] = React.useState(department);
  const [page, setPage] = React.useState(1);

  const pageSize = 4;
  const totalEmployees = attendanceKpiMock.totalEmployees;
  const totalPages = Math.ceil(totalEmployees / pageSize);

  const rows = React.useMemo(() => {
    const filtered =
      appliedDepartment === "Semua Departemen"
        ? attendanceSummaryMock
        : attendanceSummaryMock.filter((r) => r.department === appliedDepartment);

    const start = (page - 1) * pageSize;
    return filtered.slice(start, start + pageSize);
  }, [appliedDepartment, page]);

  const onApply = () => {
    setAppliedMonthLabel(monthLabel);
    setAppliedDepartment(department);
    setPage(1);
  };

  const onExportPdf = () => {
    window.alert("Export PDF (mock)");
  };

  const onExportExcel = () => {
    window.alert("Export Excel (mock)");
  };

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">{title}</h1>
        <p className="text-sm text-gray-600">Ringkasan presensi karyawan per periode</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2 bg-gray-100/70">
          <CardContent className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 items-end">
              <div>
                <div className="text-xs font-semibold tracking-wide text-gray-500">FILTER TANGGAL</div>
                <Button
                  type="button"
                  variant="outline"
                  className="mt-2 w-full justify-between rounded-full"
                  onClick={() => window.alert("Date picker (mock)")}
                >
                  <span className="flex items-center gap-2">
                    <CalendarDays className="h-4 w-4 text-gray-500" />
                    <span className="text-sm text-gray-900">{monthLabel}</span>
                  </span>
                  <span className="text-gray-400">▼</span>
                </Button>
              </div>

              <div>
                <div className="text-xs font-semibold tracking-wide text-gray-500">DEPARTEMEN</div>
                <Select value={department} onValueChange={setDepartment}>
                  <SelectTrigger className="mt-2 rounded-full">
                    <SelectValue placeholder="Pilih departemen" />
                  </SelectTrigger>
                  <SelectContent>
                    {departmentsMock.map((d) => (
                      <SelectItem key={d} value={d}>
                        {d}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="flex md:justify-end">
                <Button
                  type="button"
                  variant="primary"
                  className="w-full md:w-auto rounded-full"
                  onClick={onApply}
                >
                  <Filter className="h-4 w-4 mr-2" />
                  Terapkan
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-gray-100/70">
          <CardContent className="p-6">
            <div className="text-xs font-semibold tracking-wide text-gray-500">EXPORT LAPORAN</div>
            <div className="mt-2 flex flex-col sm:flex-row gap-3">
              <Button
                type="button"
                variant="outline"
                className="rounded-full justify-center"
                onClick={onExportPdf}
              >
                <FileText className="h-4 w-4 mr-2 text-red-600" />
                Unduh PDF
              </Button>
              <Button
                type="button"
                variant="outline"
                className="rounded-full justify-center"
                onClick={onExportExcel}
              >
                <FileSpreadsheet className="h-4 w-4 mr-2 text-green-600" />
                Export Excel
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="px-6 py-4 border-b border-gray-200">
            <div className="text-sm font-semibold text-gray-900">Ringkasan Presensi Karyawan</div>
            <div className="text-xs text-gray-500 mt-1">
              {appliedMonthLabel} • {appliedDepartment}
            </div>
          </div>
          <AttendanceSummaryTable rows={rows} />
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 px-6 py-4 border-t border-gray-200">
            <div className="text-xs text-gray-500">
              Menampilkan {rows.length} dari {totalEmployees} karyawan
            </div>
            <div className="flex justify-end">
              <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />
            </div>
          </div>
        </CardContent>
      </Card>

      <AttendanceKpiCards
        avgPresentPct={attendanceKpiMock.avgPresentPct}
        totalEmployees={attendanceKpiMock.totalEmployees}
        totalLateIncidents={attendanceKpiMock.totalLateIncidents}
        totalAbsentIncidents={attendanceKpiMock.totalAbsentIncidents}
      />
    </div>
  );
}

