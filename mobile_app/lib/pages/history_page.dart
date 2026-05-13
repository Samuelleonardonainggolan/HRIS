// lib/pages/history_page.dart
import 'package:flutter/material.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/services/sse_service.dart';
import 'package:mobile_app/models/attendance_model.dart';
import 'package:mobile_app/models/leave_request.dart';
import 'package:mobile_app/models/overtime_request.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:intl/intl.dart';
import 'dart:io';
import 'package:intl/date_symbol_data_local.dart';
import 'dart:async';

class HistoryPage extends StatefulWidget {
  const HistoryPage({super.key});
  @override
  State<HistoryPage> createState() => _HistoryPageState();
}

class _HistoryPageState extends State<HistoryPage> {
  DateTime _selectedMonth = DateTime.now();
  DateTime? _selectedDate;
  String _selectedFilter = 'Semua';
  final ScrollController _scrollController = ScrollController();

  /// Gabungan record absensi real + sintetis dari pengajuan APPROVED
  List<AttendanceRecord> _all = [];
  List<dynamic> _filtered = [];

  StreamSubscription? _sseSubscription;

  bool _isLoading = true;
  bool _localeReady = false;
  String? _error;
  User? _user;
  File? _profileImage;

  // Filter chips: semua kemungkinan status
  static const _filters = [
    'Semua',
    'Tepat Waktu',
    'Terlambat',
    'Izin',
    'Cuti',
    'Lembur',
  ];

  // Label chip → nilai AttendanceRecord.status
  static const Map<String, String?> _filterMap = {
    'Semua': null,
    'Tepat Waktu': 'Tepat Waktu',
    'Terlambat': 'Terlambat',
    'Izin': 'Izin',
    'Cuti': 'Cuti',
    'Lembur': 'Lembur',
  };

  @override
  void initState() {
    super.initState();
    ApiService.currentUser.addListener(_syncProfile);
    _sseSubscription = SSEService().events.listen((event) {
      if (mounted) _loadData();
    });
    _init();
  }

  void _syncProfile() {
    if (!mounted) return;
    setState(() => _user = ApiService.currentUser.value);
  }

  Future<void> _init() async {
    await initializeDateFormatting('id', null);
    if (mounted) setState(() => _localeReady = true);
    await Future.wait([_loadData(), _loadUser()]);
  }

  Future<void> _loadUser() async {
    try {
      final u = await ApiService.getProfile();
      if (mounted) setState(() => _user = u);
    } catch (_) {}
  }

  @override
  void dispose() {
    _scrollController.dispose();
    _sseSubscription?.cancel();
    ApiService.currentUser.removeListener(_syncProfile);
    super.dispose();
  }

  // ── Core: ambil attendance + pengajuan lalu merge ──────────────────────────
  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      // Ambil tiga sumber data secara paralel
      final results = await Future.wait([
        ApiService.getMonthlyAttendance(
          month: _selectedMonth.month,
          year: _selectedMonth.year,
        ),
        ApiService.getApprovedPengajuanByMonth(
          month: _selectedMonth.month,
          year: _selectedMonth.year,
        ),
        ApiService.getApprovedOvertimeByMonth(
          month: _selectedMonth.month,
          year: _selectedMonth.year,
        ),
      ]);

      final summary = results[0] as MonthlyAttendanceSummary;
      final pengajuan = results[1] as List<LeaveRequest>;
      final overtime = results[2] as List<OvertimeRequest>;

      print('[History] attendance records: ${summary.records.length}');
      print('[History] approved pengajuan for this month: ${pengajuan.length}');

      // Mulai dari record absensi real
      final merged = List<AttendanceRecord>.from(summary.records);

      // Kumpulkan tanggal yang sudah ada record absensi — format "yyyyMMdd"
      final existingKeys = <String>{for (final r in merged) _key(r.date)};

