// lib/models/attendance_model.dart
class AttendanceRecord {
  final String id;
  final DateTime date;
  final String clockIn;
  final String clockOut;
  final String status;
  final double workHours;
  final double overtimeHours;
  final double? faceSimilarity;

  AttendanceRecord({
    required this.id,
    required this.date,
    required this.clockIn,
    required this.clockOut,
    required this.status,
    required this.workHours,
    required this.overtimeHours,
    this.faceSimilarity,
  });

  factory AttendanceRecord.fromJson(Map<String, dynamic> json) {
    return AttendanceRecord(
      id: json['id'] ?? '',
      date: DateTime.parse(json['date'] ?? DateTime.now().toIso8601String()),
      clockIn: json['clock_in_time'] ?? '--:--',
      clockOut: json['clock_out_time'] ?? '--:--',
      status: json['status'] ?? 'Unknown',
      workHours: (json['work_hours'] ?? 0).toDouble(),
      overtimeHours: (json['overtime_hours'] ?? 0).toDouble(),
      faceSimilarity: json['face_similarity']?.toDouble(),
    );
  }
}

class MonthlyAttendanceSummary {
  final String month;
  final int year;
  final int totalDays;
  final double totalHours;
  final double overtimeHours;
  final List<AttendanceRecord> records;

  MonthlyAttendanceSummary({
    required this.month,
    required this.year,
    required this.totalDays,
    required this.totalHours,
    required this.overtimeHours,
    required this.records,
  });

  factory MonthlyAttendanceSummary.fromJson(Map<String, dynamic> json) {
    var recordsJson = json['records'] as List? ?? [];
    List<AttendanceRecord> records = recordsJson
        .map((item) => AttendanceRecord.fromJson(item))
        .toList();

    return MonthlyAttendanceSummary(
      month: json['month'] ?? '',
      year: json['year'] ?? DateTime.now().year,
      totalDays: json['total_days'] ?? 0,
      totalHours: (json['total_hours'] ?? 0).toDouble(),
      overtimeHours: (json['overtime_hours'] ?? 0).toDouble(),
      records: records,
    );
  }
}

class AttendanceProcessResult {
  final bool success;
  final String message;
  final double faceSimilarity;
  final bool locationValid;
  final double distance;
  final AttendanceRecord? attendance;

  AttendanceProcessResult({
    required this.success,
    required this.message,
    required this.faceSimilarity,
    required this.locationValid,
    required this.distance,
    this.attendance,
  });

  factory AttendanceProcessResult.fromJson(Map<String, dynamic> json) {
    return AttendanceProcessResult(
      success: json['success'] ?? false,
      message: json['message'] ?? '',
      faceSimilarity: (json['face_similarity'] ?? 0).toDouble(),
      locationValid: json['location_valid'] ?? false,
      distance: (json['distance_m'] ?? 0).toDouble(),
      attendance: json['attendance'] != null
          ? AttendanceRecord.fromJson(json['attendance'])
          : null,
    );
  }
}