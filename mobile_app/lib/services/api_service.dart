// lib/services/api_service.dart
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:flutter/foundation.dart';
import 'package:intl/intl.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../models/user_model.dart';
import '../models/auth_model.dart';
import '../models/attendance_model.dart';
import '../models/leave_request.dart';
import '../models/overtime_request.dart';
import '../models/assignment.dart';
import 'package:path_provider/path_provider.dart';

class ApiService {
  static const String baseUrl = 'http://10.252.84.48:8080/api/v1';

  static final ValueNotifier<User?> currentUser = ValueNotifier<User?>(null);

  static final Map<String, String> _headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };

  // ─── Token Management ─ ───

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
      currentUser.value = null;
    }
  }

  static Future<int> getUnreadNotificationCount() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/notifications/unread-count'), headers: headers)
          .timeout(const Duration(seconds: 15));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final unread = data['data']?['unread_count'];
        if (unread is int) return unread;
        if (unread is num) return unread.toInt();
      }

      return 0;
    } catch (e) {
      print('[API] getUnreadNotificationCount error: $e');
      return 0;
    }
  }

  static Future<List<Map<String, dynamic>>> getNotifications({int limit = 30}) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/notifications?limit=$limit'), headers: headers)
          .timeout(const Duration(seconds: 15));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final items = data['data'];
        if (items is List) {
          return items.cast<Map<String, dynamic>>();
        }
      }

      return [];
    } catch (e) {
      print('[API] getNotifications error: $e');
      return [];
    }
  }

  static Future<void> markNotificationAsRead(String id) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      await http
          .patch(Uri.parse('$baseUrl/notifications/$id/read'), headers: headers)
          .timeout(const Duration(seconds: 15));
    } catch (e) {
      print('[API] markNotificationAsRead error: $e');
    }
  }

  static Future<void> markAllNotificationsAsRead() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      await http
          .post(Uri.parse('$baseUrl/notifications/read-all'), headers: headers)
          .timeout(const Duration(seconds: 15));
    } catch (e) {
      print('[API] markAllNotificationsAsRead error: $e');
    }
  }

  // ─── Attendance ─────────────────────────────────────────────────────────────

  /// ✅ GET /api/v1/attendance/schedule-info
  static Future<WorkScheduleInfoResponse> getWorkScheduleInfo() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/attendance/schedule-info'), headers: headers)
          .timeout(const Duration(seconds: 30));

      print('[API] getWorkScheduleInfo status: ${response.statusCode}');
      print('[API] getWorkScheduleInfo body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return WorkScheduleInfoResponse.fromJson(data['data'] ?? {});
      } else {
        throw Exception('Gagal mengambil jadwal kerja');
      }
    } catch (e) {
      print('[API] getWorkScheduleInfo error: $e');
      rethrow;
    }
  }

  /// POST /api/v1/geofences/check
  /// Validasi lokasi user terhadap semua geofence aktif yang applicable dari database.
  static Future<GeoVerificationResult> checkUserInGeofence({
    required double latitude,
    required double longitude,
  }) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .post(
            Uri.parse('$baseUrl/geofences/check'),
            headers: headers,
            body: jsonEncode({'latitude': latitude, 'longitude': longitude}),
          )
          .timeout(const Duration(seconds: 30));

      print('[API] checkUserInGeofence status: ${response.statusCode}');
      print('[API] checkUserInGeofence body: ${response.body}');

      final jsonResponse = jsonDecode(response.body);
      if (response.statusCode == 200) {
        final data = (jsonResponse['data'] is Map)
            ? jsonResponse['data'] as Map<String, dynamic>
            : <String, dynamic>{};
        return GeoVerificationResult.fromJson(data);
      }

      throw Exception(
        jsonResponse['error'] ??
            jsonResponse['message'] ??
            'Gagal validasi lokasi geofence',
      );
    } catch (e) {
      print('[API] checkUserInGeofence error: $e');
      rethrow;
    }
  }

  static Future<AttendanceProcessResult> processAttendance({
    required String recordType,
    required double latitude,
    required double longitude,
    required String photoPath,
    required String liveness,
  }) async {
    try {
      print('[API] Processing attendance: $recordType, liveness: $liveness');

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
      request.fields['liveness'] = liveness;
      // ✅ PERBAIKN: Jangan kirim verify_only atau kirim false = ini adalah submission sesungguhnya
      // request.fields['verify_only'] = 'false'; // TIDAK kirim = backend default = save to DB

      request.files.add(
        await http.MultipartFile.fromPath(
          'photo',
          photoPath,
          filename:
              '${recordType}_${DateTime.now().millisecondsSinceEpoch}.jpg',
        ),
      );

      print(
        '[API] SUBMIT ATTENDANCE: record_type=$recordType, lat=$latitude, lng=$longitude, verify_only=NOT_SENT (akan disimpan ke DB)',
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

  static Future<void> startBreak() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .post(Uri.parse('$baseUrl/attendance/break/start'), headers: headers)
          .timeout(const Duration(seconds: 30));

      if (response.statusCode == 200) {
        return;
      }

      final data = jsonDecode(response.body);
      throw Exception(data['message'] ?? 'Gagal memulai break');
    } catch (e) {
      print('[API] startBreak error: $e');
      rethrow;
    }
  }

  static Future<void> endBreak() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http
          .post(Uri.parse('$baseUrl/attendance/break/end'), headers: headers)
          .timeout(const Duration(seconds: 30));

      if (response.statusCode == 200) {
        return;
      }

      final data = jsonDecode(response.body);
      throw Exception(data['message'] ?? 'Gagal mengakhiri break');
    } catch (e) {
      print('[API] endBreak error: $e');
      rethrow;
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

  static Future<void> submitPengajuan({
    required String tipePengajuanId,
    required String tanggalMulai,
    required String tanggalSelesai,
    required int totalHari,
    required String alasan,
    String? dokumenUrl,
    String? startTime,
    String? endTime,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final userId = await getUserId();
      if (userId == null) throw Exception('User ID tidak ditemukan');

      final token = await getAccessToken();
      if (token == null || token.isEmpty) {
        throw Exception('Token tidak ditemukan. Silakan login ulang.');
      }

      final request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/pengajuan'),
      );

      request.headers.addAll({
        'Authorization': 'Bearer $token',
        'Accept': 'application/json',
      });

      request.fields['user_id'] = userId;
      request.fields['request_type_id'] = tipePengajuanId;
      request.fields['tipe_pengajuan_id'] = tipePengajuanId;
      request.fields['start_date'] = tanggalMulai;
      request.fields['tanggal_mulai'] = tanggalMulai;
      request.fields['end_date'] = tanggalSelesai;
      request.fields['tanggal_selesai'] = tanggalSelesai;
      request.fields['days_total'] = totalHari.toString();
      request.fields['total_hari'] = totalHari.toString();
      request.fields['reason'] = alasan;
      request.fields['alasan'] = alasan;

      if (startTime != null) request.fields['start_time'] = startTime;
      if (endTime != null) request.fields['end_time'] = endTime;

      if (dokumenUrl != null && dokumenUrl.isNotEmpty) {
        final file = File(dokumenUrl);
        if (await file.exists()) {
          request.files.add(
            await http.MultipartFile.fromPath(
              'document',
              dokumenUrl,
              filename: dokumenUrl.split('/').last,
            ),
          );
        } else {
          request.fields['dokumen_url'] = dokumenUrl;
          request.fields['document_url'] = dokumenUrl;
        }
      }

      final streamedResponse = await request.send().timeout(
        const Duration(seconds: 30),
      );
      final response = await http.Response.fromStream(streamedResponse);

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

  static Future<List<LeaveRequest>> getMyPengajuan() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      // Fetch Leave Requests
      final response = await http
          .get(Uri.parse('$baseUrl/pengajuan'), headers: headers)
          .timeout(const Duration(seconds: 30));

      List<LeaveRequest> list = [];

      print('[API] getMyPengajuan status: ${response.statusCode}');
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        List<dynamic> rawList = [];
        if (data is List) {
          rawList = data;
        } else if (data is Map<String, dynamic>) {
          final rawData = data['data'];
          if (rawData is List) {
            rawList = rawData;
          } else if (rawData is Map<String, dynamic>) {
            rawList = rawData['data'] ?? rawData['pengajuan'] ?? [];
          } else {
            rawList = data['pengajuan'] ?? [];
          }
        }
        list.addAll(
          rawList.map((e) => LeaveRequest.fromJson(e as Map<String, dynamic>)),
        );
      } else {
        print('[API] Gagal mengambil data pengajuan (${response.statusCode})');
      }

      // Sort by start date descending
      list.sort((a, b) => b.startDate.compareTo(a.startDate));
      return list;
    } catch (e) {
      print('[API] getMyPengajuan error: $e');
      throw Exception('getMyPengajuan error: $e');
    }
  }

  static Future<List<OvertimeRequest>> getMyOvertime() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/my-overtime'), headers: headers)
          .timeout(const Duration(seconds: 30));

      List<OvertimeRequest> list = [];
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        List<dynamic> rawOvertime = [];
        if (data['data'] is List) {
          rawOvertime = data['data'];
        }
        list.addAll(
          rawOvertime.map(
            (e) => OvertimeRequest.fromJson(e as Map<String, dynamic>),
          ),
        );
        list.sort((a, b) => b.date.compareTo(a.date));
      }
      return list;
    } catch (e) {
      print('[API] getMyOvertime error: $e');
      return [];
    }
  }

  static Future<Map<String, dynamic>?> claimOvertimeReward(String id, String rewardType, {String? rewardDate, String? rewardOption}) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final body = {
        'reward_type': rewardType,
        if (rewardDate != null) 'reward_date': rewardDate,
        if (rewardOption != null) 'reward_option': rewardOption,
      };

      final response = await http
          .post(
            Uri.parse('$baseUrl/my-overtime/$id/reward'),
            headers: headers,
            body: jsonEncode(body),
          )
          .timeout(const Duration(seconds: 30));

      final responseBody = jsonDecode(response.body);
      if (response.statusCode != 200) {
        throw Exception(responseBody['message'] ?? 'Gagal mengklaim reward lembur');
      }
      return responseBody['data'] as Map<String, dynamic>?;
    } catch (e) {
      print('[API] claimOvertimeReward error: $e');
      rethrow;
    }
  }

  static Future<List<OvertimeRequest>> getApprovedOvertimeByMonth({
    required int month,
    required int year,
  }) async {
    try {
      final userId = await getUserId();
      final user = currentUser.value;

      List<OvertimeRequest> source = [];
      if (user?.isManagerDept == true) {
        source = await getMyOvertimeRequests();
      } else if (user?.isManagerHR == true) {
        source = await getOvertimeRequestsForHR();
      } else {
        source = await getAssignedOvertimeRequests();
      }

      final filtered = source.where((o) {
        final inMonth = o.date.month == month && o.date.year == year;
        if (!inMonth) return false;

        if (o.isPublished) return true;

        if (userId == null || userId.isEmpty) return false;
        final myEmployeeEntry = o.employees.where((e) => e.userId == userId).toList();
        if (myEmployeeEntry.isEmpty) return false;
        return myEmployeeEntry.any((e) => e.isAgreed || e.isRejected);
      }).toList();

      filtered.sort((a, b) => b.date.compareTo(a.date));
      return filtered;
    } catch (e) {
      print('[API] getApprovedOvertimeByMonth error: $e');
      return [];
    }
  }

  static Future<void> submitOvertime({
    required String tanggal,
    required String startTime,
    required String endTime,
    required String alasan,
    required String total,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final userId = await getUserId();
      if (userId == null) throw Exception('User ID tidak ditemukan');

      final headers = await getHeaders();

      final response = await http
          .post(
            Uri.parse('$baseUrl/my-overtime'),
            headers: headers,
            body: jsonEncode({
              'user_id': userId,
              'date': tanggal,
              'start_time': startTime,
              'end_time': endTime,
              'reason': alasan,
              'total': total,
            }),
          )
          .timeout(const Duration(seconds: 30));

      print('[API] submitOvertime status: ${response.statusCode}');
      print('[API] submitOvertime body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        return;
      } else {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal mengirimkan pengajuan lembur');
      }
    } catch (e) {
      print('[API] submitOvertime error: $e');
      rethrow;
    }
  }

  static Future<void> updateOvertime({
    required String id,
    required String tanggal,
    required String startTime,
    required String endTime,
    required String alasan,
    required String total,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      final response = await http
          .put(
            Uri.parse('$baseUrl/my-overtime/$id'),
            headers: headers,
            body: jsonEncode({
              'date': tanggal,
              'start_time': startTime,
              'end_time': endTime,
              'reason': alasan,
              'total': total,
            }),
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode == 200 || response.statusCode == 201) {
        return;
      } else {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal memperbarui pengajuan lembur');
      }
    } catch (e) {
      print('[API] updateOvertime error: $e');
      rethrow;
    }
  }

  static Future<void> cancelOvertime(String id) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      final response = await http
          .delete(Uri.parse('$baseUrl/my-overtime/$id'), headers: headers)
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal membatalkan pengajuan lembur');
      }
    } catch (e) {
      print('[API] cancelOvertime error: $e');
      rethrow;
    }
  }

  static Future<LeaveRequest> updatePengajuan({
    required String pengajuanId,
    String? tipePengajuanId,
    String? tanggalMulai,
    String? tanggalSelesai,
    int? totalHari,
    String? alasan,
    String? dokumenUrl,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final token = await getAccessToken();
      if (token == null || token.isEmpty) {
        throw Exception('Token tidak ditemukan. Silakan login ulang.');
      }

      final request = http.MultipartRequest(
        'PUT',
        Uri.parse('$baseUrl/pengajuan/$pengajuanId'),
      );

      request.headers.addAll({
        'Authorization': 'Bearer $token',
        'Accept': 'application/json',
      });

      if (tipePengajuanId != null && tipePengajuanId.isNotEmpty) {
        request.fields['tipe_pengajuan_id'] = tipePengajuanId;
      }
      if (tanggalMulai != null && tanggalMulai.isNotEmpty) {
        request.fields['tanggal_mulai'] = tanggalMulai;
      }
      if (tanggalSelesai != null && tanggalSelesai.isNotEmpty) {
        request.fields['tanggal_selesai'] = tanggalSelesai;
      }
      if (totalHari != null) {
        request.fields['total_hari'] = totalHari.toString();
      }
      if (alasan != null && alasan.isNotEmpty) {
        request.fields['alasan'] = alasan;
      }

      if (dokumenUrl != null && dokumenUrl.isNotEmpty) {
        final file = File(dokumenUrl);
        if (await file.exists()) {
          request.files.add(
            await http.MultipartFile.fromPath(
              'document',
              dokumenUrl,
              filename: dokumenUrl.split('/').last,
            ),
          );
        } else {
          request.fields['dokumen_url'] = dokumenUrl;
        }
      }

      final streamedResponse = await request.send().timeout(
        const Duration(seconds: 30),
      );
      final response = await http.Response.fromStream(streamedResponse);

      final data = jsonDecode(response.body);
      if (response.statusCode == 200) {
        final payload = (data['data'] is Map<String, dynamic>)
            ? data['data'] as Map<String, dynamic>
            : <String, dynamic>{};
        return LeaveRequest.fromJson(payload);
      }

      throw Exception(data['message'] ?? 'Gagal mengubah pengajuan');
    } catch (e) {
      print('[API] updatePengajuan error: $e');
      rethrow;
    }
  }

  static Future<void> cancelPengajuan(String pengajuanId) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .delete(
            Uri.parse('$baseUrl/pengajuan/$pengajuanId'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode == 200) {
        return;
      }

      final data = jsonDecode(response.body);
      throw Exception(data['message'] ?? 'Gagal membatalkan pengajuan');
    } catch (e) {
      print('[API] cancelPengajuan error: $e');
      rethrow;
    }
  }

  static Future<List<LeaveRequest>> getApprovedPengajuanByMonth({
    required int month,
    required int year,
  }) async {
    try {
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
      final lastDay = DateTime(
        year,
        month + 1,
        1,
      ).subtract(const Duration(days: 1));

      final approved = all.where((p) {
        if (p.statusFinal != 'APPROVED') return false;

        final s = DateTime(
          p.startDate.year,
          p.startDate.month,
          p.startDate.day,
        );
        final e = DateTime(p.endDate.year, p.endDate.month, p.endDate.day);

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

  static Future<int> getLeaveBalance() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/pengajuan/leave-balance'), headers: headers)
          .timeout(const Duration(seconds: 30));

      print('[API] getLeaveBalance status: ${response.statusCode}');
      print('[API] getLeaveBalance body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final payload = data['data'] ?? data['balance'] ?? data;
        if (payload is Map<String, dynamic>) {
          final remaining =
              payload['remaining_kuota'] ?? payload['remainingQuota'] ?? 0;
          return int.tryParse(remaining.toString()) ?? 0;
        }
      }

      final results = await Future.wait([getTipePengajuan(), getMyPengajuan()]);

      final types = results[0] as List<TipePengajuan>;
      final requests = results[1] as List<LeaveRequest>;

      final deductingTypes = types
          .where((type) => type.potongKuota)
          .map((type) => type.namaTipe.trim().toLowerCase())
          .where((name) => name.isNotEmpty)
          .toSet();

      final currentYear = DateTime.now().year;
      final usedQuota = requests
          .where((request) {
            if (request.statusFinal != 'APPROVED') return false;
            if (request.startDate.year != currentYear) return false;
            return deductingTypes.contains(request.type.trim().toLowerCase());
          })
          .fold<int>(
            0,
            (sum, request) => sum + (request.days > 0 ? request.days : 1),
          );

      const annualQuota = 12;
      final remaining = annualQuota - usedQuota;
      return remaining < 0 ? 0 : remaining;
    } catch (e) {
      print('[API] getLeaveBalance error: $e');
      rethrow;
    }
  }

  // ─── Salary ─────────────────────────────────────────────────────────────────

  /// GET /employee-basic-salaries/users/:userId/active
  static Future<Map<String, dynamic>> getActiveSalary(String userId) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(
            Uri.parse('$baseUrl/employee-basic-salaries/users/$userId/active'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      print('[API] getActiveSalary status: ${response.statusCode}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return (data['data'] as Map<String, dynamic>?) ?? {};
      } else {
        // Kembalikan gaji 0 jika belum diset
        return {'basic_salary': 0};
      }
    } catch (e) {
      print('[API] getActiveSalary error: $e');
      return {'basic_salary': 0};
    }
  }

  static Future<Map<String, dynamic>> getMyPayroll(int month, int year) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(
            Uri.parse('$baseUrl/my-payroll?month=$month&year=$year'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      print('[API] getMyPayroll status: ${response.statusCode}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return (data['data'] as Map<String, dynamic>?) ?? {};
      } else {
        throw Exception('Slip gaji tidak ditemukan');
      }
    } catch (e) {
      print('[API] getMyPayroll error: $e');
      rethrow;
    }
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
        final errorData = jsonDecode(response.body);
        // Error details from Python API are in 'error' field
        final detailedError =
            errorData['error'] ??
            errorData['message'] ??
            'Gagal mengekstrak wajah';
        throw Exception(detailedError);
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
        final user = User.fromJson(userJson);
        currentUser.value = user;
        return user;
      } else {
        throw Exception('Gagal mengambil profil (${response.statusCode})');
      }
    } catch (e) {
      throw Exception('Profile error: $e');
    }
  }

  static Future<void> updateProfile(
    Map<String, dynamic> profileData, {
    String? avatarPath,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();
      final headers = await getHeaders();
      late final http.Response response;

      if (avatarPath != null && avatarPath.isNotEmpty) {
        final token = await getAccessToken();
        if (token == null || token.isEmpty) {
          throw Exception('Token tidak ditemukan. Silakan login ulang.');
        }

        final request = http.MultipartRequest(
          'PUT',
          Uri.parse('$baseUrl/profile'),
        );
        request.headers.addAll({
          'Authorization': 'Bearer $token',
          'Accept': 'application/json',
        });

        profileData.forEach((key, value) {
          if (value == null) return;
          final textValue = value.toString().trim();
          if (textValue.isNotEmpty) {
            request.fields[key] = textValue;
          }
        });

        request.files.add(
          await http.MultipartFile.fromPath(
            'avatar',
            avatarPath,
            filename:
                'profile_${DateTime.now().millisecondsSinceEpoch}${avatarPath.contains('.') ? avatarPath.substring(avatarPath.lastIndexOf('.')) : '.jpg'}',
          ),
        );

        final streamedResponse = await request.send().timeout(
          const Duration(seconds: 30),
        );
        response = await http.Response.fromStream(streamedResponse);
      } else {
        response = await http
            .put(
              Uri.parse('$baseUrl/profile'),
              headers: headers,
              body: jsonEncode(profileData),
            )
            .timeout(const Duration(seconds: 30));
      }

      print('[API] updateProfile status: ${response.statusCode}');
      if (response.statusCode == 200) {
        await getProfile();
        return;
      }
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

  static Future<AttendanceProcessResult> confirmAttendance({
    required String recordType,
    required double latitude,
    required double longitude,
    required String photoFilename,
    required double faceSimilarity,
  }) async {
    try {
      print('[API] Confirming attendance: $recordType');

      if (await isTokenExpired()) {
        await refreshToken();
      }

      final token = await getAccessToken();
      if (token == null || token.isEmpty) {
        throw Exception('Token tidak ditemukan. Silakan login ulang.');
      }

      final headers = await getHeaders();

      final response = await http
          .post(
            Uri.parse('$baseUrl/attendance/confirm'),
            headers: headers,
            body: {
              'record_type': recordType,
              'latitude': latitude.toString(),
              'longitude': longitude.toString(),
              'photo_filename': photoFilename,
              'face_similarity': faceSimilarity.toString(),
            },
          )
          .timeout(const Duration(seconds: 30));

      print('[API] Confirm attendance status: ${response.statusCode}');
      print('[API] Confirm attendance body: ${response.body}');

      final jsonResponse = jsonDecode(response.body);

      if (response.statusCode == 200) {
        return AttendanceProcessResult.fromJson(jsonResponse);
      } else if (response.statusCode == 401) {
        await clearTokens();
        throw Exception('Sesi telah berakhir. Silakan login ulang.');
      } else {
        throw Exception(
          jsonResponse['message'] ?? 'Gagal mengkonfirmasi absensi',
        );
      }
    } catch (e) {
      print('[API] Confirm attendance error: $e');
      rethrow;
    }
  }

  // ─── Overtime Request (New Flow - Department Head Directed) ───────────────

  /// Get my draft overtime requests (for department heads)
  static Future<List<OvertimeRequest>> getMyOvertimeRequests() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      // Correct endpoint for manager department to see their dept requests
      final response = await http
          .get(Uri.parse('$baseUrl/dept-overtime-requests'), headers: headers)
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        // Fallback to my-overtime if dept-overtime fails (e.g. 403)
        return getMyOvertime();
      }

      final data = jsonDecode(response.body);
      final raw = data['data'];
      if (raw is! List) return [];

      final list = raw
          .whereType<Map<String, dynamic>>()
          .map((e) => OvertimeRequest.fromJson(e))
          .toList();

      list.sort((a, b) => b.date.compareTo(a.date));
      return list;
    } catch (e) {
      print('[API] getMyOvertimeRequests error: $e');
      return [];
    }
  }

  /// Get assigned overtime requests (for employees)
  static Future<List<OvertimeRequest>> getAssignedOvertimeRequests() async {
    return getMyOvertime();
  }

  /// Get overtime requests for HR review
  static Future<List<OvertimeRequest>> getOvertimeRequestsForHR({
    String status = 'ALL',
    String search = '',
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final uri = Uri.parse('$baseUrl/overtime-requests').replace(
        queryParameters: {
          'status': status,
          if (search.trim().isNotEmpty) 'search': search.trim(),
        },
      );

      final response = await http
          .get(uri, headers: headers)
          .timeout(const Duration(seconds: 30));

      List<OvertimeRequest> list = [];
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['data'] is List) {
          list = (data['data'] as List)
              .map((e) => OvertimeRequest.fromJson(e as Map<String, dynamic>))
              .toList();
          list.sort((a, b) => b.date.compareTo(a.date));
        }
      }
      return list;
    } catch (e) {
      print('[API] getOvertimeRequestsForHR error: $e');
      return [];
    }
  }

  /// Delete draft overtime request (kepala departemen only)
  static Future<void> deleteDraftOvertimeRequest(String requestId) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .delete(
            Uri.parse('$baseUrl/my-overtime-requests/$requestId'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal menghapus draft lembur');
      }
    } catch (e) {
      print('[API] deleteDraftOvertimeRequest error: $e');
      rethrow;
    }
  }

  /// Get employees in my department (for kepala departemen)
  static Future<List<User>> getEmployeesMyDepartment() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/employees/my-department'), headers: headers)
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        return [];
      }

      final data = jsonDecode(response.body);
      final raw = data['data'];
      if (raw is! List) return [];

      return raw
          .whereType<Map<String, dynamic>>()
          .map(User.fromJson)
          .toList();
    } catch (e) {
      print('[API] getEmployeesMyDepartment error: $e');
      return [];
    }
  }

  /// Create new overtime request (department head creates for multiple employees)
  static Future<String> createOvertimeRequest({
    required List<String> employeeIds,
    required DateTime date,
    required String startTime,
    required String endTime,
    required String reason,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      final response = await http
          .post(
            Uri.parse('$baseUrl/my-overtime-requests'),
            headers: headers,
            body: jsonEncode({
              'employee_ids': employeeIds,
              'date': date.toIso8601String().split('T')[0], // YYYY-MM-DD
              'start_time': startTime,
              'end_time': endTime,
              'reason': reason,
            }),
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode == 201) {
        final data = jsonDecode(response.body);
        return data['data']['id'] ?? '';
      } else {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal membuat pengajuan lembur');
      }
    } catch (e) {
      print('[API] createOvertimeRequest error: $e');
      rethrow;
    }
  }

  /// Submit overtime request (change from draft to submitted)
  static Future<void> submitOvertimeRequest(String requestId) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      final response = await http
          .post(
            Uri.parse('$baseUrl/my-overtime-requests/$requestId/submit'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal mengirim pengajuan lembur');
      }
    } catch (e) {
      print('[API] submitOvertimeRequest error: $e');
      rethrow;
    }
  }

  /// Employee agrees to overtime request
  static Future<void> agreeOvertimeRequest(String requestId) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      final response = await http
          .post(
            Uri.parse('$baseUrl/my-overtime/$requestId/agree'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal menyetujui pengajuan lembur');
      }
    } catch (e) {
      print('[API] agreeOvertimeRequest error: $e');
      rethrow;
    }
  }

  /// Employee rejects overtime request with optional note
  static Future<void> rejectOvertimeRequest(
    String requestId, {
    String? rejectionNote,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      final response = await http
          .post(
            Uri.parse('$baseUrl/my-overtime/$requestId/reject'),
            headers: headers,
            body: jsonEncode({
              'rejection_note': rejectionNote ?? '',
            }),
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal menolak pengajuan lembur');
      }
    } catch (e) {
      print('[API] rejectOvertimeRequest error: $e');
      rethrow;
    }
  }

  /// Publish overtime letter (HR action)
  static Future<void> publishOvertimeLetter({
    required String requestId,
    required String letterUrl,
    String? notes,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();

      final response = await http
          .post(
            Uri.parse('$baseUrl/overtime-requests/$requestId/publish-letter'),
            headers: headers,
            body: jsonEncode({
              'letter_url': letterUrl,
              'notes': notes ?? '',
            }),
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(
            err['message'] ?? 'Gagal menerbitkan surat lembur resmi');
      }
    } catch (e) {
      print('[API] publishOvertimeLetter error: $e');
      rethrow;
    }
  }

  // ─── Assignment Request (Kadep to Employee) ─────────────────────────────

  static Future<List<Assignment>> getMyAssignments() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/dept-assignments'), headers: headers)
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) return [];

      final data = jsonDecode(response.body);
      final raw = data['data'];
      if (raw is! List) return [];

      final list = raw
          .whereType<Map<String, dynamic>>()
          .map(Assignment.fromJson)
          .where((e) => e.id.isNotEmpty)
          .toList();
      list.sort((a, b) => b.date.compareTo(a.date));
      return list;
    } catch (e) {
      print('[API] getMyAssignments error: $e');
      return [];
    }
  }

  static Future<List<Assignment>> getAssignedAssignments() async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .get(Uri.parse('$baseUrl/my-assigned-assignments'), headers: headers)
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) return [];

      final data = jsonDecode(response.body);
      final raw = data['data'];
      if (raw is! List) return [];

      final list = raw
          .whereType<Map<String, dynamic>>()
          .map(Assignment.fromJson)
          .where((e) => e.id.isNotEmpty)
          .toList();
      list.sort((a, b) => b.date.compareTo(a.date));
      return list;
    } catch (e) {
      print('[API] getAssignedAssignments error: $e');
      return [];
    }
  }

  static Future<List<Assignment>> getApprovedAssignmentsByMonth({
    required int month,
    required int year,
  }) async {
    try {
      final all = await getAssignedAssignments();
      final userId = currentUser.value?.id ?? await getUserId();
      final filtered = all.where((a) {
        final inMonth = a.date.month == month && a.date.year == year;
        if (!inMonth) return false;

        if (a.status != 'submitted' && a.status != 'published') return false;

        if (userId == null || userId.isEmpty) return false;
        final myEntry = a.employees.where((e) => e.userId == userId).toList();
        if (myEntry.isEmpty) return false;
        return myEntry.any((e) => e.employeeStatus == 'agreed');
      }).toList();
      return filtered;
    } catch (e) {
      print('[API] getApprovedAssignmentsByMonth error: $e');
      return [];
    }
  }

  static Future<String> createAssignment({
    required DateTime date,
    required String startTime,
    required String endTime,
    required String reason,
    required List<String> employeeIds,
    String? notes,
    String status = 'draft',
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final me = currentUser.value ?? await getProfile();
      final departmentId = me.departmentId;

      if (departmentId.isEmpty) {
        throw Exception('Department ID user tidak ditemukan');
      }

      final response = await http
          .post(
            Uri.parse('$baseUrl/dept-assignments'),
            headers: headers,
            body: jsonEncode({
              'department_id': departmentId,
              'date': DateFormat('yyyy-MM-dd').format(date),
              'start_time': startTime,
              'end_time': endTime,
              'reason': reason,
              'status': status,
              'notes': notes ?? '',
              'employees': employeeIds
                  .map((id) => {
                        'user_id': id,
                      })
                  .toList(),
            }),
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200 && response.statusCode != 201) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal membuat penugasan');
      }

      final data = jsonDecode(response.body);
      final item = data['data'];
      if (item is Map<String, dynamic>) {
        return (item['id'] ?? '').toString();
      }
      return '';
    } catch (e) {
      print('[API] createAssignment error: $e');
      rethrow;
    }
  }

  static Future<void> submitAssignment(String assignmentId) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .post(
            Uri.parse('$baseUrl/dept-assignments/$assignmentId/submit'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal submit penugasan');
      }
    } catch (e) {
      print('[API] submitAssignment error: $e');
      rethrow;
    }
  }

  static Future<void> agreeAssignment(String assignmentId) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .post(
            Uri.parse('$baseUrl/my-assigned-assignments/$assignmentId/agree'),
            headers: headers,
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal menyetujui penugasan');
      }
    } catch (e) {
      print('[API] agreeAssignment error: $e');
      rethrow;
    }
  }

  static Future<void> rejectAssignment(
    String assignmentId, {
    String? rejectionNote,
  }) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .post(
            Uri.parse('$baseUrl/my-assigned-assignments/$assignmentId/reject'),
            headers: headers,
            body: jsonEncode({
              'rejection_note': rejectionNote ?? '',
            }),
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal menolak penugasan');
      }
    } catch (e) {
      print('[API] rejectAssignment error: $e');
      rethrow;
    }
  }

  static Future<void> useReplacementDayOff(
    String assignmentId,
    String replacementDate,
  ) async {
    try {
      if (await isTokenExpired()) await refreshToken();

      final headers = await getHeaders();
      final response = await http
          .post(
            Uri.parse('$baseUrl/my-assigned-assignments/$assignmentId/use-day-off'),
            headers: headers,
            body: jsonEncode({
              'replacement_off_date': replacementDate,
            }),
          )
          .timeout(const Duration(seconds: 30));

      if (response.statusCode != 200) {
        final err = jsonDecode(response.body);
        throw Exception(err['message'] ?? 'Gagal menetapkan hari off pengganti');
      }
    } catch (e) {
      print('[API] useReplacementDayOff error: $e');
      rethrow;
    }
  }
}

