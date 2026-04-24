"use client";

import { useCallback, useEffect, useState } from "react";
import { Users, CheckCircle, Clock, FileText } from "lucide-react";
import { StatsCard } from "@/components/stats-card";
import { MonitoringTable } from "@/components/monitoring-table";
import { ManagementPanel } from "@/components/management-panel";
import { format } from "date-fns";
import { attendanceManagerApi, type ManagerAttendanceItem } from "@/lib/api/attendance-manager";
import { leaveRequestsApi } from "@/lib/api/leave-requests";
import { employeeService } from "@/lib/api/employee";
import type { Employee } from "@/types";

function todayStr() {
  return format(new Date(), "yyyy-MM-dd");
}

function mapAttendanceToEmployee(item: ManagerAttendanceItem): Employee {
  const initials = item.full_name
    .split(" ")
    .filter(Boolean)
    .map((p) => p[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  const statusMap: Record<string, Employee["status"]> = {
    HADIR: "HADIR",
    TELAT: "TELAMBAT",
    IZIN: "IZIN",
    ALFA: "ALPHA",
  };

  const clockIn =
    item.clock_in_time && item.clock_in_time !== "--:--"
      ? `${item.clock_in_time} WIB`
      : undefined;

  return {
    id: item.id,
    name: item.full_name,
    avatar: initials,
    nik: item.payroll_number,
    department: item.department_name,
    position: item.position_name,
    checkInTime: clockIn,
    status: statusMap[item.status] ?? "HADIR",
    verified: {
      biometric: true,
      geofencing: !!item.location,
    },
  };
}

export default function ManagerHRDashboard() {
  const today = todayStr();

  // ── Loading states
  const [loadingStats, setLoadingStats]           = useState(true);
  const [loadingAttendance, setLoadingAttendance] = useState(true);
  const [loadingLeave, setLoadingLeave]           = useState(true);

  // ── Stats values
  const [totalEmployees, setTotalEmployees] = useState(0);
  const [presentToday, setPresentToday]     = useState(0);
  const [presentPct, setPresentPct]         = useState(0);
  const [lateToday, setLateToday]           = useState(0);
  const [pendingLeave, setPendingLeave]     = useState(0);

  // ── Table data
  const [employees, setEmployees] = useState<Employee[]>([]);

  /* ─── Load total karyawan aktif ─── */
  const loadTotalEmployees = useCallback(async () => {
    try {
      const list = await employeeService.getAllEmployees();
      const active = list.filter((e) => (e as { is_active?: boolean }).is_active !== false);
      setTotalEmployees(active.length);
    } catch {
      // biarkan 0
    } finally {
      setLoadingStats(false);
    }
  }, []);

  /* ─── Load presensi hari ini ─── */
  const loadTodayAttendance = useCallback(async () => {
    setLoadingAttendance(true);
    try {
      const res = await attendanceManagerApi.list({
        from: today,
        to: today,
        page: 1,
        page_size: 200,
      });

      const items = res.items ?? [];

      const hadir = items.filter((i) => i.status === "HADIR").length;
      const telat = items.filter((i) => i.status === "TELAT").length;
      const total = items.length;

      setPresentToday(hadir + telat);
      setPresentPct(total > 0 ? Math.round(((hadir + telat) / total) * 100) : 0);
      setLateToday(telat);
      setEmployees(items.map(mapAttendanceToEmployee));
    } catch {
      setEmployees([]);
    } finally {
      setLoadingAttendance(false);
    }
  }, [today]);

  /* ─── Load pengajuan izin pending ─── */
  const loadPendingLeave = useCallback(async () => {
    setLoadingLeave(true);
    try {
      const list = await leaveRequestsApi.listForManagerHR({ status: "ALL" });
      const pending = list.filter(
        (r) =>
          (r.pengajuan.status_manager_hr as string).toUpperCase() === "PENDING" &&
          (r.pengajuan.status_kepala_departemen as string).toUpperCase() === "APPROVED"
      );
      setPendingLeave(pending.length);
    } catch {
      setPendingLeave(0);
    } finally {
      setLoadingLeave(false);
    }
  }, []);

  /* ─── Initial load ─── */
  useEffect(() => {
    loadTotalEmployees();
    loadTodayAttendance();
    loadPendingLeave();
  }, [loadTotalEmployees, loadTodayAttendance, loadPendingLeave]);

  return (
    <div className="flex gap-6 p-6">
      {/* Main Content */}
      <div className="flex-1 space-y-6">
        {/* Stats Cards */}
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          {/* Total Pegawai - Blue */}
          <StatsCard
            title="TOTAL PEGAWAI"
            value={totalEmployees}
            icon={Users}
            iconColor="text-blue-600"
            iconBgColor="bg-blue-50"
            loading={loadingStats}
            trend={{
              value: 0,
              isPositive: true,
            }}
          />

          {/* Hadir Hari Ini - Green/Teal */}
          <StatsCard
            title="HADIR HARI INI"
            value={presentToday}
            icon={CheckCircle}
            iconColor="text-teal-600"
            iconBgColor="bg-teal-50"
            loading={loadingAttendance}
            badge={{
              text: `${presentPct}% Berhasil`,
              variant: "success",
            }}
          />

          {/* Terlambat - Orange/Yellow */}
          <StatsCard
            title="TERLAMBAT"
            value={lateToday}
            icon={Clock}
            iconColor="text-orange-500"
            iconBgColor="bg-orange-50"
            loading={loadingAttendance}
            link={{
              text: "Lihat Semua Log",
              href: "/dashboard/manager-hr/presensi",
            }}
          />

          {/* Pengajuan Izin - Red/Pink */}
          <StatsCard
            title="PENGAJUAN IZIN"
            value={pendingLeave}
            icon={FileText}
            iconColor="text-red-500"
            iconBgColor="bg-red-50"
            loading={loadingLeave}
            trend={{
              value: 0,
              isPositive: false,
            }}
          />
        </div>

        {/* Monitoring Table */}
        <MonitoringTable
          employees={employees}
          loading={loadingAttendance}
          emptyMessage="Belum ada data presensi hari ini"
        />
      </div>

      {/* Management Sidebar */}
      <div className="w-80">
        <ManagementPanel
          pendingLeaveCount={pendingLeave}
          loadingLeave={loadingLeave}
        />
      </div>
    </div>
  );
}
