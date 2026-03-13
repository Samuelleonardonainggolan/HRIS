// lib/models/attendance_model.dart

/// Model untuk record absensi per hari
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

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'date': date.toIso8601String(),
      'clock_in_time': clockIn,
      'clock_out_time': clockOut,
      'status': status,
      'work_hours': workHours,
      'overtime_hours': overtimeHours,
      'face_similarity': faceSimilarity,
    };
  }
}

/// Model untuk ringkasan absensi bulanan
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

  Map<String, dynamic> toJson() {
    return {
      'month': month,
      'year': year,
      'total_days': totalDays,
      'total_hours': totalHours,
      'overtime_hours': overtimeHours,
      'records': records.map((r) => r.toJson()).toList(),
    };
  }
}

/// Model untuk hasil proses absensi (response dari endpoint /attendance/process)
class AttendanceProcessResult {
  final bool success;
  final String message;
  final double faceSimilarity;
  final bool locationValid;
  final double distance;
  final Map<String, dynamic>? data;
  final FaceVerificationResult? face;
  final GeoVerificationResult? geo;

  AttendanceProcessResult({
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
    // Handle response wrapper dari Go backend
    // Format: { "status": "success", "data": { ... } } atau langsung { ... }
    
    bool isSuccess = json['status'] == 'success';
    Map<String, dynamic> responseData = json['data'] ?? json;
    
    // Ekstrak informasi face verification
    FaceVerificationResult? faceResult;
    double similarity = 0.0;
    
    if (responseData['face'] != null) {
      faceResult = FaceVerificationResult.fromJson(responseData['face']);
      similarity = faceResult.similarity;
    } else if (responseData['face_similarity'] != null) {
      similarity = responseData['face_similarity'].toDouble();
    }
    
    // Ekstrak informasi geo verification
    GeoVerificationResult? geoResult;
    bool locValid = false;
    double dist = 0.0;
    
    if (responseData['geo'] != null) {
      geoResult = GeoVerificationResult.fromJson(responseData['geo']);
      locValid = geoResult.isValid;
      dist = geoResult.distanceM;
    } else if (responseData['location_valid'] != null) {
      locValid = responseData['location_valid'] == true;
      dist = responseData['distance_m']?.toDouble() ?? 0.0;
    }

    // Tentukan apakah absensi disetujui
    bool approved = responseData['approved'] == true || 
                    responseData['decision'] == 'approved' ||
                    (isSuccess && responseData['success'] == true);

    return AttendanceProcessResult(
      success: approved,
      message: responseData['message'] ?? json['message'] ?? '',
      faceSimilarity: similarity,
      locationValid: locValid,
      distance: dist,
      data: responseData,
      face: faceResult,
      geo: geoResult,
    );
  }

  Map<String, dynamic> toJson() {
    return {
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
}

/// Model untuk hasil verifikasi wajah dari FastAPI
class FaceVerificationResult {
  final bool matched;
  final double similarity;
  final double confidence;
  final double threshold;
  final String message;

  FaceVerificationResult({
    required this.matched,
    required this.similarity,
    required this.confidence,
    required this.threshold,
    required this.message,
  });

  factory FaceVerificationResult.fromJson(Map<String, dynamic> json) {
    return FaceVerificationResult(
      matched: json['matched'] == true,
      similarity: (json['similarity'] ?? 0).toDouble(),
      confidence: (json['confidence'] ?? 0).toDouble(),
      threshold: (json['threshold'] ?? 0.6).toDouble(),
      message: json['message'] ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'matched': matched,
      'similarity': similarity,
      'confidence': confidence,
      'threshold': threshold,
      'message': message,
    };
  }
}

/// Model untuk hasil verifikasi lokasi
class GeoVerificationResult {
  final bool isValid;
  final double distanceM;
  final double radiusM;
  final double officeLat;
  final double officeLng;
  final String message;

  GeoVerificationResult({
    required this.isValid,
    required this.distanceM,
    required this.radiusM,
    required this.officeLat,
    required this.officeLng,
    required this.message,
  });

  factory GeoVerificationResult.fromJson(Map<String, dynamic> json) {
    return GeoVerificationResult(
      isValid: json['is_valid'] == true,
      distanceM: (json['distance_m'] ?? 0).toDouble(),
      radiusM: (json['radius_m'] ?? 100).toDouble(),
      officeLat: (json['office_lat'] ?? 0).toDouble(),
      officeLng: (json['office_lng'] ?? 0).toDouble(),
      message: json['message'] ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'is_valid': isValid,
      'distance_m': distanceM,
      'radius_m': radiusM,
      'office_lat': officeLat,
      'office_lng': officeLng,
      'message': message,
    };
  }
}

/// Model untuk request absensi
class AttendanceRequest {
  final String employeeId;
  final double latitude;
  final double longitude;
  final String recordType; // 'clock_in' or 'clock_out'
  final double? threshold;
  final double? radiusM;

  AttendanceRequest({
    required this.employeeId,
    required this.latitude,
    required this.longitude,
    required this.recordType,
    this.threshold,
    this.radiusM,
  });

  Map<String, dynamic> toJson() {
    return {
      'employee_id': employeeId,
      'latitude': latitude,
      'longitude': longitude,
      'record_type': recordType,
      'threshold': threshold ?? 0.6,
      'radius_m': radiusM ?? 100,
    };
  }
}

/// Model untuk ringkasan absensi di dashboard
class TodayAttendanceSummary {
  final bool isClockedIn;
  final String clockInTime;
  final String? clockOutTime;
  final String status;
  final double workHours;
  final double? faceSimilarity;

  TodayAttendanceSummary({
    required this.isClockedIn,
    required this.clockInTime,
    this.clockOutTime,
    required this.status,
    required this.workHours,
    this.faceSimilarity,
  });

  factory TodayAttendanceSummary.fromJson(Map<String, dynamic> json) {
    String clockIn = json['clock_in'] ?? '--:--';
    String? clockOut = json['clock_out'];
    
    return TodayAttendanceSummary(
      isClockedIn: clockOut == null || clockOut == '--:--',
      clockInTime: clockIn,
      clockOutTime: clockOut,
      status: json['status'] ?? 'Unknown',
      workHours: (json['work_hours'] ?? 0).toDouble(),
      faceSimilarity: json['similarity']?.toDouble(),
    );
  }
}

/// Enum untuk status absensi
enum AttendanceStatus {
  onTime('On Time'),
  late('Late'),
  absent('Absent'),
  overtime('Overtime'),
  unknown('Unknown');

  final String value;
  const AttendanceStatus(this.value);

  static AttendanceStatus fromString(String value) {
    switch (value.toLowerCase()) {
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

/// Enum untuk tipe record absensi
enum RecordType {
  clockIn('clock_in'),
  clockOut('clock_out');

  final String value;
  const RecordType(this.value);

  static RecordType fromString(String value) {
    return value == 'clock_in' ? RecordType.clockIn : RecordType.clockOut;
  }
}