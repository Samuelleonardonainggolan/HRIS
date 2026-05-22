// lib/models/overtime_request.dart

// NEW FLOW: Kepala departemen mengajukan lembur untuk karyawan mereka
// Employees kemudian bisa agree/reject
// HR bisa publish surat

class OvertimeRequest {
  final String id;
  final String departmentId;
  final String departmentName;
  final String requestedById; // Kepala departemen yang submit
  final String requestedByName; // Display name
  final DateTime date;
  final String startTime; // format HH:mm
  final String endTime; // format HH:mm
  final String reason;
  final String status; // draft|submitted|published
  final String? notes;
  final String? letterUrl;
  final List<OvertimeEmployee> employees;
  final DateTime createdAt;
  final DateTime updatedAt;

  OvertimeRequest({
    required this.id,
    required this.departmentId,
    required this.departmentName,
    required this.requestedById,
    required this.requestedByName,
    required this.date,
    required this.startTime,
    required this.endTime,
    required this.reason,
    required this.status,
    this.notes,
    this.letterUrl,
    required this.employees,
    required this.createdAt,
    required this.updatedAt,
  });

  factory OvertimeRequest.fromJson(Map<String, dynamic> json) {
    return OvertimeRequest(
      id: json['id']?.toString() ?? '',
      departmentId: json['department_id']?.toString() ?? '',
      departmentName: (json['department_name'] ?? json['departmentName'] ?? '')
          .toString(),
      requestedById: json['requested_by_id']?.toString() ?? '',
      requestedByName:
          (json['requested_by_name'] ??
                  json['requestedByName'] ??
                  json['requested_by_id'] ??
                  '')
              .toString(),
      date: _parseDate(json['date']),
      startTime: json['start_time']?.toString() ?? '',
      endTime: json['end_time']?.toString() ?? '',
      reason: json['reason']?.toString() ?? '',
      status: (json['status'] ?? 'draft').toString().toLowerCase(),
      notes: json['notes']?.toString(),
      letterUrl: json['letter_url']?.toString(),
      employees:
          (json['employees'] as List?)
              ?.whereType<Map<String, dynamic>>()
              .map(OvertimeEmployee.fromJson)
              .toList() ??
          [],
      createdAt: _parseDate(json['created_at']),
      updatedAt: _parseDate(json['updated_at']),
    );
  }

  static DateTime _parseDate(dynamic v) {
    if (v == null) return DateTime.now();
    try {
      return DateTime.parse(v.toString()).toLocal();
    } catch (_) {
      return DateTime.now();
    }
  }

  // Helper untuk UI
  String get statusDisplay {
    switch (status) {
      case 'draft':
        return 'Draft';
      case 'submitted':
        return 'Dikirim';
      case 'published':
        return 'Dipublikasikan';
      default:
        return 'Unknown';
    }
  }

  bool get isDraft => status == 'draft';
  bool get isSubmitted => status == 'submitted';
  bool get isPublished => status == 'published';

  // Get hour & minute from time string
  (int, int) getStartTimeHourMin() {
    try {
      final parts = startTime.split(':');
      return (int.parse(parts[0]), int.parse(parts[1]));
    } catch (_) {
      return (0, 0);
    }
  }

  (int, int) getEndTimeHourMin() {
    try {
      final parts = endTime.split(':');
      return (int.parse(parts[0]), int.parse(parts[1]));
    } catch (_) {
      return (0, 0);
    }
  }

  // Calculate duration in hours
  double getDurationHours() {
    try {
      final start = startTime.split(':');
      final end = endTime.split(':');
      final startMins = int.parse(start[0]) * 60 + int.parse(start[1]);
      final endMins = int.parse(end[0]) * 60 + int.parse(end[1]);
      return (endMins - startMins) / 60;
    } catch (_) {
      return 0;
    }
  }
}

class OvertimeEmployee {
  final String userId;
  final String userName; // Display name
  final String employeeStatus; // pending|agreed|rejected
  final String? rejectionNote;
  final String? letterUrl; // URL dokumen SPKL untuk karyawan ini
  final DateTime? confirmedAt;
  final OvertimeReward? reward;

  OvertimeEmployee({
    required this.userId,
    required this.userName,
    required this.employeeStatus,
    this.rejectionNote,
    this.letterUrl,
    this.confirmedAt,
    this.reward,
  });

  factory OvertimeEmployee.fromJson(Map<String, dynamic> json) {
    return OvertimeEmployee(
      userId: json['user_id']?.toString() ?? '',
      userName: (json['full_name'] ?? json['user_name'] ?? json['name'] ?? '')
          .toString(),
      employeeStatus: (json['employee_status'] ?? 'pending')
          .toString()
          .toLowerCase(),
      rejectionNote: json['rejection_note']?.toString(),
      letterUrl: json['letter_url']?.toString(),
      confirmedAt: json['confirmed_at'] != null
          ? DateTime.tryParse(json['confirmed_at'].toString())
          : null,
      reward: json['reward'] != null
          ? OvertimeReward.fromJson(json['reward'] as Map<String, dynamic>)
          : null,
    );
  }

  String get displayName => userName.isNotEmpty ? userName : userId;

  String get statusDisplay {
    switch (employeeStatus) {
      case 'pending':
        return 'Menunggu';
      case 'agreed':
        return 'Setuju';
      case 'rejected':
        return 'Tolak';
      default:
        return 'Unknown';
    }
  }

  bool get isPending => employeeStatus == 'pending';
  bool get isAgreed => employeeStatus == 'agreed';
  bool get isRejected => employeeStatus == 'rejected';
}

class OvertimeReward {
  final String rewardType; // money|time_off
  final String? rewardOption; // early_out|late_in
  final String status; // none|pending|granted|used
  final DateTime? rewardDate; // Tanggal klaim reward (terutama untuk time_off)
  final DateTime? grantedAt;
  final DateTime? usedAt;

  OvertimeReward({
    required this.rewardType,
    this.rewardOption,
    required this.status,
    this.rewardDate,
    this.grantedAt,
    this.usedAt,
  });

  factory OvertimeReward.fromJson(Map<String, dynamic> json) {
    return OvertimeReward(
      rewardType: json['reward_type']?.toString() ?? '',
      rewardOption: json['reward_option']?.toString(),
      status: json['status']?.toString() ?? 'none',
      rewardDate: json['reward_date'] != null
          ? DateTime.tryParse(json['reward_date'].toString())
          : null,
      grantedAt: json['granted_at'] != null
          ? DateTime.tryParse(json['granted_at'].toString())
          : null,
      usedAt: json['used_at'] != null
          ? DateTime.tryParse(json['used_at'].toString())
          : null,
    );
  }

  String get rewardTypeDisplay {
    switch (rewardType) {
      case 'money':
        return 'Uang Lembur';
      case 'time_off':
        return 'Jam Kerja Dipercepat';
      default:
        return 'Belum Dipilih';
    }
  }

  String get statusDisplay {
    switch (status) {
      case 'none':
        return 'Belum Dipilih';
      case 'pending':
        return 'Menunggu Persetujuan';
      case 'granted':
        return 'Diterima';
      case 'used':
        return 'Sudah Digunakan';
      default:
        return 'Unknown';
    }
  }
}
