import 'package:flutter/material.dart';
import '../models/user_model.dart';
import '../services/api_service.dart';
import 'package:shared_preferences/shared_preferences.dart';

class AuthProvider extends ChangeNotifier {
  User? _currentUser;
  bool _isLoading = false;
  String? _error;

  User? get currentUser => _currentUser;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get isLoggedIn => _currentUser != null;

  Future<bool> login(String email, String password) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      final response = await ApiService.login(email, password);
      _currentUser = response.user;
      
      // Simpan user di SharedPreferences
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('user_id', response.user.id);
      await prefs.setString('user_email', response.user.email);
      await prefs.setString('user_name', response.user.fullName);
      
      _isLoading = false;
      notifyListeners();
      return true;
    } catch (e) {
      _isLoading = false;
      _error = e.toString();
      notifyListeners();
      return false;
    }
  }

  Future<void> logout() async {
    _isLoading = true;
    notifyListeners();

    try {
      await ApiService.logout();
      _currentUser = null;
      
      final prefs = await SharedPreferences.getInstance();
      await prefs.clear();
    } catch (e) {
      debugPrint('Logout error: $e');
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> loadUser() async {
    final prefs = await SharedPreferences.getInstance();
    final userId = prefs.getString('user_id');
    
    if (userId != null) {
      try {
        final user = await ApiService.getProfile();
        _currentUser = user;
      } catch (e) {
        debugPrint('Load user error: $e');
      }
    }
    notifyListeners();
  }

  bool hasPermission(String permission) {
    return _currentUser?.hasPermission(permission) ?? false;
  }
}