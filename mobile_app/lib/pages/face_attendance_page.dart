import 'dart:io';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:mobile_app/theme/app_theme.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:intl/intl.dart';
import 'package:geolocator/geolocator.dart';
import 'package:image_picker/image_picker.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:shared_preferences/shared_preferences.dart';

class FaceAttendancePage extends StatefulWidget {
  final String type; // 'clock_in' or 'clock_out'

  const FaceAttendancePage({super.key, required this.type});

  @override
  State<FaceAttendancePage> createState() => _FaceAttendancePageState();
}

class _FaceAttendancePageState extends State<FaceAttendancePage>
    with SingleTickerProviderStateMixin {
  File? _capturedImage;
  Position? _currentPosition;
  String _locationStatus = 'Mendeteksi lokasi...';
  String _faceStatus = 'Menunggu pengambilan gambar';
  bool _isLoading = false;
  bool _isLocationValid = false;
  bool _isFaceDetected = false;
  bool _isCameraPermissionGranted = false;
  bool _isLocationPermissionGranted = false;

  late AnimationController _animationController;
  late Animation<double> _pulseAnimation;

  final ImagePicker _imagePicker = ImagePicker();

  // Koordinat IT Del Sitoluama
  final double _officeLat = 2.3561;
  final double _officeLng = 99.1431;
  final double _radiusMeters = 10000; // Radius 200 meter

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 0.8, end: 1.2).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.easeInOut),
    );

    _checkPermissions();
  }

  Future<void> _checkPermissions() async {
    await _checkLocationPermission();
    await _checkCameraPermission();
  }

  Future<void> _checkLocationPermission() async {
    setState(() {
      _locationStatus = 'Memeriksa izin lokasi...';
    });

    // Cek status lokasi menggunakan Geolocator
    LocationPermission permission = await Geolocator.checkPermission();

    if (permission == LocationPermission.denied) {
      // Minta izin
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        setState(() {
          _locationStatus = 'Izin lokasi ditolak';
          _isLocationPermissionGranted = false;
        });
        _showPermissionDialog('Lokasi');
        return;
      }
    }

    if (permission == LocationPermission.deniedForever) {
      setState(() {
        _locationStatus = 'Izin lokasi ditolak permanen';
        _isLocationPermissionGranted = false;
      });
      _showSettingsDialog('Lokasi');
      return;
    }

    // Izin diberikan
    setState(() {
      _isLocationPermissionGranted = true;
    });

    _getCurrentLocation();
  }

  Future<void> _checkCameraPermission() async {
    var status = await Permission.camera.status;

    if (status.isDenied) {
      // Minta izin
      status = await Permission.camera.request();
    }

    setState(() {
      _isCameraPermissionGranted = status.isGranted;
    });

    if (status.isPermanentlyDenied) {
      _showSettingsDialog('Kamera');
    }
  }

  Future<void> _getCurrentLocation() async {
    if (!_isLocationPermissionGranted) {
      setState(() {
        _locationStatus = 'Izin lokasi tidak diberikan';
      });
      return;
    }

    setState(() {
      _locationStatus = 'Mendapatkan lokasi...';
    });

    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        setState(() {
          _locationStatus = 'Layanan lokasi tidak aktif';
          _isLocationValid = false;
        });
        _showLocationServiceDialog();
        return;
      }

      Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 10),
      );

      setState(() {
        _currentPosition = position;
        _checkOfficeRadius(position);
      });
    } catch (e) {
      setState(() {
        _locationStatus = 'Gagal mendapatkan lokasi';
        _isLocationValid = false;
      });
      _showErrorSnackBar('Gagal mendapatkan lokasi: $e');
    }
  }

  void _checkOfficeRadius(Position position) {
    double distance = Geolocator.distanceBetween(
      position.latitude,
      position.longitude,
      _officeLat,
      _officeLng,
    );

    setState(() {
      _isLocationValid = distance <= _radiusMeters;
      if (_isLocationValid) {
        _locationStatus = 'Dalam area kampus IT Del ✓';
      } else {
        _locationStatus =
            'Di luar area kampus (${distance.toStringAsFixed(0)}m) ✗';
      }
    });
  }

  Future<void> _captureImage() async {
    if (!_isCameraPermissionGranted) {
      await _checkCameraPermission();
      if (!_isCameraPermissionGranted) {
        _showPermissionDialog('Kamera');
        return;
      }
    }

    try {
      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.camera,
        maxWidth: 1024,
        maxHeight: 1024,
        imageQuality: 85,
        preferredCameraDevice: CameraDevice.front, // Gunakan kamera depan
      );

      if (image != null) {
        setState(() {
          _capturedImage = File(image.path);
          _faceStatus = 'Memverifikasi wajah...';
          _isLoading = true;
        });

        // Simulasi verifikasi wajah
        await Future.delayed(const Duration(seconds: 2));

        // Simulasi deteksi wajah (80% berhasil)
        bool faceDetected = DateTime.now().millisecondsSinceEpoch % 10 < 8;

        setState(() {
          _isFaceDetected = faceDetected;
          _faceStatus = faceDetected
              ? 'Wajah terdeteksi ✓'
              : 'Wajah tidak terdeteksi ✗';
          _isLoading = false;
        });

        if (!faceDetected) {
          _showErrorSnackBar(
            'Wajah tidak terdeteksi. Silakan coba lagi dengan pencahayaan cukup.',
          );
        }
      }
    } catch (e) {
      setState(() {
        _faceStatus = 'Gagal mengambil gambar';
        _isLoading = false;
      });
      _showErrorSnackBar('Gagal mengakses kamera: $e');
    }
  }

  Future<void> _submitAttendance() async {
    if (_capturedImage == null || _currentPosition == null) return;

    setState(() => _isLoading = true);

    try {
      final userId = await ApiService.getUserId();
      if (userId == null || userId.isEmpty) {
        throw Exception('Sesi login telah berakhir. Silakan login ulang.');
      }

      print('📤 Submitting attendance for user: $userId');
      print(
        '📍 Location: ${_currentPosition!.latitude}, ${_currentPosition!.longitude}',
      );
      print('📷 Photo path: ${_capturedImage!.path}');

      final result = await ApiService.processAttendance(
        recordType: widget.type == 'clock_in' ? 'clock_in' : 'clock_out',
        latitude: _currentPosition!.latitude,
        longitude: _currentPosition!.longitude,
        photoPath: _capturedImage!.path,
      );

      if (result.success) {
        _showSuccessDialog(
          message: result.message,
          similarity: result.faceSimilarity,
        );
      } else {
        throw Exception(result.message);
      }
    } catch (e) {
      print('❌ Error: $e');

      String errorMsg = e.toString();
      if (errorMsg.contains('401') ||
          errorMsg.contains('Unauthorized') ||
          errorMsg.contains('sesi telah berakhir')) {
        _showErrorSnackBar('Sesi login telah berakhir. Silakan login ulang.');

        // Redirect ke login setelah 2 detik
        Future.delayed(const Duration(seconds: 2), () {
          Navigator.pushReplacementNamed(context, '/login');
        });
      } else {
        _showErrorSnackBar('Gagal melakukan absensi: $errorMsg');
      }
    } finally {
      setState(() => _isLoading = false);
    }
  }

  void _showSuccessDialog({
    required String message,
    required double similarity,
  }) {
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
              child: Icon(
                Icons.check_circle,
                color: AppTheme.successColor,
                size: 50,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              widget.type == 'clock_in'
                  ? 'Absen Masuk Berhasil!'
                  : 'Absen Pulang Berhasil!',
              style: const TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppTheme.textPrimary,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Similarity Wajah: ${(similarity * 100).toStringAsFixed(1)}%',
              style: const TextStyle(
                fontSize: 14,
                color: AppTheme.primaryColor,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              DateFormat('EEEE, dd MMMM yyyy').format(DateTime.now()),
              style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
            ),
            const SizedBox(height: 4),
            Text(
              DateFormat('HH:mm:ss').format(DateTime.now()),
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppTheme.primaryColor,
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              Navigator.pop(context, true);
            },
            style: TextButton.styleFrom(foregroundColor: AppTheme.primaryColor),
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }

  void _showPermissionDialog(String permission) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Izin $permission Diperlukan'),
        content: Text(
          'Aplikasi membutuhkan izin $permission untuk melanjutkan absensi.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              if (permission == 'Kamera') {
                _checkCameraPermission();
              } else {
                _checkLocationPermission();
              }
            },
            child: const Text('Minta Izin'),
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

  void _showSettingsDialog(String permission) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Izin $permission'),
        content: Text(
          'Izin $permission telah ditolak permanen. Silakan aktifkan di pengaturan.',
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

  void _showLocationServiceDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Text('Layanan Lokasi'),
        content: const Text(
          'Harap aktifkan layanan lokasi untuk melanjutkan absensi.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              Geolocator.openLocationSettings();
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
        content: Text(message),
        backgroundColor: AppTheme.errorColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        margin: const EdgeInsets.all(16),
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
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          widget.type == 'clock_in' ? 'Absen Masuk' : 'Absen Pulang',
          style: const TextStyle(
            color: Color(0xFF0F172A),
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(1),
          child: Container(height: 1, color: Colors.grey.shade200),
        ),
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          physics: const BouncingScrollPhysics(),
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header Info
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: widget.type == 'clock_in'
                        ? [AppTheme.successColor, const Color(0xFF10B981)]
                        : [AppTheme.errorColor, const Color(0xFFEF4444)],
                  ),
                  borderRadius: BorderRadius.circular(24),
                  boxShadow: [
                    BoxShadow(
                      color:
                          (widget.type == 'clock_in'
                                  ? AppTheme.successColor
                                  : AppTheme.errorColor)
                              .withOpacity(0.3),
                      blurRadius: 20,
                      offset: const Offset(0, 5),
                    ),
                  ],
                ),
                child: Column(
                  children: [
                    Icon(
                      widget.type == 'clock_in' ? Icons.login : Icons.logout,
                      color: Colors.white,
                      size: 40,
                    ),
                    const SizedBox(height: 12),
                    Text(
                      widget.type == 'clock_in'
                          ? 'Absen Masuk'
                          : 'Absen Pulang',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      DateFormat('EEEE, dd MMMM yyyy').format(DateTime.now()),
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.9),
                        fontSize: 14,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      DateFormat('HH:mm:ss').format(DateTime.now()),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Lokasi Section
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
                          Icons.location_on,
                          color: Color(0xFF135BEC),
                          size: 20,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Verifikasi Lokasi',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF0F172A),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Container(
                          width: 40,
                          height: 40,
                          decoration: BoxDecoration(
                            color: _isLocationValid
                                ? AppTheme.successColor.withOpacity(0.1)
                                : AppTheme.errorColor.withOpacity(0.1),
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            _isLocationValid ? Icons.check : Icons.close,
                            color: _isLocationValid
                                ? AppTheme.successColor
                                : AppTheme.errorColor,
                            size: 20,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                _locationStatus,
                                style: TextStyle(
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600,
                                  color: _isLocationValid
                                      ? AppTheme.successColor
                                      : AppTheme.errorColor,
                                ),
                              ),
                              if (_currentPosition != null) ...[
                                const SizedBox(height: 4),
                                Text(
                                  'Lat: ${_currentPosition!.latitude.toStringAsFixed(6)}',
                                  style: TextStyle(
                                    fontSize: 11,
                                    color: Colors.grey.shade600,
                                  ),
                                ),
                                Text(
                                  'Long: ${_currentPosition!.longitude.toStringAsFixed(6)}',
                                  style: TextStyle(
                                    fontSize: 11,
                                    color: Colors.grey.shade600,
                                  ),
                                ),
                              ],
                            ],
                          ),
                        ),
                        if (!_isLocationValid && _isLocationPermissionGranted)
                          IconButton(
                            icon: const Icon(
                              Icons.refresh,
                              color: Color(0xFF135BEC),
                            ),
                            onPressed: _getCurrentLocation,
                          ),
                      ],
                    ),
                    if (!_isLocationPermissionGranted)
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: Text(
                          'Izin lokasi belum diberikan. Ketuk ikon kamera untuk meminta izin.',
                          style: TextStyle(
                            fontSize: 12,
                            color: AppTheme.errorColor,
                          ),
                        ),
                      ),
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // Face Recognition Section
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
                        Icon(Icons.face, color: Color(0xFF135BEC), size: 20),
                        SizedBox(width: 8),
                        Text(
                          'Verifikasi Wajah',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF0F172A),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),

                    // Camera Preview atau Hasil Foto
                    GestureDetector(
                      onTap: _captureImage,
                      child: Container(
                        height: 200,
                        width: double.infinity,
                        decoration: BoxDecoration(
                          color: Colors.grey.shade100,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(
                            color: _capturedImage != null
                                ? AppTheme.successColor
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
                                      width: 70,
                                      height: 70,
                                      decoration: BoxDecoration(
                                        color: _isCameraPermissionGranted
                                            ? const Color(
                                                0xFF135BEC,
                                              ).withOpacity(0.1)
                                            : Colors.grey.withOpacity(0.1),
                                        shape: BoxShape.circle,
                                      ),
                                      child: Icon(
                                        _isCameraPermissionGranted
                                            ? Icons.camera_alt
                                            : Icons.camera_alt_outlined,
                                        color: _isCameraPermissionGranted
                                            ? const Color(0xFF135BEC)
                                            : Colors.grey,
                                        size: 35,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                  Text(
                                    _isCameraPermissionGranted
                                        ? 'Tap untuk mengambil foto'
                                        : 'Izin kamera diperlukan',
                                    style: TextStyle(
                                      fontSize: 14,
                                      color: _isCameraPermissionGranted
                                          ? const Color(0xFF64748B)
                                          : Colors.grey,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    _isCameraPermissionGranted
                                        ? 'Pastikan wajah Anda terlihat jelas'
                                        : 'Ketuk untuk meminta izin kamera',
                                    style: TextStyle(
                                      fontSize: 11,
                                      color: Colors.grey.shade500,
                                    ),
                                  ),
                                ],
                              ),
                      ),
                    ),

                    const SizedBox(height: 16),

                    // Status Verifikasi
                    if (_capturedImage != null) ...[
                      Row(
                        children: [
                          Container(
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                              color: _isFaceDetected
                                  ? AppTheme.successColor.withOpacity(0.1)
                                  : AppTheme.errorColor.withOpacity(0.1),
                              shape: BoxShape.circle,
                            ),
                            child: Icon(
                              _isFaceDetected ? Icons.check : Icons.close,
                              color: _isFaceDetected
                                  ? AppTheme.successColor
                                  : AppTheme.errorColor,
                              size: 16,
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  _faceStatus,
                                  style: TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                    color: _isFaceDetected
                                        ? AppTheme.successColor
                                        : AppTheme.errorColor,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  _isFaceDetected
                                      ? 'Wajah terverifikasi'
                                      : 'Silakan ambil foto ulang dengan pencahayaan cukup',
                                  style: TextStyle(
                                    fontSize: 11,
                                    color: Colors.grey.shade600,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ],

                    if (_isLoading) ...[
                      const SizedBox(height: 16),
                      const Center(
                        child: CircularProgressIndicator(
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Color(0xFF135BEC),
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // Info Tambahan
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFFEFF6FF),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFFBFDBFE)),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.info_outline,
                      color: Color(0xFF135BEC),
                      size: 20,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Petunjuk Absensi IT Del',
                            style: TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: Color(0xFF135BEC),
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            '1. Pastikan Anda dalam area kampus IT Del Sitoluama\n'
                            '2. Ambil foto dengan wajah jelas (gunakan kamera depan)\n'
                            '3. Pastikan pencahayaan cukup\n'
                            '4. Jangan gunakan aksesori yang menutupi wajah',
                            style: TextStyle(
                              fontSize: 11,
                              color: Colors.grey.shade700,
                              height: 1.4,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // Submit Button
              SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton.icon(
                  onPressed: _submitAttendance,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: widget.type == 'clock_in'
                        ? AppTheme.successColor
                        : AppTheme.errorColor,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    elevation: 4,
                  ),
                  icon: Icon(
                    widget.type == 'clock_in' ? Icons.login : Icons.logout,
                    size: 20,
                  ),
                  label: Text(
                    widget.type == 'clock_in'
                        ? 'Konfirmasi Absen Masuk'
                        : 'Konfirmasi Absen Pulang',
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),

              const SizedBox(height: 16),

              // Informasi tambahan
              Center(
                child: Text(
                  'Dengan melakukan absensi, Anda menyetujui kebijakan kampus IT Del',
                  style: TextStyle(fontSize: 10, color: Colors.grey.shade500),
                  textAlign: TextAlign.center,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
