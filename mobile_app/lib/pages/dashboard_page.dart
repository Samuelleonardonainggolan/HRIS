// lib/pages/dashboard_page.dart
import 'dart:async';
import 'package:flutter/material.dart';
import 'dart:io';
import 'package:mobile_app/pages/face_attendance_page.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/attendance_model.dart';
import 'package:mobile_app/models/user_model.dart';

class EmployeeDashboardPage extends StatefulWidget {
  const EmployeeDashboardPage({super.key});

  @override
  State<EmployeeDashboardPage> createState() => _EmployeeDashboardPageState();
}

class _EmployeeDashboardPageState extends State<EmployeeDashboardPage>
    with SingleTickerProviderStateMixin {
  int _selectedIndex = 0;
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;
  File? _profileImage;

  // ✅ State dari API
  bool isClockedIn = false;
  bool hasClockedOut = false;
  String clockInTime = "--:--";
  String clockOutTime = "--:--";
  String currentTime = "";
  bool _isLoadingAttendance = true;
  TodayAttendanceDetail? _todayAttendance;

  // ✅ Real-time clock — Timer.periodic update tiap detik (HH:mm:ss)
  late Timer _clockTimer;

  // ✅ Break state
  bool isOnBreak = false;
  DateTime? _breakStartTime;
  Timer? _breakTimer;
  String breakDuration = "00:00:00";

  // ✅ Quick stats dari API
  int _workDays = 0;
  int _leaveRemaining = 0;
  double _overtimeHours = 0;
  bool _isLoadingStats = true;

  // Timeline activities dari API
  List<Map<String, dynamic>> _activities = [];

  // User profile for header
  User? _user;
  String? _breakEndTimeStr;  // waktu selesai istirahat

  final GlobalKey<ScaffoldState> _scaffoldKey = GlobalKey<ScaffoldState>();

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1000),
    );
    _fadeAnimation = CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeOut,
    );
    _animationController.forward();

    // ✅ Real-time clock HH:mm:ss — Timer.periodic update tiap detik
    _tickClock();
    _clockTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) _tickClock();
    });

    _loadTodayAttendance(); // ✅ Load dari API
    _loadMonthlyStats();   // ✅ Load stats bulanan
    _loadUser();           // ✅ Load user untuk header
  }

  // ✅ Update currentTime HH:mm:ss tiap detik
  void _tickClock() {
    final now = DateTime.now();
    setState(() {
      currentTime =
          "${now.hour.toString().padLeft(2, '0')}:${now.minute.toString().padLeft(2, '0')}:${now.second.toString().padLeft(2, '0')}";
    });
  }

  // ✅ Break timer — update durasi tiap detik
  void _startBreakTimer() {
    _breakStartTime = DateTime.now();
    _breakTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted || !isOnBreak) { _breakTimer?.cancel(); return; }
      final e = DateTime.now().difference(_breakStartTime!);
      setState(() {
        breakDuration =
            "${e.inHours.toString().padLeft(2, '0')}:${(e.inMinutes % 60).toString().padLeft(2, '0')}:${(e.inSeconds % 60).toString().padLeft(2, '0')}";
      });
    });
  }

  void _stopBreakTimer() {
    _breakTimer?.cancel();
    _breakTimer = null;
    final now = DateTime.now();
    _breakEndTimeStr = "${now.hour.toString().padLeft(2, '0')}:${now.minute.toString().padLeft(2, '0')}";
    _breakStartTime = null;
    setState(() { breakDuration = "00:00:00"; _buildActivities(); });
  }

  // ✅ Load stats bulanan (hari kerja, sisa cuti, lembur)
  Future<void> _loadMonthlyStats() async {
    setState(() => _isLoadingStats = true);
    try {
      final now = DateTime.now();
      final summary = await ApiService.getMonthlyAttendance(
        month: now.month, year: now.year);
      if (mounted) {
        setState(() {
          _workDays      = summary.totalDays;
          _overtimeHours = summary.overtimeHours;
          _leaveRemaining = 12; // TODO: sambungkan ke endpoint leave balance jika tersedia
          _isLoadingStats = false;
        });
      }
    } catch (e) {
      if (mounted) setState(() => _isLoadingStats = false);
    }
  }


  Future<void> _loadUser() async {
    try {
      final u = await ApiService.getProfile();
      if (mounted) setState(() => _user = u);
    } catch (_) {}
  }

  // ✅ Load absensi hari ini dari backend
  Future<void> _loadTodayAttendance() async {
    setState(() => _isLoadingAttendance = true);
    try {
      final attendance = await ApiService.getTodayAttendance();
      if (mounted) {
        setState(() {
          _todayAttendance = attendance;
          if (attendance != null) {
            isClockedIn = attendance.hasClockedIn && !attendance.hasClockedOut;
            hasClockedOut = attendance.hasClockedOut;
            clockInTime = attendance.hasClockedIn ? attendance.clockInTime : "--:--";
            clockOutTime = attendance.hasClockedOut ? (attendance.clockOutTime ?? "--:--") : "--:--";
          } else {
            isClockedIn = false;
            hasClockedOut = false;
            clockInTime = "--:--";
            clockOutTime = "--:--";
          }
          _buildActivities();
          _isLoadingAttendance = false;
        });
      }
    } catch (e) {
      print('[Dashboard] Load today attendance error: $e');
      if (mounted) {
        setState(() => _isLoadingAttendance = false);
      }
    }
  }

  // ✅ Build timeline activities dari data real
  void _buildActivities() {
    _activities = [];

    if (_todayAttendance != null && _todayAttendance!.hasClockedIn) {
      _activities.add({
        'icon': Icons.login,
        'title': 'Clock In',
        'time': clockInTime,
        'status': _todayAttendance!.status,
        'color': _todayAttendance!.status == 'Late'
            ? const Color(0xFFF59E0B)
            : const Color(0xFF2ECC71),
      });
    } else {
      _activities.add({
        'icon': Icons.login,
        'title': 'Clock In',
        'time': '--:--',
        'status': 'Pending',
        'color': const Color(0xFF94A3B8),
      });
    }

    // ✅ Break mulai
    if (isOnBreak || _breakEndTimeStr != null) {
      final breakStartStr = _breakEndTimeStr != null && _breakStartTime == null
          ? '--:--'
          : (_breakStartTime != null
              ? "${_breakStartTime!.hour.toString().padLeft(2, '0')}:${_breakStartTime!.minute.toString().padLeft(2, '0')}"
              : '--:--');
      _activities.add({
        'icon': Icons.coffee_rounded,
        'title': 'Istirahat Mulai',
        'time': breakStartStr,
        'status': isOnBreak ? 'Break' : 'Selesai',
        'color': const Color(0xFFF59E0B),
      });
      // Break selesai
      if (!isOnBreak && _breakEndTimeStr != null) {
        _activities.add({
          'icon': Icons.free_breakfast_rounded,
          'title': 'Istirahat Selesai',
          'time': _breakEndTimeStr!,
          'status': 'Selesai',
          'color': const Color(0xFF059669),
        });
      }
    }

    if (_todayAttendance != null && _todayAttendance!.hasClockedOut) {
      _activities.add({
        'icon': Icons.logout,
        'title': 'Clock Out',
        'time': clockOutTime,
        'status': 'Selesai',
        'color': const Color(0xFFEF4444),
      });
    } else {
      _activities.add({
        'icon': Icons.logout,
        'title': 'Clock Out',
        'time': '--:--',
        'status': 'Pending',
        'color': const Color(0xFF94A3B8),
      });
    }
  }

  // ✅ Navigasi ke face attendance, lalu reload data setelah kembali
  Future<void> _navigateToFaceAttendance(String type) async {
    final result = await Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => FaceAttendancePage(type: type)),
    );

    // ✅ Jika berhasil (result == true), reload data dari API
    if (result == true) {
      await _loadTodayAttendance();
      await _loadMonthlyStats();
      if (mounted) {
        _showSuccessSnackBar(
          type == 'clock_in' ? "✓ Clock In Berhasil" : "✓ Clock Out Berhasil",
          type == 'clock_in' ? const Color(0xFF2ECC71) : const Color(0xFFEF4444),
        );
      }
    }
  }

  String _avatarUrl() {
    final n = Uri.encodeComponent(_user?.fullName ?? 'Employee');
    return 'https://ui-avatars.com/api/?name=$n&background=135BEC&color=fff&size=100';
  }

  @override
  void dispose() {
    _clockTimer.cancel();    // ✅ Hentikan clock timer
    _breakTimer?.cancel();   // ✅ Hentikan break timer jika aktif
    _animationController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: Scaffold(
        key: _scaffoldKey,
        backgroundColor: const Color(0xFFF8FAFC),
        body: SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              double horizontalPadding = constraints.maxWidth > 600 ? 40 : 20;
              double maxWidth = constraints.maxWidth > 600 ? 600 : double.infinity;

              return Center(
                child: Container(
                  constraints: BoxConstraints(maxWidth: maxWidth),
                  child: FadeTransition(
                    opacity: _fadeAnimation,
                    child: Column(
                      children: [
                        _buildHeader(),
                        Expanded(
                          child: RefreshIndicator(
                            onRefresh: _loadTodayAttendance,
                            child: SingleChildScrollView(
                              physics: const AlwaysScrollableScrollPhysics(),
                              padding: EdgeInsets.symmetric(horizontal: horizontalPadding),
                              child: Column(
                                children: [
                                  const SizedBox(height: 16),
                                  _buildMainClockSection(),
                                  const SizedBox(height: 20),
                                  _buildQuickStats(),
                                  const SizedBox(height: 24),
                                  _buildTodaysActivity(),
                                  const SizedBox(height: 20),
                                  _buildLiveLocationCard(),
                                  const SizedBox(height: 80),
                                ],
                              ),
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(28)),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 16, offset: const Offset(0, 4))],
      ),
      child: Row(children: [
        Stack(children: [
          Hero(
            tag: 'profile',
            child: Container(
              height: 48, width: 48,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: const LinearGradient(colors: [Color(0xFF135BEC), Color(0xFF3B7BF6)]),
                boxShadow: [BoxShadow(color: const Color(0xFF135BEC).withOpacity(0.3), blurRadius: 8, offset: const Offset(0, 2))],
              ),
              child: Padding(padding: const EdgeInsets.all(2),
                child: Container(
                  decoration: const BoxDecoration(shape: BoxShape.circle, color: Colors.white),
                  child: ClipOval(
                    child: _profileImage != null
                        ? Image.file(_profileImage!, fit: BoxFit.cover)
                        : Image.network(_avatarUrl(), fit: BoxFit.cover,
                            errorBuilder: (_, __, ___) => const Icon(Icons.person, color: Color(0xFF135BEC), size: 26)),
                  ),
                )),
            ),
          ),
          Positioned(bottom: 1, right: 1,
            child: Container(height: 12, width: 12,
              decoration: BoxDecoration(shape: BoxShape.circle, color: const Color(0xFF2ECC71), border: Border.all(color: Colors.white, width: 2)))),
        ]),
        const SizedBox(width: 12),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(_greeting(), style: TextStyle(fontSize: 12, color: Colors.grey.shade500, fontWeight: FontWeight.w500)),
          Text(_user?.fullName ?? 'Profil Saya',
            style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
            overflow: TextOverflow.ellipsis),
        ])),
        Stack(children: [
          Container(height: 44, width: 44,
            decoration: BoxDecoration(color: const Color(0xFFF1F5F9), shape: BoxShape.circle),
            child: IconButton(
              icon: const Icon(Icons.notifications_none, color: Color(0xFF475569), size: 22),
              onPressed: () {}, padding: EdgeInsets.zero)),
          Positioned(top: 9, right: 9,
            child: Container(height: 8, width: 8,
              decoration: const BoxDecoration(shape: BoxShape.circle, color: Color(0xFFEF4444)))),
        ]),
      ]),
    );
  }

  Widget _buildMainClockSection() {
    // ✅ Tentukan state tombol
    bool canClockIn = !isClockedIn && !hasClockedOut && !_isLoadingAttendance;
    bool canClockOut = isClockedIn && !hasClockedOut && !_isLoadingAttendance;

    return AnimatedContainer(
      duration: const Duration(milliseconds: 300),
      width: double.infinity,
      padding: const EdgeInsets.all(15),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: isClockedIn
              ? [const Color(0xFF059669), const Color(0xFF10B981)]
              : hasClockedOut
                  ? [const Color(0xFF7C3AED), const Color(0xFF8B5CF6)]
                  : [const Color(0xFF135BEC), const Color(0xFF3B7BF6)],
        ),
        borderRadius: BorderRadius.circular(32),
        boxShadow: [
          BoxShadow(
            color: (isClockedIn
                        ? const Color(0xFF059669)
                        : hasClockedOut
                            ? const Color(0xFF7C3AED)
                            : const Color(0xFF135BEC))
                    .withOpacity(0.3),
            blurRadius: 25,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.2),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Row(
                      children: [
                        Container(
                          height: 8,
                          width: 8,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: hasClockedOut
                                ? Colors.white70
                                : isClockedIn
                                    ? Colors.white
                                    : const Color(0xFFFCD34D),
                          ),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          hasClockedOut
                              ? "SELESAI HARI INI"
                              : isClockedIn
                                  ? "SEDANG BEKERJA"
                                  : "BELUM ABSEN",
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            letterSpacing: 0.5,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    _getCurrentDate(),
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.9),
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
              if (_isLoadingAttendance)
                const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    color: Colors.white,
                    strokeWidth: 2,
                  ),
                ),
            ],
          ),

          const SizedBox(height: 8),

          // ✅ Jam real-time HH:mm:ss WIB — selalu berjalan tiap detik
          Column(
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.baseline,
                textBaseline: TextBaseline.alphabetic,
                children: [
                  Text(
                    currentTime, // ✅ Selalu tampilkan jam live
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 44,
                      fontWeight: FontWeight.bold,
                      fontFamily: 'monospace',
                      letterSpacing: 1,
                      shadows: [
                        Shadow(
                          color: Colors.black26,
                          blurRadius: 10,
                          offset: Offset(2, 2),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 4),
                  const Text(
                    "WIB",
                    style: TextStyle(
                      color: Colors.white70,
                      fontSize: 20,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: const Text(
                  "Waktu Saat Ini",
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              // ✅ Badge jam masuk & pulang terpisah (dari API)
              if (isClockedIn || hasClockedOut) ...[
                const SizedBox(height: 8),
                Wrap(
                  alignment: WrapAlignment.center,
                  spacing: 8,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.15),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.login, color: Colors.white, size: 14),
                          const SizedBox(width: 6),
                          Text("Masuk: $clockInTime WIB",
                              style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w600)),
                        ],
                      ),
                    ),
                    if (hasClockedOut && clockOutTime != "--:--")
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                        decoration: BoxDecoration(
                          color: Colors.white.withOpacity(0.15),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(Icons.logout, color: Colors.white, size: 14),
                            const SizedBox(width: 6),
                            Text("Pulang: $clockOutTime WIB",
                                style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.w600)),
                          ],
                        ),
                      ),
                  ],
                ),
              ],
            ],
          ),

          const SizedBox(height: 8),

          Row(
            children: [
              Expanded(
                child: _buildMainActionButton(
                  icon: Icons.login,
                  label: "CLOCK IN",
                  color: Colors.white,
                  iconColor: const Color(0xFF2ECC71),
                  isEnabled: canClockIn,
                  onTap: () => _navigateToFaceAttendance('clock_in'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildMainActionButton(
                  icon: Icons.logout,
                  label: "CLOCK OUT",
                  color: Colors.white,
                  iconColor: const Color(0xFFEF4444),
                  isEnabled: canClockOut,
                  onTap: () => _navigateToFaceAttendance('clock_out'),
                ),
              ),
            ],
          ),

          const SizedBox(height: 12),

          // ✅ Break section — berjalan real-time (sama seperti versi awal)
          Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: isClockedIn && !hasClockedOut
                  ? () {
                      setState(() {
                        isOnBreak = !isOnBreak;
                        if (isOnBreak) {
                          _startBreakTimer();
                        } else {
                          _stopBreakTimer();
                        }
                        _buildActivities();
                      });
                      _showInfoSnackBar(
                        isOnBreak ? "Istirahat dimulai" : "Istirahat selesai",
                      );
                    }
                  : null,
              borderRadius: BorderRadius.circular(50),
              child: Container(
                width: double.infinity,
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
                decoration: BoxDecoration(
                  color: isClockedIn && !hasClockedOut
                      ? Colors.white.withOpacity(0.15)
                      : Colors.white.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(50),
                  border: Border.all(
                    color: Colors.white.withOpacity(
                        isClockedIn && !hasClockedOut ? 0.2 : 0.1),
                  ),
                ),
                child: Row(
                  children: [
                    Container(
                      height: 40,
                      width: 40,
                      decoration: BoxDecoration(
                        color: isOnBreak
                            ? const Color(0xFFF59E0B)
                            : Colors.white.withOpacity(0.2),
                        shape: BoxShape.circle,
                      ),
                      child: Icon(
                        isOnBreak ? Icons.free_breakfast : Icons.coffee,
                        color: isOnBreak ? Colors.white : Colors.white70,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            isOnBreak ? "Sedang Istirahat" : "Waktu Istirahat",
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: 15,
                              fontWeight: isOnBreak
                                  ? FontWeight.bold
                                  : FontWeight.w600,
                            ),
                          ),
                          if (!isOnBreak) ...[
                            const SizedBox(height: 2),
                            Text(
                              isClockedIn && !hasClockedOut
                                  ? "Ambil istirahat singkat"
                                  : "Tersedia saat jam kerja",
                              style: TextStyle(
                                color: Colors.white.withOpacity(0.7),
                                fontSize: 12,
                              ),
                            ),
                          ],
                          if (isOnBreak) ...[
                            const SizedBox(height: 2),
                            // ✅ Durasi istirahat real-time HH:mm:ss
                            Text(
                              breakDuration,
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 14,
                                fontFamily: 'monospace',
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                    if (isClockedIn && !hasClockedOut)
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 8,
                        ),
                        decoration: BoxDecoration(
                          color: isOnBreak
                              ? Colors.white
                              : Colors.white.withOpacity(0.2),
                          borderRadius: BorderRadius.circular(30),
                        ),
                        child: Text(
                          isOnBreak ? "SELESAI" : "MULAI",
                          style: TextStyle(
                            color: isOnBreak
                                ? const Color(0xFFF59E0B)
                                : Colors.white,
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ),

          // ✅ Info kerja jika sudah clock-out
          if (hasClockedOut && _todayAttendance != null) ...[
            const SizedBox(height: 12),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
              decoration: BoxDecoration(
                color: Colors.white.withOpacity(0.15),
                borderRadius: BorderRadius.circular(50),
                border: Border.all(color: Colors.white.withOpacity(0.2)),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  _buildInfoChip(
                    Icons.timer,
                    "${_todayAttendance!.workHours.toStringAsFixed(1)} jam",
                    "Jam Kerja",
                  ),
                  Container(height: 30, width: 1, color: Colors.white30),
                  _buildInfoChip(
                    Icons.verified,
                    "${(_todayAttendance!.faceSimilarity != null ? _todayAttendance!.faceSimilarity! * 100 : 0).toStringAsFixed(0)}%",
                    "Similarity",
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildInfoChip(IconData icon, String value, String label) {
    return Column(
      children: [
        Icon(icon, color: Colors.white, size: 18),
        const SizedBox(height: 4),
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 14,
            fontWeight: FontWeight.bold,
          ),
        ),
        Text(
          label,
          style: TextStyle(
            color: Colors.white.withOpacity(0.7),
            fontSize: 10,
          ),
        ),
      ],
    );
  }

  Widget _buildMainActionButton({
    required IconData icon,
    required String label,
    required Color color,
    required Color iconColor,
    required bool isEnabled,
    required VoidCallback onTap,
  }) {
    return AnimatedOpacity(
      duration: const Duration(milliseconds: 200),
      opacity: isEnabled ? 1.0 : 0.4,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: isEnabled ? onTap : null,
          borderRadius: BorderRadius.circular(20),
          child: Container(
            height: 64,
            decoration: BoxDecoration(
              color: color,
              borderRadius: BorderRadius.circular(20),
              boxShadow: isEnabled
                  ? [
                      BoxShadow(
                        color: iconColor.withOpacity(0.4),
                        blurRadius: 15,
                        offset: const Offset(0, 5),
                      ),
                    ]
                  : null,
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(icon, color: iconColor, size: 22),
                const SizedBox(width: 8),
                Text(
                  label,
                  style: TextStyle(
                    color: iconColor,
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                    letterSpacing: 0.5,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildQuickStats() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: _isLoadingStats
          ? const Center(
              child: Padding(
                padding: EdgeInsets.symmetric(vertical: 8),
                child: SizedBox(
                    height: 24,
                    width: 24,
                    child: CircularProgressIndicator(strokeWidth: 2)),
              ),
            )
          : Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildStatItem(
                  icon: Icons.today,
                  value: "$_workDays",
                  label: "Hari Kerja",
                  color: const Color(0xFF135BEC),
                ),
                Container(height: 30, width: 1, color: Colors.grey.shade200),
                _buildStatItem(
                  icon: Icons.beach_access,
                  value: "$_leaveRemaining",
                  label: "Sisa Cuti",
                  color: const Color(0xFFF59E0B),
                ),
                Container(height: 30, width: 1, color: Colors.grey.shade200),
                _buildStatItem(
                  icon: Icons.timelapse,
                  value: "${_overtimeHours.toStringAsFixed(0)}j",
                  label: "Lembur",
                  color: const Color(0xFF8B5CF6),
                ),
              ],
            ),
    );
  }

  Color _getStatusColor(String? status) {
    switch (status) {
      case 'On Time':
        return const Color(0xFF059669);
      case 'Late':
        return const Color(0xFFF59E0B);
      case 'Overtime':
        return const Color(0xFF8B5CF6);
      case 'Absent':
        return const Color(0xFFEF4444);
      default:
        return const Color(0xFF94A3B8);
    }
  }

  Widget _buildStatItem({
    required IconData icon,
    required String value,
    required String label,
    required Color color,
  }) {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            shape: BoxShape.circle,
          ),
          child: Icon(icon, color: color, size: 18),
        ),
        const SizedBox(height: 6),
        Text(
          value,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: Color(0xFF0F172A),
          ),
          overflow: TextOverflow.ellipsis,
        ),
        Text(
          label,
          style: TextStyle(fontSize: 10, color: Colors.grey.shade600),
        ),
      ],
    );
  }

  Widget _buildTodaysActivity() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                "Aktivitas Hari Ini",
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF0F172A),
                ),
              ),
              if (_isLoadingAttendance)
                const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
            ],
          ),
          const SizedBox(height: 16),
          _buildActivityTimeline(),
        ],
      ),
    );
  }

  Widget _buildActivityTimeline() {
    if (_activities.isEmpty) _buildActivities();

    return Column(
      children: List.generate(_activities.length, (index) {
        final activity = _activities[index];
        final isLast = index == _activities.length - 1;
        final isPending = activity['status'] == 'Pending';

        return Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Column(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: (activity['color'] as Color).withOpacity(0.1),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    activity['icon'] as IconData,
                    color: activity['color'] as Color,
                    size: 18,
                  ),
                ),
                if (!isLast)
                  Container(height: 30, width: 2, color: Colors.grey.shade200),
              ],
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Padding(
                padding: EdgeInsets.only(bottom: isLast ? 0 : 20),
                child: Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            activity['title'] as String,
                            style: TextStyle(
                              fontSize: 15,
                              fontWeight: FontWeight.w600,
                              color: isPending
                                  ? Colors.grey.shade400
                                  : const Color(0xFF0F172A),
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            activity['time'] as String,
                            style: TextStyle(
                              fontSize: 13,
                              color: isPending
                                  ? Colors.grey.shade400
                                  : const Color(0xFF64748B),
                            ),
                          ),
                        ],
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 12,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: isPending
                            ? Colors.grey.shade100
                            : (activity['color'] as Color).withOpacity(0.1),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        activity['status'] as String,
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          color: isPending
                              ? Colors.grey.shade400
                              : activity['color'] as Color,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        );
      }),
    );
  }

  Widget _buildLiveLocationCard() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFF135BEC).withOpacity(0.1),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              Icons.location_on,
              color: Color(0xFF135BEC),
              size: 24,
            ),
          ),
          const SizedBox(width: 16),
          const Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  "Lokasi Kantor",
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Color(0xFF0F172A),
                  ),
                ),
                SizedBox(height: 4),
                Text(
                  "Labersa Hotel - Danau Toba",
                  style: TextStyle(fontSize: 13, color: Color(0xFF64748B)),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  void _showSuccessSnackBar(String message, Color color) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.check_circle, color: Colors.white, size: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                message,
                style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
              ),
            ),
          ],
        ),
        backgroundColor: color,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 2),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  void _showInfoSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message, style: const TextStyle(fontSize: 14)),
        backgroundColor: const Color(0xFF64748B),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 1),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  String _greeting() {
    final hour = DateTime.now().hour;
    if (hour < 12) return "Selamat Pagi";
    if (hour < 15) return "Selamat Siang";
    if (hour < 18) return "Selamat Sore";
    return "Selamat Malam";
  }

  String _getCurrentDate() {
    final now = DateTime.now();
    final months = [
      'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
      'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
    ];
    final days = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu', 'Minggu'];
    return '${days[now.weekday % 7]}, ${now.day} ${months[now.month - 1]} ${now.year}';
  }
}