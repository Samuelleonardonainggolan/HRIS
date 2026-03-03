// lib/models/attendance_model.dart
class AttendanceRecord {
  final String id;
  final DateTime date;
  final String clockIn;
  final String clockOut;
  final String status;
  final double workHours;
  final double overtimeHours;

  AttendanceRecord({
    required this.id,
    required this.date,
    required this.clockIn,
    required this.clockOut,
    required this.status,
    required this.workHours,
    required this.overtimeHours,
  });
}

class LeaveRequest {
  final String id;
  final String type;
  final DateTime startDate;
  final DateTime endDate;
  final String reason;
  final String status;
  final int days;

  LeaveRequest({
    required this.id,
    required this.type,
    required this.startDate,
    required this.endDate,
    required this.reason,
    required this.status,
    required this.days,
  });
}