// lib/pages/dashboard_page.dart
import 'dart:async';
import 'package:flutter/material.dart';
import 'dart:io';
import 'package:mobile_app/pages/face_attendance_page.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/attendance_model.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/sse_service.dart';

class EmployeeDashboardPage extends StatefulWidget {
  const EmployeeDashboardPage({super.key});

  @override
  State<EmployeeDashboardPage> createState() => _EmployeeDashboardPageState();
}

class _EmployeeDashboardPageState extends State<EmployeeDashboardPage>
    with SingleTickerProviderStateMixin, WidgetsBindingObserver {
  int _selectedIndex = 0;
  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;
  File? _profileImage;

  // ✅ State untuk jadwal kerja
  WorkScheduleInfoResponse? _workScheduleInfo;
  bool _isLoadingSchedule = true;

  // State dari API
  bool isClockedIn = false;
  bool hasClockedOut = false;
  String clockInTime = "--:--";
  String clockOutTime = "--:--";
  String currentTime = "";
  bool _isLoadingAttendance = true;
  TodayAttendanceDetail? _todayAttendance;

  // Real-time clock
  late Timer _clockTimer;
  Timer? _statsRefreshTimer;
  StreamSubscription? _sseSubscription;

  // Break state
  bool isOnBreak = false;
  Timer? _breakTimer;
  String breakDuration = "00:00:00";
  bool _isBreakLoading = false;

  // Quick stats
  int _workDays = 0;
  int _leaveRemaining = 0;
  double _overtimeHours = 0;
  bool _isLoadingStats = true;

  List<Map<String, dynamic>> _activities = [];

  User? _user;

  final GlobalKey<ScaffoldState> _scaffoldKey = GlobalKey<ScaffoldState>();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    ApiService.currentUser.addListener(_syncProfile);
    _animationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1000),
    );
    _fadeAnimation = CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeOut,
    );
    _animationController.forward();

    _tickClock();
    _clockTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (mounted) _tickClock();
    });

    // ✅ Load semua data
    _loadWorkScheduleInfo();
    _loadTodayAttendance();
    _loadMonthlyStats();
    _loadUser();

    _statsRefreshTimer = Timer.periodic(const Duration(seconds: 60), (_) {
      if (mounted) {
        _loadMonthlyStats();
      }
    });

    _setupSSE();
  }

  void _setupSSE() {
    _sseSubscription = SSEService().events.listen((event) {
      if (!mounted) return;
      // Refresh data jika ada update terkait absensi, pengajuan, atau statistik
      if (event.type == 'attendance_updated' ||
          event.type == 'leave_updated' ||
          event.type == 'stats_updated') {
        _loadTodayAttendance();
        _loadWorkScheduleInfo();
        _loadMonthlyStats();
      }
    });
  }

  void _syncProfile() {
    if (!mounted) return;
    setState(() => _user = ApiService.currentUser.value);
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && mounted) {
      _loadMonthlyStats();
    }
  }

  void _tickClock() {
    final now = DateTime.now();
    setState(() {
      currentTime =
          "${now.hour.toString().padLeft(2, '0')}:${now.minute.toString().padLeft(2, '0')}:${now.second.toString().padLeft(2, '0')}";
    });
  }

  // ✅ BARU: Load work schedule info
  Future<void> _loadWorkScheduleInfo() async {
    setState(() => _isLoadingSchedule = true);
    try {
      final info = await ApiService.getWorkScheduleInfo();
      if (mounted) {
        setState(() {
          _workScheduleInfo = info;
          _isLoadingSchedule = false;
        });
      }
    } catch (e) {
      print('[Dashboard] Load work schedule error: $e');
      if (mounted) setState(() => _isLoadingSchedule = false);
    }
  }

  void _startBreakTimer() {
    _breakTimer?.cancel();
    final startedAt = _todayAttendance?.breakStartTime;
    if (startedAt == null || startedAt.isEmpty) {
      return;
    }
    final parts = startedAt.split(':');
    if (parts.length < 2) {
      return;
    }
    final now = DateTime.now();
    final breakStart = DateTime(
      now.year,
      now.month,
      now.day,
      int.tryParse(parts[0]) ?? now.hour,
      int.tryParse(parts[1]) ?? now.minute,
    );

    _breakTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted || !isOnBreak) {
        _breakTimer?.cancel();
        return;
      }
      final e = DateTime.now().difference(breakStart);
      setState(() {
        breakDuration =
            "${e.inHours.toString().padLeft(2, '0')}:${(e.inMinutes % 60).toString().padLeft(2, '0')}:${(e.inSeconds % 60).toString().padLeft(2, '0')}";
      });
    });
  }

  void _stopBreakTimer() {
    _breakTimer?.cancel();
    _breakTimer = null;
    setState(() {
      breakDuration = "00:00:00";
    });
  }

  Future<void> _loadMonthlyStats() async {
    setState(() => _isLoadingStats = true);
    try {
      final now = DateTime.now();
      final summary = await ApiService.getMonthlyAttendance(
        month: now.month,
        year: now.year,
      );
      var leaveRemaining = _leaveRemaining;
      try {
        leaveRemaining = await ApiService.getLeaveBalance();
      } catch (e) {
        print('[Dashboard] Load leave balance error: $e');
      }
      if (mounted) {
        setState(() {
          _workDays = summary.totalDays;
          _overtimeHours = summary.overtimeHours;
          _leaveRemaining = leaveRemaining;
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
            clockInTime = attendance.hasClockedIn
                ? attendance.clockInTime
                : "--:--";
            clockOutTime = attendance.hasClockedOut
                ? (attendance.clockOutTime ?? "--:--")
                : "--:--";
            isOnBreak =
                (attendance.breakStartTime?.isNotEmpty ?? false) &&
                (attendance.breakEndTime == null ||
                    attendance.breakEndTime!.isEmpty);
            if (isOnBreak) {
              _startBreakTimer();
            } else {
              _stopBreakTimer();
            }
          } else {
            isClockedIn = false;
            hasClockedOut = false;
            clockInTime = "--:--";
            clockOutTime = "--:--";
            isOnBreak = false;
            _stopBreakTimer();
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

  void _buildActivities() {
    _activities = [];

    if (_todayAttendance != null && _todayAttendance!.hasClockedIn) {
      _activities.add({
        'icon': Icons.login,
        'title': 'Clock In',
        'time': clockInTime,
        'status': _todayAttendance!.status,
        'color': _todayAttendance!.status == 'Terlambat'
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

    final breakStart = _todayAttendance?.breakStartTime;
    final breakEnd = _todayAttendance?.breakEndTime;
    if ((breakStart?.isNotEmpty ?? false) || (breakEnd?.isNotEmpty ?? false)) {
      _activities.add({
        'icon': Icons.coffee_rounded,
        'title': 'Istirahat Mulai',
        'time': (breakStart?.isNotEmpty ?? false) ? breakStart : '--:--',
        'status': isOnBreak ? 'Mulai' : 'Mulai',
        'color': const Color(0xFFF59E0B),
      });
      if (breakEnd?.isNotEmpty == true) {
        _activities.add({
          'icon': Icons.free_breakfast_rounded,
          'title': 'Istirahat Selesai',
          'time': breakEnd,
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

  Future<void> _handleBreakToggle() async {
    if (!isClockedIn || hasClockedOut || _isBreakLoading) {
      return;
    }

    setState(() => _isBreakLoading = true);
    try {
      if (isOnBreak) {
        await ApiService.endBreak();
        _showInfoSnackBar("Istirahat selesai");
      } else {
        await ApiService.startBreak();
        _showInfoSnackBar("Istirahat dimulai");
      }
      await _loadTodayAttendance();
    } catch (e) {
      if (!mounted) return;
      _showInfoSnackBar(e.toString().replaceFirst('Exception: ', ''));
    } finally {
      if (mounted) {
        setState(() => _isBreakLoading = false);
      }
    }
  }

  Future<void> _navigateToFaceAttendance(String type) async {
    final result = await Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => FaceAttendancePage(type: type)),
    );

    if (result == true) {
      await _loadTodayAttendance();
      await _loadMonthlyStats();
      if (mounted) {
        _showSuccessSnackBar(
          type == 'clock_in' ? "✓ Clock In Berhasil" : "✓ Clock Out Berhasil",
          type == 'clock_in'
              ? const Color(0xFF2ECC71)
              : const Color(0xFF2ECC71),
        );
      }
    }
  }

  String _avatarUrl() {
    final avatar = (_user?.avatar ?? '').trim();
    if (avatar.isNotEmpty) return avatar;
    final n = Uri.encodeComponent(_user?.fullName ?? 'Employee');
    return 'https://ui-avatars.com/api/?name=$n&background=135BEC&color=fff&size=100';
  }

  @override
  void dispose() {
    _clockTimer.cancel();
    _statsRefreshTimer?.cancel();
    WidgetsBinding.instance.removeObserver(this);
    ApiService.currentUser.removeListener(_syncProfile);
    _breakTimer?.cancel();
    _animationController.dispose();
    _sseSubscription?.cancel();
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
              double maxWidth = constraints.maxWidth > 600
                  ? 600
                  : double.infinity;

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
                            onRefresh: () async {
                              await _loadWorkScheduleInfo();
                              await _loadTodayAttendance();
                              await _loadMonthlyStats();
                            },
                            child: SingleChildScrollView(
                              physics: const AlwaysScrollableScrollPhysics(),
                              padding: EdgeInsets.symmetric(
                                horizontal: horizontalPadding,
                              ),
                              child: Column(
                                children: [
                                  const SizedBox(height: 16),
                                  _buildMainClockSection(),
                                  const SizedBox(height: 20),
                                  _buildQuickStats(),
                                  const SizedBox(height: 24),
                                  _buildTodaysActivity(),
                                  const SizedBox(height: 16),
                                  _buildWorkScheduleCard(), // ✅ BARU
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

  Widget _buildMainClockSection() {
    // ✅ Determine button state berdasarkan work schedule
    // Clock IN tetap bisa selama jam kerja (bukan hanya 15 menit)
    bool canClockIn =
        (_workScheduleInfo?.todaySchedule?.canClockIn ?? false) &&
        !isClockedIn &&
        !hasClockedOut &&
        !_isLoadingAttendance;

    // Clock out: bisa sejak clock in, hingga 6 jam setelah jam pulang
    // Backend mengirimkan canClockOut=true jika user sudah clock in dan belum melewati window
    bool canClockOut =
        (_workScheduleInfo?.todaySchedule?.canClockOut ?? false) &&
        isClockedIn &&
        !hasClockedOut &&
        !_isLoadingAttendance;

    String clockInButtonLabel = canClockIn
        ? "CLOCK IN"
        : (_workScheduleInfo?.todaySchedule?.isWorkDay ?? false)
        ? "CLOCK IN"
        : "CLOCK IN";

    String clockOutButtonLabel = canClockOut
        ? "CLOCK OUT"
        : (_workScheduleInfo?.todaySchedule?.isWorkDay ?? false)
        ? "CLOCK OUT"
        : "CLOCK OUT";

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
            color:
                (isClockedIn
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
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 4,
                    ),
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
          Column(
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.baseline,
                textBaseline: TextBaseline.alphabetic,
                children: [
                  Text(
                    currentTime,
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
              if (isClockedIn || hasClockedOut) ...[
                const SizedBox(height: 8),
                Wrap(
                  alignment: WrapAlignment.center,
                  spacing: 8,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 14,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.15),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(
                            Icons.login,
                            color: Colors.white,
                            size: 14,
                          ),
                          const SizedBox(width: 6),
                          Text(
                            "Masuk: $clockInTime WIB",
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ),
                    if (hasClockedOut && clockOutTime != "--:--")
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 14,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: Colors.white.withOpacity(0.15),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(
                              Icons.logout,
                              color: Colors.white,
                              size: 14,
                            ),
                            const SizedBox(width: 6),
                            Text(
                              "Pulang: $clockOutTime WIB",
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 13,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
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
                  label: clockInButtonLabel,
                  color: Colors.white,
                  iconColor: const Color(0xFF2ECC71),
                  isEnabled: canClockIn,
                  onTap: canClockIn
                      ? () => _navigateToFaceAttendance('clock_in')
                      : null,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildMainActionButton(
                  icon: Icons.logout,
                  label: clockOutButtonLabel,
                  color: Colors.white,
                  iconColor: const Color(0xFFEF4444),
                  isEnabled: canClockOut,
                  onTap: canClockOut
                      ? () => _navigateToFaceAttendance('clock_out')
                      : null,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: isClockedIn && !hasClockedOut ? _handleBreakToggle : null,
              borderRadius: BorderRadius.circular(50),
              child: Container(
                width: double.infinity,
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 14,
                ),
                decoration: BoxDecoration(
                  color: isClockedIn && !hasClockedOut
                      ? Colors.white.withOpacity(0.15)
                      : Colors.white.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(50),
                  border: Border.all(
                    color: Colors.white.withOpacity(
                      isClockedIn && !hasClockedOut ? 0.2 : 0.1,
                    ),
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
                                  ? "Break time bisa dicatat saat istirahat"
                                  : "Tersedia saat jam kerja",
                              style: TextStyle(
                                color: Colors.white.withOpacity(0.7),
                                fontSize: 12,
                              ),
                            ),
                          ],
                          if (isOnBreak) ...[
                            const SizedBox(height: 2),
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
                        child: _isBreakLoading
                            ? SizedBox(
                                height: 14,
                                width: 14,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: isOnBreak
                                      ? const Color(0xFFF59E0B)
                                      : Colors.white,
                                ),
                              )
                            : Text(
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
        ],
      ),
    );
  }

  Widget _buildMainActionButton({
    required IconData icon,
    required String label,
    required Color color,
    required Color iconColor,
    required bool isEnabled,
    VoidCallback? onTap,
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
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildQuickStats() {
    final todaySchedule = _workScheduleInfo?.todaySchedule;

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
          color: todaySchedule?.isWorkDay ?? false
              ? const Color(0xFF135BEC).withOpacity(0.3)
              : const Color(0xFF94A3B8).withOpacity(0.3),
          width: 1.5,
        ),
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
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
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
    final todaySchedule = _workScheduleInfo?.todaySchedule;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
          color: todaySchedule?.isWorkDay ?? false
              ? const Color(0xFF135BEC).withOpacity(0.3)
              : const Color(0xFF94A3B8).withOpacity(0.3),
          width: 1.5,
        ),
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

  // ✅ BARU: Build work schedule card
  Widget _buildWorkScheduleCard() {
    if (_isLoadingSchedule || _workScheduleInfo == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 12),
          child: SizedBox(
            height: 20,
            width: 20,
            child: CircularProgressIndicator(
              strokeWidth: 2,
              valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF135BEC)),
            ),
          ),
        ),
      );
    }

    final schedule = _workScheduleInfo!;
    final todaySchedule = schedule.todaySchedule;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: todaySchedule?.isWorkDay ?? false
              ? const Color(0xFF135BEC).withOpacity(0.3)
              : const Color(0xFF94A3B8).withOpacity(0.3),
          width: 1.5,
        ),
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
          // Header
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: const Color(0xFF135BEC).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  Icons.schedule_rounded,
                  color: const Color(0xFF135BEC),
                  size: 20,
                ),
              ),
              const SizedBox(width: 10),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Jadwal Kerja Anda',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                      color: Color(0xFF64748B),
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    todaySchedule?.isWorkDay ?? false
                        ? '${schedule.hariKerja.join(", ")}'
                        : 'Bukan hari kerja',
                    style: TextStyle(
                      fontSize: 12,
                      color: todaySchedule?.isWorkDay ?? false
                          ? const Color(0xFF2ECC71)
                          : const Color(0xFF94A3B8),
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 14),
          // Jadwal waktu
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F5F9),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(
                      Icons.login_rounded,
                      color: const Color(0xFF2ECC71),
                      size: 18,
                    ),
                    const SizedBox(width: 8),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Masuk',
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade600,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          schedule.waktuMulai,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF2ECC71),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                Container(width: 1, height: 40, color: Colors.grey.shade300),
                Row(
                  children: [
                    Icon(
                      Icons.logout_rounded,
                      color: const Color(0xFFEF4444),
                      size: 18,
                    ),
                    const SizedBox(width: 8),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Pulang',
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade600,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          schedule.waktuSelesai,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFFEF4444),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ],
            ),
          ),
          // Status message
          if (todaySchedule?.message != null &&
              todaySchedule!.message.isNotEmpty) ...[
            const SizedBox(height: 10),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                color: const Color(0xFFFEF3C7),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Row(
                children: [
                  Icon(
                    Icons.info_outline_rounded,
                    color: const Color(0xFFF59E0B),
                    size: 16,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          todaySchedule.message,
                          style: const TextStyle(
                            fontSize: 12,
                            color: Color(0xFFF59E0B),
                            fontWeight: FontWeight.w500,
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '🕐 Clock In: ${todaySchedule.clockInWindow}',
                          style: const TextStyle(
                            fontSize: 11,
                            color: Color(0xFFF59E0B),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        Text(
                          '🕑 Clock Out: ${todaySchedule.clockOutWindow}',
                          style: const TextStyle(
                            fontSize: 11,
                            color: Color(0xFFF59E0B),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ],
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
    final todaySchedule = _workScheduleInfo?.todaySchedule;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: todaySchedule?.isWorkDay ?? false
              ? const Color(0xFF135BEC).withOpacity(0.3)
              : const Color(0xFF94A3B8).withOpacity(0.3),
          width: 1.5,
        ),
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

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(28)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Stack(
            children: [
              Hero(
                tag: 'profile',
                child: Container(
                  height: 48,
                  width: 48,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    gradient: const LinearGradient(
                      colors: [Color(0xFF135BEC), Color(0xFF3B7BF6)],
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: const Color(0xFF135BEC).withOpacity(0.3),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(2),
                    child: Container(
                      decoration: const BoxDecoration(
                        shape: BoxShape.circle,
                        color: Colors.white,
                      ),
                      child: ClipOval(
                        child: _profileImage != null
                            ? Image.file(_profileImage!, fit: BoxFit.cover)
                            : Image.network(
                                _avatarUrl(),
                                fit: BoxFit.cover,
                                errorBuilder: (_, __, ___) => const Icon(
                                  Icons.person,
                                  color: Color(0xFF135BEC),
                                  size: 26,
                                ),
                              ),
                      ),
                    ),
                  ),
                ),
              ),
              Positioned(
                bottom: 1,
                right: 1,
                child: Container(
                  height: 12,
                  width: 12,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: const Color(0xFF2ECC71),
                    border: Border.all(color: Colors.white, width: 2),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _greeting(),
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey.shade500,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                Text(
                  _user?.fullName ?? 'Profil Saya',
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Color(0xFF0F172A),
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          Stack(
            children: [
              Container(
                height: 44,
                width: 44,
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  shape: BoxShape.circle,
                ),
                child: IconButton(
                  icon: const Icon(
                    Icons.notifications_none,
                    color: Color(0xFF475569),
                    size: 22,
                  ),
                  onPressed: () {},
                  padding: EdgeInsets.zero,
                ),
              ),
              Positioned(
                top: 9,
                right: 9,
                child: Container(
                  height: 8,
                  width: 8,
                  decoration: const BoxDecoration(
                    shape: BoxShape.circle,
                    color: Color(0xFFEF4444),
                  ),
                ),
              ),
            ],
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
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
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
        content: Row(
          children: [
            const Icon(Icons.info, color: Colors.white, size: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Text(message, style: const TextStyle(fontSize: 14)),
            ),
          ],
        ),
        backgroundColor: const Color(0xFF135BEC),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: const Duration(seconds: 3),
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
      'Januari',
      'Februari',
      'Maret',
      'April',
      'Mei',
      'Juni',
      'Juli',
      'Agustus',
      'September',
      'Oktober',
      'November',
      'Desember',
    ];
    final days = [
      'Senin',
      'Selasa',
      'Rabu',
      'Kamis',
      'Jumat',
      'Sabtu',
      'Minggu',
    ];
    return '${days[now.weekday % 7]}, ${now.day} ${months[now.month - 1]} ${now.year}';
  }
}