      // Expand setiap pengajuan APPROVED per hari dalam range-nya
      for (final p in pengajuan) {
        // Normalisasi ke midnight agar iterasi akurat
        var cur = DateTime(
          p.startDate.year,
          p.startDate.month,
          p.startDate.day,
        );
        final end = DateTime(p.endDate.year, p.endDate.month, p.endDate.day);

        print(
          '[History] expanding pengajuan ${p.id} "${p.type}" '
          '${_fmtDate(cur)} → ${_fmtDate(end)}',
        );

        while (!cur.isAfter(end)) {
          // Hanya hari yang berada di bulan yang sedang ditampilkan
          if (cur.month == _selectedMonth.month &&
              cur.year == _selectedMonth.year) {
            final k = _key(cur);
            if (!existingKeys.contains(k)) {
              final rec = AttendanceRecord.fromLeave(
                pengajuanId: p.id,
                date: cur,
                leaveType: p.type, // "Izin Sakit" / "Cuti Tahunan" dll.
                leaveKategori: p.namaKategori, // "Izin" / "Cuti" / "Lembur"
                leaveReason: p.reason,
              );
              merged.add(rec);
              existingKeys.add(k);
              print(
                '[History]   + added leave record ${_fmtDate(cur)} '
                'status=${rec.status}',
              );
            } else {
              print(
                '[History]   ~ skip ${_fmtDate(cur)} (absensi real sudah ada)',
              );
            }
            cur = cur.add(const Duration(days: 1));
          }
        }
      }

      // Merge Overtime Requests
      for (final o in overtime) {
        if (o.date.month == _selectedMonth.month && o.date.year == _selectedMonth.year) {
          final myEntry = o.employees.cast<OvertimeEmployee?>().firstWhere(
            (e) => e?.userId == ApiService.currentUser.value?.id,
            orElse: () => null,
          );
          String? rewardText = myEntry?.reward?.rewardTypeDisplay;
          if (myEntry?.reward?.rewardDate != null) {
            final dStr = DateFormat('dd/MM', 'id').format(myEntry!.reward!.rewardDate!);
            rewardText = '$rewardText ($dStr)';
          }

          final rec = AttendanceRecord.fromOvertime(
            id: o.id,
            date: o.date,
            startTime: o.startTime,
            endTime: o.endTime,
            reason: o.reason,
            overtimeHours: o.getDurationHours(),
            summary: null, // Removed approval summary as requested
            rewardInfo: rewardText,
          );
          merged.add(rec);
        }
      }

      print('[History] total merged records: ${merged.length}');

