"use client";

import { useCallback, useEffect, useState } from "react";
import {
  Users,
  CheckCircle,
  Clock,
  XCircle,
  TrendingUp,
  CalendarCheck,
  Loader2,
  RefreshCw,
  ChevronRight,
  UserCheck,
  AlertCircle,
  FileText,
} from "lucide-react";
import { format } from "date-fns";
import { id as idLocale } from "date-fns/locale";
import { useRouter } from "next/navigation";
import { authService } from "@/lib/api/auth";
import { attendanceManagerDeptApi, type ManagerDeptAttendanceItem } from "@/lib/api/attendance-manager-dept";
import { deptLeaveRequestsApi, type DeptLeaveRequestApprovalResponse } from "@/lib/api/dept-leave-requests";
import { employeeService } from "@/lib/api/employee";

/* ─────────────────────────────────────────────
   Types
───────────────────────────────────────────── */
interface DashboardStats {
  totalAnggota: number;
  hadirHariIni: number;
  menungguApproval: number;
  tidakHadir: number;
  terlambat: number;
  izinSakit: number;
}

interface AttendanceSummaryRow {
  id: string;
  name: string;
  position: string;
  clockIn: string;
  clockOut: string;
  status: "HADIR" | "TELAT" | "IZIN" | "ALFA" | string;
}

interface PendingLeaveItem {
  id: string;
  employeeName: string;
  type: string;
  startDate: string;
  endDate: string;
  reason: string;
  submittedAt: string;
}

/* ─────────────────────────────────────────────
   Helpers
───────────────────────────────────────────── */
function todayWIB(): string {
  return format(new Date(), "yyyy-MM-dd");
}

function statusLabel(s: string) {
  switch (s) {
    case "HADIR":  return "Hadir";
    case "TELAT":  return "Terlambat";
    case "IZIN":   return "Izin";
    case "ALFA":   return "Alfa";
    default:       return s;
  }
}

function statusColor(s: string) {
  switch (s) {
    case "HADIR": return "bg-green-100 text-green-700";
    case "TELAT": return "bg-yellow-100 text-yellow-700";
    case "IZIN":  return "bg-blue-100 text-blue-700";
    case "ALFA":  return "bg-red-100 text-red-700";
    default:      return "bg-gray-100 text-gray-600";
  }
}

function leaveTypeColor(type: string) {
  const t = type.toUpperCase();
  if (t.includes("SAKIT"))  return "bg-red-100 text-red-700";
  if (t.includes("TAHUN"))  return "bg-gray-100 text-gray-800";
  return "bg-blue-100 text-blue-700";
}

function formatDateShort(dateStr: string) {
  try {
    return format(new Date(dateStr), "dd MMM yyyy", { locale: idLocale });
  } catch {
    return dateStr;
  }
}

/* ─────────────────────────────────────────────
   Stat Card
───────────────────────────────────────────── */
function StatCard({
  title,
  value,
  icon: Icon,
  iconBg,
  iconColor,
  loading,
  subtitle,
}: {
  title: string;
  value: number | string;
  icon: React.ElementType;
  iconBg: string;
  iconColor: string;
  loading?: boolean;
  subtitle?: string;
}) {
  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5 flex items-center gap-4">
      <div className={`${iconBg} p-3 rounded-xl shrink-0`}>
        <Icon className={`h-6 w-6 ${iconColor}`} />
      </div>
      <div className="min-w-0">
        <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide truncate">{title}</p>
        {loading ? (
          <div className="mt-1 h-7 w-16 bg-gray-200 animate-pulse rounded" />
        ) : (
          <p className="text-2xl font-bold text-gray-900 leading-tight">{value}</p>
        )}
        {subtitle && !loading && (
          <p className="text-xs text-gray-400 mt-0.5">{subtitle}</p>
        )}
      </div>
    </div>
  );
}

