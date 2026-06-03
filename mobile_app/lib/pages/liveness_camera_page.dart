import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_mlkit_face_detection/google_mlkit_face_detection.dart';
import 'package:mobile_app/theme/app_theme.dart';

class LivenessCameraPage extends StatefulWidget {
  const LivenessCameraPage({Key? key}) : super(key: key);

  @override
  State<LivenessCameraPage> createState() => _LivenessCameraPageState();
}

class _LivenessCameraPageState extends State<LivenessCameraPage>
    with SingleTickerProviderStateMixin {
  CameraController? _cameraController;
  late FaceDetector _faceDetector;

  bool _isBusy = false;
  bool _isInitialized = false;
  int _lastProcessTime = 0;

  // Liveness steps: 0=center, 1=left, 2=right
  int _step = 0;
  bool _isSuccess = false;

  List<List<double>> _boxesSeq = [];

  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  // Config per step
  static const _steps = [
    {
      'icon': Icons.face_outlined,
      'title': 'Posisikan wajah Anda',
      'subtitle': 'Pastikan wajah berada di dalam lingkaran',
    },
    {
      'icon': Icons.arrow_back_outlined,
      'title': 'Tengok ke Kiri',
      'subtitle': 'Putar kepala perlahan ke arah kiri',
    },
    {
      'icon': Icons.arrow_forward_outlined,
      'title': 'Tengok ke Kanan',
      'subtitle': 'Putar kepala perlahan ke arah kanan',
    },
  ];

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 2),
    )..repeat(reverse: true);
    _pulseAnimation =
        Tween<double>(begin: 1.0, end: 1.04).animate(_pulseController);

    _initializeCamera();
  }

  Future<void> _initializeCamera() async {
    try {
      final cameras = await availableCameras();
      final frontCamera = cameras.firstWhere(
        (c) => c.lensDirection == CameraLensDirection.front,
        orElse: () => cameras.first,
      );

      _cameraController = CameraController(
        frontCamera,
        ResolutionPreset.medium,
        enableAudio: false,
        imageFormatGroup: Platform.isAndroid
            ? ImageFormatGroup.nv21
            : ImageFormatGroup.bgra8888,
      );

      await _cameraController!.initialize();
      if (!mounted) return;

      setState(() => _isInitialized = true);
      _cameraController!.startImageStream(_processCameraImage);
    } catch (e) {
      debugPrint('Camera init error: $e');
    }
  }

  Future<void> _processCameraImage(CameraImage image) async {
    if (_isBusy || _isSuccess || _cameraController == null) return;

    final now = DateTime.now().millisecondsSinceEpoch;
    if (now - _lastProcessTime < 300) return;
    _lastProcessTime = now;

    _isBusy = true;
    try {
      final inputImage = _buildInputImage(image);
      if (inputImage == null) return;

      final faces = await _faceDetector.processImage(inputImage);
      if (!mounted) return;

      if (faces.isNotEmpty) {
        final face = faces.first;
        final yaw = face.headEulerAngleY ?? 0.0;
        final rect = face.boundingBox;

        _boxesSeq.add([rect.left, rect.top, rect.right, rect.bottom]);
        if (_boxesSeq.length > 30) _boxesSeq.removeAt(0);

        setState(() {
          if (_step == 0 && yaw > -10 && yaw < 10) {
            _step = 1;
          } else if (_step == 1 && yaw > 20) {
            _step = 2;
          } else if (_step == 2 && yaw < -20) {
            _onLivenessSuccess();
          }
        });
      }
    } catch (e) {
      debugPrint('Frame processing error: $e');
    } finally {
      _isBusy = false;
    }
  }

  void _onLivenessSuccess() async {
    if (_isSuccess) return;
    _isSuccess = true;
    _pulseController.stop();

    try {
      await _cameraController!.stopImageStream();
      await Future.delayed(const Duration(milliseconds: 300));
      final XFile photo = await _cameraController!.takePicture();

      if (!mounted) return;
      Navigator.pop(context, {
        'photoPath': photo.path,
        'liveness': jsonEncode(_boxesSeq),
      });
    } catch (e) {
      debugPrint('Capture error: $e');
      if (mounted) Navigator.pop(context);
    }
  }

  InputImage? _buildInputImage(CameraImage image) {
    if (_cameraController == null || image.planes.isEmpty) return null;

    final rotation = InputImageRotationValue.fromRawValue(
        _cameraController!.description.sensorOrientation);
    if (rotation == null) return null;

    final format = InputImageFormatValue.fromRawValue(image.format.raw);
    if (format == null) return null;

    final WriteBuffer allBytes = WriteBuffer();
    for (final Plane plane in image.planes) {
      allBytes.putUint8List(plane.bytes);
    }

    return InputImage.fromBytes(
      bytes: allBytes.done().buffer.asUint8List(),
      metadata: InputImageMetadata(
        size: Size(image.width.toDouble(), image.height.toDouble()),
        rotation: rotation,
        format: format,
        bytesPerRow: image.planes[0].bytesPerRow,
      ),
    );
  }

  @override
  void dispose() {
    _pulseController.dispose();
    _cameraController?.dispose();
    _faceDetector.close();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: SystemUiOverlayStyle.light,
      child: Scaffold(
        backgroundColor: Colors.black,
        body: _isInitialized ? _buildCamera() : _buildLoading(),
      ),
    );
  }

  Widget _buildLoading() {
    return Container(
      color: const Color(0xFF0D1117),
      child: const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(
              color: Colors.white,
              strokeWidth: 2,
            ),
            SizedBox(height: 16),
            Text(
              'Membuka kamera...',
              style: TextStyle(color: Colors.white70, fontSize: 14),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCamera() {
    final size = MediaQuery.of(context).size;
    return Stack(
      children: [
        // ── 1. Camera Preview fills the screen perfectly ──
        SizedBox.expand(
          child: FittedBox(
            fit: BoxFit.cover,
            child: SizedBox(
              width: _cameraController!.value.previewSize!.height,
              height: _cameraController!.value.previewSize!.width,
              child: CameraPreview(_cameraController!),
            ),
          ),
        ),

        // ── 2. Dark mask with circle cutout ──
        CustomPaint(
          size: size,
          painter: _CircleMaskPainter(
            isSuccess: _isSuccess,
            step: _step,
          ),
        ),

        // ── 3. Pulse ring on the circle ──
        Positioned.fill(
          child: Center(
            child: LayoutBuilder(builder: (ctx, constraints) {
              final radius = size.width * 0.42;
              return AnimatedBuilder(
                animation: _pulseAnimation,
                builder: (_, __) => _isSuccess
                    ? const SizedBox.shrink()
                    : Container(
                        width: radius * 2 * _pulseAnimation.value,
                        height: radius * 2 * _pulseAnimation.value,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          border: Border.all(
                            color: Colors.white.withOpacity(0.3),
                            width: 2,
                          ),
                        ),
                      ),
              );
            }),
          ),
        ),

        // ── 4. Header ──
        SafeArea(
          child: Padding(
            padding:
                const EdgeInsets.symmetric(horizontal: 16.0, vertical: 12.0),
            child: Row(
              children: [
                _glassButton(
                  icon: Icons.close,
                  onTap: () => Navigator.pop(context),
                ),
                const Spacer(),
                _glassPill(
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.verified_user_outlined,
                        color: Colors.white,
                        size: 16,
                      ),
                      const SizedBox(width: 6),
                      const Text(
                        'Verifikasi Wajah',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                ),
                const Spacer(),
                const SizedBox(width: 44),
              ],
            ),
          ),
        ),

        // ── 5. Step progress dots (below circle area) ──
        Positioned(
          left: 0,
          right: 0,
          bottom: 180,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(3, (i) {
              final active = i <= _step;
              final done = i < _step;
              return AnimatedContainer(
                duration: const Duration(milliseconds: 300),
                margin: const EdgeInsets.symmetric(horizontal: 5),
                width: done ? 24 : (active ? 24 : 8),
                height: 8,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(4),
                  color: done || active
                      ? Colors.white
                      : Colors.white.withOpacity(0.3),
                ),
              );
            }),
          ),
        ),

        // ── 6. Instruction card at the bottom ──
        Positioned(
          left: 20,
          right: 20,
          bottom: 40,
          child: AnimatedSwitcher(
            duration: const Duration(milliseconds: 350),
            transitionBuilder: (child, anim) =>
                FadeTransition(opacity: anim, child: child),
            child: _buildCard(key: ValueKey(_isSuccess ? 'done' : _step)),
          ),
        ),
      ],
    );
  }

  Widget _buildCard({required Key key}) {
    if (_isSuccess) {
      return _InstructionCard(
        key: key,
        icon: Icons.check_circle_rounded,
        iconColor: const Color(0xFF22C55E),
        title: 'Verifikasi Berhasil!',
        subtitle: 'Mengambil foto...',
        bgColor: const Color(0xFF0F2A1A),
        borderColor: const Color(0xFF22C55E),
      );
    }

    final info = _steps[_step];
    return _InstructionCard(
      key: key,
      icon: info['icon'] as IconData,
      iconColor: Colors.white,
      title: info['title'] as String,
      subtitle: info['subtitle'] as String,
      bgColor: Colors.black.withOpacity(0.65),
      borderColor: Colors.white.withOpacity(0.2),
    );
  }

  Widget _glassButton({required IconData icon, required VoidCallback onTap}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 44,
        height: 44,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: Colors.black.withOpacity(0.45),
          border: Border.all(color: Colors.white.withOpacity(0.15)),
        ),
        child: Icon(icon, color: Colors.white, size: 20),
      ),
    );
  }

  Widget _glassPill({required Widget child}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.45),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white.withOpacity(0.15)),
      ),
      child: child,
    );
  }
}

