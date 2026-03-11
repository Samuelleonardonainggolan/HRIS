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
      // FIX #2 & #3: Handle both direct response and wrapped response
      final Map<String, dynamic> responseData = json['data'] ?? json;
      
      // Extract user data
      final userData = responseData['user'] ?? responseData;
      final user = User.fromJson(userData);
      
      // FIX #2: Convert expiresIn dari Unix timestamp ke seconds jika perlu
      int expiresIn = responseData['expires_in'] ?? 3600;
      
      // Jika expiresIn adalah Unix timestamp besar (>1000000000), convert ke seconds
      if (expiresIn > 1000000000) {
        expiresIn = expiresIn - DateTime.now().millisecondsSinceEpoch ~/ 1000;
        if (expiresIn < 0) expiresIn = 3600; // Default 1 hour
      }

      // Get requiresFaceRegistration from response
      bool? requiresFaceRegistration = responseData['requires_face_registration'];
      
      // If not found in responseData, check in data field
      if (requiresFaceRegistration == null && json['data'] != null) {
        requiresFaceRegistration = json['data']['requires_face_registration'];
      }

      final response = LoginResponse(
        user: user,
        accessToken: responseData['access_token'] ?? '',
        refreshToken: responseData['refresh_token'] ?? '',
        expiresIn: expiresIn,
        requiresFaceRegistration: requiresFaceRegistration,
      );
      
      print('[LoginResponse] Parsed successfully - User ID: ${user.id}, requiresFaceRegistration: $requiresFaceRegistration');
      return response;
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

  LoginRequest({
    required this.email,
    required this.password,
  });

  Map<String, dynamic> toJson() {
    return {
      'email': email,
      'password': password,
    };
  }
}

class RefreshTokenRequest {
  final String refreshToken;

  RefreshTokenRequest({required this.refreshToken});

  Map<String, dynamic> toJson() {
    return {
      'refresh_token': refreshToken,
    };
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
