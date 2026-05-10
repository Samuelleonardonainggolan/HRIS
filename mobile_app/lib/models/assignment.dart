// lib/models/assignment.dart

class Assignment {
  final String id;
  final String departmentId;
  final String departmentName;
  final String requestedById;
  final String requestedByName;
  final DateTime date;
  final String reason;
  final String status;
  final String? notes;
  final String startTime;
  final String endTime;
  final List<AssignmentEmployee> employees;
  final DateTime createdAt;
  final DateTime updatedAt;

  Assignment({
    required this.id,
    required this.departmentId,
    required this.departmentName,
    required this.requestedById,
    required this.requestedByName,
    required this.date,
    required this.reason,
    required this.status,
    this.notes,
    required this.startTime,
    required this.endTime,
    required this.employees,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Assignment.fromJson(Map<String, dynamic> json) {
    DateTime parseDate(dynamic v) {
      if (v == null) return DateTime.now();
      try {
        return DateTime.parse(v.toString()).toLocal();
      } catch (_) {
        return DateTime.now();
      }
    }

    final emps = (json['employees'] as List?)
            ?.whereType<Map<String, dynamic>>()
            .map(AssignmentEmployee.fromJson)
            .toList() ??
        [];

    return Assignment(
      id: (json['id'] ?? '').toString(),
      departmentId: (json['department_id'] ?? '').toString(),
      departmentName: (json['department_name'] ?? json['departmentName'] ?? '').toString(),
      requestedById: (json['requested_by_id'] ?? '').toString(),
      requestedByName: (json['requested_by_name'] ?? json['requestedByName'] ?? '').toString(),
      date: parseDate(json['date']),
      reason: (json['reason'] ?? '').toString(),
      status: (json['status'] ?? 'draft').toString().toLowerCase(),
      notes: json['notes']?.toString(),
      startTime: (json['start_time'] ?? json['shift_start'] ?? json['startTime'] ?? '').toString(),
      endTime: (json['end_time'] ?? json['shift_end'] ?? json['endTime'] ?? '').toString(),
      employees: emps,
      createdAt: parseDate(json['created_at'] ?? json['createdAt']),
      updatedAt: parseDate(json['updated_at'] ?? json['updatedAt']),
    );
  }

  String get statusDisplay {
    switch (status) {
      case 'draft':
        return 'Draft';
      case 'submitted':
        return 'Dikirim';
      case 'published':
        return 'Dipublikasikan';
      case 'cancelled':
        return 'Dibatalkan';
      default:
        return status;
    }
  }

  bool get isDraft => status == 'draft';
  bool get isSubmitted => status == 'submitted';
  bool get isPublished => status == 'published';
}

class AssignmentEmployee {
  final String userId;
  final String fullName;
  final String payrollNumber;
  final String positionName;
  final String originalStartTime;
  final String originalEndTime;
  final String assignedStartTime;
  final String assignedEndTime;
  final String employeeStatus;
  final String? rejectionNote;
  final DateTime? confirmedAt;
  final bool dayOffEligible;
  final String dayOffStatus;
  final DateTime? dayOffGrantedAt;
  final DateTime? dayOffUsedAt;
  final DateTime? replacementOffDate;

  AssignmentEmployee({
    required this.userId,
    required this.fullName,
    required this.payrollNumber,
    required this.positionName,
    required this.originalStartTime,
    required this.originalEndTime,
    required this.assignedStartTime,
    required this.assignedEndTime,
    required this.employeeStatus,
    this.rejectionNote,
    this.confirmedAt,
    required this.dayOffEligible,
    required this.dayOffStatus,
    this.dayOffGrantedAt,
    this.dayOffUsedAt,
    this.replacementOffDate,
  });

  factory AssignmentEmployee.fromJson(Map<String, dynamic> json) {
    DateTime? parseMaybeDate(dynamic v) {
      if (v == null) return null;
      try {
        return DateTime.parse(v.toString()).toLocal();
      } catch (_) {
        return null;
      }
    }

    return AssignmentEmployee(
      userId: (json['user_id'] ?? json['userId'] ?? '').toString(),
      fullName: (json['full_name'] ?? json['fullName'] ?? json['name'] ?? '').toString(),
      payrollNumber: (json['payroll_number'] ?? '').toString(),
      positionName: (json['position_name'] ?? '').toString(),
      originalStartTime: (json['original_start_time'] ?? '').toString(),
      originalEndTime: (json['original_end_time'] ?? '').toString(),
      assignedStartTime: (json['assigned_start_time'] ?? json['assignedStartTime'] ?? '').toString(),
      assignedEndTime: (json['assigned_end_time'] ?? json['assignedEndTime'] ?? '').toString(),
      employeeStatus: (json['employee_status'] ?? 'pending').toString().toLowerCase(),
      rejectionNote: json['rejection_note']?.toString(),
      confirmedAt: parseMaybeDate(json['confirmed_at'] ?? json['confirmedAt']),
      dayOffEligible: (json['day_off_eligible'] == true) || (json['dayOffEligible'] == true),
      dayOffStatus: (json['day_off_status'] ?? json['dayOffStatus'] ?? '').toString(),
      dayOffGrantedAt: parseMaybeDate(json['day_off_granted_at'] ?? json['dayOffGrantedAt']),
      dayOffUsedAt: parseMaybeDate(json['day_off_used_at'] ?? json['dayOffUsedAt']),
      replacementOffDate: parseMaybeDate(json['replacement_off_date'] ?? json['replacementOffDate']),
    );
  }
}
