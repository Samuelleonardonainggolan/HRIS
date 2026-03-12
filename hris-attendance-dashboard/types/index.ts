// types/index.ts

// ==================== EMPLOYEE TYPES ====================

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
  // Additional fields for employee management
  birthDate?: string;
  religion?: string;
  education?: string;
  yearEnrolled?: string;
  officeEmail?: string;
  isActive?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateEmployeeRequest {
  payroll_number: string;
  nik?: string;
  email: string;
  office_email?: string;
  password?: string;
  full_name: string;
  birth_date: string;
  religion: string;
  last_education: string;
  year_enrolled: string;
  employment_status: string;
  department_id: string;
  position_id: string;
  phone: string;
  phone_number?: string;
  address: string;
  role: string;
  is_active?: boolean;
}

export interface UpdateEmployeeRequest {
  fullName?: string;
  birthDate?: string;
  religion?: string;
  lastEducation?: string;
  yearEnrolled?: string;
  employmentStatus?: string;
  department?: string;
  position?: string;
  officeEmail?: string;
  phoneNumber?: string;
  address?: string;
  status?: 'AKTIF' | 'NONAKTIF';
}

export interface EmployeeFilter {
  search?: string;
  department?: string;
  position?: string;
  status?: 'AKTIF' | 'NONAKTIF' | 'HADIR' | 'TELAMBAT' | 'IZIN' | 'ALPHA';
  employmentStatus?: string;
}

// ==================== DASHBOARD STATS ====================

export interface Stats {
  totalEmployees: number;
  totalEmployeesTrend: number;
  presentToday: number;
  presentPercentage: number;
  lateToday: number;
  leaveRequests: number;
  leaveRequestsTrend: number;
}

// ==================== NAVIGATION ====================

export interface NavItem {
  name: string;
  href: string;
  icon: string;
  active?: boolean;
  children?: NavItem[];
  badge?: number | string;
}

// ==================== DEPARTMENT TYPES ====================

