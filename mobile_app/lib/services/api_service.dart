import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../models/user_model.dart';
import '../models/auth_model.dart';
import '../models/attendance_model.dart';

class ApiService {
  static const String baseUrl = 'http://10.198.142.218:8080/api/v1'; // Untuk emulator Android
  // static const String baseUrl = 'http://localhost:8080/api/v1'; // Untuk web
  //static const String baseUrl = 'http://192.168.1.100:8080/api/v1'; // Untuk device fisik (ganti dengan IP komputer)

  static final Map<String, String> _headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  };

  // Token management
  static Future<void> saveTokens(String accessToken, String refreshToken, int expiresIn) async {
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
    final expiryTime = loginTime + (expiresIn * 1000); // Convert to milliseconds
    
    return now >= expiryTime;
  }

  static Future<void> clearTokens() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('access_token');
    await prefs.remove('refresh_token');
    await prefs.remove('expires_in');
    await prefs.remove('login_time');
  }

  static Future<Map<String, String>> getHeaders() async {
    final token = await getAccessToken();
    if (token != null) {
      return {
        ..._headers,
        'Authorization': 'Bearer $token',
      };
    }
    return _headers;
  }

  // Auth APIs
  static Future<LoginResponse> login(String email, String password) async {
    try {
      print('[API] Attempting login for: $email');
      
      final response = await http.post(
        Uri.parse('$baseUrl/auth/login'),
        headers: _headers,
        body: jsonEncode({
          'email': email,
          'password': password,
        }),
      );

      print('[API] Login response status: ${response.statusCode}');
      print('[API] Login response body: ${response.body}');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        
        // FIX #1: Handle Golang response wrapper dengan "data" field
        // Golang mengirim: { "status": "success", "message": "...", "data": { "user": {...}, "access_token": "..." } }
        final Map<String, dynamic> responseData = data['data'] ?? data;
        
        final loginResponse = LoginResponse.fromJson(responseData);
        
        // Save tokens
        await saveTokens(
          loginResponse.accessToken,
          loginResponse.refreshToken,
          loginResponse.expiresIn,
        );
        
        print('[API] Login successful for user: ${loginResponse.user.id}');
        return loginResponse;
      } else {
        // FIX #4: Improved error handling
        try {
          final error = jsonDecode(response.body);
          final errorMessage = error['message'] ?? error['error'] ?? 'Login failed';
          throw Exception('Login error: $errorMessage');
        } catch (e) {
          throw Exception('Login failed with status ${response.statusCode}');
        }
      }
    } catch (e) {
      print('[API] Login exception: $e');
      throw Exception('Connection error: $e');
    }
  }

  // FIX #5: Add register method yang kurang
  static Future<User> register({
    required String nik,
    required String email,
    required String password,
    required String fullName,
    required String role,
    required String department,
    required String position,
    String? phone,
    String? address,
  }) async {
    try {
      print('[API] Attempting registration for: $email');
      
      final response = await http.post(
        Uri.parse('$baseUrl/auth/register'),
        headers: _headers,
        body: jsonEncode({
          'nik': nik,
          'email': email,
          'password': password,
          'full_name': fullName,
          'role': role,
          'department': department,
          'position': position,
          'phone': phone,
          'address': address,
        }),
      );

      print('[API] Register response status: ${response.statusCode}');
      print('[API] Register response body: ${response.body}');

      if (response.statusCode == 201 || response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final Map<String, dynamic> responseData = data['data'] ?? data;
        
        final user = User.fromJson(responseData);
        print('[API] Registration successful for: ${user.email}');
        return user;
      } else {
        try {
          final error = jsonDecode(response.body);
          final errorMessage = error['message'] ?? error['error'] ?? 'Registration failed';
          throw Exception('Registration error: $errorMessage');
        } catch (e) {
          throw Exception('Registration failed with status ${response.statusCode}');
        }
      }
    } catch (e) {
      print('[API] Registration exception: $e');
      throw Exception('Registration error: $e');
    }
  }

  static Future<LoginResponse> refreshToken() async {
    try {
      final refreshToken = await getRefreshToken();
      if (refreshToken == null) throw Exception('No refresh token');

      final response = await http.post(
        Uri.parse('$baseUrl/auth/refresh'),
        headers: _headers,
        body: jsonEncode({
          'refresh_token': refreshToken,
        }),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        // FIX #1: Handle Golang response wrapper
        final Map<String, dynamic> responseData = data['data'] ?? data;
        final loginResponse = LoginResponse.fromJson(responseData);
        
        await saveTokens(
          loginResponse.accessToken,
          loginResponse.refreshToken,
          loginResponse.expiresIn,
        );
        
        return loginResponse;
      } else {
        throw Exception('Failed to refresh token');
      }
    } catch (e) {
      throw Exception('Refresh token error: $e');
    }
  }

  static Future<void> logout() async {
    try {
      final headers = await getHeaders();
      await http.post(
        Uri.parse('$baseUrl/auth/logout'),
        headers: headers,
      );
    } finally {
      await clearTokens();
    }
  }

  // Attendance APIs
  static Future<Map<String, dynamic>> clockIn({
    required String employeeId,
    required double latitude,
    required double longitude,
    required String photoPath,
  }) async {
    try {
      // Check if token expired
      if (await isTokenExpired()) {
        await refreshToken();
      }

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/attendance/clock-in'),
      );

      final headers = await getHeaders();
      request.headers.addAll(headers);

      // Add fields
      request.fields['employee_id'] = employeeId;
      request.fields['latitude'] = latitude.toString();
      request.fields['longitude'] = longitude.toString();

      // Add photo
      request.files.add(await http.MultipartFile.fromPath('photo', photoPath));

      final response = await request.send();
      final responseData = await response.stream.bytesToString();

      if (response.statusCode == 200 || response.statusCode == 201) {
        return jsonDecode(responseData);
      } else {
        throw Exception('Clock in failed: ${response.reasonPhrase}');
      }
    } catch (e) {
      throw Exception('Clock in error: $e');
    }
  }

  static Future<Map<String, dynamic>> clockOut({
    required String employeeId,
    required double latitude,
    required double longitude,
    required String photoPath,
  }) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/attendance/clock-out'),
      );

      final headers = await getHeaders();
      request.headers.addAll(headers);

      request.fields['employee_id'] = employeeId;
      request.fields['latitude'] = latitude.toString();
      request.fields['longitude'] = longitude.toString();
      request.files.add(await http.MultipartFile.fromPath('photo', photoPath));

      final response = await request.send();
      final responseData = await response.stream.bytesToString();

      if (response.statusCode == 200 || response.statusCode == 201) {
        return jsonDecode(responseData);
      } else {
        throw Exception('Clock out failed: ${response.reasonPhrase}');
      }
    } catch (e) {
      throw Exception('Clock out error: $e');
    }
  }

  static Future<List<AttendanceRecord>> getAttendanceHistory({
    int? month,
    int? year,
    String? status,
  }) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      var url = '$baseUrl/attendance/history';
      var params = <String, String>{};
      
      if (month != null) params['month'] = month.toString();
      if (year != null) params['year'] = year.toString();
      if (status != null) params['status'] = status;
      
      if (params.isNotEmpty) {
        url += '?' + Uri(queryParameters: params).query;
      }

      final headers = await getHeaders();
      final response = await http.get(
        Uri.parse(url),
        headers: headers,
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['data'] != null) {
          return (data['data'] as List)
              .map((item) => AttendanceRecord.fromJson(item))
              .toList();
        }
        return [];
      } else {
        throw Exception('Failed to get history');
      }
    } catch (e) {
      throw Exception('History error: $e');
    }
  }

  static Future<Map<String, dynamic>> getTodayAttendance() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http.get(
        Uri.parse('$baseUrl/attendance/today'),
        headers: headers,
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      } else {
        throw Exception('Failed to get today attendance');
      }
    } catch (e) {
      throw Exception('Today attendance error: $e');
    }
  }

  // Profile APIs
  static Future<User> getProfile() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http.get(
        Uri.parse('$baseUrl/profile'),
        headers: headers,
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return User.fromJson(data['user'] ?? data);
      } else {
        throw Exception('Failed to get profile');
      }
    } catch (e) {
      throw Exception('Profile error: $e');
    }
  }

  static Future<User> updateProfile(Map<String, dynamic> profileData) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http.put(
        Uri.parse('$baseUrl/profile'),
        headers: headers,
        body: jsonEncode(profileData),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return User.fromJson(data['user'] ?? data);
      } else {
        throw Exception('Failed to update profile');
      }
    } catch (e) {
      throw Exception('Update profile error: $e');
    }
  }

  static Future<Map<String, dynamic>> changePassword({
    required String oldPassword,
    required String newPassword,
  }) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http.post(
        Uri.parse('$baseUrl/profile/change-password'),
        headers: headers,
        body: jsonEncode({
          'old_password': oldPassword,
          'new_password': newPassword,
        }),
      );

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      } else {
        throw Exception('Failed to change password');
      }
    } catch (e) {
      throw Exception('Change password error: $e');
    }
  }

  // Request APIs (Cuti, Izin, etc)
  static Future<Map<String, dynamic>> submitLeaveRequest({
    required String type,
    required DateTime startDate,
    required DateTime endDate,
    required String reason,
    String? attachmentPath,
  }) async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('$baseUrl/requests/leave'),
      );

      final headers = await getHeaders();
      request.headers.addAll(headers);

      request.fields['type'] = type;
      request.fields['start_date'] = startDate.toIso8601String();
      request.fields['end_date'] = endDate.toIso8601String();
      request.fields['reason'] = reason;

      if (attachmentPath != null) {
        request.files.add(await http.MultipartFile.fromPath('attachment', attachmentPath));
      }

      final response = await request.send();
      final responseData = await response.stream.bytesToString();

      if (response.statusCode == 200 || response.statusCode == 201) {
        return jsonDecode(responseData);
      } else {
        throw Exception('Failed to submit request');
      }
    } catch (e) {
      throw Exception('Submit request error: $e');
    }
  }

  static Future<List<LeaveRequest>> getLeaveRequests() async {
    try {
      if (await isTokenExpired()) {
        await refreshToken();
      }

      final headers = await getHeaders();
      final response = await http.get(
        Uri.parse('$baseUrl/requests/leave'),
        headers: headers,
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['data'] != null) {
          return (data['data'] as List)
              .map((item) => LeaveRequest.fromJson(item))
              .toList();
        }
        return [];
      } else {
        throw Exception('Failed to get leave requests');
      }
    } catch (e) {
      throw Exception('Leave requests error: $e');
    }
  }
}
