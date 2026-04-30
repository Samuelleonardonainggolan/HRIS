// lib/models/leave_request.dart
class LeaveRequest {
  final String id;
  final String type; // nama_tipe dari backend, misal "Izin Sakit"
  final String namaKategori; // "Izin" | "Cuti" | "Lembur"
  final DateTime startDate; // tanggal_mulai
  final DateTime endDate; // tanggal_selesai
  final String reason; // alasan
  final String
  status; // sudah di-map ke Indonesia: Menunggu/Disetujui/Ditolak/Dibatalkan
  final String statusFinal; // raw: PENDING/APPROVED/REJECTED/CANCELLED
  final String statusKepala; // status_kepala_departemen
  final String statusManagerHr; // status_manager_hr
  final String? kepalaDepartemenName;
  final String? managerHrName;
  final String? rejectionReasonKepalaDept;
  final String? rejectionReasonManagerHr;
  final int days; // total_hari
  final String? total; // total / ringkasan kompensasi untuk lembur
  final String? startTime; // lembur: jam mulai
  final String? endTime; // lembur: jam selesai
  final String? dokumenUrl;
  final DateTime createdAt;

  LeaveRequest({
    required this.id,
    required this.type,
    required this.namaKategori,
    required this.startDate,
    required this.endDate,
    required this.reason,
    required this.status,
    required this.statusFinal,
    required this.statusKepala,
    required this.statusManagerHr,
    this.kepalaDepartemenName,
    this.managerHrName,
    this.rejectionReasonKepalaDept,
    this.rejectionReasonManagerHr,
    required this.days,
    this.total,
    this.startTime,
    this.endTime,
    this.dokumenUrl,
    required this.createdAt,
  });

  /// Parse langsung dari response backend PengajuanIzinCuti
  factory LeaveRequest.fromJson(Map<String, dynamic> json) {
    final isOvertimePayload =
      json['date'] != null || json['start_time'] != null || json['end_time'] != null;
    final rawStatus =
        (json['final_status'] ??
                json['status_final'] ??
                _reverseMapStatus((json['status'] ?? 'Menunggu').toString()))
            .toString()
            .toUpperCase();

    final typeName =
      (json['type_name'] ?? json['nama_tipe'] ?? json['type'] ?? (isOvertimePayload ? 'Lembur' : ''))
            .toString();

    final categoryName = (json['category_name'] ?? json['nama_kategori'] ?? '')
      .toString();

    final startRaw =
      json['start_date'] ?? json['tanggal_mulai'] ?? json['date'] ?? json['start_date'];
    final endRaw =
      json['end_date'] ?? json['tanggal_selesai'] ?? json['date'] ?? json['end_date'];

    final daysRaw = json['days_total'] ?? json['total_hari'] ?? json['days'] ?? (isOvertimePayload ? 1 : null);

    return LeaveRequest(
      id: (json['id'] ?? '').toString(),
      type: typeName,
      namaKategori: categoryName.isNotEmpty
          ? categoryName
          : _guessKategori(typeName),
      startDate: _parseDate(startRaw),
      endDate: _parseDate(endRaw),
      reason: (json['reason'] ?? json['alasan'] ?? '').toString(),
      status: _mapStatus(rawStatus),
      statusFinal: rawStatus,
      statusKepala: (json['status_kepala_departemen'] ?? 'PENDING')
          .toString()
          .toUpperCase(),
      statusManagerHr: (json['status_manager_hr'] ?? 'PENDING')
          .toString()
          .toUpperCase(),
      kepalaDepartemenName: _normalizeApproverDisplay(
        json['kepala_departemen_id'],
      ),
      managerHrName: _normalizeApproverDisplay(json['manager_hr_id']),
      rejectionReasonKepalaDept:
          (json['rejection_reason_kepala_dept'] ??
                  json['rejectionReasonKepalaDept'])
              ?.toString(),
      rejectionReasonManagerHr:
          (json['rejection_reason_manager_hr'] ??
                  json['rejectionReasonManagerHr'] ??
                  json['rejection_reason'])
              ?.toString(),
      days: _toInt(daysRaw),
      total: (json['total'] ?? json['total_overtime'] ?? json['duration'])?.toString(),
      startTime: json['start_time']?.toString(),
      endTime: json['end_time']?.toString(),
      dokumenUrl: (json['document_url'] ?? json['dokumen_url'])?.toString(),
      createdAt: _parseDate(json['created_at']),
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

  static int _toInt(dynamic v) {
    if (v is int) return v;
    if (v is num) return v.toInt();
    return int.tryParse(v?.toString() ?? '') ?? 0;
  }

  static String _mapStatus(String raw) {
    switch (raw) {
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

  static String _reverseMapStatus(String id) {
    switch (id.toLowerCase()) {
      case 'disetujui':
        return 'APPROVED';
      case 'ditolak':
        return 'REJECTED';
      case 'dibatalkan':
        return 'CANCELLED';
      default:
        return 'PENDING';
    }
  }

  static String _guessKategori(String namaTipe) {
    final n = namaTipe.toLowerCase();
    if (n.contains('cuti')) return 'Cuti';
    if (n.contains('lembur')) return 'Lembur';
    return 'Izin';
  }

  bool get isOvertime => namaKategori.trim().toLowerCase() == 'lembur';

  String get durationLabel {
    if (isOvertime) {
      final t = (total ?? '').trim();
      if (t.isNotEmpty) return t;
      if ((startTime ?? '').trim().isNotEmpty && (endTime ?? '').trim().isNotEmpty) {
        return '${startTime!.trim()} - ${endTime!.trim()}';
      }
      return '— Jam';
    }
    return '$days Hari';
  }

  static String? _normalizeApproverDisplay(dynamic rawValue) {
    final value = (rawValue ?? '').toString().trim();
    if (value.isEmpty) return null;

    final objectIdPattern = RegExp(r'^[a-fA-F0-9]{24}$');
    if (objectIdPattern.hasMatch(value)) {
      return null;
    }

    return value;
  }

  /// Apakah pengajuan ini sudah disetujui
  bool get isApproved => statusFinal == 'APPROVED';

  String? get primaryRejectionReason {
    final hr = rejectionReasonManagerHr?.trim();
    if (hr != null && hr.isNotEmpty) return hr;
    final kepala = rejectionReasonKepalaDept?.trim();
    if (kepala != null && kepala.isNotEmpty) return kepala;
    return null;
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'type': type,
    'start_date': startDate.toIso8601String(),
    'end_date': endDate.toIso8601String(),
    'reason': reason,
    'status': status,
    'status_final': statusFinal,
    'status_kepala_departemen': statusKepala,
    'status_manager_hr': statusManagerHr,
    'kepala_departemen_id': kepalaDepartemenName,
    'manager_hr_id': managerHrName,
    'rejection_reason_kepala_dept': rejectionReasonKepalaDept,
    'rejection_reason_manager_hr': rejectionReasonManagerHr,
    'days': days,
    'total': total,
    'start_time': startTime,
    'end_time': endTime,
    'created_at': createdAt.toIso8601String(),
  };
}
