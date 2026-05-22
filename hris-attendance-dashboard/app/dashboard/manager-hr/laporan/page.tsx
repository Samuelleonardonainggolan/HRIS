"use client";

import { useMemo, useState } from "react";
import {
  CalendarDays,
  ChevronDown,
  Download,
  Search,
  Clock3,
  LogOut,
  Calendar,
  FileCheck2,
  Timer,
} from "lucide-react";
import { generateAttendanceReportPDF } from "@/lib/utils/pdf-generator";
import { useEffect, useCallback } from "react";
import { getAttendanceActivityReport, AttendanceActivityReportRow } from "@/lib/api/reports";
import { employeeService } from "@/lib/api/employee";
import { toast } from "react-hot-toast";
import { Skeleton } from "@/components/ui/skeleton";

import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

type Period = { label: string; value: string }; // YYYY-MM
type EventType = "all" | "late" | "missing_clock_out" | "leave" | "permission" | "overtime";
type ApprovalStatus = "all" | "pending" | "approved" | "rejected";
type Impact = "deduction" | "no_impact" | "addition" | "needs_review";

type EmployeeEventRow = {
  id: string;
  date: string; // YYYY-MM-DD
  dayLabel: string; // "Kamis"
  employee: { name: string; nik: string; initials: string };
  department: string;
  position: string;

  type: Exclude<EventType, "all">;

  // time fields (optional depending on type)
  scheduledIn?: string; // "09:00"
  actualIn?: string; // "09:15"
  scheduledOut?: string; // "18:00"
  actualOut?: string | null; // null/undefined => missing

  // overtime fields
  overtimeStart?: string; // "17:00"
  overtimeEnd?: string; // "19:00"
  overtimeHours?: number;

  // leave/permission fields
  dateRange?: string; // "12 Okt - 14 Okt (3 hari)"

  // computed details
  lateMinutes?: number;

  approvalStatus?: string;
  impact: Impact;

  note?: string;
};

function formatDateID(dateISO: string) {
  // Handles "2023-10-12" or ISO string from API
  const d = new Date(dateISO);
  return new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(d);
}

function cn(...classes: Array<string | false | undefined | null>) {
  return classes.filter(Boolean).join(" ");
}

function typeLabel(t: EmployeeEventRow["type"]) {
  switch (t) {
    case "late":
      return "Terlambat";
    case "missing_clock_out":
      return "Lupa Clock-out";
    case "leave":
      return "Cuti";
    case "permission":
      return "Izin";
    case "overtime":
      return "Lembur";
  }
}

function statusBadge(status?: string) {
  if (!status) return null;
  const s = status.toLowerCase();
  const map: Record<string, { cls: string; text: string }> = {
    pending: { cls: "bg-yellow-50 text-yellow-700 border border-yellow-200", text: "Pending" },
    approved: { cls: "bg-emerald-50 text-emerald-700 border border-emerald-200", text: "Approved" },
    rejected: { cls: "bg-rose-50 text-rose-700 border border-rose-200", text: "Rejected" },
  };
  const m = map[s] || { cls: "bg-gray-50 text-gray-700 border border-gray-200", text: status };
  return <Badge className={cn("rounded-full font-medium", m.cls)}>{m.text}</Badge>;
}

function impactBadge(impact: Impact) {
  const map: Record<Impact, { cls: string; text: string }> = {
    deduction: { cls: "bg-rose-50 text-rose-700 border border-rose-200", text: "Potong Gaji" },
    addition: { cls: "bg-blue-50 text-blue-700 border border-blue-200", text: "Tambah" },
    no_impact: { cls: "bg-gray-50 text-gray-700 border border-gray-200", text: "No Impact" },
    needs_review: { cls: "bg-orange-50 text-orange-700 border border-orange-200", text: "Butuh Klarifikasi" },
  };
  const m = map[impact];
  return <Badge className={cn("rounded-full font-medium", m.cls)}>{m.text}</Badge>;
}

