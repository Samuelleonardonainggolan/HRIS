"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useAuth } from "@/contexts/AuthContext";
import {
  LayoutDashboard,
  Users,
  Fingerprint,
  MapPin,
  ClipboardCheck,
  Clock,
  Shield,
  FileText,
  Bell,
  Building2,
  UserCog,
  Camera,
  Calendar,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";

const iconMap: { [key: string]: React.ElementType } = {
  LayoutDashboard,
  Users,
  Fingerprint,
  MapPin,
  ClipboardCheck,
  Clock,
  Shield,
  FileText,
  Bell,
  Building2,
  UserCog,
  Camera,
  Calendar,
};

interface NavItem {
  name: string;
  href: string;
  icon: string;
}

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuth();

  // Get base URL for current role
  const getRoleBasePath = () => {
    switch (user?.role) {
      case "manager_hr":
        return "/dashboard/manager-hr";
      case "manager_departemen":
        return "/dashboard/manager-dept";
      case "admin_departemen":
        return "/dashboard/admin-dept";
      case "staf":
        return "/dashboard/staf";
      default:
        return "/dashboard";
    }
  };

  // Navigation items based on role
  const getNavigationItems = (): NavItem[] => {
    const basePath = getRoleBasePath();

    switch (user?.role) {
      case "manager_hr":
        return [
          { name: "Dashboard", href: basePath, icon: "LayoutDashboard" },
          { name: "Pegawai", href: `${basePath}/pegawai`, icon: "Users" },
          { name: "Departemen", href: `${basePath}/departemen`, icon: "Building2" },
          { name: "Biometrik", href: `${basePath}/biometrik`, icon: "Fingerprint" },
          { name: "Geofencing", href: `${basePath}/geofencing`, icon: "MapPin" },
          { name: "Presensi", href: `${basePath}/presensi`, icon: "ClipboardCheck" },
          { name: "Jam Kerja", href: `${basePath}/jam-kerja`, icon: "Clock" },
          { name: "Keamanan", href: `${basePath}/keamanan`, icon: "Shield" },
          { name: "Audit Log", href: `${basePath}/audit-log`, icon: "FileText" },
          { name: "Notifikasi", href: `${basePath}/notifikasi`, icon: "Bell" },
        ];

      case "manager_departemen":
        return [
          { name: "Dashboard", href: basePath, icon: "LayoutDashboard" },
          { name: "Tim Saya", href: `${basePath}/tim`, icon: "Users" },
          { name: "Presensi", href: `${basePath}/presensi`, icon: "ClipboardCheck" },
          { name: "Persetujuan", href: `${basePath}/persetujuan`, icon: "FileText" },
          { name: "Laporan", href: `${basePath}/laporan`, icon: "FileText" },
          { name: "Notifikasi", href: `${basePath}/notifikasi`, icon: "Bell" },
        ];

      case "admin_departemen":
        return [
          { name: "Dashboard", href: basePath, icon: "LayoutDashboard" },
          { name: "Kelola Staf", href: `${basePath}/staf`, icon: "UserCog" },
          { name: "Presensi", href: `${basePath}/presensi`, icon: "ClipboardCheck" },
          { name: "Laporan", href: `${basePath}/laporan`, icon: "FileText" },
          { name: "Notifikasi", href: `${basePath}/notifikasi`, icon: "Bell" },
        ];

      case "staf":
        return [
          { name: "Dashboard", href: basePath, icon: "LayoutDashboard" },
          { name: "Presensi", href: `${basePath}/presensi`, icon: "Camera" },
          { name: "Riwayat", href: `${basePath}/riwayat`, icon: "Calendar" },
          { name: "Notifikasi", href: `${basePath}/notifikasi`, icon: "Bell" },
        ];

      default:
        return [];
    }
  };

  const navigationItems = getNavigationItems();

  // Get user initials
  const getUserInitials = () => {
    if (!user?.full_name) return "U";
    const names = user.full_name.split(" ");
    if (names.length >= 2) {
      return names[0][0] + names[1][0];
    }
    return names[0][0];
  };

  // Get role display name
  const getRoleDisplay = () => {
    switch (user?.role) {
      case "manager_hr":
        return "Manager HR";
      case "manager_departemen":
        return "Manager Departemen";
      case "admin_departemen":
        return "Admin Departemen";
      case "staf":
        return "Staf";
      default:
        return "User";
    }
  };

  return (
    <div className="flex h-screen w-64 flex-col border-r border-gray-200 bg-white">
      {/* Header */}
      <div className="flex items-center gap-3 border-b border-gray-200 px-6 py-4">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-600 text-white font-bold">
          HR
        </div>
        <div>
          <h2 className="text-sm font-semibold text-gray-900">Dashboard HRIS</h2>
          <p className="text-xs text-gray-500">{user?.department || "System"}</p>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 px-3 py-4 overflow-y-auto">
        {navigationItems.map((item) => {
          const Icon = iconMap[item.icon];
          const isActive = pathname === item.href;

          return (
            <Link
              key={item.name}
              href={item.href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors",
                isActive
                  ? "bg-blue-50 text-blue-600"
                  : "text-gray-700 hover:bg-gray-100"
              )}
            >
              <Icon className="h-5 w-5" />
              <span>{item.name}</span>
            </Link>
          );
        })}
      </nav>

      {/* User Profile */}
      <div className="border-t border-gray-200 p-4">
        <div className="flex items-center gap-3">
          <Avatar>
            <AvatarFallback className="bg-blue-600 text-white">
              {getUserInitials()}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-gray-900 truncate">
              {user?.full_name || "User"}
            </p>
            <p className="text-xs text-gray-500 truncate">
              {getRoleDisplay()}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}