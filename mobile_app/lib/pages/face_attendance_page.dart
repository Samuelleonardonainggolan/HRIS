import 'dart:async';
import 'dart:io';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:flutter/material.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:intl/intl.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:geolocator/geolocator.dart';
import 'package:image_picker/image_picker.dart';
import 'package:permission_handler/permission_handler.dart';

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
  bool _isFaceVerified = false;
  bool _attendanceSuccess = false;
  double _faceSimilarity = 0.0;
  String _userId = '';
  bool _isCameraPermissionGranted = false;
  bool _isLocationPermissionGranted = false;
  Map<String, dynamic>? _verificationResult;
  String? _verificationMessage;
  bool _isFaceDetected = false;
  List<double>? _faceEmbedding;
  String? _errorMessage;

  // ✅ Real-time clock
  late Timer _clockTimer;
  String _currentTime = '';
  String _currentDate = '';

  late AnimationController _animationController;
  late Animation<double> _pulseAnimation;

  final ImagePicker _imagePicker = ImagePicker();

  // Koordinat Labersa Hotel (sesuai backend)
  final double _officeLat = 2.3561;
  final double _officeLng = 99.1431;
  final double _radiusMeters = 10000;
  final double _similarityThreshold = 0.6;

  @override
  void initState() {
    super.initState();

    // Init locale untuk format tanggal Indonesia
    initializeDateFormatting('id', null);

    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 0.8, end: 1.2).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.easeInOut),
    );

    // ✅ Mulai timer real-time — update tiap detik
    _updateClock();
    _clockTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) _updateClock();
    });

    _loadUserId();
    _checkPermissions();
  }

  // ✅ Update jam dan tanggal secara real-time
  void _updateClock() {
    final now = DateTime.now();
    setState(() {
      _currentTime = DateFormat('HH:mm:ss').format(now);
      _currentDate = DateFormat('EEEE, dd MMMM yyyy', 'id').format(now);
    });
  }

  Future<void> _loadUserId() async {
    final userId = await ApiService.getUserId();
    setState(() {
      _userId = userId ?? '';
    });

    if (_userId.isEmpty) {
      _showErrorSnackBar('Sesi login telah berakhir. Silakan login ulang.');
      Future.delayed(const Duration(seconds: 2), () {
        if (mounted) Navigator.pushReplacementNamed(context, '/login');
      });
    } else {
      print('✅ User ID loaded: $_userId');
    }
  }

  Future<void> _checkPermissions() async {
    await _checkLocationPermission();
    await _checkCameraPermission();
  }

  // ─── Accessory Warning ────────────────────────────────────────────────────

  void _showAccessoryWarningDialog(String message) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        shape:
            RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Row(
          children: [
            Icon(Icons.warning_amber_rounded,
                color: AppTheme.errorColor, size: 28),
            SizedBox(width: 8),
            Expanded(
              child: Text(
                'Perhatian',
                style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppTheme.errorColor),
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
              child: const Icon(Icons.face_retouching_off,
                  color: AppTheme.errorColor, size: 60),
            ),
            const SizedBox(height: 20),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w500,
                  color: AppTheme.textPrimary),
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppTheme.warningColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Text(
                'Pastikan wajah terlihat jelas tanpa halangan untuk absensi yang akurat.',
                textAlign: TextAlign.center,
                style:
                    TextStyle(fontSize: 13, color: AppTheme.warningColor),
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
                _isFaceDetected = false;
                _faceEmbedding = null;
                _errorMessage = null;
                _isFaceVerified = false;
                _faceStatus = 'Menunggu pengambilan gambar';
              });
            },
            style: TextButton.styleFrom(
              foregroundColor: AppTheme.primaryColor,
              minimumSize: const Size(double.infinity, 48),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12)),
            ),
            child: const Text('Ambil Foto Ulang',
                style:
                    TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  // ─── Location ─────────────────────────────────────────────────────────────

  Future<void> _checkLocationPermission() async {
    setState(() => _locationStatus = 'Memeriksa izin lokasi...');

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
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

    setState(() => _isLocationPermissionGranted = true);
    _getCurrentLocation();
  }

  Future<void> _checkCameraPermission() async {
    var status = await Permission.camera.status;
    if (status.isDenied) status = await Permission.camera.request();
    setState(() => _isCameraPermissionGranted = status.isGranted);
    if (status.isPermanentlyDenied) _showSettingsDialog('Kamera');
  }

  Future<void> _getCurrentLocation() async {
    if (!_isLocationPermissionGranted) {
      setState(() => _locationStatus = 'Izin lokasi tidak diberikan');
      return;
    }

    setState(() => _locationStatus = 'Mendapatkan lokasi...');

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
        timeLimit: const Duration(seconds: 15),
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
      _locationStatus = _isLocationValid
          ? 'Dalam area kantor ✓'
          : 'Di luar area kantor (${distance.toStringAsFixed(0)}m) ✗';
    });
  }

  // ─── Camera & Face Verification ───────────────────────────────────────────

  Future<void> _captureImage() async {
    if (_attendanceSuccess) {
      _showInfoSnackBar('Anda sudah melakukan konfirmasi absensi');
      return;
    }

    if (!_isCameraPermissionGranted) {
      await _checkCameraPermission();
      if (!_isCameraPermissionGranted) {
        _showPermissionDialog('Kamera');
        return;
      }
    }

    if (_userId.isEmpty) {
      _showErrorSnackBar('Sesi login telah berakhir. Silakan login ulang.');
      if (mounted) Navigator.pushReplacementNamed(context, '/login');
      return;
    }

    try {
      final XFile? image = await _imagePicker.pickImage(
        source: ImageSource.camera,
        maxWidth: 512,
        maxHeight: 512,
        imageQuality: 85,
        preferredCameraDevice: CameraDevice.front,
      );

      if (image != null) {
        setState(() {
          _capturedImage = File(image.path);
          _faceStatus = 'Memverifikasi wajah...';
          _isLoading = true;
          _isFaceVerified = false;
          _verificationResult = null;
          _verificationMessage = null;
        });

        try {
          final result = await _verifyFaceOnly(image.path);

          setState(() {
            _verificationResult = result;
            _faceSimilarity =
                (result['similarity'] as num?)?.toDouble() ?? 0.0;
            _isFaceVerified = result['matched'] == true;
            _faceStatus = _isFaceVerified
                ? '✓ Wajah terverifikasi (${(_faceSimilarity * 100).toStringAsFixed(1)}%)'
                : '✗ ${result['message'] ?? 'Wajah tidak cocok'}';
            _isLoading = false;
          });

          if (!_isFaceVerified) {
            _showErrorSnackBar(result['message'] ?? 'Verifikasi gagal');
          }
        } catch (e) {
          setState(() {
            _isFaceVerified = false;
            _isLoading = false;
            _faceStatus = '✗ Gagal verifikasi wajah';
          });

          _handleVerificationError(e.toString());
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

  // ✅ Parsing error yang lebih akurat — rambut TIDAK ditandai sebagai topi
  void _handleVerificationError(String errorMsg) {
    // Ekstrak pesan bersih dari backend
    String clean = errorMsg;
    if (errorMsg.contains('"message":"')) {
      final m = RegExp(r'"message":"([^"]+)"').firstMatch(errorMsg);
      if (m != null) clean = m.group(1) ?? errorMsg;
    } else if (errorMsg.contains('message:')) {
      final parts = errorMsg.split('message:');
      if (parts.length > 1) clean = parts[1].trim();
    }

    print('🧹 Clean error: $clean');

    // Kacamata
    if (clean.contains('kacamata') ||
        clean.contains('glasses') ||
        clean.contains('bingkai kacamata') ||
        clean.contains('distorsi tekstur') ||
        clean.contains('refleksi') ||
        clean.contains('frame kacamata')) {
      _showAccessoryWarningDialog(
        'Terdeteksi kacamata.\nHarap lepas kacamata Anda (termasuk kacamata bening).',
      );
    }
    // Masker
    else if (clean.contains('masker') || clean.contains('mask')) {
      _showAccessoryWarningDialog(
        'Terdeteksi masker.\nHarap lepas masker untuk melanjutkan absensi.',
      );
    }
    // Topi/aksesoris kepala — HANYA jika bukan rambut biasa
    else if ((clean.contains('topi') ||
            clean.contains('hat') ||
            clean.contains('aksesoris kepala') ||
            clean.contains('tepi tajam di kepala')) &&
        !clean.contains('rambut')) {
      _showAccessoryWarningDialog(
        'Terdeteksi topi atau penutup kepala.\nHarap lepas topi/aksesoris kepala.',
      );
    }
    // Multiple faces
    else if (clean.contains('lebih dari 1 wajah') ||
        clean.contains('multiple faces') ||
        RegExp(r'\d wajah').hasMatch(clean)) {
      _showAccessoryWarningDialog(
        'Terdeteksi lebih dari satu wajah.\nPastikan hanya Anda sendiri yang terlihat dalam frame.',
      );
    }
    // Tidak ada wajah
    else if (clean.contains('tidak ada wajah') ||
        clean.contains('no face') ||
        clean.contains('Tidak ada wajah')) {
      _showErrorSnackBar(
          'Tidak ada wajah terdeteksi. Arahkan kamera ke wajah Anda.');
    }
    // Warna mencolok di kepala — bisa rambut berwarna, skip aja
    else if (clean.contains('warna mencolok')) {
      // Tidak tampilkan warning untuk warna rambut yang dianggap mencolok
      _showErrorSnackBar(
          'Gagal verifikasi wajah. Pastikan wajah terlihat jelas dan pencahayaan cukup.');
    }
    // Error lainnya
    else {
      _showErrorSnackBar('Gagal verifikasi: $clean');
    }
  }

  Future<Map<String, dynamic>> _verifyFaceOnly(String imagePath) async {
    final token = await ApiService.getAccessToken();
    if (token == null) throw Exception('Token tidak ditemukan');

    var request = http.MultipartRequest(
      'POST',
      Uri.parse('${ApiService.baseUrl}/attendance/process'),
    );

    request.headers.addAll({'Authorization': 'Bearer $token'});
    request.fields['record_type'] =
        widget.type == 'clock_in' ? 'clock_in' : 'clock_out';
    request.fields['latitude'] =
        (_currentPosition?.latitude ?? 0).toString();
    request.fields['longitude'] =
        (_currentPosition?.longitude ?? 0).toString();

    request.files.add(
      await http.MultipartFile.fromPath(
        'photo',
        imagePath,
        filename: 'verify_${DateTime.now().millisecondsSinceEpoch}.jpg',
      ),
    );

    print('📤 Verifikasi ke /attendance/process...');

    final streamedResponse = await request.send();
    final response = await http.Response.fromStream(streamedResponse);

    print('Status: ${response.statusCode}');
    print('Body: ${response.body}');

    if (response.statusCode == 200) {
      final json = jsonDecode(response.body);
      final data = json['data'] as Map<String, dynamic>? ?? json;

      return {
        'success': true,
        'matched': true,
        'similarity': (data['face_similarity'] as num?)?.toDouble() ?? 0.0,
        'message': data['message'] ?? json['message'] ?? '',
      };
    } else if (response.statusCode == 400) {
      final error = jsonDecode(response.body);
      final errorMsg = error['message']?.toString() ?? '';
      final similarity =
          (error['data']?['face_similarity'] as num?)?.toDouble() ?? 0.0;

      // Sudah clock-out
      if (widget.type == 'clock_out' &&
          errorMsg.contains('already clocked out')) {
        return {
          'success': true,
          'matched': true,
          'similarity': similarity > 0 ? similarity : 0.9,
          'message': 'Anda sudah melakukan clock out hari ini',
        };
      }
      // Belum clock-in
      if (widget.type == 'clock_out' &&
          errorMsg.contains('no clock in record')) {
        return {
          'success': false,
          'matched': false,
          'similarity': 0.0,
          'message': 'Anda belum melakukan clock in hari ini',
        };
      }
      // Sudah clock-in
      if (errorMsg.contains('already clocked in')) {
        return {
          'success': true,
          'matched': true,
          'similarity': similarity > 0 ? similarity : 0.9,
          'message': 'Wajah terverifikasi (sudah pernah absen)',
        };
      }
      // Face mismatch
      if (errorMsg.contains('face does not match') ||
          errorMsg.contains('tidak cocok') ||
          errorMsg.contains('Wajah tidak cocok')) {
        return {
          'success': false,
          'matched': false,
          'similarity': similarity,
          'message': errorMsg,
        };
      }
      // No face
      if (errorMsg.contains('no face') ||
          errorMsg.contains('tidak ada wajah')) {
        return {
          'success': false,
          'matched': false,
          'similarity': 0.0,
          'message': 'Tidak ada wajah terdeteksi',
        };
      }

      throw Exception(errorMsg);
    } else {
      final error = jsonDecode(response.body);
      throw Exception(error['message'] ?? 'Verifikasi gagal');
    }
  }

  // ─── Confirm Attendance ───────────────────────────────────────────────────

  Future<void> _confirmAttendance() async {
    if (_attendanceSuccess) {
      _showInfoSnackBar('Anda sudah melakukan konfirmasi absensi');
      return;
    }
    if (_capturedImage == null) {
      _showErrorSnackBar('Silakan ambil foto terlebih dahulu');
      return;
    }
    if (!_isLocationValid) {
      _showErrorSnackBar('Anda harus berada dalam area kantor');
      return;
    }
    if (!_isFaceVerified) {
      _showErrorSnackBar('Verifikasi wajah belum berhasil');
      return;
    }
    if (_userId.isEmpty) {
      _showErrorSnackBar('Sesi login telah berakhir');
      if (mounted) Navigator.pushReplacementNamed(context, '/login');
      return;
    }

    setState(() => _isLoading = true);

    try {
      final result = await ApiService.processAttendance(
        recordType: widget.type == 'clock_in' ? 'clock_in' : 'clock_out',
        latitude: _currentPosition!.latitude,
        longitude: _currentPosition!.longitude,
        photoPath: _capturedImage!.path,
      );

      if (result.success) {
        setState(() => _attendanceSuccess = true);
        _showSuccessDialog(
          title: widget.type == 'clock_in'
              ? 'Absen Masuk Berhasil!'
              : 'Absen Pulang Berhasil!',
          similarity: result.faceSimilarity,
        );
      } else {
        throw Exception(result.message);
      }
    } catch (e) {
      final msg = e.toString();
      if (msg.contains('401') || msg.contains('Unauthorized')) {
        _showErrorSnackBar('Sesi login telah berakhir. Silakan login ulang.');
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) Navigator.pushReplacementNamed(context, '/login');
        });
      } else if (msg.contains('already clocked in')) {
        _showInfoSnackBar('Anda sudah melakukan clock in hari ini');
        setState(() => _attendanceSuccess = true);
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) Navigator.pop(context, true);
        });
      } else if (msg.contains('no clock in record')) {
        _showErrorSnackBar('Anda belum melakukan clock in hari ini');
      } else if (msg.contains('already clocked out')) {
        _showInfoSnackBar('Anda sudah melakukan clock out hari ini');
        setState(() => _attendanceSuccess = true);
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) Navigator.pop(context, true);
        });
      } else {
        _showErrorSnackBar('Gagal melakukan absensi: $msg');
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ─── Dialogs & Snackbars ──────────────────────────────────────────────────

  void _showSuccessDialog({required String title, required double similarity}) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) {
        final subtitle = widget.type == 'clock_in'
            ? 'Selamat bekerja!'
            : 'Selamat beristirahat!';
        return AlertDialog(
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
          title: Column(
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: AppTheme.successColor.withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.check_circle,
                    color: AppTheme.successColor, size: 50),
              ),
              const SizedBox(height: 16),
              Text(title,
                  style: const TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      color: AppTheme.textPrimary)),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(subtitle,
                  style:
                      TextStyle(fontSize: 14, color: Colors.grey.shade600)),
              const SizedBox(height: 12),
              Container(
                padding: const EdgeInsets.symmetric(
                    horizontal: 16, vertical: 8),
                decoration: BoxDecoration(
                  color: AppTheme.primaryColor.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(50),
                ),
                child: Text(
                  'Similarity ${(similarity * 100).toStringAsFixed(1)}%',
                  style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppTheme.primaryColor),
                ),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.access_time,
                      size: 16, color: Colors.grey.shade600),
                  const SizedBox(width: 8),
                  // ✅ Tampilkan waktu saat ini (sudah real-time dari state)
                  Text(_currentTime,
                      style: const TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: AppTheme.textPrimary)),
                ],
              ),
              const SizedBox(height: 4),
              Text(_currentDate,
                  style:
                      TextStyle(fontSize: 12, color: Colors.grey.shade600)),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.pop(ctx);
                Navigator.pop(context, true);
              },
              style: TextButton.styleFrom(
                foregroundColor: AppTheme.primaryColor,
                minimumSize: const Size(double.infinity, 48),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12)),
              ),
              child: const Text('OK',
                  style: TextStyle(
                      fontSize: 16, fontWeight: FontWeight.bold)),
            ),
          ],
        );
      },
    );
  }

  void _showInfoSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(
      content: Row(children: [
        const Icon(Icons.info, color: Colors.white, size: 20),
        const SizedBox(width: 12),
        Expanded(child: Text(message)),
      ]),
      backgroundColor: const Color(0xFF135BEC),
      behavior: SnackBarBehavior.floating,
      shape:
          RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      margin: const EdgeInsets.all(16),
      duration: const Duration(seconds: 3),
    ));
  }

  void _showErrorSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(
      content: Text(message),
      backgroundColor: AppTheme.errorColor,
      behavior: SnackBarBehavior.floating,
      shape:
          RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      margin: const EdgeInsets.all(16),
      duration: const Duration(seconds: 3),
    ));
  }

  void _showPermissionDialog(String permission) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape:
            RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Izin $permission Diperlukan'),
        content: Text(
            'Aplikasi membutuhkan izin $permission untuk melanjutkan absensi.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Batal')),
          TextButton(
            onPressed: () {
              Navigator.pop(ctx);
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
              Navigator.pop(ctx);
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
      builder: (ctx) => AlertDialog(
        shape:
            RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Izin $permission'),
        content: Text(
            'Izin $permission telah ditolak permanen. Silakan aktifkan di pengaturan.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Batal')),
          TextButton(
            onPressed: () {
              Navigator.pop(ctx);
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
      builder: (ctx) => AlertDialog(
        shape:
            RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Text('Layanan Lokasi'),
        content:
            const Text('Harap aktifkan layanan lokasi untuk melanjutkan absensi.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Batal')),
          TextButton(
            onPressed: () {
              Navigator.pop(ctx);
              Geolocator.openLocationSettings();
            },
            child: const Text('Buka Pengaturan'),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _clockTimer.cancel(); // ✅ Hentikan timer saat widget dihancurkan
    _animationController.dispose();
    super.dispose();
  }

  // ─── Build ────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final bool canConfirm = _isLocationValid &&
        _isFaceVerified &&
        !_isLoading &&
        !_attendanceSuccess;

    final Color headerColor = widget.type == 'clock_in'
        ? AppTheme.successColor
        : AppTheme.errorColor;

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
              color: Color(0xFF0F172A), fontWeight: FontWeight.bold),
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
              // ── Header dengan jam real-time ───────────────────────────
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                    colors: [headerColor, headerColor.withOpacity(0.8)],
                  ),
                  borderRadius: BorderRadius.circular(24),
                  boxShadow: [
                    BoxShadow(
                      color: headerColor.withOpacity(0.3),
                      blurRadius: 20,
                      offset: const Offset(0, 5),
                    ),
                  ],
                ),
                child: Column(
                  children: [
                    Icon(
                      widget.type == 'clock_in'
                          ? Icons.login
                          : Icons.logout,
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
                          fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 4),
                    // ✅ Tanggal real-time
                    Text(
                      _currentDate,
                      style: TextStyle(
                          color: Colors.white.withOpacity(0.9),
                          fontSize: 13),
                    ),
                    const SizedBox(height: 6),
                    // ✅ Jam real-time berubah tiap detik
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 20, vertical: 8),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(30),
                      ),
                      child: Text(
                        _currentTime,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 28,
                          fontWeight: FontWeight.bold,
                          fontFamily: 'monospace',
                          letterSpacing: 2,
                        ),
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // ── Lokasi ────────────────────────────────────────────────
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [
                    BoxShadow(
                        color: Colors.black.withOpacity(0.02),
                        blurRadius: 10,
                        offset: const Offset(0, 2))
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Row(
                      children: [
                        Icon(Icons.location_on,
                            color: Color(0xFF135BEC), size: 20),
                        SizedBox(width: 8),
                        Text('Verifikasi Lokasi',
                            style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                color: Color(0xFF0F172A))),
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
                            _isLocationValid
                                ? Icons.check
                                : Icons.close,
                            color: _isLocationValid
                                ? AppTheme.successColor
                                : AppTheme.errorColor,
                            size: 20,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment:
                                CrossAxisAlignment.start,
                            children: [
                              Text(
                                _locationStatus,
                                style: TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                    color: _isLocationValid
                                        ? AppTheme.successColor
                                        : AppTheme.errorColor),
                              ),
                              if (_currentPosition != null) ...[
                                const SizedBox(height: 4),
                                Text(
                                  'Lat: ${_currentPosition!.latitude.toStringAsFixed(6)}',
                                  style: TextStyle(
                                      fontSize: 11,
                                      color: Colors.grey.shade600),
                                ),
                                Text(
                                  'Long: ${_currentPosition!.longitude.toStringAsFixed(6)}',
                                  style: TextStyle(
                                      fontSize: 11,
                                      color: Colors.grey.shade600),
                                ),
                              ],
                            ],
                          ),
                        ),
                        if (!_isLocationValid &&
                            _isLocationPermissionGranted)
                          IconButton(
                            icon: const Icon(Icons.refresh,
                                color: Color(0xFF135BEC)),
                            onPressed: _getCurrentLocation,
                          ),
                      ],
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // ── Face Recognition ──────────────────────────────────────
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [
                    BoxShadow(
                        color: Colors.black.withOpacity(0.02),
                        blurRadius: 10,
                        offset: const Offset(0, 2))
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Row(
                      children: [
                        Icon(Icons.face,
                            color: Color(0xFF135BEC), size: 20),
                        SizedBox(width: 8),
                        Text('Verifikasi Wajah',
                            style: TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                                color: Color(0xFF0F172A))),
                      ],
                    ),
                    const SizedBox(height: 16),

                    GestureDetector(
                      onTap: _isLoading ? null : _captureImage,
                      child: Container(
                        height: 200,
                        width: double.infinity,
                        decoration: BoxDecoration(
                          color: Colors.grey.shade100,
                          borderRadius: BorderRadius.circular(16),
                          border: Border.all(
                            color: _capturedImage != null
                                ? (_isFaceVerified
                                    ? AppTheme.successColor
                                    : AppTheme.errorColor)
                                : Colors.grey.shade300,
                            width: 2,
                          ),
                        ),
                        child: _capturedImage != null
                            ? ClipRRect(
                                borderRadius:
                                    BorderRadius.circular(14),
                                child: Image.file(_capturedImage!,
                                    fit: BoxFit.cover),
                              )
                            : Column(
                                mainAxisAlignment:
                                    MainAxisAlignment.center,
                                children: [
                                  ScaleTransition(
                                    scale: _pulseAnimation,
                                    child: Container(
                                      width: 70,
                                      height: 70,
                                      decoration: BoxDecoration(
                                        color: _isCameraPermissionGranted
                                            ? const Color(0xFF135BEC)
                                                .withOpacity(0.1)
                                            : Colors.grey
                                                .withOpacity(0.1),
                                        shape: BoxShape.circle,
                                      ),
                                      child: Icon(
                                        Icons.camera_alt,
                                        color:
                                            _isCameraPermissionGranted
                                                ? const Color(
                                                    0xFF135BEC)
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
                                        color:
                                            _isCameraPermissionGranted
                                                ? const Color(
                                                    0xFF64748B)
                                                : Colors.grey),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    'Pastikan wajah terlihat jelas',
                                    style: TextStyle(
                                        fontSize: 11,
                                        color: Colors.grey.shade500),
                                  ),
                                ],
                              ),
                      ),
                    ),

                    if (_capturedImage != null) ...[
                      const SizedBox(height: 16),
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: _isFaceVerified
                              ? AppTheme.successColor.withOpacity(0.1)
                              : AppTheme.errorColor.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          children: [
                            Icon(
                              _isFaceVerified
                                  ? Icons.check_circle
                                  : Icons.error,
                              color: _isFaceVerified
                                  ? AppTheme.successColor
                                  : AppTheme.errorColor,
                            ),
                            const SizedBox(width: 12),
                            Expanded(
                              child: Column(
                                crossAxisAlignment:
                                    CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    _faceStatus,
                                    style: TextStyle(
                                        fontSize: 14,
                                        fontWeight: FontWeight.w600,
                                        color: _isFaceVerified
                                            ? AppTheme.successColor
                                            : AppTheme.errorColor),
                                  ),
                                  if (_faceSimilarity > 0)
                                    Text(
                                      'Similarity: ${(_faceSimilarity * 100).toStringAsFixed(1)}%',
                                      style: TextStyle(
                                          fontSize: 12,
                                          color:
                                              Colors.grey.shade600),
                                    ),
                                ],
                              ),
                            ),
                            // Tombol retake
                            if (!_isLoading && !_attendanceSuccess)
                              IconButton(
                                icon: const Icon(Icons.refresh,
                                    color: Color(0xFF135BEC),
                                    size: 20),
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
                        child: CircularProgressIndicator(
                          valueColor: AlwaysStoppedAnimation<Color>(
                              Color(0xFF135BEC)),
                        ),
                      ),
                    ],
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // ── Info ──────────────────────────────────────────────────
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFFEFF6FF),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFFBFDBFE)),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.info_outline,
                        color: Color(0xFF135BEC), size: 20),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Petunjuk Absensi',
                            style: TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.bold,
                                color: Color(0xFF135BEC)),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            '1. Pastikan Anda dalam area kantor\n'
                            '2. Ambil foto selfie tanpa kacamata/masker\n'
                            '3. Klik Konfirmasi jika wajah terverifikasi',
                            style: TextStyle(
                                fontSize: 11,
                                color: Colors.grey.shade700,
                                height: 1.4),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // ── Tombol Konfirmasi ─────────────────────────────────────
              SizedBox(
                width: double.infinity,
                height: 56,
                child: ElevatedButton.icon(
                  onPressed: canConfirm ? _confirmAttendance : null,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: headerColor,
                    foregroundColor: Colors.white,
                    disabledBackgroundColor: Colors.grey.shade300,
                    shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16)),
                    elevation: 4,
                  ),
                  icon: Icon(
                      widget.type == 'clock_in'
                          ? Icons.login
                          : Icons.logout,
                      size: 20),
                  label: Text(
                    widget.type == 'clock_in'
                        ? (_attendanceSuccess
                            ? 'Sudah Absen Masuk'
                            : 'Konfirmasi Absen Masuk')
                        : (_attendanceSuccess
                            ? 'Sudah Absen Pulang'
                            : 'Konfirmasi Absen Pulang'),
                    style: const TextStyle(
                        fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                ),
              ),

              if (_attendanceSuccess) ...[
                const SizedBox(height: 12),
                Center(
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 16, vertical: 8),
                    decoration: BoxDecoration(
                      color: AppTheme.successColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(50),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.check_circle,
                            color: AppTheme.successColor, size: 16),
                        const SizedBox(width: 8),
                        Text(
                          'Absensi hari ini sudah tercatat',
                          style: TextStyle(
                              fontSize: 12,
                              color: AppTheme.successColor,
                              fontWeight: FontWeight.w600),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}