function typeIcon(t: EmployeeEventRow["type"]) {
  const common = "h-4 w-4";
  switch (t) {
    case "late":
      return <Clock3 className={cn(common, "text-orange-600")} />;
    case "missing_clock_out":
      return <LogOut className={cn(common, "text-rose-600")} />;
    case "leave":
      return <Calendar className={cn(common, "text-blue-600")} />;
    case "permission":
      return <FileCheck2 className={cn(common, "text-indigo-600")} />;
    case "overtime":
      return <Timer className={cn(common, "text-emerald-600")} />;
  }
}

function buildDetail(r: EmployeeEventRow) {
  switch (r.type) {
    case "late":
      return `In ${r.actualIn ?? "-"} • Late ${r.lateMinutes ?? 0}m (Sched ${r.scheduledIn ?? "-"})`;
    case "missing_clock_out":
      return `In ${r.actualIn ?? "-"} • Out — (Sched ${r.scheduledOut ?? "-"})`;
    case "leave":
      return `Cuti • ${r.dateRange ?? "-"}`;
    case "permission":
      return `Izin • ${r.dateRange ?? "-"}`;
    case "overtime":
      return `${r.overtimeStart ?? "-"}–${r.overtimeEnd ?? "-"} • ${r.overtimeHours ?? 0} jam`;
  }
}

function mapApiRowToRow(apiRow: any): EmployeeEventRow {
  return {
    id: apiRow.id,
    date: apiRow.date,
    dayLabel: apiRow.day_label,
    employee: {
      name: apiRow.employee_name,
      nik: apiRow.employee_nik,
      initials: apiRow.employee_initials,
    },
    department: apiRow.department_name,
    position: apiRow.position_name,
    type: apiRow.type,
    scheduledIn: apiRow.scheduled_in,
    actualIn: apiRow.actual_in,
    scheduledOut: apiRow.scheduled_out,
    actualOut: apiRow.actual_out,
    overtimeStart: apiRow.overtime_start,
    overtimeEnd: apiRow.overtime_end,
    overtimeHours: apiRow.overtime_hours,
    dateRange: apiRow.date_range,
    lateMinutes: apiRow.late_minutes,
    approvalStatus: apiRow.approval_status,
    impact: apiRow.impact as Impact,
    note: apiRow.note,
  };
}

function exportCSV(rows: EmployeeEventRow[], filename: string) {
  const headers = [
    "date",
    "employee_name",
    "nik",
    "department",
    "position",
    "type",
    "detail",
    "approval_status",
    "impact",
  ];

  const escape = (v: any) => {
    const s = String(v ?? "");
    if (s.includes(",") || s.includes('"') || s.includes("\n")) return `"${s.replaceAll('"', '""')}"`;
    return s;
  };

  const lines = [
    headers.join(","),
    ...rows.map((r) =>
      [
        r.date,
        r.employee.name,
        r.employee.nik,
        r.department,
        r.position,
        r.type,
        buildDetail(r),
        r.approvalStatus ?? "",
        r.impact,
      ]
        .map(escape)
        .join(",")
    ),
  ];

  const blob = new Blob([lines.join("\n")], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);

  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  a.click();

  URL.revokeObjectURL(url);
}

