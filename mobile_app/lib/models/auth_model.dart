// lib/models/auth_model.dart
import 'user_model.dart';

class LoginResponse {
  final User user;
  final String accessToken;
  final String refreshToken;
  final int expiresIn;
  final bool? requiresFaceRegistration;

  LoginResponse({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
    required this.expiresIn,
    this.requiresFaceRegistration,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    print('[LoginResponse] Parsing response: $json');

    try {
      // Handle response wrapper
      final Map<String, dynamic> responseData = json['data'] ?? json;

      // Extract user data
      final userData = responseData['user'] ?? responseData;
      final user = User.fromJson(userData);

      // Parse expiresIn
      int expiresIn = responseData['expires_in'] ?? 3600;
      if (expiresIn > 1000000000) {
        expiresIn = expiresIn - DateTime.now().millisecondsSinceEpoch ~/ 1000;
        if (expiresIn < 0) expiresIn = 3600;
      }

      // Dapatkan requiresFaceRegistration - CARI DI BEBERAPA TEMPAT
      bool requiresFaceRegistration = false;

      // 1. Cek di responseData langsung
      if (responseData.containsKey('requires_face_registration')) {
        requiresFaceRegistration =
            responseData['requires_face_registration'] == true;
        print(
          '[LoginResponse] Found in responseData: $requiresFaceRegistration',
        );
      }
      // 2. Cek di json.data
      else if (json.containsKey('data') && json['data'] != null) {
        if (json['data'].containsKey('requires_face_registration')) {
          requiresFaceRegistration =
              json['data']['requires_face_registration'] == true;
          print(
            '[LoginResponse] Found in json.data: $requiresFaceRegistration',
          );
        }
      }
      // 3. Cek di root json
      else if (json.containsKey('requires_face_registration')) {
        requiresFaceRegistration = json['requires_face_registration'] == true;
        print('[LoginResponse] Found in root json: $requiresFaceRegistration');
      }

      print(
        '[LoginResponse] FINAL requiresFaceRegistration: $requiresFaceRegistration',
      );

      return LoginResponse(
        user: user,
        accessToken: responseData['access_token'] ?? '',
        refreshToken: responseData['refresh_token'] ?? '',
        expiresIn: expiresIn,
        requiresFaceRegistration: requiresFaceRegistration,
      );
    } catch (e) {
      print('[LoginResponse] Parse error: $e');
      rethrow;
    }
  }

  Map<String, dynamic> toJson() {
    return {
      'user': user.toJson(),
      'access_token': accessToken,
      'refresh_token': refreshToken,
      'expires_in': expiresIn,
      'requires_face_registration': requiresFaceRegistration,
    };
  }
}

class LoginRequest {
  final String email;
  final String password;

  LoginRequest({required this.email, required this.password});

  Map<String, dynamic> toJson() {
    return {'email': email, 'password': password};
  }
}

class RefreshTokenRequest {
  final String refreshToken;

  RefreshTokenRequest({required this.refreshToken});

  Map<String, dynamic> toJson() {
    return {'refresh_token': refreshToken};
  }
}

class RegisterRequest {
  final String nik;
  final String email;
  final String password;
  final String fullName;
  final String role;
  final String department;
  final String position;
  final String? phone;
  final String? address;

  RegisterRequest({
    required this.nik,
    required this.email,
    required this.password,
    required this.fullName,
    required this.role,
    required this.department,
    required this.position,
    this.phone,
    this.address,
  });

  Map<String, dynamic> toJson() {
    return {
      'nik': nik,
      'email': email,
      'password': password,
      'full_name': fullName,
      'role': role,
      'department': department,
      'position': position,
      'phone': phone,
      'address': address,
    };
  }
}
