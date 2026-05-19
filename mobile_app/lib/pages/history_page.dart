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
import 'package:mobile_app/models/assignment.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';
import 'package:flutter/services.dart';
import 'package:mobile_app/widgets/app_sidebar.dart';
import 'package:mobile_app/widgets/app_header.dart';

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
    'Penugasan',
  ];

  // Label chip → nilai AttendanceRecord.status
  static const Map<String, String?> _filterMap = {
    'Semua': null,
    'Tepat Waktu': 'Tepat Waktu',
    'Terlambat': 'Terlambat',
    'Izin': 'Izin',
    'Cuti': 'Cuti',
    'Lembur': 'Lembur',
    'Penugasan': 'Penugasan',
  };

  @override
  void initState() {
    super.initState();
    ApiService.currentUser.addListener(_syncProfile);
    _sseSubscription = SSEService().events.listen((event) {
      if (mounted) _loadData(silent: true);
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
  Future<void> _loadData({bool silent = false}) async {
    if (!silent) {
      setState(() {
        _isLoading = true;
        _error = null;
      });
    } else {
      setState(() {
        _error = null;
      });
    }
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
        ApiService.getApprovedAssignmentsByMonth(
          month: _selectedMonth.month,
          year: _selectedMonth.year,
        ),
      ]);

      final summary = results[0] as MonthlyAttendanceSummary;
      final pengajuan = results[1] as List<LeaveRequest>;
      final overtime = results[2] as List<OvertimeRequest>;
      final assignments = results[3] as List<Assignment>;

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
            // Hapus penanda "Absent/Unknown" jika ada pengajuan yang disetujui untuk hari ini
            merged.removeWhere((r) =>
                _key(r.date) == k &&
                (r.status == 'Absent' || r.status == 'Unknown'));

            final rec = AttendanceRecord.fromLeave(
              pengajuanId: p.id,
              date: cur,
              leaveType: p.type,
              leaveKategori: p.namaKategori,
              leaveReason: p.reason,
            );
            merged.add(rec);
            print('[History] + added leave record ${_fmtDate(cur)}');
            cur = cur.add(const Duration(days: 1));
          } else {
            cur = cur.add(const Duration(days: 1));
          }
        }
      }

      // Merge Overtime Requests
      for (final o in overtime) {
        if (o.date.month == _selectedMonth.month && o.date.year == _selectedMonth.year) {
          final k = _key(o.date);
          merged.removeWhere((r) =>
              _key(r.date) == k &&
              (r.status == 'Absent' || r.status == 'Unknown'));

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

      // Merge Assignments
      for (final a in assignments) {
        if (a.date.month == _selectedMonth.month && a.date.year == _selectedMonth.year) {
          final k = _key(a.date);
          merged.removeWhere((r) =>
              _key(r.date) == k &&
              (r.status == 'Absent' || r.status == 'Unknown'));

          final myEntry = a.employees.firstWhere(
            (e) => e.userId == ApiService.currentUser.value?.id,
            orElse: () => a.employees.first, // Fallback (should not happen if filtered correctly)
          );

          final rec = AttendanceRecord.fromAssignment(
            id: a.id,
            date: a.date,
            startTime: myEntry.assignedStartTime,
            endTime: myEntry.assignedEndTime,
            reason: a.reason,
            rewardInfo: null, // Assignments use replacement day off which is separate
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

    // Grouping by date for daily summary cards
    final Map<String, List<AttendanceRecord>> dailyMap = {};
    for (final r in raw) {
      final k = _key(r.date);
      dailyMap.putIfAbsent(k, () => []).add(r);
    }

    final List<MapEntry<String, List<AttendanceRecord>>> dailyGroups =
        dailyMap.entries.toList()..sort((a, b) => b.key.compareTo(a.key));

    setState(() {
      _filtered = dailyGroups;
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
  int get _cntPenugasan => _all.where((r) => r.status == 'Penugasan').length;

  String _fmt(DateTime dt, String p) =>
      _localeReady ? DateFormat(p, 'id').format(dt) : '';





  // ═════════════════════════════════════════════════════════════════════════
  // BUILD
  // ═════════════════════════════════════════════════════════════════════════
  @override
  Widget build(BuildContext context) {
    if (!_localeReady) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }
    return Scaffold(
      endDrawer: const AppSidebar(),
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: Column(
          children: [
            const AppHeader(),
            _buildMonthBanner(),
            _buildFilterChips(),
            Expanded(child: _buildBody()),
          ],
        ),
      ),
      floatingActionButton: _isLoading || _error != null
          ? null
          : FloatingActionButton.extended(
              onPressed: _exportToPDF,
              backgroundColor: const Color(0xFF135BEC),
              elevation: 4,
              icon: const Icon(
                Icons.picture_as_pdf_rounded,
                color: Colors.white,
                size: 20,
              ),
              label: const Text(
                'Laporan PDF',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.bold,
                  letterSpacing: 0.2,
                ),
              ),
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
            children: [
              Expanded(child: _stat('ONTIME', '$_cntHadir', Colors.white)),
              _vDiv(),
              Expanded(child: _stat('TELAT', '$_cntLate', const Color(0xFFFCD34D))),
              _vDiv(),
              Expanded(child: _stat('IZIN', '$_cntIzin', Colors.white)),
              _vDiv(),
              Expanded(child: _stat('CUTI', '$_cntCuti', const Color(0xFFD8B4FE))),
              _vDiv(),
              Expanded(child: _stat('LEMBUR', '$_cntLembur', const Color(0xFFFBBF24))),
              _vDiv(),
              Expanded(child: _stat('TUGAS', '$_cntPenugasan', const Color(0xFF0EA5E9))),
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
          final group = _filtered[i] as MapEntry<String, List<AttendanceRecord>>;
          return _buildDailySummaryCard(group.value);
        },
      ),
    );
  }

  // ── Daily Summary Card (Unified) ──────────────────────────────────────────
  Widget _buildDailySummaryCard(List<AttendanceRecord> records) {
    if (records.isEmpty) return const SizedBox.shrink();
    final date = records.first.date;

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Day Header inside card
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: const Color(0xFF135BEC).withOpacity(0.03),
              borderRadius: const BorderRadius.vertical(top: Radius.circular(20)),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.calendar_today_rounded,
                  size: 14,
                  color: const Color(0xFF135BEC).withOpacity(0.7),
                ),
                const SizedBox(width: 8),
                Text(
                  _fmt(date, 'EEEE, dd MMMM yyyy'),
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                    color: Color(0xFF1E293B),
                    letterSpacing: 0.1,
                  ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: records.map((r) => _buildActivityItem(r)).toList(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildActivityItem(AttendanceRecord r) {
    return r.isLeaveRecord ? _buildLeaveItem(r) : _buildAttendanceItem(r);
  }

  Widget _buildAttendanceItem(AttendanceRecord r) {
    final sc = _statusColor(r.status);
    final sl = _statusLabel(r.status);
    final hasBreak = (r.breakStart?.isNotEmpty ?? false) || (r.breakEnd?.isNotEmpty ?? false);

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Colors.grey.shade100),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              _typeBadge(Icons.fingerprint_rounded, 'ABSENSI', sc),
              const Spacer(),
              _statusChip(sl, sc),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(child: _timeSmall('Masuk', r.clockIn, const Color(0xFF2ECC71))),
              Expanded(child: _timeSmall('Pulang', r.clockOut, const Color(0xFFEF4444))),
            ],
          ),
          if (hasBreak) ...[
            const SizedBox(height: 8),
            Row(
              children: [
                Expanded(child: _timeSmall('Istirahat In', r.breakStart ?? '-', const Color(0xFFF59E0B))),
                Expanded(child: _timeSmall('Istirahat Out', r.breakEnd ?? '-', const Color(0xFF059669))),
              ],
            ),
          ],
          const SizedBox(height: 8),
          Row(
            children: [
              Icon(Icons.location_on_outlined, size: 11, color: Colors.grey.shade400),
              const SizedBox(width: 4),
              Text(
                r.location.isNotEmpty ? r.location : 'Area Hotel',
                style: TextStyle(fontSize: 10, color: Colors.grey.shade500),
              ),
              const Spacer(),
              Text(
                '${r.workHours.toStringAsFixed(1)} jam kerja',
                style: const TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: Color(0xFF135BEC)),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildLeaveItem(AttendanceRecord r) {
    final sc = _statusColor(r.status);
    final sl = _statusLabel(r.status);
    final icon = _leaveIcon(r.leaveKategori ?? '');

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: Colors.grey.shade100),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              _typeBadge(icon, (r.leaveKategori ?? r.status).toUpperCase(), sc),
              const Spacer(),
              _statusChip(sl, sc),
            ],
          ),
          const SizedBox(height: 10),
          Text(
            r.leaveType ?? r.status,
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
          ),
          if (r.leaveReason != null && r.leaveReason!.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              r.leaveReason!,
              style: TextStyle(fontSize: 11.5, color: Colors.grey.shade600),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ],
          if (r.status == 'Lembur' || r.status == 'Penugasan') ...[
            const SizedBox(height: 8),
            Row(
              children: [
                _buildMiniInfo(Icons.access_time_rounded, '${r.clockIn} - ${r.clockOut}', const Color(0xFF1E293B)),
                if (r.overtimeHours > 0) ...[
                  const SizedBox(width: 12),
                  _buildMiniInfo(Icons.timer_outlined, '${r.overtimeHours.toStringAsFixed(1)} jam', const Color(0xFF1E293B)),
                ],
              ],
            ),
            if (r.rewardInfo != null) ...[
              const SizedBox(height: 6),
              _buildMiniInfo(Icons.stars_rounded, r.rewardInfo!, const Color(0xFFD97706)),
            ],
          ],
        ],
      ),
    );
  }

  Widget _typeBadge(IconData icon, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(6)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 10, color: color),
          const SizedBox(width: 4),
          Text(label, style: TextStyle(fontSize: 9, fontWeight: FontWeight.w900, color: color, letterSpacing: 0.5)),
        ],
      ),
    );
  }

  Widget _statusChip(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(color: color.withOpacity(0.05), borderRadius: BorderRadius.circular(12), border: Border.all(color: color.withOpacity(0.1))),
      child: Text(label, style: TextStyle(fontSize: 9, fontWeight: FontWeight.w800, color: color)),
    );
  }

  Widget _timeSmall(String label, String value, Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label.toUpperCase(), style: TextStyle(fontSize: 8, color: Colors.grey.shade400, fontWeight: FontWeight.w700)),
        const SizedBox(height: 2),
        Text(value.isEmpty || value == '--:--' ? '-' : value, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Color(0xFF1E293B))),
      ],
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
      case 'Penugasan':
        return const Color(0xFF0EA5E9);
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
      case 'Penugasan':
        return 'PENUGASAN';
      default:
        return s.toUpperCase();
    }
  }

  IconData _leaveIcon(String k) {
    switch (k.trim().toLowerCase()) {
      case 'lembur':
        return Icons.timelapse_rounded;
      case 'penugasan':
        return Icons.assignment_turned_in_rounded;
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

  // ── PDF Export ────────────────────────────────────────────────────────────
  Future<void> _exportToPDF() async {
    try {
      final pdf = pw.Document();
      final monthName = _fmt(_selectedMonth, 'MMMM yyyy');
      final userName = _user?.fullName ?? 'Karyawan';
      final payroll = _user?.nik ?? '-';

      // Group activities by date
      final Map<String, List<AttendanceRecord>> grouped = {};
      for (final r in _all) {
        final key = DateFormat('yyyy-MM-dd').format(r.date);
        grouped.putIfAbsent(key, () => []).add(r);
      }
      final sortedKeys = grouped.keys.toList()..sort((a, b) => b.compareTo(a));

      final times = pw.Font.times();
      final timesBold = pw.Font.timesBold();

      pdf.addPage(
        pw.MultiPage(
          pageFormat: PdfPageFormat.a4.landscape,
          margin: const pw.EdgeInsets.all(20),
          header: (context) => pw.Column(
            children: [
              pw.Row(
                mainAxisAlignment: pw.MainAxisAlignment.center,
                children: [
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.center,
                    children: [
                      pw.Text(
                        'PT. Labersa Hutahaean',
                        style: pw.TextStyle(
                          font: timesBold,
                          fontSize: 16,
                          color: PdfColor.fromInt(0xFF988300), // Golden
                        ),
                      ),
                      pw.SizedBox(height: 2),
                      pw.Text(
                        'HEAD OFFICE - WILAYAH TOBA',
                        style: pw.TextStyle(
                          font: timesBold,
                          fontSize: 12,
                          color: PdfColor.fromInt(0xFF006400), // Dark Green
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              pw.SizedBox(height: 5),
              pw.Divider(thickness: 0.5, color: PdfColors.black),
              pw.SizedBox(height: 10),
              pw.Text(
                'LAPORAN AKTIVITAS KARYAWAN',
                style: pw.TextStyle(
                  font: timesBold,
                  fontSize: 14,
                  color: PdfColors.black,
                ),
              ),
              pw.SizedBox(height: 2),
              pw.Text(
                'Periode: $monthName',
                style: pw.TextStyle(font: times, fontSize: 11),
              ),
              pw.SizedBox(height: 15),
            ],
          ),
          footer: (context) => pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              pw.Text(
                'Dicetak pada: ${DateFormat('dd/MM/yyyy HH:mm').format(DateTime.now())}',
                style: pw.TextStyle(font: times, fontSize: 8, color: PdfColors.grey600),
              ),
              pw.Text(
                'Halaman ${context.pageNumber} dari ${context.pagesCount}',
                style: pw.TextStyle(font: times, fontSize: 8, color: PdfColors.grey600),
              ),
            ],
          ),
          build: (context) => [
            for (final dateStr in sortedKeys) ...[
              pw.Container(
                width: double.infinity,
                padding: const pw.EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: const pw.BoxDecoration(
                  color: PdfColors.grey200,
                  borderRadius: pw.BorderRadius.all(pw.Radius.circular(4)),
                ),
                child: pw.Text(
                  DateFormat('EEEE, dd MMMM yyyy', 'id').format(DateTime.parse(dateStr)),
                  style: pw.TextStyle(font: timesBold, fontSize: 10, color: PdfColors.blue900),
                ),
              ),
              pw.SizedBox(height: 4),
              pw.Table(
                columnWidths: {
                  0: const pw.FlexColumnWidth(2),
                  1: const pw.FlexColumnWidth(2),
                  2: const pw.FlexColumnWidth(2),
                  3: const pw.FlexColumnWidth(1),
                  4: const pw.FlexColumnWidth(3),
                },
                border: pw.TableBorder.all(color: PdfColors.grey300, width: 0.5),
                children: [
                  pw.TableRow(
                    decoration: const pw.BoxDecoration(color: PdfColors.grey100),
                    children: [
                      _pdfCell('Jenis Aktivitas', isHeader: true, font: timesBold),
                      _pdfCell('Masuk', isHeader: true, font: timesBold),
                      _pdfCell('Pulang', isHeader: true, font: timesBold),
                      _pdfCell('Jam', isHeader: true, font: timesBold),
                      _pdfCell('Keterangan', isHeader: true, font: timesBold),
                    ],
                  ),
                  for (final r in grouped[dateStr]!)
                    pw.TableRow(
                      children: [
                        _pdfCell(r.status, font: times),
                        _pdfCell(r.clockIn.isEmpty ? '-' : r.clockIn, font: times),
                        _pdfCell(r.clockOut.isEmpty || r.clockOut == '--:--' ? '-' : r.clockOut, font: times),
                        _pdfCell('${(r.status == 'Lembur' ? r.overtimeHours : r.workHours).toStringAsFixed(1)} h', font: times),
                        _pdfCell(r.leaveReason ?? r.rewardInfo ?? r.location ?? '-', font: times),
                      ],
                    ),
                ],
              ),
              pw.SizedBox(height: 12),
            ],
            
            pw.SizedBox(height: 20),
            pw.Divider(thickness: 1, color: PdfColors.blue900),
            pw.SizedBox(height: 10),
            pw.Text('RINGKASAN BULANAN', style: pw.TextStyle(font: timesBold, fontSize: 11, color: PdfColors.blue900)),
            pw.SizedBox(height: 8),
            pw.Table(
              columnWidths: const {
                0: pw.FractionColumnWidth(1 / 6),
                1: pw.FractionColumnWidth(1 / 6),
                2: pw.FractionColumnWidth(1 / 6),
                3: pw.FractionColumnWidth(1 / 6),
                4: pw.FractionColumnWidth(1 / 6),
                5: pw.FractionColumnWidth(1 / 6),
              },
              border: pw.TableBorder.all(color: PdfColors.grey400, width: 0.5),
              children: [
                pw.TableRow(
                  decoration: const pw.BoxDecoration(color: PdfColors.grey200),
                  children: [
                    _pdfCell('Ontime', isHeader: true, font: timesBold, align: pw.Alignment.center),
                    _pdfCell('Telat', isHeader: true, font: timesBold, align: pw.Alignment.center),
                    _pdfCell('Izin', isHeader: true, font: timesBold, align: pw.Alignment.center),
                    _pdfCell('Cuti', isHeader: true, font: timesBold, align: pw.Alignment.center),
                    _pdfCell('Lembur', isHeader: true, font: timesBold, align: pw.Alignment.center),
                    _pdfCell('Tugas', isHeader: true, font: timesBold, align: pw.Alignment.center),
                  ],
                ),
                pw.TableRow(
                  children: [
                    _pdfCell('$_cntHadir', font: times, align: pw.Alignment.center),
                    _pdfCell('$_cntLate', font: times, align: pw.Alignment.center),
                    _pdfCell('$_cntIzin', font: times, align: pw.Alignment.center),
                    _pdfCell('$_cntCuti', font: times, align: pw.Alignment.center),
                    _pdfCell('$_cntLembur', font: times, align: pw.Alignment.center),
                    _pdfCell('$_cntPenugasan', font: times, align: pw.Alignment.center),
                  ],
                ),
              ],
            ),
          ],
        ),
      );

      await Printing.layoutPdf(
        onLayout: (PdfPageFormat format) async => pdf.save(),
        name: 'Laporan_${userName}_${monthName.replaceAll(' ', '_')}.pdf',
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Gagal export PDF: $e')));
    }
  }

  pw.Widget _pdfCell(String text, {bool isHeader = false, PdfColor? color, pw.Alignment align = pw.Alignment.centerLeft, pw.Font? font}) {
    return pw.Container(
      alignment: align,
      padding: const pw.EdgeInsets.symmetric(horizontal: 4, vertical: 6),
      child: pw.Text(
        text,
        style: pw.TextStyle(
          font: font,
          fontSize: 8,
          fontWeight: isHeader ? pw.FontWeight.bold : pw.FontWeight.normal,
          color: color ?? PdfColors.black,
        ),
      ),
    );
  }
}