      if (mounted) {
        setState(() {
          _all = merged;
          _applyFilter();
          _isLoading = false;
        });
      }
    } catch (e, st) {
      print('[History] _loadData error: $e\n$st');
      if (mounted) {
        setState(() {
          _error = 'Gagal memuat data: $e';
          _isLoading = false;
        });
      }
    }
  }

  /// Key unik per hari — format "yyyyMMdd"
  String _key(DateTime d) =>
      '${d.year}${d.month.toString().padLeft(2, '0')}${d.day.toString().padLeft(2, '0')}';

  String _fmtDate(DateTime d) =>
      '${d.year}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';

  String _overtimeApprovalSummary(OvertimeRequest o) {
    final agreed = o.employees.where((e) => e.isAgreed).length;
    final rejected = o.employees.where((e) => e.isRejected).length;
    final pending = o.employees.where((e) => e.isPending).length;

    if (o.isPublished) {
      return 'SPKL dipublikasikan • $agreed setuju, $rejected menolak, $pending menunggu';
    }
    if (o.isSubmitted) {
      return 'Menunggu respons karyawan • $agreed setuju, $rejected menolak, $pending menunggu';
    }
    return 'Draft lembur • $agreed setuju, $rejected menolak, $pending menunggu';
  }

  // ── Filter ────────────────────────────────────────────────────────────────
  void _applyFilter() {
    final statusFilter = _filterMap[_selectedFilter];
    final raw = _all.where((r) {
      final mOk =
          r.date.month == _selectedMonth.month &&
          r.date.year == _selectedMonth.year;
      final dOk = _selectedDate == null || _isSameDay(r.date, _selectedDate!);
      bool sOk;
      if (statusFilter == null) {
        sOk = true;
      } else if (statusFilter == 'Lembur') {
        sOk = r.status == 'Lembur' || r.status == 'Overtime';
      } else {
        sOk = r.status == statusFilter;
      }
      return mOk && dOk && sOk;
    }).toList()
      ..sort((a, b) => b.date.compareTo(a.date));

    final List<dynamic> grouped = [];
    DateTime? lastDate;

    for (final r in raw) {
      final curDate = DateTime(r.date.year, r.date.month, r.date.day);
      if (lastDate == null || !_isSameDay(lastDate, curDate)) {
        grouped.add(curDate);
        lastDate = curDate;
      }
      grouped.add(r);
    }

    setState(() {
      _filtered = grouped;
    });
  }

  void _changeMonth(int delta) {
    setState(() {
      _selectedMonth = DateTime(
        _selectedMonth.year,
        _selectedMonth.month + delta,
        1,
      );
      _selectedDate = null;
    });
    _loadData();
  }

  Future<void> _pickMonthYear() async {
    final picked = await showModalBottomSheet<DateTime>(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (sheetContext) {
        var tempYear = _selectedMonth.year;
        var tempMonth = _selectedMonth.month;
        final monthLabels = List.generate(
          12,
          (i) => DateFormat(
            'MMM',
            'id',
          ).format(DateTime(2000, i + 1, 1)).toUpperCase(),
        );

        return StatefulBuilder(
          builder: (context, setModalState) {
            return SafeArea(
              child: Container(
                margin: const EdgeInsets.fromLTRB(12, 0, 12, 12),
                padding: const EdgeInsets.fromLTRB(16, 14, 16, 16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Pilih Periode',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w700,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                    const SizedBox(height: 12),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        TextButton.icon(
                          onPressed: () => setModalState(() => tempYear -= 1),
                          icon: const Icon(Icons.arrow_back_rounded, size: 16),
                          label: const Text(''),
                          style: TextButton.styleFrom(
                            foregroundColor: const Color(0xFF334155),
                          ),
                        ),
                        Text(
                          '$tempYear',
                          style: const TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.w700,
                            color: Color(0xFF0F172A),
                          ),
                        ),
                        TextButton.icon(
                          onPressed: () => setModalState(() => tempYear += 1),
                          label: const Text(''),
                          icon: const Icon(
                            Icons.arrow_forward_rounded,
                            size: 16,
                          ),
                          style: TextButton.styleFrom(
                            foregroundColor: const Color(0xFF334155),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    GridView.builder(
                      shrinkWrap: true,
                      physics: const NeverScrollableScrollPhysics(),
                      itemCount: 12,
                      gridDelegate:
                          const SliverGridDelegateWithFixedCrossAxisCount(
                            crossAxisCount: 3,
                            mainAxisSpacing: 8,
                            crossAxisSpacing: 8,
                            childAspectRatio: 2.4,
                          ),
                      itemBuilder: (_, i) {
                        final month = i + 1;
                        final selected = month == tempMonth;
                        return InkWell(
                          onTap: () => setModalState(() => tempMonth = month),
                          borderRadius: BorderRadius.circular(12),
                          child: Container(
                            alignment: Alignment.center,
                            decoration: BoxDecoration(
                              color: selected
                                  ? const Color(0xFF135BEC).withOpacity(0.1)
                                  : const Color(0xFFF8FAFC),
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(
                                color: selected
                                    ? const Color(0xFF135BEC).withOpacity(0.3)
                                    : const Color(0xFFE2E8F0),
                              ),
                            ),
                            child: Text(
                              monthLabels[i],
                              style: TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.w700,
                                color: selected
                                    ? const Color(0xFF135BEC)
                                    : const Color(0xFF475569),
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                    const SizedBox(height: 14),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: () {
                          Navigator.of(
                            sheetContext,
                          ).pop(DateTime(tempYear, tempMonth, 1));
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFF135BEC),
                          foregroundColor: Colors.white,
                        ),
                        child: const Text('Pilih Periode'),
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        );
      },
    );

    if (picked == null || !mounted) return;
    setState(() {
      _selectedMonth = DateTime(picked.year, picked.month, 1);
      _selectedDate = null;
    });
    _loadData();
  }

  bool _isSameDay(DateTime a, DateTime b) {
    return a.year == b.year && a.month == b.month && a.day == b.day;
  }

  Future<void> _pickDateFilter() async {
    final firstDay = DateTime(_selectedMonth.year, _selectedMonth.month, 1);
    final lastDay = DateTime(_selectedMonth.year, _selectedMonth.month + 1, 0);
    final initialDate = _selectedDate ?? _selectedMonth;

    final picked = await showDatePicker(
      context: context,
      initialDate: initialDate.isBefore(firstDay)
          ? firstDay
          : (initialDate.isAfter(lastDay) ? lastDay : initialDate),
      firstDate: firstDay,
      lastDate: lastDay,
      helpText: 'Pilih hanya tanggal di ${_fmt(_selectedMonth, 'MMMM yyyy')}',
      locale: const Locale('id', 'ID'),
    );

    if (picked == null || !mounted) return;
    setState(() {
      _selectedDate = DateTime(picked.year, picked.month, picked.day);
    });
    _applyFilter();
  }

  void _clearDateFilter() {
    if (_selectedDate == null) return;
    setState(() {
      _selectedDate = null;
    });
    _applyFilter();
  }

  // ── Statistik banner (dari _all yang sudah merged) ────────────────────────
  int get _cntHadir => _all.where((r) => r.status == 'Tepat Waktu').length;
  int get _cntLate => _all.where((r) => r.status == 'Terlambat').length;
  int get _cntIzin => _all.where((r) => r.status == 'Izin').length;
  int get _cntCuti => _all.where((r) => r.status == 'Cuti').length;
  int get _cntLembur =>
      _all.where((r) => r.status == 'Lembur' || r.status == 'Overtime').length;

  String _fmt(DateTime dt, String p) =>
      _localeReady ? DateFormat(p, 'id').format(dt) : '';

  String _greeting() {
    final h = DateTime.now().hour;
    if (h < 12) return 'Selamat Pagi';
    if (h < 15) return 'Selamat Siang';
    if (h < 18) return 'Selamat Sore';
    return 'Selamat Malam';
  }

  String _avatarUrl() {
    final avatar = (_user?.avatar ?? '').trim();
    if (avatar.isNotEmpty) return avatar;
    final n = Uri.encodeComponent(_user?.fullName ?? 'Employee');
    return 'https://ui-avatars.com/api/?name=$n&background=135BEC&color=fff&size=100';
  }

  // ═════════════════════════════════════════════════════════════════════════
  // BUILD
  // ═════════════════════════════════════════════════════════════════════════
  @override
  Widget build(BuildContext context) {
    if (!_localeReady) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: Column(
          children: [
            _buildHeader(),
            _buildMonthBanner(),
            _buildFilterChips(),
            Expanded(child: _buildBody()),
          ],
        ),
      ),
    );
  }

  // ── Header ────────────────────────────────────────────────────────────────
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
                tag: 'profile_history',
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
          ValueListenableBuilder<bool>(
            valueListenable: SSEService().hasNewOvertime,
            builder: (context, hasOvertime, _) {
              return ValueListenableBuilder<bool>(
                valueListenable: SSEService().hasNewAssignment,
                builder: (context, hasAssignment, _) {
                  return ValueListenableBuilder<bool>(
                    valueListenable: SSEService().hasNewLeaveRequest,
                    builder: (context, hasLeave, _) {
                      final hasNew = hasOvertime || hasAssignment || hasLeave;
                      return Stack(
                        children: [
                          Container(
                            height: 44,
                            width: 44,
                            decoration: const BoxDecoration(
                              color: Color(0xFFF1F5F9),
                              shape: BoxShape.circle,
                            ),
                            child: IconButton(
                              icon: const Icon(
                                Icons.notifications_none,
                                color: Color(0xFF475569),
                                size: 22,
                              ),
                              onPressed: () {
                                SSEService().hasNewOvertime.value = false;
                                SSEService().hasNewAssignment.value = false;
                                SSEService().hasNewLeaveRequest.value = false;
                              },
                              padding: EdgeInsets.zero,
                            ),
                          ),
                          if (hasNew)
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
                      );
                    },
                  );
                },
              );
            },
          ),
        ],
      ),
    );
  }

  // ── Month banner ──────────────────────────────────────────────────────────
  Widget _buildMonthBanner() {
    final isCurrent =
        _selectedMonth.month == DateTime.now().month &&
        _selectedMonth.year == DateTime.now().year;
    return Container(
      margin: const EdgeInsets.fromLTRB(16, 16, 16, 0),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFF135BEC), Color(0xFF2563EB)],
        ),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF135BEC).withOpacity(0.3),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              InkWell(
                onTap: () => _changeMonth(-1),
                borderRadius: BorderRadius.circular(10),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 6,
                    vertical: 4,
                  ),
                  child: Column(
                    children: const [
                      Icon(
                        Icons.arrow_back_rounded,
                        color: Colors.white,
                        size: 18,
                      ),
                    ],
                  ),
                ),
              ),
              Material(
                color: Colors.transparent,
                child: InkWell(
                  onTap: _pickMonthYear,
                  borderRadius: BorderRadius.circular(10),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 4,
                    ),
                    child: Column(
                      children: [
                        Text(
                          'Periode Saat Ini',
                          style: TextStyle(
                            color: Colors.white.withOpacity(0.8),
                            fontSize: 11,
                          ),
                        ),
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              _fmt(_selectedMonth, 'MMMM yyyy'),
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 18,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(width: 2),
                            const Icon(
                              Icons.keyboard_arrow_down_rounded,
                              color: Colors.white,
                              size: 18,
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              InkWell(
                onTap: isCurrent ? null : () => _changeMonth(1),
                borderRadius: BorderRadius.circular(10),
                child: Opacity(
                  opacity: isCurrent ? 0.4 : 1,
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 4,
                    ),
                    child: Column(
                      children: const [
                        Icon(
                          Icons.arrow_forward_rounded,
                          color: Colors.white,
                          size: 18,
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          // 5 statistik — semuanya dari _all (sudah merged)
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _stat('TEPAT WAKTU', '$_cntHadir', Colors.white),
              _vDiv(),
              _stat('TERLAMBAT', '$_cntLate', const Color(0xFFFCD34D)),
              _vDiv(),
              _stat('IZIN', '$_cntIzin', Colors.white),
              _vDiv(),
              _stat('CUTI', '$_cntCuti', const Color(0xFFD8B4FE)),
              _vDiv(),
              _stat('LEMBUR', '$_cntLembur', const Color(0xFFFBBF24)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _vDiv() => Container(height: 28, width: 1, color: Colors.white24);

  Widget _stat(String label, String value, Color color) => Column(
    children: [
      Text(
        value,
        style: TextStyle(
          color: color,
          fontSize: 18,
          fontWeight: FontWeight.bold,
        ),
      ),
      Text(
        label,
        style: TextStyle(
          color: Colors.white.withOpacity(0.7),
          fontSize: 8,
          fontWeight: FontWeight.w600,
          letterSpacing: 0.3,
        ),
      ),
    ],
  );

  // ── Filter chips ──────────────────────────────────────────────────────────
  Widget _buildFilterChips() {
    final hasDateFilter = _selectedDate != null;

    return Container(
      margin: const EdgeInsets.fromLTRB(16, 12, 16, 0),
      padding: const EdgeInsets.fromLTRB(12, 10, 10, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Expanded(
            child: Stack(
              alignment: Alignment.center,
              children: [
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 18),
                  child: Row(
                    children: _filters.map((f) {
                      final sel = _selectedFilter == f;
                      return Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: GestureDetector(
                          onTap: () {
                            setState(() => _selectedFilter = f);
                            _applyFilter();
                          },
                          child: AnimatedContainer(
                            duration: const Duration(milliseconds: 180),
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 9,
                            ),
                            decoration: BoxDecoration(
                              color: sel
                                  ? const Color(0xFF135BEC).withOpacity(0.08)
                                  : const Color(0xFFF8FAFC),
                              borderRadius: BorderRadius.circular(18),
                              border: Border.all(
                                color: sel
                                    ? const Color(0xFF135BEC).withOpacity(0.25)
                                    : Colors.grey.shade200,
                              ),
                            ),
                            child: Text(
                              f,
                              style: TextStyle(
                                fontSize: 13,
                                fontWeight: FontWeight.w600,
                                color: sel
                                    ? const Color(0xFF135BEC)
                                    : Colors.grey.shade700,
                              ),
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ),
                Positioned(
                  left: 0,
                  child: IgnorePointer(
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.9),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: Colors.grey.shade200),
                      ),
                      child: Row(
                        children: [
                          Icon(
                            Icons.arrow_back_ios_new_rounded,
                            color: Colors.grey.shade500,
                            size: 12,
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
                Positioned(
                  right: 0,
                  child: IgnorePointer(
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.9),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: Colors.grey.shade200),
                      ),
                      child: Row(
                        children: [
                          Icon(
                            Icons.arrow_forward_ios_rounded,
                            color: Colors.grey.shade500,
                            size: 12,
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Material(
            color: hasDateFilter
                ? const Color(0xFF135BEC).withOpacity(0.08)
                : const Color(0xFFF8FAFC),
            borderRadius: BorderRadius.circular(14),
            child: InkWell(
              onTap: _pickDateFilter,
              onLongPress: hasDateFilter ? _clearDateFilter : null,
              borderRadius: BorderRadius.circular(14),
              child: Tooltip(
                message: hasDateFilter
                    ? 'Tanggal: ${_fmt(_selectedDate!, 'dd MMM yyyy')} (tekan lama untuk reset)'
                    : 'Filter tanggal',
                child: Container(
                  width: 42,
                  height: 42,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(
                      color: hasDateFilter
                          ? const Color(0xFF135BEC).withOpacity(0.3)
                          : Colors.grey.shade200,
                    ),
                  ),
                  child: Stack(
                    alignment: Alignment.center,
                    children: [
                      Icon(
                        Icons.calendar_today_rounded,
                        size: 16,
                        color: hasDateFilter
                            ? const Color(0xFF135BEC)
                            : const Color(0xFF475569),
                      ),
                      if (hasDateFilter)
                        Positioned(
                          top: 8,
                          right: 8,
                          child: Container(
                            width: 6,
                            height: 6,
                            decoration: const BoxDecoration(
                              color: Color(0xFF135BEC),
                              shape: BoxShape.circle,
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ── Body ──────────────────────────────────────────────────────────────────
  Widget _buildBody() {
    if (_isLoading) return const Center(child: CircularProgressIndicator());

    if (_error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.error_outline, size: 52, color: Colors.red.shade300),
            const SizedBox(height: 12),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24),
              child: Text(
                _error!,
                style: TextStyle(color: Colors.grey.shade600),
                textAlign: TextAlign.center,
              ),
            ),
            const SizedBox(height: 12),
            ElevatedButton.icon(
              onPressed: _loadData,
              icon: const Icon(Icons.refresh),
              label: const Text('Coba Lagi'),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF135BEC),
                foregroundColor: Colors.white,
              ),
            ),
          ],
        ),
      );
    }

    if (_filtered.isEmpty) {
      final dateLabel = _selectedDate == null
          ? ''
          : ' pada ${_fmt(_selectedDate!, 'dd MMM yyyy')}';
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.calendar_today_outlined,
              size: 52,
              color: Colors.grey.shade300,
            ),
            const SizedBox(height: 12),
            Text(
              'Tidak ada data "$_selectedFilter"$dateLabel\nuntuk ${_fmt(_selectedMonth, 'MMMM yyyy')}',
              style: TextStyle(color: Colors.grey.shade500),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadData,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 80),
        itemCount: _filtered.length,
        itemBuilder: (_, i) {
          final item = _filtered[i];
          if (item is DateTime) {
            return _buildDateHeader(item);
          }
          final r = item as AttendanceRecord;
          return r.isLeaveRecord ? _buildLeaveCard(r) : _buildAttendanceCard(r);
        },
      ),
    );
  }

  // ── Header Tanggal (Daily Log Group) ──────────────────────────────────────
  Widget _buildDateHeader(DateTime date) {
    return Padding(
      padding: const EdgeInsets.only(top: 20, bottom: 10, left: 4),
      child: Row(
        children: [
          Icon(
            Icons.calendar_today_rounded,
            size: 16,
            color: Colors.grey.shade600,
          ),
          const SizedBox(width: 8),
          Text(
            _fmt(date, 'EEE, dd MMM yyyy'),
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w700,
              color: Colors.grey.shade700,
              letterSpacing: 0.2,
            ),
          ),
        ],
      ),
    );
  }

  // ── Card absensi real ─────────────────────────────────────────────────────
  Widget _buildAttendanceCard(AttendanceRecord r) {
    final sc = _statusColor(r.status);
    final sl = _statusLabel(r.status);
    final hasBreak =
        (r.breakStart?.isNotEmpty ?? false) ||
        (r.breakEnd?.isNotEmpty ?? false);

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: Colors.grey.shade200),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      r.shiftName.isNotEmpty ? r.shiftName : 'Shift Kerja',
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                  ],
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 5,
                  ),
                  decoration: BoxDecoration(
                    color: sc.withOpacity(0.07),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: sc.withOpacity(0.16)),
                  ),
                  child: Text(
                    sl,
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      color: sc.withOpacity(0.85),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Row(
              children: [
                Expanded(
                  child: _timeChip(
                    Icons.login_rounded,
                    'JAM MASUK',
                    r.clockIn.isEmpty ? '--:--' : '${r.clockIn} WIB',
                    const Color(0xFF2ECC71),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: _timeChip(
                    Icons.logout_rounded,
                    'JAM KELUAR',
                    (r.clockOut.isEmpty || r.clockOut == '--:--')
                        ? '-'
                        : '${r.clockOut} WIB',
                    const Color(0xFFEF4444),
                  ),
                ),
              ],
            ),
          ),
          if (hasBreak) ...[
            const SizedBox(height: 8),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Row(
                children: [
                  Expanded(
                    child: _timeChip(
                      Icons.coffee_rounded,
                      'ISTIRAHAT MULAI',
                      r.breakStart?.isNotEmpty == true
                          ? '${r.breakStart} WIB'
                          : '-',
                      const Color(0xFFF59E0B),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _timeChip(
                      Icons.free_breakfast_rounded,
                      'ISTIRAHAT SELESAI',
                      r.breakEnd?.isNotEmpty == true
                          ? '${r.breakEnd} WIB'
                          : '-',
                      const Color(0xFF059669),
                    ),
                  ),
                ],
              ),
            ),
          ],
          const SizedBox(height: 10),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            decoration: const BoxDecoration(
              color: Color(0xFFF8FAFC),
              borderRadius: BorderRadius.vertical(bottom: Radius.circular(18)),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(
                      Icons.location_on_outlined,
                      size: 13,
                      color: Colors.grey.shade400,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      r.location.isNotEmpty ? r.location : 'Area Hotel',
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey.shade500,
                      ),
                    ),
                  ],
                ),
                Row(
                  children: [
                    const Icon(
                      Icons.timer_outlined,
                      size: 13,
                      color: Color(0xFF135BEC),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      '${r.workHours.toStringAsFixed(1)} jam',
                      style: const TextStyle(
                        fontSize: 12,
                        color: Color(0xFF135BEC),
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    if (r.faceSimilarity != null && r.faceSimilarity! > 0) ...[
                      const SizedBox(width: 8),
                      Icon(
                        Icons.face_retouching_natural,
                        size: 13,
                        color: Colors.grey.shade400,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        'FACE VERIFIED',
                        style: TextStyle(
                          fontSize: 10,
                          color: Colors.grey.shade500,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ── Card izin/cuti/lembur (dari pengajuan APPROVED) ───────────────────────
  Widget _buildLeaveCard(AttendanceRecord r) {
    final sc = _statusColor(r.status);
    final sl = _statusLabel(r.status);
    final icon = _leaveIcon(r.leaveKategori ?? '');

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: Colors.grey.shade200),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Ikon kategori
                Container(
                  height: 44,
                  width: 44,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade100,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(icon, color: sc, size: 22),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SizedBox(height: 2),
                      Text(
                        r.leaveType ?? r.status,
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFF0F172A),
                        ),
                      ),
                      if (r.leaveType != 'Lembur' && r.leaveReason != null &&
                          r.leaveReason!.isNotEmpty) ...[
                        const SizedBox(height: 3),
                        Text(
                          r.leaveReason!,
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey.shade500,
                          ),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                      if (r.leaveType == 'Lembur') ...[
                        const SizedBox(height: 8),
                        // Row 1: Time and Total
                        Row(
                          children: [
                            _buildMiniInfo(
                              Icons.access_time_filled_rounded,
                              '${r.clockIn} - ${r.clockOut}',
                              const Color(0xFF135BEC),
                            ),
                            const SizedBox(width: 12),
                            _buildMiniInfo(
                              Icons.hourglass_bottom_rounded,
                              '${r.overtimeHours.toStringAsFixed(1)} Jam',
                              const Color(0xFF6366F1),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        // Row 2: Alasan
                        if (r.leaveReason != null && r.leaveReason!.isNotEmpty)
                          _buildMiniInfo(
                            Icons.chat_bubble_rounded,
                            r.leaveReason!,
                            Colors.grey.shade600,
                          ),
                        // Row 3: Reward
                        if (r.rewardInfo != null) ...[
                          const SizedBox(height: 6),
                          _buildMiniInfo(
                            Icons.stars_rounded,
                            r.rewardInfo!,
                            const Color(0xFFD97706),
                          ),
                        ],
                      ],
                      
                    ],
                  ),
                ),
                const SizedBox(width: 10),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 5,
                  ),
                  decoration: BoxDecoration(
                    color: sc.withOpacity(0.07),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: sc.withOpacity(0.16)),
                  ),
                  child: Text(
                    sl,
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                      color: sc.withOpacity(0.85),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ── Helpers ───────────────────────────────────────────────────────────────
  Widget _timeChip(IconData icon, String label, String value, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 12, color: color),
              const SizedBox(width: 4),
              Text(
                label,
                style: TextStyle(
                  fontSize: 9,
                  color: Colors.grey.shade600,
                  fontWeight: FontWeight.w600,
                  letterSpacing: 0.3,
                ),
              ),
            ],
          ),
          const SizedBox(height: 4),
          Text(
            value,
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.bold,
              color: const Color(0xFF0F172A),
            ),
          ),
        ],
      ),
    );
  }

  Color _statusColor(String s) {
    switch (s) {
      case 'Tepat Waktu':
        return const Color(0xFF059669);
      case 'Terlambat':
        return const Color(0xFFD97706);
      case 'Absent':
        return const Color(0xFF94A3B8);
      case 'Overtime':
        return const Color(0xFF7C3AED);
      case 'Izin':
        return const Color(0xFF2563EB);
      case 'Cuti':
        return const Color(0xFF6366F1);
      case 'Lembur':
        return const Color(0xFFD97706);
      default:
        return Colors.grey;
    }
  }

  String _statusLabel(String s) {
    switch (s) {
      case 'Tepat Waktu':
        return 'TEPAT WAKTU';
      case 'Terlambat':
        return 'TERLAMBAT';
      case 'Absent':
        return 'ABSEN';
      case 'Overtime':
        return 'LEMBUR';
      case 'Izin':
        return 'IZIN';
      case 'Cuti':
        return 'CUTI';
      case 'Lembur':
        return 'LEMBUR';
      default:
        return s.toUpperCase();
    }
  }

  IconData _leaveIcon(String k) {
    switch (k.trim().toLowerCase()) {
      case 'cuti':
        return Icons.beach_access_rounded;
      case 'lembur':
        return Icons.timelapse_rounded;
      default:
        return Icons.assignment_late_rounded;
    }
  }

  Widget _buildMiniInfo(IconData icon, String text, Color color) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 13, color: color),
        const SizedBox(width: 5),
        Flexible(
          child: Text(
            text,
            style: TextStyle(
              fontSize: 11.5,
              color: color,
              fontWeight: FontWeight.w500,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }
}
