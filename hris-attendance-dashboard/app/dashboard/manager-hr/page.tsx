// app/dashboard/manager-hr/page.tsx
import { Users, CheckCircle, Clock, FileText } from "lucide-react";
import { StatsCard } from "@/components/stats-card";
import { MonitoringTable } from "@/components/monitoring-table";
import { ManagementPanel } from "@/components/management-panel";
import { statsData, employeesData } from "@/lib/mock-data";

export default function ManagerHRDashboard() {
  return (
    <div className="flex gap-6 p-6">
      {/* Main Content */}
      <div className="flex-1 space-y-6">
        {/* Stats Cards */}
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          {/* Total Pegawai - Blue */}
          <StatsCard
            title="TOTAL PEGAWAI"
            value={statsData.totalEmployees}
            icon={Users}
            iconColor="text-blue-600"
            iconBgColor="bg-blue-50"
            trend={{
              value: statsData.totalEmployeesTrend,
              isPositive: true,
            }}
          />
          
          {/* Hadir Hari Ini - Green/Teal */}
          <StatsCard
            title="HADIR HARI INI"
            value={statsData.presentToday}
            icon={CheckCircle}
            iconColor="text-teal-600"
            iconBgColor="bg-teal-50"
            badge={{
              text: `${statsData.presentPercentage}% Berhasil`,
              variant: "success",
            }}
          />
          
          {/* Terlambat - Orange/Yellow */}
          <StatsCard
            title="TERLAMBAT"
            value={statsData.lateToday}
            icon={Clock}
            iconColor="text-orange-500"
            iconBgColor="bg-orange-50"
            link={{
              text: "Lihat Semua Log",
              href: "/dashboard/manager-hr/presensi",
            }}
          />
          
          {/* Pengajuan Izin - Red/Pink */}
          <StatsCard
            title="PENGAJUAN IZIN"
            value={statsData.leaveRequests}
            icon={FileText}
            iconColor="text-red-500"
            iconBgColor="bg-red-50"
            trend={{
              value: statsData.leaveRequestsTrend,
              isPositive: false,
            }}
          />
        </div>

        {/* Monitoring Table */}
        <MonitoringTable employees={employeesData} />
      </div>

      {/* Management Sidebar */}
      <div className="w-80">
        <ManagementPanel />
      </div>
    </div>
  );
}