// ── Instruction Card ──────────────────────────────────────
class _InstructionCard extends StatelessWidget {
  final IconData icon;
  final Color iconColor;
  final String title;
  final String subtitle;
  final Color bgColor;
  final Color borderColor;

  const _InstructionCard({
    required Key key,
    required this.icon,
    required this.iconColor,
    required this.title,
    required this.subtitle,
    required this.bgColor,
    required this.borderColor,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        decoration: BoxDecoration(
          color: bgColor,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: borderColor, width: 1.5),
        ),
        child: Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: Colors.white.withOpacity(0.1),
              ),
              child: Icon(icon, color: iconColor, size: 26),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.65),
                      fontSize: 13,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Circle Mask Painter ────────────────────────────────────
class _CircleMaskPainter extends CustomPainter {
  final bool isSuccess;
  final int step;

  const _CircleMaskPainter({required this.isSuccess, required this.step});

  @override
  void paint(Canvas canvas, Size size) {
    final cx = size.width / 2;
    final cy = size.height * 0.42;
    final radius = size.width * 0.42;

    // Dark overlay
    final bgPath = Path()
      ..addRect(Rect.fromLTWH(0, 0, size.width, size.height));
    final circlePath = Path()
      ..addOval(Rect.fromCircle(center: Offset(cx, cy), radius: radius));
    final maskPath =
        Path.combine(PathOperation.difference, bgPath, circlePath);

    canvas.drawPath(
        maskPath,
        Paint()
          ..color = Colors.black.withOpacity(0.72)
          ..blendMode = BlendMode.srcOver);

    // Circle border
    final borderColor = isSuccess
        ? const Color(0xFF22C55E)
        : Colors.white.withOpacity(0.8);
    canvas.drawCircle(
      Offset(cx, cy),
      radius,
      Paint()
        ..color = borderColor
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2.5,
    );
  }

  @override
  bool shouldRepaint(_CircleMaskPainter old) =>
      old.isSuccess != isSuccess || old.step != step;
}
