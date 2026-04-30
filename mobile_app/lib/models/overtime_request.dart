// lib/models/overtime_request.dart

class OvertimeRequest {
  final String id;
  final String userId;
  final DateTime date;
  final DateTime startTime;
  final DateTime endTime;
  final String reason;
  final String total;
  final String statusKepalaDepartemen;
  final String? kepalaDepartemenId;
  final String statusManagerHr;
  final String? managerHrId;
  final String finalStatus;
  final String? rejectionReasonKepalaDept;
  final String? rejectionReasonManagerHr;
  final DateTime createdAt;
  final DateTime updatedAt;

  OvertimeRequest({
    required this.id,
    required this.userId,
    required this.date,
    required this.startTime,
    required this.endTime,
    required this.reason,
    required this.total,
    required this.statusKepalaDepartemen,
    this.kepalaDepartemenId,
    required this.statusManagerHr,
    this.managerHrId,
    required this.finalStatus,
    this.rejectionReasonKepalaDept,
    this.rejectionReasonManagerHr,
    required this.createdAt,
    required this.updatedAt,
  });

  factory OvertimeRequest.fromJson(Map<String, dynamic> json) {
    return OvertimeRequest(
      id: json['id']?.toString() ?? '',
      userId: json['user_id']?.toString() ?? '',
      date: _parseDate(json['date']),
      startTime: _parseDate(json['start_time']),
      endTime: _parseDate(json['end_time']),
      reason: json['reason']?.toString() ?? '',
      total: json['total']?.toString() ?? '',
      statusKepalaDepartemen: (json['status_kepala_departemen'] ?? 'PENDING').toString().toUpperCase(),
      kepalaDepartemenId: _normalizeString(json['kepala_departemen_id']),
      statusManagerHr: (json['status_manager_hr'] ?? 'PENDING').toString().toUpperCase(),
      managerHrId: _normalizeString(json['manager_hr_id']),
      finalStatus: (json['final_status'] ?? 'PENDING').toString().toUpperCase(),
      rejectionReasonKepalaDept: json['rejection_reason_kepala_dept']?.toString(),
      rejectionReasonManagerHr: json['rejection_reason_manager_hr']?.toString(),
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

  static String? _normalizeString(dynamic value) {
    final str = value?.toString().trim();
    return (str != null && str.isNotEmpty) ? str : null;
  }

  /// Helper untuk UI
  String get statusDisplay {
    switch (finalStatus) {
      case 'APPROVED':
        return 'Disetujui';
      case 'REJECTED':
        return 'Ditolak';
      case 'CANCELLED':
        return 'Dibatalkan';
      default:
        return 'Menunggu';
    }
  }

  bool get isApproved => finalStatus == 'APPROVED';
  bool get isPending => finalStatus == 'PENDING';
  bool get isRejected => finalStatus == 'REJECTED';

  String? get primaryRejectionReason {
    final hr = rejectionReasonManagerHr?.trim();
    if (hr != null && hr.isNotEmpty) return hr;
    final kepala = rejectionReasonKepalaDept?.trim();
    if (kepala != null && kepala.isNotEmpty) return kepala;
    return null;
  }
}
