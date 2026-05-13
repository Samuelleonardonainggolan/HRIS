import { authService } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.startsWith("/")
    ? process.env.NEXT_PUBLIC_API_URL
    : "/api/v1";

export type ReportType = "late" | "missing_clock_out" | "leave" | "permission" | "overtime";

export type AttendanceActivityReportRow = {
  id: string;
  date: string;
  dayLabel: string;
  employee_name: string;
  employee_nik: string;
  employee_initials: string;
  department_name: string;
  position_name: string;
  type: ReportType;
  
  scheduled_in?: string;
  actual_in?: string;
  scheduled_out?: string;
  actual_out?: string;
  
  overtime_start?: string;
  overtime_end?: string;
  overtime_hours?: number;
  
  date_range?: string;
  
  late_minutes?: number;
  
  approval_status?: string;
  impact: "deduction" | "no_impact" | "addition" | "needs_review";
  note?: string;
};

export type AttendanceActivitySummaryItem = {
  events: number;
  unique: number;
  hours?: number;
};

export type AttendanceActivityReportSummary = {
  late: AttendanceActivitySummaryItem;
  missing: AttendanceActivitySummaryItem;
  leave: AttendanceActivitySummaryItem;
  permission: AttendanceActivitySummaryItem;
  overtime: AttendanceActivitySummaryItem;
  total: number;
};

export type AttendanceActivityReportResponse = {
  rows: AttendanceActivityReportRow[];
  summary: AttendanceActivityReportSummary;
};

function buildUrl(path: string, params?: Record<string, string | undefined>) {
  const base = API_BASE.startsWith("http")
    ? API_BASE
    : `${typeof window !== "undefined" ? window.location.origin : "http://localhost:3000"}${API_BASE}`;
  const url = new URL(`${base}${path}`);
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      if (value) url.searchParams.set(key, value);
    }
  }
  return url.toString();
}

export const getAttendanceActivityReport = async (params: {
  period?: string;
  department_id?: string;
  type?: string;
  status?: string;
  search?: string;
}) => {
  const res = await fetch(
    buildUrl("/reports/attendance-activity", params),
    {
      method: "GET",
      headers: authService.getAuthHeaders(),
    }
  );
  
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || data.message || "Gagal mengambil data laporan");
  }
  
  return data.data as AttendanceActivityReportResponse;
};
