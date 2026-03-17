export type AttendanceEmployeeSummary = {
  nik: string;
  name: string;
  department: string;
  workDays: number;
  present: number;
  late: number;
  sickLeave: number;
  absent: number;
};

export type AttendanceKpi = {
  avgPresentPct: number;
  totalEmployees: number;
  totalLateIncidents: number;
  totalAbsentIncidents: number;
};

export const departmentsMock = [
  "Semua Departemen",
  "IT Support",
  "Human Resource",
  "Marketing",
  "Finance",
];

export const attendanceSummaryMock: AttendanceEmployeeSummary[] = [
  {
    nik: "ID-99210",
    name: "Rizky Darmawan",
    department: "IT Support",
    workDays: 22,
    present: 21,
    late: 1,
    sickLeave: 0,
    absent: 0,
  },
  {
    nik: "ID-99215",
    name: "Siti Aminah",
    department: "Human Resource",
    workDays: 22,
    present: 22,
    late: 0,
    sickLeave: 0,
    absent: 0,
  },
  {
    nik: "ID-99302",
    name: "Bambang Pamungkas",
    department: "Marketing",
    workDays: 22,
    present: 18,
    late: 2,
    sickLeave: 2,
    absent: 2,
  },
  {
    nik: "ID-99310",
    name: "Dewi Lestari",
    department: "Finance",
    workDays: 22,
    present: 21,
    late: 0,
    sickLeave: 1,
    absent: 0,
  },
];

export const attendanceKpiMock: AttendanceKpi = {
  avgPresentPct: 92.4,
  totalEmployees: 120,
  totalLateIncidents: 14,
  totalAbsentIncidents: 5,
};

