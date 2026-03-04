class LeaveRequest {
  final String id;
  final String type;
  final DateTime startDate;
  final DateTime endDate;
  final String reason;
  final String status;
  final int days;
  final String? attachmentUrl;
  final DateTime? createdAt;
  final String? approvedBy;
  final DateTime? approvedAt;

  LeaveRequest({
    required this.id,
    required this.type,
    required this.startDate,
    required this.endDate,
    required this.reason,
    required this.status,
    required this.days,
    this.attachmentUrl,
    this.createdAt,
    this.approvedBy,
    this.approvedAt,
  });

  factory LeaveRequest.fromJson(Map<String, dynamic> json) {
    return LeaveRequest(
      id: json['id'] ?? json['_id'] ?? '',
      type: json['type'] ?? '',
      startDate: DateTime.parse(json['start_date'] ?? json['startDate'] ?? DateTime.now().toIso8601String()),
      endDate: DateTime.parse(json['end_date'] ?? json['endDate'] ?? DateTime.now().toIso8601String()),
      reason: json['reason'] ?? '',
      status: json['status'] ?? 'Pending',
      days: json['days'] ?? 0,
      attachmentUrl: json['attachment_url'],
      createdAt: json['created_at'] != null ? DateTime.parse(json['created_at']) : null,
      approvedBy: json['approved_by'],
      approvedAt: json['approved_at'] != null ? DateTime.parse(json['approved_at']) : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'start_date': startDate.toIso8601String(),
      'end_date': endDate.toIso8601String(),
      'reason': reason,
      'days': days,
    };
  }
}

class AttendanceRecord {
  final String id;
  final DateTime date;
  final String clockIn;
  final String clockOut;
  final String status;
  final double workHours;
  final double overtimeHours;
  final String? photoUrl;
  final double? latitude;
  final double? longitude;
  final String? notes;

  AttendanceRecord({
    required this.id,
    required this.date,
    required this.clockIn,
    required this.clockOut,
    required this.status,
    required this.workHours,
    required this.overtimeHours,
    this.photoUrl,
    this.latitude,
    this.longitude,
    this.notes,
  });

  factory AttendanceRecord.fromJson(Map<String, dynamic> json) {
    return AttendanceRecord(
      id: json['id'] ?? json['_id'] ?? '',
      date: DateTime.parse(json['date'] ?? DateTime.now().toIso8601String()),
      clockIn: json['clock_in'] ?? json['clockIn'] ?? '--:--',
      clockOut: json['clock_out'] ?? json['clockOut'] ?? '--:--',
      status: json['status'] ?? 'Pending',
      workHours: (json['work_hours'] ?? json['workHours'] ?? 0).toDouble(),
      overtimeHours: (json['overtime_hours'] ?? json['overtimeHours'] ?? 0).toDouble(),
      photoUrl: json['photo_url'],
      latitude: json['latitude']?.toDouble(),
      longitude: json['longitude']?.toDouble(),
      notes: json['notes'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'date': date.toIso8601String(),
      'clock_in': clockIn,
      'clock_out': clockOut,
      'status': status,
      'work_hours': workHours,
      'overtime_hours': overtimeHours,
      'notes': notes,
    };
  }
}

class ClockInRequest {
  final String employeeId;
  final double latitude;
  final double longitude;
  final String photoPath;

  ClockInRequest({
    required this.employeeId,
    required this.latitude,
    required this.longitude,
    required this.photoPath,
  });
}

class ClockOutRequest {
  final String employeeId;
  final double latitude;
  final double longitude;
  final String photoPath;

  ClockOutRequest({
    required this.employeeId,
    required this.latitude,
    required this.longitude,
    required this.photoPath,
  });
}

class AttendanceSummary {
  final int totalDays;
  final int presentDays;
  final int absentDays;
  final int lateDays;
  final double totalWorkHours;
  final double totalOvertimeHours;

  AttendanceSummary({
    required this.totalDays,
    required this.presentDays,
    required this.absentDays,
    required this.lateDays,
    required this.totalWorkHours,
    required this.totalOvertimeHours,
  });

  factory AttendanceSummary.fromJson(Map<String, dynamic> json) {
    return AttendanceSummary(
      totalDays: json['total_days'] ?? 0,
      presentDays: json['present_days'] ?? 0,
      absentDays: json['absent_days'] ?? 0,
      lateDays: json['late_days'] ?? 0,
      totalWorkHours: (json['total_work_hours'] ?? 0).toDouble(),
      totalOvertimeHours: (json['total_overtime_hours'] ?? 0).toDouble(),
    );
  }
}