// ─── Model TipePengajuan ────────────────────────────────────────────────────

class TipePengajuan {
  final String id;
  final String namaTipe;
  final String namaKategori;
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
    final categoryName = (json['category_name'] ?? json['nama_kategori'] ?? '')
        .toString();
    final normalizedCategory = categoryName.trim().toLowerCase();

    return TipePengajuan(
      id: (json['id'] ?? '').toString(),
      namaTipe: (json['type_name'] ?? json['nama_tipe'] ?? '').toString(),
      namaKategori: categoryName,
      potongKuota:
          json['quota_deduction'] == true ||
          json['potong_kuota'] == true ||
          normalizedCategory == 'izin' ||
          normalizedCategory == 'cuti',
      wajibLampiran:
          json['attachment_required'] == true || json['wajib_lampiran'] == true,
    );
  }

  @override
  String toString() => namaTipe;
}

// ─── Work Schedule Models ───────────────────────────────────────────────────

class WorkScheduleInfoResponse {
  final String userId;
  final List<String> hariKerja;
  final String waktuMulai; // HH:mm
  final String waktuSelesai; // HH:mm
  final bool aktif;
  final TodayScheduleInfo? todaySchedule;

  WorkScheduleInfoResponse({
    required this.userId,
    required this.hariKerja,
    required this.waktuMulai,
    required this.waktuSelesai,
    required this.aktif,
    this.todaySchedule,
  });

