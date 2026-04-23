// lib/models/attendance_model.dart

/// Record absensi per hari.
/// Bisa berasal dari dua sumber:
///  • [isLeaveRecord]=false → clock in/out real dari /attendance/monthly
///  • [isLeaveRecord]=true  → sintetis dari pengajuan izin/cuti yang APPROVED
class AttendanceRecord {
  final String id;
  final DateTime date;
  final String clockIn;
  final String clockOut;
  final String status;
  final double workHours;
  final double overtimeHours;
  final double? faceSimilarity;
  final String shiftName;
  final String location;
  final String? breakStart;
  final String? breakEnd;

  // ── Field khusus record izin/cuti dari pengajuan APPROVED ─────────────────
  final bool isLeaveRecord;
  final String? leaveType; // "Izin Sakit", "Cuti Tahunan", dll.
  final String? leaveKategori; // "Izin" | "Cuti" | "Lembur"
  final String? leaveReason;

  const AttendanceRecord({
    required this.id,
    required this.date,
    required this.clockIn,
    required this.clockOut,
    required this.status,
    required this.workHours,
    required this.overtimeHours,
    this.faceSimilarity,
    this.shiftName = '',
    this.location = '',
    this.breakStart,
    this.breakEnd,
    this.isLeaveRecord = false,
    this.leaveType,
    this.leaveKategori,
    this.leaveReason,
  });

