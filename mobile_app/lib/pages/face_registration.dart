// lib/pages/face_registration.dart
import 'dart:io';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:mobile_app/main_navigation.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:mobile_app/login.dart';

class FaceRegistrationPage extends StatefulWidget {
  final String userId;

  const FaceRegistrationPage({super.key, required this.userId});

  @override
  State<FaceRegistrationPage> createState() => _FaceRegistrationPageState();
}

class _FaceRegistrationPageState extends State<FaceRegistrationPage>
    with SingleTickerProviderStateMixin {
  File? _capturedImage;
  bool _isLoading = false;
  bool _isCameraPermissionGranted = false;
  bool _isFaceDetected = false;
  List<double>? _faceEmbedding;
  String? _errorMessage;
  // ✅ Simpan path foto untuk dikirim ke API
  String? _capturedImagePath;

  late AnimationController _animationController;
  late Animation<double> _pulseAnimation;

  final ImagePicker _imagePicker = ImagePicker();

  @override
  void initState() {
    super.initState();
    print('🔵 FaceRegistrationPage untuk user: ${widget.userId}');

    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 0.9, end: 1.1).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.easeInOut),
    );

    _checkCameraPermission();
  }

  Future<void> _checkCameraPermission() async {
    var status = await Permission.camera.status;

    if (status.isDenied) {
      status = await Permission.camera.request();
    }

    setState(() {
      _isCameraPermissionGranted = status.isGranted;
    });

    if (status.isPermanentlyDenied) {
      _showSettingsDialog();
    }
  }

  Future<void> _captureImage() async {
    if (!_isCameraPermissionGranted) {
      await _checkCameraPermission();
      if (!_isCameraPermissionGranted) {
        _showPermissionDialog();
        return;
      }
    }

    setState(() {
      _errorMessage = null;
      _isFaceDetected = false;
      _faceEmbedding = null;
      _capturedImagePath = null;
      _isLoading = true;
    });

    try {
      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.camera,
        maxWidth: 512,
        maxHeight: 512,
        imageQuality: 90,
        preferredCameraDevice: CameraDevice.front,
      );

      if (image != null) {
        setState(() {
          _capturedImage = File(image.path);
          _capturedImagePath = image.path; // ✅ Simpan path
        });

        try {
          print('📤 Mengirim foto untuk ekstraksi embedding...');
          final embedding = await ApiService.extractFaceEmbedding(image.path);

          setState(() {
            _faceEmbedding = embedding;
            _isFaceDetected = true;
            _isLoading = false;
          });

          print('✅ Embedding real diterima, panjang: ${embedding.length}');
        } catch (e) {
          setState(() {
            _isLoading = false;
            _errorMessage = e.toString();
          });

          String errorMsg = e.toString();
          String cleanErrorMsg = errorMsg;

          if (errorMsg.contains('"message":"')) {
            final RegExp regex = RegExp(r'"message":"([^"]+)"');
            final match = regex.firstMatch(errorMsg);
            if (match != null) {
              cleanErrorMsg = match.group(1) ?? errorMsg;
            }
          } else if (errorMsg.contains('message:')) {
            final parts = errorMsg.split('message:');
            if (parts.length > 1) {
              cleanErrorMsg = parts[1].trim();
            }
          }

          print('🧹 Clean error: $cleanErrorMsg');

          if (cleanErrorMsg.contains('tidak ada wajah') ||
              cleanErrorMsg.contains('no face') ||
              cleanErrorMsg.contains('Tidak ada wajah')) {
            _showNoFaceDialog();
          } else if (cleanErrorMsg.contains('Hanya satu wajah') ||
              cleanErrorMsg.contains('lebih dari 1 wajah') ||
              cleanErrorMsg.contains('multiple faces') ||
              cleanErrorMsg.contains('Terdeteksi') &&
                  RegExp(r'\d+\s*wajah').hasMatch(cleanErrorMsg)) {
            _showMultipleFacesDialog();
          } else if (cleanErrorMsg.contains('kacamata') ||
              cleanErrorMsg.contains('glasses') ||
              cleanErrorMsg.contains('Terdeteksi kacamata') ||
              cleanErrorMsg.contains('masker') ||
              cleanErrorMsg.contains('mask') ||
              cleanErrorMsg.contains('Terdeteksi masker') ||
              cleanErrorMsg.contains('topi') ||
              cleanErrorMsg.contains('hat') ||
              cleanErrorMsg.contains('Terdeteksi topi') ||
              cleanErrorMsg.contains('aksesoris kepala') ||
              cleanErrorMsg.contains('bingkai kacamata') ||
              cleanErrorMsg.contains('distorsi tekstur') ||
              cleanErrorMsg.contains('refleksi')) {
            _showAccessoryWarningDialog();
          } else {
            _showErrorSnackBar('Gagal: $cleanErrorMsg');
          }
        }
      } else {
        setState(() => _isLoading = false);
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString();
      });
      _showErrorSnackBar('Gagal mengakses kamera: $e');
    }
  }

  Future<void> _registerFace() async {
    if (_capturedImage == null || _capturedImagePath == null) {
      _showErrorSnackBar('Silakan ambil foto terlebih dahulu');
      return;
    }

    if (!_isFaceDetected || _faceEmbedding == null) {
      _showErrorSnackBar('Wajah tidak terdeteksi. Silakan coba lagi.');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      print('📸 Registering face for user: ${widget.userId}');
      print('📏 Embedding length: ${_faceEmbedding!.length}');

      // ✅ FIX: Gunakan registerFace() dengan photoPath saja
      // Backend Go akan re-extract embedding dari foto
      await ApiService.registerFace(
        userId: widget.userId,
        photoPath: _capturedImagePath!,
      );

      print('✅ Face registered successfully');

      if (!mounted) return;
      _showSuccessDialog();
    } catch (e) {
      print('❌ Error registering face: $e');
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString();
      });
      _showErrorSnackBar('Gagal registrasi wajah: $e');
    }
  }

  void _showSuccessDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: AppTheme.successColor.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.face_retouching_natural,
                color: AppTheme.successColor,
                size: 50,
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              'Registrasi Wajah Berhasil!',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppTheme.textPrimary,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              'Data wajah Anda telah tersimpan. Anda sekarang dapat menggunakan fitur absensi.',
              textAlign: TextAlign.center,
              style: TextStyle(fontSize: 14, color: AppTheme.textSecondary),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              Navigator.pushReplacement(
                context,
                MaterialPageRoute(
                  builder: (context) => const MainNavigationPage(),
                ),
              );
            },
            style: TextButton.styleFrom(foregroundColor: AppTheme.primaryColor),
            child: const Text('Lanjut ke Dashboard'),
          ),
        ],
      ),
    );
  }

  void _showPermissionDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Text('Izin Kamera Diperlukan'),
        content: const Text(
          'Aplikasi membutuhkan izin kamera untuk registrasi wajah.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              _checkCameraPermission();
            },
            child: const Text('Minta Izin'),
          ),
        ],
      ),
    );
  }

  void _showSettingsDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Text('Izin Kamera'),
        content: const Text(
          'Izin kamera telah ditolak permanen. Silakan aktifkan di pengaturan.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              openAppSettings();
            },
            child: const Text('Buka Pengaturan'),
          ),
        ],
      ),
    );
  }

  void _showErrorSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.error_outline, color: Colors.white, size: 20),
            const SizedBox(width: 8),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: AppTheme.errorColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        margin: const EdgeInsets.all(16),
        duration: const Duration(seconds: 4),
      ),
    );
  }

  void _showNoFaceDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Row(
          children: [
            Icon(
              Icons.face_retouching_off,
              color: AppTheme.errorColor,
              size: 28,
            ),
            SizedBox(width: 8),
            Expanded(
              child: Text(
                'Wajah Tidak Terdeteksi',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppTheme.errorColor,
                ),
              ),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: AppTheme.errorColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                Icons.face_retouching_off,
                color: AppTheme.errorColor,
                size: 60,
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Tidak ada wajah terdeteksi.\nArahkan kamera ke wajah Anda dengan benar.',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w500,
                color: AppTheme.textPrimary,
              ),
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppTheme.warningColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Text(
                'Pastikan wajah Anda terlihat jelas di tengah frame dengan pencahayaan yang cukup.',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 13, color: AppTheme.warningColor),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              setState(() {
                _capturedImage = null;
                _capturedImagePath = null;
                _isFaceDetected = false;
                _faceEmbedding = null;
                _errorMessage = null;
              });
            },
            style: TextButton.styleFrom(
              foregroundColor: AppTheme.primaryColor,
              minimumSize: const Size(double.infinity, 48),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'Ambil Foto Ulang',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
    );
  }

  void _showAccessoryWarningDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Row(
          children: [
            Icon(
              Icons.warning_amber_rounded,
              color: AppTheme.errorColor,
              size: 28,
            ),
            SizedBox(width: 8),
            Expanded(
              child: Text(
                'Aksesoris Terdeteksi',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppTheme.errorColor,
                ),
              ),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: AppTheme.errorColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                Icons.no_photography,
                color: AppTheme.errorColor,
                size: 60,
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Terdeteksi aksesoris (kacamata, masker, topi, dll).\nHarap lepas semua aksesoris Anda.',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w500,
                color: AppTheme.textPrimary,
              ),
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppTheme.warningColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Text(
                'Untuk keamanan dan keakuratan sistem, harap lepas semua aksesoris sebelum melanjutkan registrasi.',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 13, color: AppTheme.warningColor),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              setState(() {
                _capturedImage = null;
                _capturedImagePath = null;
                _isFaceDetected = false;
                _faceEmbedding = null;
                _errorMessage = null;
              });
            },
            style: TextButton.styleFrom(
              foregroundColor: AppTheme.primaryColor,
              minimumSize: const Size(double.infinity, 48),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'Ambil Foto Ulang',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
    );
  }

  void _showMultipleFacesDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Row(
          children: [
            Icon(Icons.group, color: AppTheme.errorColor, size: 28),
            SizedBox(width: 8),
            Expanded(
              child: Text(
                'Lebih dari Satu Wajah',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: AppTheme.errorColor,
                ),
              ),
            ),
          ],
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: AppTheme.errorColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: const Icon(
                Icons.group,
                color: AppTheme.errorColor,
                size: 60,
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Terdeteksi lebih dari satu wajah dalam frame.\nPastikan hanya Anda sendiri yang terlihat.',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w500,
                color: AppTheme.textPrimary,
              ),
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppTheme.warningColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Text(
                'Registrasi hanya untuk satu orang. Pastikan tidak ada orang lain di belakang atau di samping Anda.',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 13, color: AppTheme.warningColor),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              setState(() {
                _capturedImage = null;
                _capturedImagePath = null;
                _isFaceDetected = false;
                _faceEmbedding = null;
                _errorMessage = null;
              });
            },
            style: TextButton.styleFrom(
              foregroundColor: AppTheme.primaryColor,
              minimumSize: const Size(double.infinity, 48),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'Ambil Foto Ulang',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Color(0xFF0F172A)),
          onPressed: () {
            Navigator.pushAndRemoveUntil(
              context,
              MaterialPageRoute(
                builder: (context) => const EmployeeLoginPage(),
              ),
              (route) => false,
            );
          },
        ),
        title: const Text(
          'Registrasi Wajah',
          style: TextStyle(
            color: Color(0xFF0F172A),
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          physics: const BouncingScrollPhysics(),
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [Color(0xFF135BEC), Color(0xFF0F3B9E)],
                  ),
                  borderRadius: BorderRadius.circular(24),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF135BEC).withOpacity(0.3),
                      blurRadius: 20,
                      offset: const Offset(0, 5),
                    ),
                  ],
                ),
                child: Column(
                  children: [
                    const Icon(
                      Icons.face_retouching_natural,
                      color: Colors.white,
                      size: 50,
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      'Registrasi Wajah Pertama Kali',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      'Silakan registrasi wajah untuk menggunakan fitur absensi.',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.white70, fontSize: 14),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Camera Section
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.02),
                      blurRadius: 10,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Row(
                      children: [
                        Icon(
                          Icons.camera_alt,
                          color: Color(0xFF135BEC),
                          size: 20,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Ambil Foto Wajah',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF0F172A),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),

                    GestureDetector(
                      onTap: _isLoading ? null : _captureImage,
                      child: Container(
                        height: 250,
                        width: double.infinity,
                        decoration: BoxDecoration(
                          color: Colors.grey.shade100,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(
                            color: _capturedImage != null
                                ? (_isFaceDetected
                                      ? AppTheme.successColor
                                      : AppTheme.warningColor)
                                : Colors.grey.shade300,
                            width: 2,
                          ),
                        ),
                        child: _capturedImage != null
                            ? ClipRRect(
                                borderRadius: BorderRadius.circular(14),
                                child: Image.file(
                                  _capturedImage!,
                                  fit: BoxFit.cover,
                                ),
                              )
                            : Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  ScaleTransition(
                                    scale: _pulseAnimation,
                                    child: Container(
                                      width: 80,
                                      height: 80,
                                      decoration: BoxDecoration(
                                        color: const Color(
                                          0xFF135BEC,
                                        ).withOpacity(0.1),
                                        shape: BoxShape.circle,
                                      ),
                                      child: const Icon(
                                        Icons.camera_alt,
                                        color: Color(0xFF135BEC),
                                        size: 40,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(height: 16),
                                  const Text(
                                    'Tap untuk mengambil foto',
                                    style: TextStyle(
                                      fontSize: 14,
                                      color: Color(0xFF64748B),
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    'Pastikan: 1 wajah, tanpa aksesoris',
                                    style: TextStyle(
                                      fontSize: 11,
                                      color: Colors.grey.shade500,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ],
                              ),
                      ),
                    ),

                    const SizedBox(height: 16),

                    if (_capturedImage != null) ...[
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: _isFaceDetected
                              ? AppTheme.successColor.withOpacity(0.1)
                              : (_errorMessage != null
                                    ? AppTheme.errorColor.withOpacity(0.1)
                                    : AppTheme.warningColor.withOpacity(0.1)),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          children: [
                            Icon(
                              _isFaceDetected
                                  ? Icons.check_circle
                                  : (_errorMessage != null
                                        ? Icons.error
                                        : Icons.error_outline),
                              color: _isFaceDetected
                                  ? AppTheme.successColor
                                  : (_errorMessage != null
                                        ? AppTheme.errorColor
                                        : AppTheme.warningColor),
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    _isFaceDetected
                                        ? 'Wajah terdeteksi — siap didaftarkan'
                                        : (_errorMessage != null
                                              ? 'Gagal deteksi wajah'
                                              : 'Memproses...'),
                                    style: TextStyle(
                                      fontSize: 14,
                                      fontWeight: FontWeight.w600,
                                      color: _isFaceDetected
                                          ? AppTheme.successColor
                                          : (_errorMessage != null
                                                ? AppTheme.errorColor
                                                : AppTheme.warningColor),
                                    ),
                                  ),
                                  if (_isFaceDetected && _faceEmbedding != null)
                                    Text(
                                      // 'Embedding: ${_faceEmbedding!.length} dimensi',
                                      'Lanjutkan Simpan Data Wajah',
                                      style: TextStyle(
                                        fontSize: 11,
                                        color: Colors.grey.shade600,
                                      ),
                                    )
                                  else if (_errorMessage != null)
                                    Text(
                                      _errorMessage!,
                                      style: const TextStyle(
                                        fontSize: 11,
                                        color: AppTheme.errorColor,
                                      ),
                                      maxLines: 3,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                ],
                              ),
                            ),
                            // Tombol retake
                            if (!_isLoading)
                              IconButton(
                                icon: const Icon(
                                  Icons.refresh,
                                  color: Color(0xFF135BEC),
                                  size: 20,
                                ),
                                onPressed: _captureImage,
                                tooltip: 'Ambil ulang',
                              ),
                          ],
                        ),
                      ),
                    ],

                    if (_isLoading) ...[
                      const SizedBox(height: 16),
                      const Center(
                        child: Column(
                          children: [
                            CircularProgressIndicator(
                              valueColor: AlwaysStoppedAnimation<Color>(
                                Color(0xFF135BEC),
                              ),
                            ),
                            SizedBox(height: 8),
                            Text(
                              'Memproses wajah...',
                              style: TextStyle(
                                fontSize: 12,
                                color: Color(0xFF64748B),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // Guidelines
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFEBEE),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: AppTheme.errorColor),
                ),
                child: const Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(
                          Icons.warning_amber_rounded,
                          color: AppTheme.errorColor,
                          size: 20,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'SYARAT REGISTRASI WAJIB:',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: AppTheme.errorColor,
                          ),
                        ),
                      ],
                    ),
                    SizedBox(height: 8),
                    Text(
                      '✓ HANYA SATU ORANG dalam frame\n'
                      '✓ LEPAS KACAMATA (termasuk bening)\n'
                      '✓ LEPAS MASKER\n'
                      '✓ LEPAS TOPI/AKSESORIS KEPALA\n'
                      '✓ Wajah terlihat jelas, pencahayaan cukup\n'
                      '✓ Ekspresi normal (tidak tersenyum lebar)',
                      style: TextStyle(
                        fontSize: 12,
                        color: Color(0xFFB71C1C),
                        height: 1.5,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Register Button
              SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton.icon(
                  onPressed:
                      (_isFaceDetected && _faceEmbedding != null && !_isLoading)
                      ? _registerFace
                      : null,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF135BEC),
                    foregroundColor: Colors.white,
                    disabledBackgroundColor: Colors.grey.shade300,
                    disabledForegroundColor: Colors.grey.shade600,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: 4,
                  ),
                  icon: const Icon(Icons.save, size: 20),
                  label: const Text(
                    'Simpan Data Wajah',
                    style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