/* ─────────────────────────────────────────────
   Main Page
───────────────────────────────────────────── */
export default function ManagerDeptDashboard() {
  const router = useRouter();
  const user = authService.getUser();
  const today = todayWIB();

  const [loadingStats, setLoadingStats]         = useState(true);
  const [loadingAttendance, setLoadingAttendance] = useState(true);
  const [loadingLeave, setLoadingLeave]           = useState(true);
  const [error, setError]                         = useState<string | null>(null);

  const [stats, setStats] = useState<DashboardStats>({
    totalAnggota: 0,
    hadirHariIni: 0,
    menungguApproval: 0,
    tidakHadir: 0,
    terlambat: 0,
    izinSakit: 0,
  });

  const [attendanceRows, setAttendanceRows] = useState<AttendanceSummaryRow[]>([]);
  const [pendingLeave, setPendingLeave]     = useState<PendingLeaveItem[]>([]);
  const [lastRefresh, setLastRefresh]       = useState<Date>(new Date());

  /* ── Load jumlah anggota tim ── */
  const loadTeamCount = useCallback(async () => {
    try {
      const employees = await employeeService.getEmployeesMyDepartment();
      setStats((prev) => ({ ...prev, totalAnggota: employees.length }));
    } catch {
      // fallback: biarkan 0
    } finally {
      setLoadingStats(false);
    }
  }, []);

  /* ── Load presensi hari ini ── */
  const loadTodayAttendance = useCallback(async () => {
    setLoadingAttendance(true);
    try {
      const res = await attendanceManagerDeptApi.list({
        from: today,
        to: today,
        page: 1,
        page_size: 100,
      });

      const items: ManagerDeptAttendanceItem[] = res.items ?? [];

      const rows: AttendanceSummaryRow[] = items.map((it) => ({
        id: it.id,
        name: it.full_name,
        position: it.position_name ?? "-",
        clockIn: it.clock_in_time ?? "--:--",
        clockOut: it.clock_out_time ?? "--:--",
        status: it.status,
      }));

      setAttendanceRows(rows);

      // Hitung statistik dari data presensi
      const hadir   = items.filter((i) => i.status === "HADIR").length;
      const telat   = items.filter((i) => i.status === "TELAT").length;
      const izin    = items.filter((i) => i.status === "IZIN").length;
      const alfa    = items.filter((i) => i.status === "ALFA").length;

      setStats((prev) => ({
        ...prev,
        hadirHariIni: hadir + telat, // hadir tepat + telat = total masuk
        terlambat: telat,
        izinSakit: izin,
        tidakHadir: alfa,
      }));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memuat data presensi");
    } finally {
      setLoadingAttendance(false);
    }
  }, [today]);

  /* ── Load pengajuan izin pending ── */
  const loadPendingLeave = useCallback(async () => {
    setLoadingLeave(true);
    try {
      const res: DeptLeaveRequestApprovalResponse[] = await deptLeaveRequestsApi.list({ status: "ALL" });

      const pending = res.filter(
        (r) => (r.pengajuan.status_kepala_departemen as string).toUpperCase() === "PENDING"
      );

      const mapped: PendingLeaveItem[] = pending.map((r) => ({
        id: r.pengajuan.id,
        employeeName: r.employee?.full_name ?? "Karyawan",
        type: r.pengajuan.type_name ?? "Izin",
        startDate: r.pengajuan.start_date,
        endDate: r.pengajuan.end_date,
        reason: r.pengajuan.reason ?? "-",
        submittedAt: r.pengajuan.created_at,
      }));

      setPendingLeave(mapped);
      setStats((prev) => ({ ...prev, menungguApproval: mapped.length }));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memuat data pengajuan");
    } finally {
      setLoadingLeave(false);
    }
  }, []);

  /* ── Initial load ── */
  useEffect(() => {
    loadTeamCount();
    loadTodayAttendance();
    loadPendingLeave();
  }, [loadTeamCount, loadTodayAttendance, loadPendingLeave]);

  /* ── Refresh all ── */
  const handleRefresh = () => {
    setError(null);
    setLastRefresh(new Date());
    loadTeamCount();
    loadTodayAttendance();
    loadPendingLeave();
  };

  const isLoading = loadingStats || loadingAttendance || loadingLeave;

  /* ─────────────────────────────────────────────
     Render
  ───────────────────────────────────────────── */
  return (
    <div className="p-6 space-y-6 bg-gray-50 min-h-full">

      {/* ── Header ── */}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">
            Dashboard Kepala Departemen
          </h1>
          <p className="text-sm text-gray-500 mt-0.5">
            Selamat datang,&nbsp;
            <span className="font-medium text-gray-700">
              {user?.full_name ?? "Kepala Departemen"}
            </span>
            &nbsp;·&nbsp;
            {user?.department_name ?? user?.department ?? ""}
          </p>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <p className="text-xs text-gray-400 hidden sm:block">
            Diperbarui&nbsp;{format(lastRefresh, "HH:mm", { locale: idLocale })}
          </p>
          <button
            onClick={handleRefresh}
            disabled={isLoading}
            className="flex items-center gap-1.5 text-sm text-gray-600 hover:text-gray-900 border border-gray-200 hover:border-gray-300 bg-white rounded-xl px-3 py-2 transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
            Refresh
          </button>
        </div>
      </div>

      {/* ── Error banner ── */}
      {error && (
        <div className="rounded-xl border border-red-200 bg-red-50 px-4 py-3 flex items-center gap-2 text-sm text-red-700">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{error}</span>
          <button
            onClick={() => setError(null)}
            className="ml-auto text-red-400 hover:text-red-600"
          >
            ✕
          </button>
        </div>
      )}

      {/* ── Tanggal hari ini ── */}
      <div className="flex items-center gap-2 text-sm text-gray-500">
        <CalendarCheck className="h-4 w-4 text-indigo-500" />
        <span>
          Data hari ini:&nbsp;
          <strong className="text-gray-800">
            {format(new Date(), "EEEE, dd MMMM yyyy", { locale: idLocale })}
          </strong>
        </span>
      </div>

      {/* ── Stats Grid ── */}
      <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-4">
        <StatCard
          title="Anggota Tim"
          value={stats.totalAnggota}
          icon={Users}
          iconBg="bg-blue-50"
          iconColor="text-blue-600"
          loading={loadingStats}
        />
        <StatCard
          title="Hadir Hari Ini"
          value={stats.hadirHariIni}
          icon={CheckCircle}
          iconBg="bg-green-50"
          iconColor="text-green-600"
          loading={loadingAttendance}
          subtitle="tepat waktu + terlambat"
        />
        <StatCard
          title="Terlambat"
          value={stats.terlambat}
          icon={TrendingUp}
          iconBg="bg-yellow-50"
          iconColor="text-yellow-600"
          loading={loadingAttendance}
        />
        <StatCard
          title="Menunggu Approval"
          value={stats.menungguApproval}
          icon={Clock}
          iconBg="bg-orange-50"
          iconColor="text-orange-500"
          loading={loadingLeave}
          subtitle="pengajuan izin/cuti"
        />
        <StatCard
          title="Alfa / Tidak Hadir"
          value={stats.tidakHadir}
          icon={XCircle}
          iconBg="bg-red-50"
          iconColor="text-red-600"
          loading={loadingAttendance}
        />
      </div>

      {/* ── Content Grid ── */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">

        {/* ── Presensi Hari Ini ── */}
        <div className="xl:col-span-2 bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
          <div className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
            <div className="flex items-center gap-2">
              <UserCheck className="h-4 w-4 text-indigo-500" />
              <h2 className="text-sm font-semibold text-gray-900">Presensi Tim Hari Ini</h2>
            </div>
            <button
              onClick={() => router.push("/dashboard/manager-dept/presensi")}
              className="flex items-center gap-1 text-xs text-indigo-600 hover:text-indigo-700 font-medium"
            >
              Lihat Semua <ChevronRight className="h-3.5 w-3.5" />
            </button>
          </div>

          {loadingAttendance ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="h-6 w-6 animate-spin text-indigo-400" />
            </div>
          ) : attendanceRows.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-gray-400 gap-2">
              <CalendarCheck className="h-10 w-10" />
              <p className="text-sm">Belum ada data presensi hari ini</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="bg-gray-50 border-b border-gray-100">
                    <th className="px-5 py-3 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">
                      Nama
                    </th>
                    <th className="px-5 py-3 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">
                      Jabatan
                    </th>
                    <th className="px-5 py-3 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">
                      Masuk
                    </th>
                    <th className="px-5 py-3 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">
                      Keluar
                    </th>
                    <th className="px-5 py-3 text-left text-[11px] font-semibold uppercase tracking-wider text-gray-500">
                      Status
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-50">
                  {attendanceRows.slice(0, 8).map((row) => (
                    <tr key={row.id} className="hover:bg-gray-50 transition-colors">
                      <td className="px-5 py-3">
                        <div className="flex items-center gap-2.5">
                          <div className="h-7 w-7 rounded-full bg-indigo-100 flex items-center justify-center shrink-0">
                            <span className="text-[10px] font-bold text-indigo-600">
                              {row.name
                                .split(" ")
                                .filter(Boolean)
                                .map((p) => p[0])
                                .join("")
                                .slice(0, 2)
                                .toUpperCase()}
                            </span>
                          </div>
                          <span className="text-sm font-medium text-gray-900 truncate max-w-[120px]">
                            {row.name}
                          </span>
                        </div>
                      </td>
                      <td className="px-5 py-3 text-sm text-gray-500 truncate max-w-[120px]">
                        {row.position}
                      </td>
                      <td className="px-5 py-3 text-sm text-gray-700 font-mono">
                        {row.clockIn ?? "--:--"}
                      </td>
                      <td className="px-5 py-3 text-sm text-gray-700 font-mono">
                        {row.clockOut ?? "--:--"}
                      </td>
                      <td className="px-5 py-3">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-[11px] font-semibold ${statusColor(row.status)}`}
                        >
                          {statusLabel(row.status)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {attendanceRows.length > 8 && (
                <div className="px-5 py-3 border-t border-gray-100 text-center">
                  <button
                    onClick={() => router.push("/dashboard/manager-dept/presensi")}
                    className="text-xs text-indigo-600 hover:text-indigo-700 font-medium"
                  >
                    +{attendanceRows.length - 8} karyawan lainnya — Lihat semua
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* ── Pengajuan Izin Pending ── */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden flex flex-col">
          <div className="flex items-center justify-between px-5 py-4 border-b border-gray-100">
            <div className="flex items-center gap-2">
              <FileText className="h-4 w-4 text-orange-500" />
              <h2 className="text-sm font-semibold text-gray-900">
                Pengajuan Izin Menunggu
              </h2>
              {pendingLeave.length > 0 && (
                <span className="ml-1 inline-flex items-center justify-center h-5 w-5 rounded-full bg-orange-500 text-white text-[10px] font-bold">
                  {pendingLeave.length}
                </span>
              )}
            </div>
            <button
              onClick={() => router.push("/dashboard/manager-dept/persetujuan-izin-cuti")}
              className="flex items-center gap-1 text-xs text-indigo-600 hover:text-indigo-700 font-medium"
            >
              Kelola <ChevronRight className="h-3.5 w-3.5" />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto">
            {loadingLeave ? (
              <div className="flex items-center justify-center py-16">
                <Loader2 className="h-6 w-6 animate-spin text-orange-400" />
              </div>
            ) : pendingLeave.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-16 text-gray-400 gap-2">
                <CheckCircle className="h-10 w-10 text-green-300" />
                <p className="text-sm font-medium text-gray-500">Tidak ada pengajuan</p>
                <p className="text-xs text-gray-400">Semua pengajuan sudah diproses</p>
              </div>
            ) : (
              <div className="divide-y divide-gray-50">
                {pendingLeave.map((item) => (
                  <div
                    key={item.id}
                    className="px-5 py-4 hover:bg-gray-50 transition-colors cursor-pointer"
                    onClick={() => router.push("/dashboard/manager-dept/persetujuan-izin-cuti")}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-semibold text-gray-900 truncate">
                          {item.employeeName}
                        </p>
                        <div className="flex items-center gap-1.5 mt-1 flex-wrap">
                          <span
                            className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold ${leaveTypeColor(item.type)}`}
                          >
                            {item.type}
                          </span>
                          <span className="text-[11px] text-gray-500">
                            {formatDateShort(item.startDate)}
                            {item.startDate !== item.endDate &&
                              ` – ${formatDateShort(item.endDate)}`}
                          </span>
                        </div>
                        <p className="text-xs text-gray-400 mt-1 line-clamp-1">
                          {item.reason}
                        </p>
                      </div>
                      <span className="shrink-0 inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-orange-100 text-orange-700">
                        Pending
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {pendingLeave.length > 0 && (
            <div className="px-5 py-3 border-t border-gray-100">
              <button
                onClick={() => router.push("/dashboard/manager-dept/persetujuan-izin-cuti")}
                className="w-full py-2 rounded-xl bg-orange-50 hover:bg-orange-100 text-orange-700 text-sm font-semibold transition-colors"
              >
                Proses Semua Pengajuan
              </button>
            </div>
          )}
        </div>
      </div>

      {/* ── Aksi Cepat ── */}
      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-5">
        <h2 className="text-sm font-semibold text-gray-900 mb-4">Aksi Cepat</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
          {[
            {
              label: "Data Karyawan",
              desc: "Kelola anggota tim",
              icon: Users,
              color: "text-blue-600",
              bg: "bg-blue-50 hover:bg-blue-100",
              href: "/dashboard/manager-dept/karyawan",
            },
            {
              label: "Presensi Tim",
              desc: "Rekap kehadiran",
              icon: CalendarCheck,
              color: "text-green-600",
              bg: "bg-green-50 hover:bg-green-100",
              href: "/dashboard/manager-dept/presensi",
            },
            {
              label: "Izin & Cuti",
              desc: `${stats.menungguApproval} menunggu persetujuan`,
              icon: FileText,
              color: "text-orange-600",
              bg: "bg-orange-50 hover:bg-orange-100",
              href: "/dashboard/manager-dept/persetujuan-izin-cuti",
            },
            {
              label: "Jam Kerja",
              desc: "Atur jadwal tim",
              icon: Clock,
              color: "text-purple-600",
              bg: "bg-purple-50 hover:bg-purple-100",
              href: "/dashboard/manager-dept/jam-kerja",
            },
          ].map((action) => (
            <button
              key={action.href}
              onClick={() => router.push(action.href)}
              className={`flex items-center gap-3 p-4 rounded-xl ${action.bg} transition-colors text-left`}
            >
              <div className="shrink-0">
                <action.icon className={`h-5 w-5 ${action.color}`} />
              </div>
              <div className="min-w-0">
                <p className={`text-sm font-semibold ${action.color}`}>{action.label}</p>
                <p className="text-xs text-gray-500 mt-0.5 truncate">{action.desc}</p>
              </div>
              <ChevronRight className="h-4 w-4 text-gray-400 ml-auto shrink-0" />
            </button>
          ))}
        </div>
      </div>

    </div>
  );
}
