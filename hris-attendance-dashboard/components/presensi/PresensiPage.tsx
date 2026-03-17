"use client";

import { useMemo, useState } from "react";
import { Card } from "@/components/ui/card";
import { PresensiKpiCards } from "@/components/presensi/PresensiKpiCards";
import { PresensiPagination } from "@/components/presensi/PresensiPagination";
import { PresensiTable } from "@/components/presensi/PresensiTable";
import { PresensiToolbar } from "@/components/presensi/PresensiToolbar";
import type { PresensiEmployeeSummary, PresensiKpi } from "@/components/presensi/types";

function buildCsv(rows: PresensiEmployeeSummary[]) {
  const header = [
    "NIK",
    "Nama",
    "Departemen",
    "Hari Kerja",
    "Hadir",
    "Telat",
    "Izin/Sakit",
    "Alpa",
    "Persentase",
  ];

  const lines = rows.map((r) => {
    const percent = r.workDays > 0 ? (r.present / r.workDays) * 100 : 0;
    return [
      r.nik,
      r.name,
      r.department,
      String(r.workDays),
      String(r.present),
      String(r.late),
      String(r.leaveSick),
      String(r.absent),
      percent.toFixed(1) + "%",
    ]
      .map((v) => `"${String(v).replaceAll('"', '""')}"`)
      .join(",");
  });

  return [header.join(","), ...lines].join("\n");
}

function downloadBlob(filename: string, content: BlobPart, mime: string) {
  const blob = new Blob([content], { type: mime });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

export function PresensiPage() {
  const [monthRangeLabel, setMonthRangeLabel] = useState(
    "Mei 2024 (01/05 - 31/05)"
  );
  const [department, setDepartment] = useState("Semua Departemen");
  const [page, setPage] = useState(1);

  const allRows = useMemo<PresensiEmployeeSummary[]>(
    () => [
      {
        nik: "ID-99210",
        name: "Rizky Darmawan",
        department: "IT Support",
        workDays: 22,
        present: 21,
        late: 1,
        leaveSick: 0,
        absent: 0,
      },
      {
        nik: "ID-99215",
        name: "Siti Aminah",
        department: "Human Resource",
        workDays: 22,
        present: 22,
        late: 0,
        leaveSick: 0,
        absent: 0,
      },
      {
        nik: "ID-99302",
        name: "Bambang Pamungkas",
        department: "Marketing",
        workDays: 22,
        present: 18,
        late: 2,
        leaveSick: 2,
        absent: 2,
      },
      {
        nik: "ID-99310",
        name: "Dewi Lestari",
        department: "Finance",
        workDays: 22,
        present: 21,
        late: 0,
        leaveSick: 1,
        absent: 0,
      },
    ],
    []
  );

  const departments = useMemo(() => {
    const set = new Set<string>(["Semua Departemen"]);
    for (const r of allRows) set.add(r.department);
    return Array.from(set);
  }, [allRows]);

  const filteredRows = useMemo(() => {
    if (department === "Semua Departemen") return allRows;
    return allRows.filter((r) => r.department === department);
  }, [allRows, department]);

  const totalEmployees = 120;
  const pageSize = 10;
  const totalPages = Math.max(1, Math.ceil(totalEmployees / pageSize));

  const kpi = useMemo<PresensiKpi>(() => {
    const averageAttendancePercent = 92.4;
    const totalLateIncidents = 14;
    const totalAbsentIncidents = 5;
    return {
      averageAttendancePercent,
      totalEmployees,
      totalLateIncidents,
      totalAbsentIncidents,
    };
  }, [totalEmployees]);

  const onApply = () => {
    setPage(1);
  };

  const onExportExcel = () => {
    const csv = buildCsv(filteredRows);
    downloadBlob("presensi.csv", csv, "text/csv;charset=utf-8");
  };

  const onExportPdf = () => {
    const html = `
      <html>
        <head>
          <meta charset="utf-8" />
          <title>Presensi</title>
          <style>
            body { font-family: Arial, sans-serif; padding: 24px; }
            h1 { font-size: 18px; margin: 0 0 12px 0; }
            .meta { color: #6b7280; font-size: 12px; margin-bottom: 16px; }
            table { width: 100%; border-collapse: collapse; }
            th, td { border: 1px solid #e5e7eb; padding: 8px; font-size: 12px; }
            th { background: #f9fafb; text-align: left; }
          </style>
        </head>
        <body>
          <h1>Ringkasan Presensi</h1>
          <div class="meta">${monthRangeLabel} • ${department}</div>
          <table>
            <thead>
              <tr>
                <th>NIK</th>
                <th>Nama</th>
                <th>Departemen</th>
                <th>Hari Kerja</th>
                <th>Hadir</th>
                <th>Telat</th>
                <th>Izin/Sakit</th>
                <th>Alpa</th>
                <th>Persentase</th>
              </tr>
            </thead>
            <tbody>
              ${filteredRows
                .map((r) => {
                  const percent = r.workDays > 0 ? (r.present / r.workDays) * 100 : 0;
                  return `
                    <tr>
                      <td>${r.nik}</td>
                      <td>${r.name}</td>
                      <td>${r.department}</td>
                      <td>${r.workDays}</td>
                      <td>${r.present}</td>
                      <td>${r.late}</td>
                      <td>${r.leaveSick}</td>
                      <td>${r.absent}</td>
                      <td>${percent.toFixed(1)}%</td>
                    </tr>
                  `;
                })
                .join("")}
            </tbody>
          </table>
        </body>
      </html>
    `;

    const win = window.open("", "_blank");
    if (!win) return;
    win.document.open();
    win.document.write(html);
    win.document.close();
    win.focus();
    win.print();
  };

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Presensi</h1>
        <div className="text-sm text-gray-600">Ringkasan presensi per karyawan</div>
      </div>

      <PresensiToolbar
        monthRangeLabel={monthRangeLabel}
        selectedDepartment={department}
        departments={departments}
        onChangeMonthRangeLabel={setMonthRangeLabel}
        onChangeDepartment={setDepartment}
        onApply={onApply}
        onExportPdf={onExportPdf}
        onExportExcel={onExportExcel}
      />

      <Card className="overflow-hidden">
        <PresensiTable rows={filteredRows} />
        <div className="flex flex-col gap-3 border-t border-gray-200 bg-gray-50 px-6 py-4 md:flex-row md:items-center md:justify-between">
          <div className="text-xs text-gray-600">
            Menampilkan {filteredRows.length} dari {totalEmployees} karyawan
          </div>
          <PresensiPagination page={page} totalPages={totalPages} onChange={setPage} />
        </div>
      </Card>

      <PresensiKpiCards kpi={kpi} />
    </div>
  );
}