export default function ReportsPresensiAktivitasPage() {
  // Mock period options
  const periods: Period[] = [
    { label: "Oktober 2023", value: "2023-10" },
    { label: "September 2023", value: "2023-09" },
    { label: "Agustus 2023", value: "2023-08" },
  ];

  const [period, setPeriod] = useState<Period>(periods[0]);
  const [department, setDepartment] = useState<{ id: string; name: string }>({ id: "all", name: "Semua Departemen" });
  const [eventType, setEventType] = useState<EventType>("all");
  const [approvalStatus, setApprovalStatus] = useState<ApprovalStatus>("all");
  const [query, setQuery] = useState<string>("");

  const [loading, setLoading] = useState(true);
  const [allRows, setAllRows] = useState<EmployeeEventRow[]>([]);
  const [summary, setSummary] = useState<any>(null);
  const [departments, setDepartments] = useState<Array<{ id: string; name: string }>>([{ id: "all", name: "Semua Departemen" }]);

  const [detailOpen, setDetailOpen] = useState(false);
  const [selected, setSelected] = useState<EmployeeEventRow | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getAttendanceActivityReport({
        period: period.value,
        department_id: department.id === "all" ? "" : department.id,
        type: eventType === "all" ? "" : eventType,
        status: approvalStatus === "all" ? "" : approvalStatus,
        search: query,
      });
      setAllRows(data.rows.map(mapApiRowToRow));
      setSummary(data.summary);
    } catch (err) {
      console.error(err);
      toast.error("Gagal mengambil data laporan");
    } finally {
      setLoading(false);
    }
  }, [period, department, eventType, approvalStatus, query]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    const loadDepts = async () => {
      try {
        const data = await employeeService.getAllDepartments();
        setDepartments([{ id: "all", name: "Semua Departemen" }, ...data.map((d: any) => ({ id: d.id, name: d.nama_departemen || d.name }))]);
      } catch (err) {
        console.error(err);
      }
    };
    loadDepts();
  }, []);

  const filtered = allRows; // Filtering is now done on the backend

  const tabItems: Array<{ key: EventType; label: string }> = [
    { key: "all", label: "Semua" },
    { key: "late", label: "Terlambat" },
    { key: "missing_clock_out", label: "Lupa Clock-out" },
    { key: "leave", label: "Cuti" },
    { key: "permission", label: "Izin" },
    { key: "overtime", label: "Lembur" },
  ];

  const openDetail = (r: EmployeeEventRow) => {
    setSelected(r);
    setDetailOpen(true);
  };

  const onExport = () => {
    const filename = `laporan-presensi-aktivitas_${period.value}_${eventType}.csv`;
    exportCSV(filtered, filename);
  };

  const [isExportingPdf, setIsExportingPdf] = useState(false);

  const onExportPdf = async () => {
    if (filtered.length === 0) {
      toast.error("Tidak ada data untuk diekspor");
      return;
    }
    try {
      setIsExportingPdf(true);
      const headers = ["Tanggal", "Karyawan", "NIK", "Jenis", "Detail", "Status"];
      const body = filtered.map((r) => [
        formatDateID(r.date),
        r.employee.name,
        r.employee.nik,
        typeLabel(r.type),
        buildDetail(r),
        r.approvalStatus ?? "-",
      ]);

      const blob = await generateAttendanceReportPDF({
        title: "Laporan Aktivitas Presensi",
        period: period.label,
        headers,
        body,
      });

      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `laporan-presensi-aktivitas_${period.value}_${eventType}.pdf`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      toast.success("PDF berhasil diunduh");
    } catch (err) {
      console.error("PDF export error:", err);
      toast.error("Gagal mengekspor PDF. Silakan coba lagi.");
    } finally {
      setIsExportingPdf(false);
    }
  };

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Laporan Aktivitas</h1>
          <p className="text-gray-600">
            Pantau karyawan yang telat clock-in, lupa clock-out, cuti, izin, dan lembur.
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* Period picker */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="rounded-xl gap-2">
                <CalendarDays className="h-4 w-4" />
                {period.label}
                <ChevronDown className="h-4 w-4 text-gray-500" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {periods.map((p) => (
                <DropdownMenuItem key={p.value} onClick={() => setPeriod(p)}>
                  {p.label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>

          {/* Export (dropdown) */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl gap-2">
                <Download className="h-4 w-4" />
                Export
                <ChevronDown className="h-4 w-4 opacity-90" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={onExport}>Export CSV</DropdownMenuItem>
              <DropdownMenuItem onClick={onExportPdf}>Export PDF</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          <KpiCard
            icon={<Clock3 className="h-5 w-5 text-orange-600" />}
            title="Total Terlambat"
            value={loading ? "..." : `${summary?.late?.events ?? 0}`}
            sub={loading ? "..." : `${summary?.late?.unique ?? 0} karyawan unik`}
          />
          <KpiCard
            icon={<LogOut className="h-5 w-5 text-rose-600" />}
            title="Lupa Clock-out"
            value={loading ? "..." : `${summary?.missing?.events ?? 0}`}
            sub={loading ? "..." : `${summary?.missing?.unique ?? 0} karyawan unik`}
          />
          <KpiCard
            icon={<Calendar className="h-5 w-5 text-blue-600" />}
            title="Total Cuti"
            value={loading ? "..." : `${summary?.leave?.events ?? 0}`}
            sub={loading ? "..." : `${summary?.leave?.unique ?? 0} karyawan unik`}
          />
          <KpiCard
            icon={<FileCheck2 className="h-5 w-5 text-indigo-600" />}
            title="Total Izin"
            value={loading ? "..." : `${summary?.permission?.events ?? 0}`}
            sub={loading ? "..." : `${summary?.permission?.unique ?? 0} karyawan unik`}
          />
          <KpiCard
            icon={<Timer className="h-5 w-5 text-emerald-600" />}
            title="Total Lembur"
            value={loading ? "..." : `${summary?.overtime?.events ?? 0}`}
            sub={loading ? "..." : `${summary?.overtime?.hours?.toFixed(1) ?? 0} jam • ${summary?.overtime?.unique ?? 0} karyawan`}
          />
      </div>

      {/* Filter bar */}
      <Card className="rounded-2xl">
        <CardContent className="p-5">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 items-end">
            <div>
              <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">Departemen</div>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="w-full justify-between rounded-xl">
                    {department.name}
                    <ChevronDown className="h-4 w-4 text-gray-500" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-56 max-h-72 overflow-y-auto">
                  {departments.map((d) => (
                    <DropdownMenuItem key={d.id} onClick={() => setDepartment(d)}>
                      {d.name}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div>
              <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">Tipe Aktivitas</div>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="w-full justify-between rounded-xl">
                    {tabItems.find((t) => t.key === eventType)?.label ?? "Semua"}
                    <ChevronDown className="h-4 w-4 text-gray-500" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-56">
                  {tabItems.map((t) => (
                    <DropdownMenuItem key={t.key} onClick={() => setEventType(t.key)}>
                      {t.label}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div>
              <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">Status</div>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="w-full justify-between rounded-xl">
                    {approvalStatus === "all"
                      ? "Semua Status"
                      : approvalStatus.charAt(0).toUpperCase() + approvalStatus.slice(1)}
                    <ChevronDown className="h-4 w-4 text-gray-500" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-56">
                  <DropdownMenuItem onClick={() => setApprovalStatus("all")}>Semua Status</DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setApprovalStatus("pending")}>Pending</DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setApprovalStatus("approved")}>Approved</DropdownMenuItem>
                  <DropdownMenuItem onClick={() => setApprovalStatus("rejected")}>Rejected</DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>

            <div>
              <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">Pencarian</div>
              <div className="relative">
                <Search className="h-4 w-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
                <input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Cari nama / NIK..."
                  className="w-full rounded-xl border border-gray-200 bg-white pl-9 pr-3 py-2.5 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
            </div>
          </div>

          {/* Active filter chips */}
          <div className="mt-4 flex flex-wrap gap-2">
            <FilterChip label={period.label} />
            {department.id !== "all" && <FilterChip label={department.name} onClear={() => setDepartment({ id: "all", name: "Semua Departemen" })} />}
            {eventType !== "all" && (
              <FilterChip
                label={`Tipe: ${tabItems.find((t) => t.key === eventType)?.label}`}
                onClear={() => setEventType("all")}
              />
            )}
            {approvalStatus !== "all" && (
              <FilterChip label={`Status: ${approvalStatus}`} onClear={() => setApprovalStatus("all")} />
            )}
            {query.trim() && <FilterChip label={`Cari: ${query.trim()}`} onClear={() => setQuery("")} />}
          </div>
        </CardContent>
      </Card>

      {/* Tabs */}
      <div className="flex items-center gap-2 flex-wrap">
        {tabItems.map((t) => {
          const active = eventType === t.key;
          return (
            <button
              key={t.key}
              type="button"
              onClick={() => setEventType(t.key)}
              className={cn(
                "px-4 py-2 rounded-xl text-sm font-medium border transition-colors",
                active
                  ? "bg-blue-600 text-white border-blue-600"
                  : "bg-white text-gray-700 border-gray-200 hover:bg-gray-50"
              )}
            >
              {t.label}
            </button>
          );
        })}
      </div>

      {/* Table */}
      <Card className="rounded-2xl">
        <CardContent className="p-0">
          <div className="p-5 border-b border-gray-100 flex items-center justify-between">
            <div className="font-semibold text-gray-900">
              Data Aktivitas <span className="text-gray-500 font-medium">({filtered.length})</span>
            </div>
          </div>

          {loading ? (
            <div className="p-10 space-y-4">
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
            </div>
          ) : filtered.length === 0 ? (
            <div className="p-10 text-center">
              <div className="text-sm font-semibold text-gray-900">Tidak ada data</div>
              <div className="text-sm text-gray-600 mt-1">
                Coba ubah filter atau periode.
              </div>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full">
                <thead className="sticky top-0 bg-gray-50 border-b border-gray-100">
                  <tr>
                    <Th>Tanggal</Th>
                    <Th>Karyawan</Th>
                    <Th>Departemen / Posisi</Th>
                    <Th>Jenis</Th>
                    <Th>Detail Aktivitas</Th>
                    <Th>Status</Th>
                    <Th>Impact</Th>
                    <Th className="text-right">Aksi</Th>
                  </tr>
                </thead>

                <tbody className="divide-y divide-gray-100 bg-white">
                  {filtered.map((r) => (
                    <tr key={r.id} className="hover:bg-gray-50 transition-colors">
                      <Td>
                        <div className="text-sm font-medium text-gray-900">{formatDateID(r.date)}</div>
                        <div className="text-xs text-gray-500">{r.dayLabel}</div>
                      </Td>

                      <Td>
                        <div className="flex items-center gap-3">
                          <div className="h-9 w-9 rounded-full bg-gray-100 flex items-center justify-center text-xs font-semibold text-gray-700">
                            {r.employee.initials}
                          </div>
                          <div>
                            <div className="text-sm font-semibold text-gray-900">{r.employee.name}</div>
                            <div className="text-xs text-gray-500">{r.employee.nik}</div>
                          </div>
                        </div>
                      </Td>

                      <Td>
                        <div className="text-sm font-medium text-gray-900">{r.department}</div>
                        <div className="text-xs text-gray-500">{r.position}</div>
                      </Td>

                      <Td>
                        <div className="flex items-center gap-2">
                          {typeIcon(r.type)}
                          <span className="text-sm font-medium text-gray-900">{typeLabel(r.type)}</span>
                        </div>
                      </Td>

                      <Td>
                        <div className="text-sm text-gray-900">{buildDetail(r)}</div>
                      </Td>

                      <Td>{statusBadge(r.approvalStatus as any)}</Td>

                      <Td>{impactBadge(r.impact)}</Td>

                      <Td className="text-right">
                        <Button variant="link" className="text-blue-600 px-0" onClick={() => openDetail(r)}>
                          Detail
                        </Button>
                      </Td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination (mock) */}
          <div className="p-4 border-t border-gray-100 flex items-center justify-between text-sm text-gray-600">
            <div>Menampilkan 1 – {Math.min(filtered.length, 10)} dari {filtered.length} data</div>
            <div className="flex items-center gap-2">
              <Button variant="outline" className="rounded-xl" disabled>
                Sebelumnya
              </Button>
              <div className="px-3 py-2 rounded-xl border border-gray-200 bg-white text-gray-900">1</div>
              <Button variant="outline" className="rounded-xl" disabled>
                Selanjutnya
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Detail dialog */}
      <Dialog open={detailOpen} onOpenChange={setDetailOpen}>
        <DialogContent className="sm:max-w-[640px] rounded-2xl">
          <DialogHeader>
            <DialogTitle>Detail Aktivitas</DialogTitle>
          </DialogHeader>

          {selected && (
            <div className="space-y-4">
              <div className="rounded-xl border border-gray-100 p-4">
                <div className="flex items-center gap-3">
                  <div className="h-10 w-10 rounded-full bg-gray-100 flex items-center justify-center text-xs font-semibold text-gray-700">
                    {selected.employee.initials}
                  </div>
                  <div className="flex-1">
                    <div className="text-sm font-semibold text-gray-900">{selected.employee.name}</div>
                    <div className="text-xs text-gray-500">
                      {selected.employee.nik} • {selected.department} • {selected.position}
                    </div>
                  </div>
                  <div className="text-xs text-gray-500">{formatDateID(selected.date)}</div>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                <InfoCard label="Jenis" value={typeLabel(selected.type)} />
                <InfoCard label="Detail" value={buildDetail(selected)} />
                <InfoCard label="Status" value={selected.approvalStatus ?? "-"} />
                <InfoCard label="Impact" value={selected.impact} />
              </div>

              <div>
                <div className="text-[11px] font-semibold text-gray-500 uppercase mb-2">Catatan</div>
                <textarea
                  rows={3}
                  defaultValue={selected.note ?? ""}
                  placeholder="Tambahkan catatan (opsional)..."
                  className="w-full rounded-xl border border-gray-200 bg-white px-3 py-2.5 text-sm
                             focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
                <div className="mt-3 flex justify-end">
                  <Button className="bg-blue-600 hover:bg-blue-700 text-white rounded-xl">
                    Simpan Catatan
                  </Button>
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

function KpiCard(props: { icon: React.ReactNode; title: string; value: string; sub: string }) {
  return (
    <Card className="rounded-2xl">
      <CardContent className="p-5">
        <div className="flex items-center justify-between">
          <div className="h-10 w-10 rounded-xl bg-gray-50 border border-gray-100 flex items-center justify-center">
            {props.icon}
          </div>
        </div>
        <div className="mt-4 text-xs font-semibold uppercase tracking-wider text-gray-500">{props.title}</div>
        <div className="mt-1 text-2xl font-bold text-gray-900">{props.value}</div>
        <div className="mt-1 text-xs text-gray-500">{props.sub}</div>
      </CardContent>
    </Card>
  );
}

function FilterChip(props: { label: string; onClear?: () => void }) {
  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-gray-200 bg-white px-3 py-1 text-xs text-gray-700">
      {props.label}
      {props.onClear && (
        <button
          type="button"
          onClick={props.onClear}
          className="text-gray-400 hover:text-gray-700"
          aria-label={`Hapus filter ${props.label}`}
        >
          ×
        </button>
      )}
    </span>
  );
}

function Th(props: React.PropsWithChildren<{ className?: string }>) {
  return (
    <th
      className={cn(
        "px-6 py-3 text-left text-xs font-semibold uppercase tracking-wider text-gray-500 whitespace-nowrap",
        props.className
      )}
    >
      {props.children}
    </th>
  );
}

function Td(props: React.PropsWithChildren<{ className?: string }>) {
  return <td className={cn("px-6 py-4 align-top", props.className)}>{props.children}</td>;
}

function InfoCard(props: { label: string; value: string }) {
  return (
    <div className="rounded-xl border border-gray-100 p-4">
      <div className="text-[11px] font-semibold text-gray-500 uppercase">{props.label}</div>
      <div className="mt-1 text-sm font-semibold text-gray-900">{props.value}</div>
    </div>
  );
}