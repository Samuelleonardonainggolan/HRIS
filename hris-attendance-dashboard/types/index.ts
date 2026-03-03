export interface Employee {
  id: string;
  name: string;
  avatar?: string;
  nik: string;
  department: string;
  position: string;
  checkInTime?: string;
  status: 'AKTIF' | 'NONAKTIF' | 'HADIR' | 'TELAMBAT' | 'IZIN' | 'ALPHA';
  verified?: {
    biometric: boolean;
    geofencing: boolean;
  };
  email?: string;
  phone?: string;
  address?: string;
  joinDate?: string;
  employmentStatus?: string;
  workYears?: number;
}

export interface Stats {
  totalEmployees: number;
  totalEmployeesTrend: number;
  presentToday: number;
  presentPercentage: number;
  lateToday: number;
  leaveRequests: number;
  leaveRequestsTrend: number;
}

export interface NavItem {
  name: string;
  href: string;
  icon: string;
  active?: boolean;
}