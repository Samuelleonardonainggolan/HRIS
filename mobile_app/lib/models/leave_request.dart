// lib/models/leave_request.dart
class LeaveRequest {
  final String id;
  final String type;         // nama_tipe dari backend, misal "Izin Sakit"
  final String namaKategori; // "Izin" | "Cuti" | "Lembur"
  final DateTime startDate;  // tanggal_mulai
  final DateTime endDate;    // tanggal_selesai
  final String reason;       // alasan
  final String status;       // sudah di-map ke Indonesia: Menunggu/Disetujui/Ditolak
  final String statusFinal;  // raw: PENDING/APPROVED/REJECTED
  final String statusKepala; // status_kepala_departemen
  final String statusManagerHr; // status_manager_hr
  final int days;            // total_hari
  final String? startTime;   // lembur: jam mulai
  final String? endTime;     // lembur: jam selesai
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
    // Deteksi apakah ini format backend baru (ada tanggal_mulai) atau format lama
    final isBackend = json['tanggal_mulai'] != null || json['nama_tipe'] != null;

    if (isBackend) {
      final rawStatus = (json['status_final'] ?? 'PENDING').toString().toUpperCase();
      return LeaveRequest(
        id:             json['id'] ?? '',
        type:           json['nama_tipe'] ?? json['type'] ?? '',
        namaKategori:   json['nama_kategori'] ?? _guessKategori(json['nama_tipe'] ?? ''),
        startDate:      _parseDate(json['tanggal_mulai']),
        endDate:        _parseDate(json['tanggal_selesai']),
        reason:         json['alasan'] ?? '',
        status:         _mapStatus(rawStatus),
        statusFinal:    rawStatus,
        statusKepala:   (json['status_kepala_departemen'] ?? 'PENDING').toString().toUpperCase(),
        statusManagerHr: (json['status_manager_hr'] ?? 'PENDING').toString().toUpperCase(),
        days:           json['total_hari'] ?? 0,
        startTime:      json['start_time'],
        endTime:        json['end_time'],
        dokumenUrl:     json['dokumen_url'],
      );
    }

    // Format lama (internal mapping dari api_service)
    final rawStatus = (json['status'] ?? 'Menunggu').toString();
    return LeaveRequest(
      id:             json['id'] ?? '',
      type:           json['type'] ?? '',
      namaKategori:   json['nama_kategori'] ?? _guessKategori(json['type'] ?? ''),
      startDate:      _parseDate(json['start_date']),
      endDate:        _parseDate(json['end_date']),
      reason:         json['reason'] ?? '',
      status:         rawStatus,
      statusFinal:    _reverseMapStatus(rawStatus),
      statusKepala:   _reverseMapStatus(rawStatus),
      statusManagerHr: _reverseMapStatus(rawStatus),
      days:           json['days'] ?? 0,
      startTime:      json['start_time'],
      endTime:        json['end_time'],
      dokumenUrl:     json['dokumen_url'],
    );
  }

  static DateTime _parseDate(dynamic v) {
    if (v == null) return DateTime.now();
    try { return DateTime.parse(v.toString()); } catch (_) { return DateTime.now(); }
  }

  static String _mapStatus(String raw) {
    switch (raw) {
      case 'APPROVED': return 'Disetujui';
      case 'REJECTED': return 'Ditolak';
      default:         return 'Menunggu';
    }
  }

  static String _reverseMapStatus(String id) {
    switch (id.toLowerCase()) {
      case 'disetujui': return 'APPROVED';
      case 'ditolak':   return 'REJECTED';
      default:           return 'PENDING';
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
    'id':         id,
    'type':       type,
    'start_date': startDate.toIso8601String(),
    'end_date':   endDate.toIso8601String(),
    'reason':     reason,
    'status':     status,
    'days':       days,
    'start_time': startTime,
    'end_time':   endTime,
  };
}