  factory WorkScheduleInfoResponse.fromJson(Map<String, dynamic> json) {
    return WorkScheduleInfoResponse(
      userId: json['user_id']?.toString() ?? '',
      hariKerja: List<String>.from(json['hari_kerja'] ?? []),
      waktuMulai: json['waktu_mulai']?.toString() ?? '08:00',
      waktuSelesai: json['waktu_selesai']?.toString() ?? '17:00',
      aktif: json['aktif'] == true,
      todaySchedule: json['today_schedule'] != null
          ? TodayScheduleInfo.fromJson(json['today_schedule'])
          : null,
    );
  }

  Map<String, dynamic> toJson() => {
    'user_id': userId,
    'hari_kerja': hariKerja,
    'waktu_mulai': waktuMulai,
    'waktu_selesai': waktuSelesai,
    'aktif': aktif,
    'today_schedule': todaySchedule?.toJson(),
  };
}

class TodayScheduleInfo {
  final bool isWorkDay;
  final String clockInWindow; // "HH:mm - HH:mm"
  final String clockOutWindow; // "HH:mm - HH:mm" ✅ DIUBAH
  final bool canClockIn;
  final bool canClockOut;
  final String message;

  TodayScheduleInfo({
    required this.isWorkDay,
    required this.clockInWindow,
    required this.clockOutWindow,
    required this.canClockIn,
    required this.canClockOut,
    required this.message,
  });

  factory TodayScheduleInfo.fromJson(Map<String, dynamic> json) {
    return TodayScheduleInfo(
      isWorkDay: json['is_work_day'] == true,
      clockInWindow: json['clock_in_window']?.toString() ?? '',
      clockOutWindow: json['clock_out_window']?.toString() ?? '',
      canClockIn: json['can_clock_in'] == true,
      canClockOut: json['can_clock_out'] == true,
      message: json['message']?.toString() ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
    'is_work_day': isWorkDay,
    'clock_in_window': clockInWindow,
    'clock_out_window': clockOutWindow,
    'can_clock_in': canClockIn,
    'can_clock_out': canClockOut,
    'message': message,
  };
}