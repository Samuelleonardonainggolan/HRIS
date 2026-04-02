// lib/services/api_service.dart
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../models/user_model.dart';
import '../models/auth_model.dart';
import '../models/attendance_model.dart';
import '../models/leave_request.dart';
import 'package:path_provider/path_provider.dart';

class ApiService {
  // ✅ Ganti dengan IP yang sesuai environment Anda
  static const String baseUrl = 'http://10.218.68.218:8080/api/v1';

  static final Map<String, String> _headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };

  // ─── Token Management ───────────────────────────────────────────────────────

  static Future<void> saveTokens(
    String accessToken,
    String refreshToken,
    int expiresIn,
  ) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('access_token', accessToken);
    await prefs.setString('refresh_token', refreshToken);
    await prefs.setInt('expires_in', expiresIn);
    await prefs.setInt('login_time', DateTime.now().millisecondsSinceEpoch);
  }

  static Future<String?> getAccessToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString('access_token');
  }

  static Future<String?> getRefreshToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString('refresh_token');
  }

  static Future<bool> isTokenExpired() async {
    final prefs = await SharedPreferences.getInstance();
    final loginTime = prefs.getInt('login_time');
    final expiresIn = prefs.getInt('expires_in');
    if (loginTime == null || expiresIn == null) return true;
    final now = DateTime.now().millisecondsSinceEpoch;
    final expiryTime = loginTime + (expiresIn * 1000);
    return now >= expiryTime;
  }

  static Future<void> clearTokens() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('access_token');
    await prefs.remove('refresh_token');
    await prefs.remove('expires_in');
    await prefs.remove('login_time');
    await prefs.remove('user_id');
  }

  static Future<Map<String, String>> getHeaders() async {
    final token = await getAccessToken();
    if (token != null && token.isNotEmpty) {
      return {..._headers, 'Authorization': 'Bearer $token'};
    }
    return _headers;
  }

  static Future<void> saveUserId(String userId) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('user_id', userId);
    print('[API] User ID saved: $userId');
  }

  static Future<String?> getUserId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString('user_id');
  }

  // ─── Auth ───────────────────────────────────────────────────────────────────

  static Future<LoginResponse> login(String email, String password) async {
    try {
      print('[API] Login untuk: $email');
      final response = await http
          .post(
            Uri.parse('$baseUrl/auth/login'),
            headers: _headers,
            body: jsonEncode({'email': email, 'password': password}),
          )
          .timeout(const Duration(seconds: 30));

      print('[API] Login status: ${response.statusCode}');
      print('[API] Login body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final Map<String, dynamic> responseData = data['data'] ?? data;
        final loginResponse = LoginResponse.fromJson(responseData);

        await saveTokens(
          loginResponse.accessToken,
          loginResponse.refreshToken,
          loginResponse.expiresIn,
        );
        await saveUserId(loginResponse.user.id);

        print('[API] Login sukses untuk user: ${loginResponse.user.id}');
        return loginResponse;
      } else {
        final error = jsonDecode(response.body);
        throw Exception(error['message'] ?? 'Login gagal');
      }
    } catch (e) {
      print('[API] Login error: $e');
      throw Exception('Login error: $e');
    }
  }

  static Future<LoginResponse> refreshToken() async {
    try {
      final token = await getRefreshToken();
      if (token == null) throw Exception('No refresh token');

      final response = await http.post(
        Uri.parse('$baseUrl/auth/refresh'),
        headers: _headers,
        body: jsonEncode({'refresh_token': token}),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final Map<String, dynamic> responseData = data['data'] ?? data;
        final loginResponse = LoginResponse.fromJson(responseData);
        await saveTokens(
          loginResponse.accessToken,
          loginResponse.refreshToken,
          loginResponse.expiresIn,
        );
        return loginResponse;
      } else {
        throw Exception('Gagal refresh token');
      }
    } catch (e) {
      throw Exception('Refresh token error: $e');
    }
  }

  static Future<void> logout() async {
    try {
      final headers = await getHeaders();
      await http.post(Uri.parse('$baseUrl/auth/logout'), headers: headers);
    } finally {
      await clearTokens();
    }
  }

  // ─── Attendance ─────────────────────────────────────────────────────────────

  static Future<AttendanceProcessResult> processAttendance({
    required String recordType,
    required double latitude,
    required double longitude,
    required String photoPath,
  }) async {
    try {
      print('[API] Processing attendance: $recordType');

      if (await isTokenExpired()) {
        print('[API] Token expired, refreshing...');
        await refreshToken();
      }

      final token = await getAccessToken();
      if (token == null || token.isEmpty) {
        throw Exception('Token tidak ditemukan. Silakan login ulang.');
      }

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/attendance/process'),
      );

      request.headers.addAll({
        'Authorization': 'Bearer $token',
        'Accept': 'application/json',
      });

      request.fields['record_type'] = recordType;
      request.fields['latitude'] = latitude.toString();
      request.fields['longitude'] = longitude.toString();

      request.files.add(
        await http.MultipartFile.fromPath(
          'photo',
          photoPath,
          filename:
              '${recordType}_${DateTime.now().millisecondsSinceEpoch}.jpg',
        ),
      );

      print(
        '[API] Sending: record_type=$recordType, lat=$latitude, lng=$longitude',
      );

      final streamedResponse = await request.send().timeout(
        const Duration(seconds: 60),
      );
      final response = await http.Response.fromStream(streamedResponse);

      print('[API] Process attendance status: ${response.statusCode}');
      print('[API] Process attendance body: ${response.body}');

      final jsonResponse = jsonDecode(response.body);

      if (response.statusCode == 200) {
        return AttendanceProcessResult.fromJson(jsonResponse);
      } else if (response.statusCode == 401) {
        await clearTokens();
        throw Exception('Sesi telah berakhir. Silakan login ulang.');
      } else if (response.statusCode == 400) {
        final message = jsonResponse['message'] ?? 'Absensi gagal';
        return AttendanceProcessResult(
          success: false,
          message: message,
          faceSimilarity:
              jsonResponse['data']?['face_similarity']?.toDouble() ?? 0.0,
          locationValid: jsonResponse['data']?['location_valid'] ?? false,
          distance: jsonResponse['data']?['distance_m']?.toDouble() ?? 0.0,
        );
      } else {
        throw Exception(jsonResponse['message'] ?? 'Gagal memproses absensi');
      }
    } catch (e) {
      print('[API] Process attendance error: $e');
      rethrow;
    }
  }

  static Future<TodayAttendanceDetail?> getTodayAttendance() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/attendance/today'), headers: headers)
          .timeout(const Duration(seconds: 30));

      print('[API] Today attendance status: ${response.statusCode}');
      print('[API] Today attendance body: ${response.body}');

      if (response.statusCode == 200) {
        final jsonResponse = jsonDecode(response.body);
        if (jsonResponse['data'] == null) {
          return null;
        }
        return TodayAttendanceDetail.fromJson(jsonResponse['data']);
      }
      return null;
    } catch (e) {
      print('[API] Get today attendance error: $e');
      return null;
    }
  }

  static Future<MonthlyAttendanceSummary> getMonthlyAttendance({
    int? month,
    int? year,
  }) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final now = DateTime.now();
      final queryMonth = month ?? now.month;
      final queryYear = year ?? now.year;

      final headers = await getHeaders();
      final response = await http
          .get(
            Uri.parse(
              '$baseUrl/attendance/monthly?month=$queryMonth&year=$queryYear',
            ),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      print('[API] Monthly attendance status: ${response.statusCode}');
      print('[API] Monthly attendance body: ${response.body}');

      if (response.statusCode == 200) {
        final jsonResponse = jsonDecode(response.body);
        return MonthlyAttendanceSummary.fromJson(jsonResponse['data'] ?? {});
      } else {
        throw Exception('Gagal mengambil data absensi bulanan');
      }
    } catch (e) {
      print('[API] Get monthly attendance error: $e');
      throw Exception('Get monthly attendance error: $e');
    }
  }

  // ─── Pengajuan Izin/Cuti ────────────────────────────────────────────────────

  /// GET /api/v1/pengajuan/tipe
  /// Mengambil daftar tipe pengajuan (Izin Sakit, Cuti Tahunan, dll.)
  static Future<List<TipePengajuan>> getTipePengajuan() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/pengajuan/tipe'), headers: headers)
          .timeout(const Duration(seconds: 30));

      print('[API] GetTipePengajuan status: ${response.statusCode}');
      print('[API] GetTipePengajuan body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final List<dynamic> list = data['data'] ?? data['tipe'] ?? [];
        return list.map((e) => TipePengajuan.fromJson(e)).toList();
      } else {
        throw Exception(
          'Gagal mengambil tipe pengajuan (${response.statusCode})',
        );
      }
    } catch (e) {
      print('[API] GetTipePengajuan error: $e');
      throw Exception('GetTipePengajuan error: $e');
    }
  }

  /// POST /api/v1/pengajuan
  /// Mengirimkan pengajuan izin/cuti baru
  static Future<void> submitPengajuan({
    required String tipePengajuanId,
    required String tanggalMulai, // format "yyyy-MM-dd"
    required String tanggalSelesai, // format "yyyy-MM-dd"
    required int totalHari,
    required String alasan,
    String? dokumenUrl,
    // Lembur-specific (opsional)
    String? startTime,
    String? endTime,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final userId = await getUserId();
      if (userId == null) throw Exception('User ID tidak ditemukan');

      final headers = await getHeaders();

      final body = <String, dynamic>{
        'user_id': userId,
        'tipe_pengajuan_id': tipePengajuanId,
        'tanggal_mulai': tanggalMulai,
        'tanggal_selesai': tanggalSelesai,
        'total_hari': totalHari,
        'alasan': alasan,
        if (dokumenUrl != null && dokumenUrl.isNotEmpty)
          'dokumen_url': dokumenUrl,
        if (startTime != null) 'start_time': startTime,
        if (endTime != null) 'end_time': endTime,
      };

      final response = await http
          .post(
            Uri.parse('$baseUrl/pengajuan'),
            headers: headers,
            body: jsonEncode(body),
          )
          .timeout(const Duration(seconds: 30));

      print('[API] submitPengajuan status: ${response.statusCode}');
      print('[API] submitPengajuan body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        return;
      } else {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal mengirimkan pengajuan');
      }
    } catch (e) {
      print('[API] submitPengajuan error: $e');
      rethrow;
    }
  }

  /// GET /api/v1/pengajuan — ambil riwayat pengajuan user
  static Future<List<LeaveRequest>> getMyPengajuan() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/pengajuan'), headers: headers)
          .timeout(const Duration(seconds: 30));

      print('[API] getMyPengajuan status: ${response.statusCode}');
      print('[API] getMyPengajuan body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        // Backend bisa return data sebagai List, atau Map dengan key data/pengajuan.
        List<dynamic> list;
        if (data is List) {
          list = data;
        } else if (data is Map<String, dynamic>) {
          final rawData = data['data'];
          if (rawData is List) {
            list = rawData;
          } else if (rawData is Map<String, dynamic>) {
            list = rawData['data'] ?? rawData['pengajuan'] ?? [];
          } else {
            list = data['pengajuan'] ?? [];
          }
        } else {
          list = [];
        }
        return list
            .map((e) => LeaveRequest.fromJson(e as Map<String, dynamic>))
            .toList()
          ..sort((a, b) => b.startDate.compareTo(a.startDate));
      } else {
        throw Exception(
          'Gagal mengambil data pengajuan (${response.statusCode})',
        );
      }
    } catch (e) {
      print('[API] getMyPengajuan error: $e');
      throw Exception('getMyPengajuan error: $e');
    }
  }

  /// Ambil semua pengajuan user yang APPROVED dan memiliki overlap dengan bulan [month]/[year]
  /// Filter dilakukan di client (bukan query param backend) karena
  /// endpoint GET /pengajuan tidak mendukung filter status/bulan.
  /// Jika gagal, kembalikan [] agar history page tetap tampil.
  static Future<List<LeaveRequest>> getApprovedPengajuanByMonth({
    required int month,
    required int year,
  }) async {
    try {
      // Reuse getMyPengajuan — sudah handle berbagai format response backend
      final all = await getMyPengajuan();

      print(
        '[API] getApprovedPengajuanByMonth: total pengajuan = ${all.length}',
      );
      for (final p in all) {
        print(
          '[API]   id=${p.id} type=${p.type} status=${p.statusFinal}'
          ' start=${p.startDate} end=${p.endDate}',
        );
      }

      final firstDay = DateTime(year, month, 1);
      // Hari terakhir bulan: hari pertama bulan berikutnya dikurangi 1 hari
      final lastDay = DateTime(
        year,
        month + 1,
        1,
      ).subtract(const Duration(days: 1));

      final approved = all.where((p) {
        if (p.statusFinal != 'APPROVED') return false;

        // Normalisasi tanggal ke midnight untuk perbandingan yang akurat
        final s = DateTime(
          p.startDate.year,
          p.startDate.month,
          p.startDate.day,
        );
        final e = DateTime(p.endDate.year, p.endDate.month, p.endDate.day);

        // Overlap: range pengajuan harus memiliki setidaknya satu hari yang berada di bulan yang ditampilkan
        // Kondisi overlap: mulai <= lastDay DAN selesai >= firstDay
        // Ini akan menangkap kasus:
        // - pengajuan yang dimulai sebelum bulan ini dan berakhir di bulan ini
        // - pengajuan yang dimulai di bulan ini dan berakhir setelah bulan ini
        // - pengajuan yang sepenuhnya di dalam bulan ini
        final hasOverlap = !s.isAfter(lastDay) && !e.isBefore(firstDay);

        print(
          '[API]   checking: ${p.id} s=${s} e=${e} firstDay=$firstDay lastDay=$lastDay overlap=$hasOverlap',
        );

        return hasOverlap;
      }).toList();

      print(
        '[API] getApprovedPengajuanByMonth: approved & overlap = ${approved.length}',
      );
      return approved;
    } catch (e) {
      print('[API] getApprovedPengajuanByMonth error: $e');
      return [];
    }
  }

  /// Konversi respons backend pengajuan_izin_cuti → format LeaveRequest Flutter
  static Map<String, dynamic> _mapPengajuanToLeave(Map<String, dynamic> json) {
    // Mapping status backend (PENDING/APPROVED/REJECTED) ke bahasa Indonesia
    final rawStatus =
        json['status_final'] ?? json['status_kepala_departemen'] ?? 'PENDING';
    String status;
    switch ((rawStatus as String).toUpperCase()) {
      case 'APPROVED':
        status = 'Disetujui';
        break;
      case 'REJECTED':
        status = 'Ditolak';
        break;
      default:
        status = 'Menunggu';
    }

    return {
      'id': json['id'] ?? '',
      'type': json['nama_tipe'] ?? json['tipe'] ?? 'Izin',
      'start_date': json['tanggal_mulai'] ?? DateTime.now().toIso8601String(),
      'end_date': json['tanggal_selesai'] ?? DateTime.now().toIso8601String(),
      'reason': json['alasan'] ?? '',
      'status': status,
      'days': json['total_hari'] ?? 0,
    };
  }

  // ─── Face Registration ──────────────────────────────────────────────────────

  static Future<void> registerFace({
    required String userId,
    required String photoPath,
  }) async {
    try {
      print('[API] Registering face for user: $userId');

      if (await isTokenExpired()) {
        await refreshToken();
      }

      final token = await getAccessToken();

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/face/register'),
      );

      request.headers.addAll({
        'Authorization': 'Bearer $token',
        'Accept': 'application/json',
      });

      request.fields['user_id'] = userId;

      request.files.add(
        await http.MultipartFile.fromPath(
          'photo',
          photoPath,
          filename: 'face_${DateTime.now().millisecondsSinceEpoch}.jpg',
        ),
      );

      print('[API] Sending face registration...');
      final streamedResponse = await request.send().timeout(
        const Duration(seconds: 60),
      );
      final response = await http.Response.fromStream(streamedResponse);

      print('[API] Register face status: ${response.statusCode}');
      print('[API] Register face body: ${response.body}');

      if (response.statusCode == 200) {
        print('[API] Face registered successfully');
        return;
      } else {
        final error = jsonDecode(response.body);
        throw Exception(error['message'] ?? 'Gagal mendaftarkan wajah');
      }
    } catch (e) {
      print('[API] Register face error: $e');
      rethrow;
    }
  }

  static Future<void> registerFaceWithBytes({
    required String userId,
    required List<double> faceEmbedding,
    required String faceImageBase64,
  }) async {
    try {
      print('[API] Registering face (with bytes) for user: $userId');

      if (await isTokenExpired()) {
        await refreshToken();
      }

      final token = await getAccessToken();
      final imageBytes = base64Decode(faceImageBase64);
      final tempDir = await getTemporaryDirectory();
      final tempFile = File(
        '${tempDir.path}/face_${DateTime.now().millisecondsSinceEpoch}.jpg',
      );
      await tempFile.writeAsBytes(imageBytes);

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/face/register'),
      );
      request.headers.addAll({
        'Authorization': 'Bearer $token',
        'Accept': 'application/json',
      });
      request.fields['user_id'] = userId;
      request.files.add(
        await http.MultipartFile.fromPath(
          'photo',
          tempFile.path,
          filename: 'face.jpg',
        ),
      );

      final streamedResponse = await request.send().timeout(
        const Duration(seconds: 60),
      );
      final response = await http.Response.fromStream(streamedResponse);

      await tempFile.delete();

      print('[API] Register face status: ${response.statusCode}');
      print('[API] Register face body: ${response.body}');

      if (response.statusCode == 200) {
        print('[API] Face registered successfully');
        return;
      } else {
        final error = jsonDecode(response.body);
        throw Exception(error['message'] ?? 'Gagal mendaftarkan wajah');
      }
    } catch (e) {
      print('[API] Register face (bytes) error: $e');
      rethrow;
    }
  }

  static Future<Map<String, dynamic>> checkFaceStatus() async {
    try {
      print('[API] Checking face status...');

      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/face/status'), headers: headers)
          .timeout(const Duration(seconds: 30));

      print('[API] Face status: ${response.statusCode}');
      print('[API] Face status body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        bool hasFaceRegistered = false;

        if (data['data'] != null) {
          hasFaceRegistered = data['data']['has_face_registered'] == true;
        } else if (data['has_face_registered'] != null) {
          hasFaceRegistered = data['has_face_registered'] == true;
        }

        return {'has_face_registered': hasFaceRegistered};
      } else {
        return {'has_face_registered': false};
      }
    } catch (e) {
      print('[API] Face status error: $e');
      return {'has_face_registered': false};
    }
  }

  static Future<List<double>> extractFaceEmbedding(String imagePath) async {
    try {
      print('[API] Extracting face embedding dari: $imagePath');

      if (await isTokenExpired()) {
        await refreshToken();
      }

      final token = await getAccessToken();
      if (token == null || token.isEmpty) {
        throw Exception('Token tidak ditemukan');
      }

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/face/extract-embedding'),
      );

      request.headers.addAll({'Authorization': 'Bearer $token'});

      final userId = await getUserId();
      if (userId != null) request.fields['employee_id'] = userId;

      request.files.add(
        await http.MultipartFile.fromPath(
          'photo',
          imagePath,
          filename: 'face_${DateTime.now().millisecondsSinceEpoch}.jpg',
        ),
      );

      final streamedResponse = await request.send().timeout(
        const Duration(seconds: 60),
      );
      final response = await http.Response.fromStream(streamedResponse);

      print('[API] Extract embedding status: ${response.statusCode}');
      print('[API] Extract embedding body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        List<double> embedding = [];

        if (data['data'] != null && data['data']['embedding'] != null) {
          embedding = List<double>.from(data['data']['embedding']);
        } else if (data['embedding'] != null) {
          embedding = List<double>.from(data['embedding']);
        } else {
          throw Exception('Embedding tidak ditemukan dalam response');
        }

        if (embedding.isEmpty) {
          throw Exception('Embedding kosong');
        }

        return embedding;
      } else {
        final error = jsonDecode(response.body);
        throw Exception(error['message'] ?? 'Gagal mengekstrak wajah');
      }
    } catch (e) {
      print('[API] Extract embedding error: $e');
      rethrow;
    }
  }

  // ─── Profile ─────────────────────────────────────────────────────────────────

  static Future<User> getProfile() async {
    try {
      if (await isTokenExpired()) await refreshToken();
      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/profile'), headers: headers)
          .timeout(const Duration(seconds: 30));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final userJson = data['data'] ?? data['user'] ?? data;
        return User.fromJson(userJson);
      } else {
        throw Exception('Gagal mengambil profil (${response.statusCode})');
      }
    } catch (e) {
      throw Exception('Profile error: $e');
    }
  }

  static Future<void> updateProfile(Map<String, dynamic> profileData) async {
    try {
      if (await isTokenExpired()) await refreshToken();
      final headers = await getHeaders();
      final response = await http
          .put(
            Uri.parse('$baseUrl/profile'),
            headers: headers,
            body: jsonEncode(profileData),
          )
          .timeout(const Duration(seconds: 30));
      print('[API] updateProfile status: ${response.statusCode}');
      if (response.statusCode == 200) return;
      if (response.statusCode == 401) {
        await clearTokens();
        throw Exception('Sesi berakhir, silakan login ulang.');
      }
      final err = jsonDecode(response.body);
      throw Exception(
        err['error'] ?? err['message'] ?? 'Gagal memperbarui profil',
      );
    } catch (e) {
      print('[API] updateProfile error: $e');
      rethrow;
    }
  }

  static Future<void> changePassword({
    required String oldPassword,
    required String newPassword,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();
      final headers = await getHeaders();
      final response = await http
          .post(
            Uri.parse('$baseUrl/profile/change-password'),
            headers: headers,
            body: jsonEncode({
              'old_password': oldPassword,
              'new_password': newPassword,
            }),
          )
          .timeout(const Duration(seconds: 30));
      print('[API] changePassword status: ${response.statusCode}');
      if (response.statusCode == 200) return;
      if (response.statusCode == 401) {
        await clearTokens();
        throw Exception('Sesi berakhir, silakan login ulang.');
      }
      final err = jsonDecode(response.body);
      throw Exception(
        err['error'] ?? err['message'] ?? 'Gagal mengubah password',
      );
    } catch (e) {
      print('[API] changePassword error: $e');
      rethrow;
    }
  }

  static Future<List<AttendanceRecord>> getAttendanceHistory({
    int? month,
    int? year,
    String? status,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final now = DateTime.now();
      final queryMonth = month ?? now.month;
      final queryYear = year ?? now.year;

      final headers = await getHeaders();
      final response = await http.get(
        Uri.parse(
          '$baseUrl/attendance/monthly?month=$queryMonth&year=$queryYear',
        ),
        headers: headers,
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final Map<String, dynamic> responseData = data['data'] ?? {};
        final List<dynamic> records = responseData['records'] ?? [];
        return records.map((item) => AttendanceRecord.fromJson(item)).toList();
      } else {
        throw Exception('Gagal mengambil history');
      }
    } catch (e) {
      throw Exception('History error: $e');
    }
  }
}

// ─── Model TipePengajuan (untuk dropdown di NewRequestPage) ──────────────────

class TipePengajuan {
  final String id;
  final String namaTipe;
  final String namaKategori; // "Izin" atau "Cuti"
  final bool potongKuota;
  final bool wajibLampiran;

  const TipePengajuan({
    required this.id,
    required this.namaTipe,
    required this.namaKategori,
    required this.potongKuota,
    required this.wajibLampiran,
  });

  factory TipePengajuan.fromJson(Map<String, dynamic> json) {
    return TipePengajuan(
      id: json['id'] ?? '',
      namaTipe: json['nama_tipe'] ?? '',
      namaKategori: json['nama_kategori'] ?? '',
      potongKuota: json['potong_kuota'] == true,
      wajibLampiran: json['wajib_lampiran'] == true,
    );
  }

  @override
  String toString() => namaTipe;
}
