// lib/models/leave_request.dart
class LeaveRequest {
  final String id;
  final String type; // nama_tipe dari backend, misal "Izin Sakit"
  final String namaKategori; // "Izin" | "Cuti" | "Lembur"
  final DateTime startDate; // tanggal_mulai
  final DateTime endDate; // tanggal_selesai
  final String reason; // alasan
  final String status; // sudah di-map ke Indonesia: Menunggu/Disetujui/Ditolak
  final String statusFinal; // raw: PENDING/APPROVED/REJECTED
  final String statusKepala; // status_kepala_departemen
  final String statusManagerHr; // status_manager_hr
  final int days; // total_hari
  final String? startTime; // lembur: jam mulai
  final String? endTime; // lembur: jam selesai
  final String? dokumenUrl;

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
    required this.days,
    this.startTime,
    this.endTime,
    this.dokumenUrl,
  });

  /// Parse langsung dari response backend PengajuanIzinCuti
  factory LeaveRequest.fromJson(Map<String, dynamic> json) {
    final rawStatus =
        (json['final_status'] ??
                json['status_final'] ??
                _reverseMapStatus((json['status'] ?? 'Menunggu').toString()))
            .toString()
            .toUpperCase();

    final typeName =
        (json['type_name'] ?? json['nama_tipe'] ?? json['type'] ?? '')
            .toString();

    final categoryName = (json['category_name'] ?? json['nama_kategori'] ?? '')
        .toString();

    final startRaw =
        json['start_date'] ?? json['tanggal_mulai'] ?? json['start_date'];
    final endRaw =
        json['end_date'] ?? json['tanggal_selesai'] ?? json['end_date'];

    final daysRaw = json['days_total'] ?? json['total_hari'] ?? json['days'];

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
      days: _toInt(daysRaw),
      startTime: json['start_time']?.toString(),
      endTime: json['end_time']?.toString(),
      dokumenUrl: (json['document_url'] ?? json['dokumen_url'])?.toString(),
    );
  }

  static DateTime _parseDate(dynamic v) {
    if (v == null) return DateTime.now();
    try {
      return DateTime.parse(v.toString());
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

  /// Apakah pengajuan ini sudah disetujui
  bool get isApproved => statusFinal == 'APPROVED';

  Map<String, dynamic> toJson() => {
    'id': id,
    'type': type,
    'start_date': startDate.toIso8601String(),
    'end_date': endDate.toIso8601String(),
    'reason': reason,
    'status': status,
    'days': days,
    'start_time': startTime,
    'end_time': endTime,
  };
}
