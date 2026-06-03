import 'dart:async';
import 'dart:io';
import 'dart:convert';
import 'package:camera/camera.dart';
import 'package:http/http.dart' as http;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_mlkit_face_detection/google_mlkit_face_detection.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:mobile_app/services/api_service.dart';
import 'dart:math';
import 'package:intl/intl.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:geolocator/geolocator.dart';
import 'package:permission_handler/permission_handler.dart';

enum LivenessChallenge {
  lookLeft,
  lookRight,
  smile,
  blink,
}

class FaceAttendancePage extends StatefulWidget {
  final String type; // 'clock_in' or 'clock_out'

  const FaceAttendancePage({super.key, required this.type});

  @override
  State<FaceAttendancePage> createState() => _FaceAttendancePageState();
}

class _FaceAttendancePageState extends State<FaceAttendancePage>
    with SingleTickerProviderStateMixin {
  // ─── Location ───────────────────────────────────────────────────────────────
  Position? _currentPosition;
  String _locationStatus = 'Mendeteksi lokasi...';
  bool _isLocationValid = false;
  bool _isLocationPermissionGranted = false;

  // ─── Status ──────────────────────────────────────────────────────────────────
  String _faceStatus = 'Arahkan wajah ke kamera';
  bool _isLoading = false;
  bool _isFaceVerified = false;
  bool _attendanceSuccess = false;
  double _faceSimilarity = 0.0;
  String _userId = '';
  bool _isCameraPermissionGranted = false;
  Map<String, dynamic>? _verificationResult;
  File? _capturedImage;

  // ─── Liveness inline ────────────────────────────────────────────────────────
  CameraController? _cameraController;
  late FaceDetector _faceDetector;
  bool _cameraInitialized = false;
  bool _isLivenessStarted = false;
  bool _isBusy = false;
  int _lastProcessTime = 0;

  // 0=center, 1=left, 2=right, 3=done
  int _livenessStep = 0;
  bool _livenessCompleted = false;
  String _livenessData = 'false';
  List<List<double>> _boxesSeq = [];

  // ─── Clock ───────────────────────────────────────────────────────────────────
  late Timer _clockTimer;
  String _currentTime = '';
  String _currentDate = '';

  // ─── Animation ───────────────────────────────────────────────────────────────
  late AnimationController _pulseCtrl;
  late Animation<double> _pulseAnim;

  List<LivenessChallenge> _activeChallenges = [];

  String get _currentStepLabel {
    if (_livenessStep >= _activeChallenges.length) return 'Selesai!';
    switch (_activeChallenges[_livenessStep]) {
      case LivenessChallenge.lookLeft:
        return 'Tengok ke Kiri';
      case LivenessChallenge.lookRight:
        return 'Tengok ke Kanan';
      case LivenessChallenge.smile:
        return 'Tersenyum Lebar';
      case LivenessChallenge.blink:
        return 'Berkedip / Tutup Mata';
    }
  }

  IconData get _currentStepIcon {
    if (_livenessStep >= _activeChallenges.length) return Icons.check;
    switch (_activeChallenges[_livenessStep]) {
      case LivenessChallenge.lookLeft:
        return Icons.arrow_back;
      case LivenessChallenge.lookRight:
        return Icons.arrow_forward;
      case LivenessChallenge.smile:
        return Icons.sentiment_very_satisfied;
      case LivenessChallenge.blink:
        return Icons.visibility_off;
    }
  }

  // ────────────────────────────────────────────────────────────────────────────
  @override
  void initState() {
    super.initState();
    initializeDateFormatting('id', null);

    _pulseCtrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);
    _pulseAnim = Tween<double>(begin: 0.95, end: 1.05).animate(
      CurvedAnimation(parent: _pulseCtrl, curve: Curves.easeInOut),
    );

    _updateClock();
    _clockTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) _updateClock();
    });

    _faceDetector = FaceDetector(
      options: FaceDetectorOptions(
        enableTracking: true,
        enableClassification: true, // For smile and blink detection
        performanceMode: FaceDetectorMode.fast,
      ),
    );

    _loadUserId();
    _checkPermissions();
  }

  @override
  void dispose() {
    _clockTimer.cancel();
    _pulseCtrl.dispose();
    _cameraController?.dispose();
    _faceDetector.close();
    super.dispose();
  }

  // ─── Clock ───────────────────────────────────────────────────────────────────
  void _updateClock() {
    final now = DateTime.now();
    setState(() {
      _currentTime = DateFormat('HH:mm:ss').format(now);
      _currentDate = DateFormat('EEEE, dd MMMM yyyy', 'id').format(now);
    });
  }

  // ─── Permissions ─────────────────────────────────────────────────────────────
  Future<void> _loadUserId() async {
    final userId = await ApiService.getUserId();
    setState(() => _userId = userId ?? '');
    if (_userId.isEmpty) {
      _showErrorSnackBar('Sesi login telah berakhir. Silakan login ulang.');
      Future.delayed(const Duration(seconds: 2), () {
        if (mounted) Navigator.pushReplacementNamed(context, '/login');
      });
    }
  }

  Future<void> _checkPermissions() async {
    await _checkLocationPermission();
    await _checkCameraPermission();
  }

  Future<void> _checkLocationPermission() async {
    setState(() => _locationStatus = 'Memeriksa izin lokasi...');
    LocationPermission perm = await Geolocator.checkPermission();
    if (perm == LocationPermission.denied) {
      perm = await Geolocator.requestPermission();
      if (perm == LocationPermission.denied) {
        setState(() {
          _locationStatus = 'Izin lokasi ditolak';
          _isLocationPermissionGranted = false;
        });
        return;
      }
    }
    if (perm == LocationPermission.deniedForever) {
      setState(() {
        _locationStatus = 'Izin lokasi ditolak permanen';
        _isLocationPermissionGranted = false;
      });
      return;
    }
    setState(() => _isLocationPermissionGranted = true);
    _getCurrentLocation();
  }

  Future<void> _checkCameraPermission() async {
    var status = await Permission.camera.status;
    if (status.isDenied) status = await Permission.camera.request();
    setState(() => _isCameraPermissionGranted = status.isGranted);
    if (status.isGranted) _initCamera();
  }

  // ─── Camera / Liveness ───────────────────────────────────────────────────────
  Future<void> _initCamera() async {
    try {
      final cameras = await availableCameras();
      final front = cameras.firstWhere(
        (c) => c.lensDirection == CameraLensDirection.front,
        orElse: () => cameras.first,
      );
      _cameraController = CameraController(
        front,
        ResolutionPreset.medium,
        enableAudio: false,
        imageFormatGroup: Platform.isAndroid
            ? ImageFormatGroup.nv21
            : ImageFormatGroup.bgra8888,
      );
      await _cameraController!.initialize();
      if (!mounted) return;
      setState(() => _cameraInitialized = true);
      // Removed auto-start of image stream; it's now started manually.
    } catch (e) {
      debugPrint('Camera init error: $e');
    }
  }

  Future<void> _processFrame(CameraImage image) async {
    if (_isBusy || _livenessCompleted || !mounted) return;
    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - _lastProcessTime < 300) return;
    _lastProcessTime = now;
    _isBusy = true;

    try {
      final rotation = InputImageRotationValue.fromRawValue(
          _cameraController!.description.sensorOrientation);
      if (rotation == null) return;
      final format = InputImageFormatValue.fromRawValue(image.format.raw);
      if (format == null) return;

      final buf = WriteBuffer();
      for (final p in image.planes) buf.putUint8List(p.bytes);

      final inputImage = InputImage.fromBytes(
        bytes: buf.done().buffer.asUint8List(),
        metadata: InputImageMetadata(
          size: Size(image.width.toDouble(), image.height.toDouble()),
          rotation: rotation,
          format: format,
          bytesPerRow: image.planes[0].bytesPerRow,
        ),
      );

      final faces = await _faceDetector.processImage(inputImage);
      if (!mounted) return;

      if (faces.isNotEmpty) {
        final face = faces.first;
        final yaw = face.headEulerAngleY ?? 0.0;
        final rect = face.boundingBox;
        _boxesSeq.add([rect.left, rect.top, rect.right, rect.bottom]);
        if (_boxesSeq.length > 30) _boxesSeq.removeAt(0);

        if (_livenessStep < 3 && _activeChallenges.isNotEmpty) {
          final challenge = _activeChallenges[_livenessStep];
          bool passed = false;

          switch (challenge) {
            case LivenessChallenge.lookLeft:
              if (yaw > 20) passed = true;
              break;
            case LivenessChallenge.lookRight:
              if (yaw < -20) passed = true;
              break;
            case LivenessChallenge.smile:
              final smileProb = face.smilingProbability ?? 0.0;
              if (smileProb > 0.65) passed = true;
              break;
            case LivenessChallenge.blink:
              final leftEye = face.leftEyeOpenProbability ?? 1.0;
              final rightEye = face.rightEyeOpenProbability ?? 1.0;
              // Detect blink if both eyes are closed (probability < 0.2)
              if (leftEye < 0.2 && rightEye < 0.2) passed = true;
              break;
          }

          if (passed) {
            if (_livenessStep == 2) {
              _onLivenessComplete();
            } else {
              setState(() => _livenessStep++);
            }
          }
        }
      }
    } catch (e) {
      debugPrint('Frame err: $e');
    } finally {
      _isBusy = false;
    }
  }

  Future<void> _onLivenessComplete() async {
    if (_livenessCompleted) return;
    setState(() {
      _livenessCompleted = true;
      _livenessStep = 3;
      _faceStatus = 'Mengambil foto...';
    });

    try {
      await _cameraController!.stopImageStream();
      await Future.delayed(const Duration(milliseconds: 200));
      final XFile photo = await _cameraController!.takePicture();
      _livenessData = jsonEncode(_boxesSeq);

      if (!mounted) return;
      setState(() {
        _capturedImage = File(photo.path);
        _faceStatus = 'Memverifikasi wajah...';
        _isLoading = true;
      });

      await _runVerification(photo.path);
    } catch (e) {
      debugPrint('Capture err: $e');
      if (mounted) {
        setState(() {
          _faceStatus = 'Gagal mengambil foto';
          _livenessCompleted = false;
          _livenessStep = 0;
          _isLoading = false;
        });
        _cameraController?.startImageStream(_processFrame);
      }
    }
  }

  void _resetLiveness() {
    _generateRandomChallenges();
    setState(() {
      _capturedImage = null;
      _livenessCompleted = false;
      _isLivenessStarted = false; // Require manual start again
      _livenessStep = 0;
      _livenessData = 'false';
      _boxesSeq = [];
      _isFaceVerified = false;
      _faceStatus = 'Arahkan wajah ke kamera';
      _isLoading = false;
    });
  }

  void _generateRandomChallenges() {
    final random = Random();
    final allChallenges = [
      LivenessChallenge.lookLeft,
      LivenessChallenge.lookRight,
      LivenessChallenge.smile,
      LivenessChallenge.blink,
    ];
    allChallenges.shuffle(random);
    _activeChallenges = allChallenges.take(3).toList();
  }

  // ─── Location ───────────────────────────────────────────────────────────────
  Future<void> _getCurrentLocation() async {
    if (!_isLocationPermissionGranted) return;
    setState(() => _locationStatus = 'Mendapatkan lokasi...');
    try {
      bool enabled = await Geolocator.isLocationServiceEnabled();
      if (!enabled) {
        setState(() {
          _locationStatus = 'Layanan lokasi tidak aktif';
          _isLocationValid = false;
        });
        return;
      }
      Position pos = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 15),
      );
      setState(() {
        _currentPosition = pos;
        _locationStatus = 'Memvalidasi geofence...';
      });
      await _checkGeofence(pos);
    } catch (e) {
      setState(() {
        _locationStatus = 'Gagal mendapatkan lokasi';
        _isLocationValid = false;
      });
    }
  }

  Future<void> _checkGeofence(Position pos) async {
    try {
      final result = await ApiService.checkUserInGeofence(
        latitude: pos.latitude,
        longitude: pos.longitude,
      );
      if (!mounted) return;
      setState(() {
        _isLocationValid = result.isValid;
        if (result.isValid) {
          final name = (result.geofenceName ?? '').trim();
          _locationStatus = name.isNotEmpty
              ? 'Dalam area $name (${result.distanceM.toStringAsFixed(0)}m) ✓'
              : 'Dalam area geofence ✓';
        } else {
          _locationStatus =
              result.message.isNotEmpty ? result.message : 'Di luar area geofence';
        }
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLocationValid = false;
        _locationStatus = 'Gagal validasi geofence';
      });
    }
  }

  // ─── Face Verification ────────────────────────────────────────────────────────
  Future<void> _runVerification(String imagePath) async {
    try {
      final token = await ApiService.getAccessToken();
      if (token == null) throw Exception('Token tidak ditemukan');

      var req = http.MultipartRequest(
        'POST',
        Uri.parse('${ApiService.baseUrl}/attendance/process'),
      );
      req.headers.addAll({'Authorization': 'Bearer $token'});
      req.fields['record_type'] =
          widget.type == 'clock_in' ? 'clock_in' : 'clock_out';
      req.fields['verify_only'] = 'true';
      req.fields['latitude'] =
          (_currentPosition?.latitude ?? 0).toString();
      req.fields['longitude'] =
          (_currentPosition?.longitude ?? 0).toString();
      req.fields['liveness'] = _livenessData;
      req.files.add(await http.MultipartFile.fromPath(
        'photo',
        imagePath,
        filename: 'verify_${DateTime.now().millisecondsSinceEpoch}.jpg',
      ));

      final streamed = await req.send();
      final resp = await http.Response.fromStream(streamed);

      if (resp.statusCode == 200) {
        final json = jsonDecode(resp.body);
        final data = (json['data'] as Map<String, dynamic>?) ?? json;
        final similarity = (data['face_similarity'] as num?)?.toDouble() ?? 0.0;
        setState(() {
          _isFaceVerified = true;
          _faceSimilarity = similarity;
          _faceStatus = '✓ Wajah terverifikasi';
          _isLoading = false;
        });
      } else {
        final err = jsonDecode(resp.body);
        final msg = err['message']?.toString() ?? 'Verifikasi gagal';
        throw Exception(msg);
      }
    } catch (e) {
      String clean = e.toString().replaceFirst('Exception: ', '');
      if (!mounted) return;
      setState(() {
        _isFaceVerified = false;
        _faceStatus = '✗ $clean';
        _isLoading = false;
      });
      _showErrorSnackBar(clean);
      // Let user retry
      await Future.delayed(const Duration(seconds: 2));
      if (mounted) _resetLiveness();
    }
  }

  // ─── Confirm Attendance ──────────────────────────────────────────────────────
  Future<void> _confirmAttendance() async {
    if (_attendanceSuccess || _capturedImage == null || !_isFaceVerified) return;
    setState(() => _isLoading = true);
    try {
      final result = await ApiService.processAttendance(
        recordType: widget.type == 'clock_in' ? 'clock_in' : 'clock_out',
        latitude: _currentPosition!.latitude,
        longitude: _currentPosition!.longitude,
        photoPath: _capturedImage!.path,
        liveness: _livenessData,
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
        _showErrorSnackBar(result.message);
      }
    } catch (e) {
      _showErrorSnackBar('Gagal melakukan absensi: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ─── Dialogs / Snackbars ─────────────────────────────────────────────────────
  void _showSuccessDialog({required String title, required double similarity}) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
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
            Text(
              widget.type == 'clock_in' ? 'Selamat bekerja!' : 'Selamat beristirahat!',
              style: TextStyle(fontSize: 14, color: Colors.grey.shade600),
            ),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.access_time, size: 16, color: Colors.grey.shade600),
                const SizedBox(width: 8),
                Text(_currentTime,
                    style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: AppTheme.textPrimary)),
              ],
            ),
            const SizedBox(height: 4),
            Text(_currentDate,
                style: TextStyle(fontSize: 12, color: Colors.grey.shade600)),
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
              shape:
                  RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            ),
            child: const Text('OK',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          ),
        ],
      ),
    );
  }

  void _showErrorSnackBar(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(
      content: Text(message),
      backgroundColor: AppTheme.errorColor,
      behavior: SnackBarBehavior.floating,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      margin: const EdgeInsets.all(16),
      duration: const Duration(seconds: 3),
    ));
  }

  void _showInfoSnackBar(String message) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(
      content: Row(children: [
        const Icon(Icons.info, color: Colors.white, size: 20),
        const SizedBox(width: 12),
        Expanded(child: Text(message)),
      ]),
      backgroundColor: const Color(0xFF135BEC),
      behavior: SnackBarBehavior.floating,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      margin: const EdgeInsets.all(16),
    ));
  }

  void _showSettingsDialog(String perm) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Izin $perm'),
        content: Text('Izin $perm ditolak permanen. Aktifkan di Pengaturan.'),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx), child: const Text('Batal')),
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

  // ─── Build ────────────────────────────────────────────────────────────────────
  @override
  Widget build(BuildContext context) {
    final headerColor = widget.type == 'clock_in'
        ? AppTheme.successColor
        : AppTheme.errorColor;
    final canConfirm =
        _isLocationValid && _isFaceVerified && !_isLoading && !_attendanceSuccess;

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
              // ── Header / Clock ──────────────────────────────────────────────
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
                        offset: const Offset(0, 5)),
                  ],
                ),
                child: Column(
                  children: [
                    Icon(
                        widget.type == 'clock_in' ? Icons.login : Icons.logout,
                        color: Colors.white,
                        size: 40),
                    const SizedBox(height: 12),
                    Text(
                      widget.type == 'clock_in' ? 'Absen Masuk' : 'Absen Pulang',
                      style: const TextStyle(
                          color: Colors.white,
                          fontSize: 20,
                          fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 4),
                    Text(_currentDate,
                        style: TextStyle(
                            color: Colors.white.withOpacity(0.9), fontSize: 13)),
                    const SizedBox(height: 6),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 20, vertical: 8),
                      decoration: BoxDecoration(
                          color: Colors.white.withOpacity(0.2),
                          borderRadius: BorderRadius.circular(30)),
                      child: Text(
                        _currentTime,
                        style: const TextStyle(
                            color: Colors.white,
                            fontSize: 28,
                            fontWeight: FontWeight.bold,
                            fontFamily: 'monospace',
                            letterSpacing: 2),
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // ── Location ────────────────────────────────────────────────────
              _buildLocationCard(),

              const SizedBox(height: 16),

              // ── Inline Camera / Face Verification ──────────────────────────
              _buildCameraCard(),

              const SizedBox(height: 16),

              // ── Info ────────────────────────────────────────────────────────
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
                      child: Text(
                        _isFaceVerified
                            ? '✅ Verifikasi wajah berhasil!\nKlik tombol di bawah untuk konfirmasi absensi.'
                            : '1. Pastikan Anda dalam area geofence\n2. Ikuti instruksi liveness (lurus, kiri, kanan)\n3. Tanpa kacamata/masker',
                        style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade700,
                            height: 1.4),
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 24),

              // ── Confirm Button ───────────────────────────────────────────────
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
                      widget.type == 'clock_in' ? Icons.login : Icons.logout,
                      size: 20),
                  label: Text(
                    _attendanceSuccess
                        ? '✅ Absensi Berhasil'
                        : (_isFaceVerified
                            ? 'Konfirmasi Absensi'
                            : 'Tunggu Verifikasi Wajah'),
                    style: const TextStyle(
                        fontSize: 14, fontWeight: FontWeight.bold),
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
                      border: Border.all(
                          color: AppTheme.successColor.withOpacity(0.3)),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.check_circle,
                            color: AppTheme.successColor, size: 16),
                        SizedBox(width: 8),
                        Text('Data tersimpan di database',
                            style: TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.w600,
                                color: AppTheme.successColor)),
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

  // ─── Location Card ────────────────────────────────────────────────────────────
  Widget _buildLocationCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 10,
              offset: const Offset(0, 2))
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(Icons.location_on, color: Color(0xFF135BEC), size: 20),
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
                    _isLocationValid ? Icons.check : Icons.close,
                    color: _isLocationValid
                        ? AppTheme.successColor
                        : AppTheme.errorColor,
                    size: 20),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  _locationStatus,
                  style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: _isLocationValid
                          ? AppTheme.successColor
                          : AppTheme.errorColor),
                ),
              ),
              if (!_isLocationValid && _isLocationPermissionGranted)
                IconButton(
                  icon: const Icon(Icons.refresh, color: Color(0xFF135BEC)),
                  onPressed: _getCurrentLocation,
                ),
            ],
          ),
        ],
      ),
    );
  }

  // ─── Camera Card (Inline Liveness) ───────────────────────────────────────────
  Widget _buildCameraCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 10,
              offset: const Offset(0, 2))
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.face, color: Color(0xFF135BEC), size: 20),
              const SizedBox(width: 8),
              const Expanded(
                child: Text('Verifikasi Wajah',
                    style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A))),
              ),
              if (_capturedImage != null && !_attendanceSuccess)
                GestureDetector(
                  onTap: _resetLiveness,
                  child: Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                    decoration: BoxDecoration(
                      color: const Color(0xFF135BEC).withOpacity(0.1),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.refresh,
                            color: Color(0xFF135BEC), size: 14),
                        SizedBox(width: 4),
                        Text('Ulang',
                            style: TextStyle(
                                color: Color(0xFF135BEC),
                                fontSize: 12,
                                fontWeight: FontWeight.w600)),
                      ],
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 16),

          // ── Camera / Result preview box ──────────────────────────────────
          ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 400),
              height: 280,
              width: double.infinity,
              decoration: BoxDecoration(
                color: Colors.black,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(
                  color: _isFaceVerified
                      ? AppTheme.successColor
                      : (_capturedImage != null
                          ? AppTheme.errorColor
                          : const Color(0xFF135BEC).withOpacity(0.3)),
                  width: 2,
                ),
              ),
              child: _buildCameraContent(),
            ),
          ),

          const SizedBox(height: 12),

          // ── Status bar below camera ───────────────────────────────────────
          _buildStatusBar(),
        ],
      ),
    );
  }

  Widget _buildCameraContent() {
    // Show captured image after liveness done
    if (_capturedImage != null) {
      return Stack(
        fit: StackFit.expand,
        children: [
          Image.file(_capturedImage!, fit: BoxFit.cover),
          if (_isLoading)
            Container(
              color: Colors.black.withOpacity(0.5),
              child: const Center(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    CircularProgressIndicator(color: Colors.white),
                    SizedBox(height: 12),
                    Text('Memverifikasi...',
                        style: TextStyle(color: Colors.white, fontSize: 14)),
                  ],
                ),
              ),
            ),
          if (_isFaceVerified)
            Container(
              color: Colors.black.withOpacity(0.3),
              child: const Center(
                child: Icon(Icons.check_circle_rounded,
                    color: Colors.white, size: 64),
              ),
            ),
        ],
      );
    }

    // No camera permission
    if (!_isCameraPermissionGranted) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.camera_alt_outlined, color: Colors.white54, size: 48),
            const SizedBox(height: 12),
            const Text('Izin kamera diperlukan', style: TextStyle(color: Colors.white70, fontSize: 14)),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: _checkCameraPermission,
              style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFF135BEC)),
              child: const Text('Berikan Izin'),
            )
          ],
        ),
      );
    }

    // Wait for location to be valid first
    if (!_isLocationValid) {
      return const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.location_off_outlined, color: Colors.white54, size: 48),
            SizedBox(height: 12),
            Text('Verifikasi lokasi Anda terlebih dahulu', style: TextStyle(color: Colors.white70, fontSize: 14)),
          ],
        ),
      );
    }

    // Camera not ready yet
    if (!_cameraInitialized || _cameraController == null) {
      return const Center(child: CircularProgressIndicator(color: Colors.white));
    }

    // Manual Start Screen
    if (!_isLivenessStarted) {
      return Stack(
        fit: StackFit.expand,
        children: [
          // Background preview, slightly blurred or darkened
          FittedBox(
            fit: BoxFit.cover,
            child: SizedBox(
              width: _cameraController!.value.previewSize!.height,
              height: _cameraController!.value.previewSize!.width,
              child: CameraPreview(_cameraController!),
            ),
          ),
          Container(color: Colors.black.withOpacity(0.6)),
          Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.white.withOpacity(0.1),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.face_retouching_natural, color: Colors.white, size: 48),
                ),
                const SizedBox(height: 16),
                const Text('Siap untuk verifikasi wajah?', 
                  style: TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold)
                ),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () {
                    _generateRandomChallenges();
                    setState(() {
                      _livenessStep = 0;
                      _isLivenessStarted = true;
                    });
                    _cameraController!.startImageStream(_processFrame);
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF135BEC),
                    padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
                  ),
                  child: const Text('Mulai Verifikasi', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                ),
              ],
            ),
          ),
        ],
      );
    }

    // Live camera with liveness overlay
    return Stack(
      fit: StackFit.expand,
      children: [
        // Camera preview – properly scaled
        FittedBox(
          fit: BoxFit.cover,
          child: SizedBox(
            width: _cameraController!.value.previewSize!.height,
            height: _cameraController!.value.previewSize!.width,
            child: CameraPreview(_cameraController!),
          ),
        ),

        // Dark vignette overlay
        Container(
          decoration: BoxDecoration(
            gradient: RadialGradient(
              center: Alignment.center,
              radius: 0.9,
              colors: [
                Colors.transparent,
                Colors.black.withOpacity(0.45),
              ],
            ),
          ),
        ),

        // Step instruction at top
        Positioned(
          top: 12,
          left: 12,
          right: 12,
          child: AnimatedSwitcher(
            duration: const Duration(milliseconds: 300),
            child: Container(
              key: ValueKey(_livenessStep),
              padding:
                  const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                color: Colors.black.withOpacity(0.6),
                borderRadius: BorderRadius.circular(12),
                border:
                    Border.all(color: Colors.white.withOpacity(0.15)),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    _currentStepIcon,
                    color: Colors.white,
                    size: 18,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    _currentStepLabel,
                    style: const TextStyle(
                        color: Colors.white,
                        fontSize: 13,
                        fontWeight: FontWeight.w600),
                  ),
                ],
              ),
            ),
          ),
        ),

        // Step progress dots at bottom
        Positioned(
          bottom: 12,
          left: 0,
          right: 0,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(3, (i) {
              final done = i < _livenessStep;
              final active = i == _livenessStep;
              return AnimatedContainer(
                duration: const Duration(milliseconds: 300),
                margin: const EdgeInsets.symmetric(horizontal: 4),
                width: active ? 20 : 8,
                height: 8,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(4),
                  color: done || active
                      ? Colors.white
                      : Colors.white.withOpacity(0.35),
                ),
              );
            }),
          ),
        ),
      ],
    );
  }

  Widget _buildStatusBar() {
    Color bg, fg;
    IconData icon;
    String text;

    if (_isLoading) {
      return const SizedBox.shrink();
    } else if (_isFaceVerified) {
      bg = AppTheme.successColor.withOpacity(0.1);
      fg = AppTheme.successColor;
      icon = Icons.check_circle;
      text = '✓ Wajah terverifikasi';
    } else if (_capturedImage != null) {
      bg = AppTheme.errorColor.withOpacity(0.1);
      fg = AppTheme.errorColor;
      icon = Icons.error;
      text = _faceStatus;
    } else {
      bg = const Color(0xFF135BEC).withOpacity(0.08);
      fg = const Color(0xFF135BEC);
      icon = Icons.info_outline;
      text = 'Langkah ${_livenessStep + 1}/3 – ${_livenessStep < 3 ? _stepLabels[_livenessStep] : "Selesai"}';
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
          color: bg, borderRadius: BorderRadius.circular(12)),
      child: Row(
        children: [
          Icon(icon, color: fg, size: 18),
          const SizedBox(width: 10),
          Expanded(
              child: Text(text,
                  style: TextStyle(
                      color: fg,
                      fontSize: 13,
                      fontWeight: FontWeight.w600))),
        ],
      ),
    );
  }
}
