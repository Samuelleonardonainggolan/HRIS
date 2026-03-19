export type PresensiEmployeeSummary = {
  nik: string;
  name: string;
  department: string;
  workDays: number;
  present: number;
  late: number;
  leaveSick: number;
  absent: number;
};

export type PresensiKpi = {
  averageAttendancePercent: number;
  totalEmployees: number;
  totalLateIncidents: number;
  totalAbsentIncidents: number;
};

