// lib/pages/face_registration.dart
import 'dart:io';
import 'dart:convert';
import 'dart:math'; // Tambahkan import ini untuk sqrt
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:mobile_app/main_navigation.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:permission_handler/permission_handler.dart';

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

  late AnimationController _animationController;
  late Animation<double> _pulseAnimation;

  final ImagePicker _imagePicker = ImagePicker();

  @override
  void initState() {
    super.initState();
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
          _isLoading = true;
        });

        // TODO: Ganti dengan pemanggilan face recognition service yang sebenarnya
        // Untuk sekarang, kita generate dummy embedding yang unik per user
        await Future.delayed(const Duration(seconds: 2));

        // Generate embedding yang berbeda untuk setiap user
        // Menggunakan kombinasi userId dan timestamp sebagai seed
        final seed = widget.userId.hashCode +
            DateTime.now().millisecondsSinceEpoch.hashCode;

        // Generate 128-dimensional embedding dengan nilai antara -1 dan 1
        // (seperti hasil dari face recognition model seperti FaceNet)
        final dummyEmbedding = List.generate(128, (index) {
          // Gunakan pseudo-random yang konsisten
          final value = ((seed + index * 31) % 1000) / 500.0 - 1.0;
          return value;
        });

        // Normalisasi L2 (panjang vector = 1) seperti embedding face pada umumnya
        final norm = dummyEmbedding.fold(0.0, (sum, val) => sum + val * val);
        // PERBAIKAN: Gunakan sqrt dari dart:math
        final normalizedEmbedding =
            dummyEmbedding.map((v) => v / sqrt(norm)).toList();

        setState(() {
          _faceEmbedding = normalizedEmbedding;
          _isFaceDetected = true;
          _isLoading = false;
        });

        print('✅ Face embedding generated for user: ${widget.userId}');
        print('📊 Embedding length: ${normalizedEmbedding.length}');
        print('📊 Sample values: ${normalizedEmbedding.take(5).join(", ")}...');
        print('📊 Norm: $norm, sqrt: ${sqrt(norm)}');
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      _showErrorSnackBar('Gagal mengakses kamera: $e');
    }
  }

  Future<void> _registerFace() async {
    if (_capturedImage == null) {
      _showErrorSnackBar('Silakan ambil foto terlebih dahulu');
      return;
    }

    if (!_isFaceDetected || _faceEmbedding == null) {
      _showErrorSnackBar('Wajah tidak terdeteksi. Silakan coba lagi.');
      return;
    }

    setState(() {
      _isLoading = true;
    });

    try {
      // Convert image to base64
      final bytes = await _capturedImage!.readAsBytes();
      final base64Image = base64Encode(bytes);

      print('📸 Registering face for user: ${widget.userId}');
      print('📏 Embedding length: ${_faceEmbedding!.length}');
      print('📷 Image size: ${bytes.length} bytes');

      // Register face to backend
      await ApiService.registerFace(
        userId: widget.userId,
        faceEmbedding: _faceEmbedding!,
        faceImage: base64Image,
      );

      print('✅ Face registered successfully');

      if (!mounted) return;

      // Show success dialog
      _showSuccessDialog();
    } catch (e) {
      print('❌ Error registering face: $e');
      setState(() {
        _isLoading = false;
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
              'Data wajah Anda telah tersimpan. Anda sekarang dapat menggunakan fitur absensi face recognition.',
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
              // Header Info
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
                    const SizedBox(height: 8),
                    const Text(
                      'Ini adalah login pertama Anda. Silakan registrasi wajah untuk menggunakan fitur absensi.',
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

                    // Camera Preview
                    GestureDetector(
                      onTap: _captureImage,
                      child: Container(
                        height: 250,
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
                                    'Pastikan wajah Anda terlihat jelas',
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

                    // Face Detection Status
                    if (_capturedImage != null) ...[
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: _isFaceDetected
                              ? AppTheme.successColor.withOpacity(0.1)
                              : AppTheme.warningColor.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          children: [
                            Icon(
                              _isFaceDetected
                                  ? Icons.check_circle
                                  : Icons.error_outline,
                              color: _isFaceDetected
                                  ? AppTheme.successColor
                                  : AppTheme.warningColor,
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    _isFaceDetected
                                        ? 'Wajah terdeteksi'
                                        : 'Memproses deteksi wajah...',
                                    style: TextStyle(
                                      fontSize: 14,
                                      fontWeight: FontWeight.w600,
                                      color: _isFaceDetected
                                          ? AppTheme.successColor
                                          : AppTheme.warningColor,
                                    ),
                                  ),
                                  if (!_isFaceDetected)
                                    const Text(
                                      'Mohon tunggu sebentar',
                                      style: TextStyle(
                                        fontSize: 11,
                                        color: AppTheme.textSecondary,
                                      ),
                                    ),
                                ],
                              ),
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
                            Color(0xFF135BEC),
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),

              const SizedBox(height: 16),

              // Info Guidelines
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFFEFF6FF),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFFBFDBFE)),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Row(
                      children: [
                        Icon(
                          Icons.info_outline,
                          color: Color(0xFF135BEC),
                          size: 20,
                        ),
                        SizedBox(width: 8),
                        Text(
                          'Petunjuk Registrasi',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF135BEC),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    const Text(
                      '• Gunakan kamera depan untuk foto selfie\n'
                      '• Pastikan wajah terlihat jelas dan pencahayaan cukup\n'
                      '• Jangan gunakan aksesori yang menutupi wajah\n'
                      '• Posisikan wajah di tengah frame\n'
                      '• Ekspresi wajah normal (tidak tersenyum berlebihan)',
                      style: TextStyle(
                        fontSize: 12,
                        color: Color(0xFF334155),
                        height: 1.5,
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
                  onPressed: _isFaceDetected ? _registerFace : null,
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