export interface Department {
  id: string;
  code: string;
  name: string;
  description?: string;
  managerId?: string;
  managerName?: string;
  totalEmployees?: number;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateDepartmentRequest {
  code: string;
  name: string;
  description?: string;
  managerId?: string;
}

export interface UpdateDepartmentRequest {
  name?: string;
  description?: string;
  managerId?: string;
  isActive?: boolean;
}

// ==================== POSITION TYPES ====================

export interface Position {
  id: string;
  departmentId: string;
  departmentName?: string;
  code: string;
  name: string;
  level: number;
  grade: string;
  description?: string;
  responsibilities?: string[];
  requirements?: string[];
  salaryRange?: {
    min: number;
    max: number;
    currency: string;
  };
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreatePositionRequest {
  departmentId: string;
  code: string;
  name: string;
  level: number;
  grade: string;
  description?: string;
  responsibilities?: string[];
  requirements?: string[];
  salaryRange?: {
    min: number;
    max: number;
    currency: string;
  };
}

export interface UpdatePositionRequest {
  name?: string;
  level?: number;
  grade?: string;
  description?: string;
  responsibilities?: string[];
  requirements?: string[];
  salaryRange?: {
    min: number;
    max: number;
    currency: string;
  };
  isActive?: boolean;
}

// ==================== ATTENDANCE TYPES ====================

export interface Attendance {
  id: string;
  employeeId: string;
  employeeName: string;
  employeeNik: string;
  date: string;
  checkInTime?: string;
  checkOutTime?: string;
  status: 'HADIR' | 'TELAMBAT' | 'IZIN' | 'ALPHA' | 'CUTI';
  workHours?: number;
  overtime?: number;
  location?: {
    latitude: number;
    longitude: number;
    address: string;
  };
  verified: {
    biometric: boolean;
    geofencing: boolean;
  };
  notes?: string;
  approvedBy?: string;
  approvedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CheckInRequest {
  latitude: number;
  longitude: number;
  faceEmbedding?: number[];
  photoUrl?: string;
}

export interface CheckOutRequest {
  latitude: number;
  longitude: number;
  notes?: string;
}

export interface AttendanceFilter {
  startDate?: string;
  endDate?: string;
  employeeId?: string;
  department?: string;
  status?: 'HADIR' | 'TELAMBAT' | 'IZIN' | 'ALPHA' | 'CUTI';
}

// ==================== LEAVE REQUEST TYPES ====================

export interface LeaveRequest {
  id: string;
  employeeId: string;
  employeeName: string;
  employeeNik: string;
  leaveType: 'CUTI' | 'SAKIT' | 'IZIN';
  startDate: string;
  endDate: string;
  totalDays: number;
  reason: string;
  attachment?: string;
  status: 'PENDING' | 'APPROVED' | 'REJECTED';
  approvedBy?: string;
  approvedAt?: string;
  rejectionReason?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateLeaveRequest {
  leaveType: 'CUTI' | 'SAKIT' | 'IZIN';
  startDate: string;
  endDate: string;
  reason: string;
  attachment?: string;
}

export interface ApproveLeaveRequest {
  status: 'APPROVED' | 'REJECTED';
  rejectionReason?: string;
}

// ==================== FACE RECOGNITION TYPES ====================

export interface FaceEmbedding {
  id: string;
  userId: string;
  faceEmbedding: number[];
  faceImageUrl?: string;
  isActive: boolean;
  registeredAt: string;
  lastUpdatedAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface RegisterFaceRequest {
  userId: string;
  faceEmbedding: number[];
  faceImageUrl?: string;
}

export interface FaceRecognitionRequest {
  faceEmbedding: number[];
  threshold?: number;
}

export interface FaceRecognitionResponse {
  success: boolean;
  userId?: string;
  similarity?: number;
  message: string;
}

// ==================== AUTH TYPES ====================

export interface User {
  id: string;
  nik: string;
  email: string;
  fullName: string;
  role: 'manager_hr' | 'manager_departemen' | 'admin_departemen' | 'staf';
  department: string;
  position: string;
  phone?: string;
  avatar?: string;
  joinDate: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface AuthState {
  user: User | null;
  loading: boolean;
  isAuthenticated: boolean;
}

// ==================== API RESPONSE TYPES ====================

export interface ApiResponse<T = unknown> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
}

export interface PaginatedResponse<T = unknown> {
  success: boolean;
  message: string;
  data: T[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
  };
}

export interface ErrorResponse {
  success: false;
  message: string;
  error: string;
  statusCode?: number;
}

// ==================== FORM TYPES ====================

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface FormError {
  field: string;
  message: string;
}

export interface FormState<T = unknown> {
  data: T;
  errors: FormError[];
  isSubmitting: boolean;
  isValid: boolean;
}

// ==================== TABLE TYPES ====================

export interface TableColumn<T = unknown> {
  key: keyof T | string;
  label: string;
  sortable?: boolean;
  render?: (item: T) => React.ReactNode;
  className?: string;
}

export interface TableSort {
  key: string;
  direction: 'asc' | 'desc';
}

export interface TablePagination {
  page: number;
  limit: number;
  total: number;
}

// ==================== NOTIFICATION TYPES ====================

export interface Notification {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  title: string;
  message: string;
  read: boolean;
  createdAt: string;
}

// ==================== REPORT TYPES ====================

export interface AttendanceReport {
  month: string;
  year: number;
  department: string;
  totalEmployees: number;
  totalPresent: number;
  totalLate: number;
  totalAbsent: number;
  totalLeave: number;
  attendanceRate: number;
}

export interface EmployeeReport {
  department: string;
  totalEmployees: number;
  activeEmployees: number;
  inactiveEmployees: number;
  newHires: number;
  resignations: number;
}

// ==================== UTILITY TYPES ====================

export type Status = 'idle' | 'loading' | 'success' | 'error';

export type ActionType = 'create' | 'read' | 'update' | 'delete';

export type Role = 'manager_hr' | 'manager_departemen' | 'admin_departemen' | 'staf';

export type EmploymentStatus = 'TETAP' | 'KONTRAK' | 'MAGANG' | 'OUTSOURCING';

export type AttendanceStatus = 'HADIR' | 'TELAMBAT' | 'IZIN' | 'ALPHA' | 'CUTI';

export type LeaveType = 'CUTI' | 'SAKIT' | 'IZIN';

export type ApprovalStatus = 'PENDING' | 'APPROVED' | 'REJECTED';

// ==================== CONSTANTS ====================

export const ROLES = {
  MANAGER_HR: 'manager_hr',
  MANAGER_DEPT: 'manager_departemen',
  ADMIN_DEPT: 'admin_departemen',
  STAFF: 'staf',
} as const;

export const EMPLOYMENT_STATUS = {
  PERMANENT: 'TETAP',
  CONTRACT: 'KONTRAK',
  INTERN: 'MAGANG',
  OUTSOURCING: 'OUTSOURCING',
} as const;

export const ATTENDANCE_STATUS = {
  PRESENT: 'HADIR',
  LATE: 'TELAMBAT',
  PERMISSION: 'IZIN',
  ABSENT: 'ALPHA',
  LEAVE: 'CUTI',
} as const;

export const LEAVE_TYPES = {
  ANNUAL: 'CUTI',
  SICK: 'SAKIT',
  PERMISSION: 'IZIN',
} as const;

export const APPROVAL_STATUS = {
  PENDING: 'PENDING',
  APPROVED: 'APPROVED',
  REJECTED: 'REJECTED',
} as const;