  /// Parse dari JSON backend (/attendance/monthly)
  factory AttendanceRecord.fromJson(Map<String, dynamic> json) {
    return AttendanceRecord(
      id: json['id']?.toString() ?? '',
      date: _parseDate(json['date']),
      clockIn:
          json['clock_in_time']?.toString() ??
          json['clock_in']?.toString() ??
          '--:--',
      clockOut:
          json['clock_out_time']?.toString() ??
          json['clock_out']?.toString() ??
          '--:--',
      status: normalizeAttendanceStatusLabel(json['status']?.toString()),
      workHours: (json['work_hours'] as num?)?.toDouble() ?? 0,
      overtimeHours: (json['overtime_hours'] as num?)?.toDouble() ?? 0,
      faceSimilarity: (json['face_similarity'] as num?)?.toDouble(),
      shiftName:
          json['shift_name']?.toString() ?? json['shift']?.toString() ?? '',
      location:
          json['location']?.toString() ??
          (json['clock_in_location'] is Map
              ? (json['clock_in_location'] as Map)['name']?.toString() ?? ''
              : ''),
      breakStart:
          json['break_start']?.toString() ??
          json['break_start_time']?.toString(),
      breakEnd:
          json['break_end']?.toString() ?? json['break_end_time']?.toString(),
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

  /// Buat record sintetis untuk SATU HARI dari pengajuan APPROVED.
  factory AttendanceRecord.fromLeave({
    required String pengajuanId,
    required DateTime date,
    required String leaveType,
    required String leaveKategori,
    required String leaveReason,
  }) {
    final d = DateTime(date.year, date.month, date.day);
    return AttendanceRecord(
      id:
          '${pengajuanId}_'
          '${d.year}${d.month.toString().padLeft(2, '0')}${d.day.toString().padLeft(2, '0')}',
      date: d,
      clockIn: '--:--',
      clockOut: '--:--',
      status: _kategoriToStatus(leaveKategori),
      workHours: 0,
      overtimeHours: 0,
      isLeaveRecord: true,
      leaveType: leaveType,
      leaveKategori: leaveKategori,
      leaveReason: leaveReason,
    );
  }

  static String _kategoriToStatus(String k) {
    switch (k.trim().toLowerCase()) {
      case 'cuti':
        return 'Cuti';
      case 'lembur':
        return 'Lembur';
      default:
        return 'Izin';
    }
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'date': date.toIso8601String(),
    'clock_in_time': clockIn,
    'clock_out_time': clockOut,
    'status': status,
    'work_hours': workHours,
    'overtime_hours': overtimeHours,
    'face_similarity': faceSimilarity,
    'shift_name': shiftName,
    'location': location,
    'break_start': breakStart,
    'break_end': breakEnd,
  };
}

class MonthlyAttendanceSummary {
  final String month;
  final int year;
  final int totalDays;
  final double totalHours;
  final double overtimeHours;
  final List<AttendanceRecord> records;

  const MonthlyAttendanceSummary({
    required this.month,
    required this.year,
    required this.totalDays,
    required this.totalHours,
    required this.overtimeHours,
    required this.records,
  });

  factory MonthlyAttendanceSummary.fromJson(Map<String, dynamic> json) {
    final raw = json['records'];
    final list = raw is List ? raw : <dynamic>[];
    return MonthlyAttendanceSummary(
      month: json['month']?.toString() ?? '',
      year: (json['year'] as int?) ?? DateTime.now().year,
      totalDays: (json['total_days'] as int?) ?? 0,
      totalHours: (json['total_hours'] as num?)?.toDouble() ?? 0,
      overtimeHours: (json['overtime_hours'] as num?)?.toDouble() ?? 0,
      records: list
          .map((e) => AttendanceRecord.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() => {
    'month': month,
    'year': year,
    'total_days': totalDays,
    'total_hours': totalHours,
    'overtime_hours': overtimeHours,
    'records': records.map((r) => r.toJson()).toList(),
  };
}

class AttendanceProcessResult {
  final bool success;
  final String message;
  final double faceSimilarity;
  final bool locationValid;
  final double distance;
  final Map<String, dynamic>? data;
  final FaceVerificationResult? face;
  final GeoVerificationResult? geo;

  const AttendanceProcessResult({
    required this.success,
    required this.message,
    required this.faceSimilarity,
    required this.locationValid,
    required this.distance,
    this.data,
    this.face,
    this.geo,
  });

  factory AttendanceProcessResult.fromJson(Map<String, dynamic> json) {
    final isSuccess = json['status'] == 'success';
    final rd = (json['data'] is Map)
        ? json['data'] as Map<String, dynamic>
        : json;

    FaceVerificationResult? faceR;
    double sim = 0;
    if (rd['face'] is Map) {
      faceR = FaceVerificationResult.fromJson(
        rd['face'] as Map<String, dynamic>,
      );
      sim = faceR.similarity;
    } else {
      sim = (rd['face_similarity'] as num?)?.toDouble() ?? 0;
    }

    GeoVerificationResult? geoR;
    bool locV = false;
    double dist = 0;
    if (rd['geo'] is Map) {
      geoR = GeoVerificationResult.fromJson(rd['geo'] as Map<String, dynamic>);
      locV = geoR.isValid;
      dist = geoR.distanceM;
    } else {
      locV = rd['location_valid'] == true;
      dist = (rd['distance_m'] as num?)?.toDouble() ?? 0;
    }

    final approved =
        rd['approved'] == true ||
        rd['decision'] == 'approved' ||
        (isSuccess && rd['success'] == true);

    return AttendanceProcessResult(
      success: approved,
      message: rd['message']?.toString() ?? json['message']?.toString() ?? '',
      faceSimilarity: sim,
      locationValid: locV,
      distance: dist,
      data: rd is Map<String, dynamic> ? rd : null,
      face: faceR,
      geo: geoR,
    );
  }

  Map<String, dynamic> toJson() => {
    'success': success,
    'message': message,
    'face_similarity': faceSimilarity,
    'location_valid': locationValid,
    'distance_m': distance,
    'data': data,
    'face': face?.toJson(),
    'geo': geo?.toJson(),
  };
}

class FaceVerificationResult {
  final bool matched;
  final double similarity, confidence, threshold;
  final String message;
  const FaceVerificationResult({
    required this.matched,
    required this.similarity,
    required this.confidence,
    required this.threshold,
    required this.message,
  });
  factory FaceVerificationResult.fromJson(Map<String, dynamic> j) =>
      FaceVerificationResult(
        matched: j['matched'] == true,
        similarity: (j['similarity'] as num?)?.toDouble() ?? 0,
        confidence: (j['confidence'] as num?)?.toDouble() ?? 0,
        threshold: (j['threshold'] as num?)?.toDouble() ?? 0.6,
        message: j['message']?.toString() ?? '',
      );
  Map<String, dynamic> toJson() => {
    'matched': matched,
    'similarity': similarity,
    'confidence': confidence,
    'threshold': threshold,
    'message': message,
  };
}

class GeoVerificationResult {
  final bool isValid;
  final double distanceM, radiusM, officeLat, officeLng;
  final String message;
  final String? geofenceName;
  const GeoVerificationResult({
    required this.isValid,
    required this.distanceM,
    required this.radiusM,
    required this.officeLat,
    required this.officeLng,
    required this.message,
    this.geofenceName,
  });
  factory GeoVerificationResult.fromJson(Map<String, dynamic> j) =>
      GeoVerificationResult(
        isValid: j['is_valid'] == true || j['is_within_geofence'] == true,
        distanceM:
            (j['distance_m'] as num?)?.toDouble() ??
            (j['distance'] as num?)?.toDouble() ??
            0,
        radiusM:
            (j['radius_m'] as num?)?.toDouble() ??
            ((j['geofence'] is Map)
                ? (((j['geofence'] as Map<String, dynamic>)['radius'] as num?)
                          ?.toDouble() ??
                      0)
                : 0),
        officeLat:
            (j['office_lat'] as num?)?.toDouble() ??
            ((j['geofence'] is Map)
                ? (((j['geofence'] as Map<String, dynamic>)['latitude'] as num?)
                          ?.toDouble() ??
                      0)
                : 0),
        officeLng:
            (j['office_lng'] as num?)?.toDouble() ??
            ((j['geofence'] is Map)
                ? (((j['geofence'] as Map<String, dynamic>)['longitude']
                              as num?)
                          ?.toDouble() ??
                      0)
                : 0),
        message: j['message']?.toString() ?? '',
        geofenceName: (j['geofence'] is Map)
            ? (j['geofence'] as Map<String, dynamic>)['name']?.toString()
            : null,
      );
  Map<String, dynamic> toJson() => {
    'is_valid': isValid,
    'distance_m': distanceM,
    'radius_m': radiusM,
    'office_lat': officeLat,
    'office_lng': officeLng,
    'message': message,
    'geofence_name': geofenceName,
  };
}

class AttendanceRequest {
  final String employeeId, recordType;
  final double latitude, longitude;
  final double? threshold, radiusM;
  const AttendanceRequest({
    required this.employeeId,
    required this.latitude,
    required this.longitude,
    required this.recordType,
    this.threshold,
    this.radiusM,
  });
  Map<String, dynamic> toJson() => {
    'employee_id': employeeId,
    'latitude': latitude,
    'longitude': longitude,
    'record_type': recordType,
    'threshold': threshold ?? 0.75,
    'radius_m': radiusM ?? 100,
  };
}

class TodayAttendanceSummary {
  final bool isClockedIn;
  final String clockInTime;
  final String? clockOutTime;
  final String status;
  final double workHours;
  final double? faceSimilarity;
  const TodayAttendanceSummary({
    required this.isClockedIn,
    required this.clockInTime,
    this.clockOutTime,
    required this.status,
    required this.workHours,
    this.faceSimilarity,
  });
  factory TodayAttendanceSummary.fromJson(Map<String, dynamic> j) {
    final ci = j['clock_in']?.toString() ?? '--:--';
    final co = j['clock_out']?.toString();
    return TodayAttendanceSummary(
      isClockedIn: co == null || co == '--:--',
      clockInTime: ci,
      clockOutTime: co,
      status: normalizeAttendanceStatusLabel(j['status']?.toString()),
      workHours: (j['work_hours'] as num?)?.toDouble() ?? 0,
      faceSimilarity: (j['similarity'] as num?)?.toDouble(),
    );
  }
}

enum AttendanceStatus {
  onTime('Tepat Waktu'),
  late('Terlambat'),
  absent('Absent'),
  overtime('Overtime'),
  unknown('Unknown');

  final String value;
  const AttendanceStatus(this.value);
  static AttendanceStatus fromString(String v) {
    switch (v.toLowerCase()) {
      case 'on time':
        return AttendanceStatus.onTime;
      case 'late':
        return AttendanceStatus.late;
      case 'absent':
        return AttendanceStatus.absent;
      case 'overtime':
        return AttendanceStatus.overtime;
      default:
        return AttendanceStatus.unknown;
    }
  }
}

enum RecordType {
  clockIn('clock_in'),
  clockOut('clock_out');

  final String value;
  const RecordType(this.value);
  static RecordType fromString(String v) =>
      v == 'clock_in' ? RecordType.clockIn : RecordType.clockOut;
}

class TodayAttendanceDetail {
  final String id;
  final DateTime date;
  final String clockInTime;
  final String? clockOutTime;
  final String status;
  final double workHours;
  final double? overtimeHours, faceSimilarity;
  final String? breakStartTime;
  final String? breakEndTime;
  final bool hasClockedIn, hasClockedOut;

  const TodayAttendanceDetail({
    required this.id,
    required this.date,
    required this.clockInTime,
    this.clockOutTime,
    required this.status,
    required this.workHours,
    this.overtimeHours,
    this.faceSimilarity,
    this.breakStartTime,
    this.breakEndTime,
    required this.hasClockedIn,
    required this.hasClockedOut,
  });

  factory TodayAttendanceDetail.fromJson(Map<String, dynamic> j) {
    final ci = j['clock_in_time']?.toString() ?? '--:--';
    final co = j['clock_out_time']?.toString();
    return TodayAttendanceDetail(
      id: j['id']?.toString() ?? '',
      date: DateTime.tryParse(j['date']?.toString() ?? '') ?? DateTime.now(),
      clockInTime: ci,
      clockOutTime: co,
      status: normalizeAttendanceStatusLabel(j['status']?.toString()),
      workHours: (j['work_hours'] as num?)?.toDouble() ?? 0,
      overtimeHours: (j['overtime_hours'] as num?)?.toDouble(),
      faceSimilarity: (j['face_similarity'] as num?)?.toDouble(),
      breakStartTime: j['break_start_time']?.toString(),
      breakEndTime: j['break_end_time']?.toString(),
      hasClockedIn: ci.isNotEmpty && ci != '--:--',
      hasClockedOut: co != null && co.isNotEmpty && co != '--:--',
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'date': date.toIso8601String(),
    'clock_in_time': clockInTime,
    'clock_out_time': clockOutTime,
    'status': status,
    'work_hours': workHours,
    'overtime_hours': overtimeHours,
    'face_similarity': faceSimilarity,
    'break_start_time': breakStartTime,
    'break_end_time': breakEndTime,
  };
}

String normalizeAttendanceStatusLabel(String? raw) {
  final value = (raw ?? '').trim();
  final lower = value.toLowerCase();

  switch (lower) {
    case 'on time':
    case 'tepat waktu':
      return 'Tepat Waktu';
    case 'late':
    case 'terlambat':
      return 'Terlambat';
    default:
      return value.isEmpty ? 'Unknown' : value;
  }
}
