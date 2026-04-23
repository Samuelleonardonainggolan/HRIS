// lib/pages/history_page.dart
import 'package:flutter/material.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/attendance_model.dart';
import 'package:mobile_app/models/leave_request.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:intl/intl.dart';
import 'dart:io';
import 'package:intl/date_symbol_data_local.dart';

class HistoryPage extends StatefulWidget {
  const HistoryPage({super.key});
  @override
  State<HistoryPage> createState() => _HistoryPageState();
}

class _HistoryPageState extends State<HistoryPage> {
  DateTime _selectedMonth = DateTime.now();
  String _selectedFilter = 'Semua';

  /// Gabungan record absensi real + sintetis dari pengajuan APPROVED
  List<AttendanceRecord> _all = [];
  List<AttendanceRecord> _filtered = [];
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
    _init();
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

  // ── Core: ambil attendance + pengajuan lalu merge ──────────────────────────
  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      // Ambil dua sumber data secara paralel
      final results = await Future.wait([
        ApiService.getMonthlyAttendance(
          month: _selectedMonth.month,
          year: _selectedMonth.year,
        ),
        ApiService.getApprovedPengajuanByMonth(
          month: _selectedMonth.month,
          year: _selectedMonth.year,
        ),
      ]);

      final summary = results[0] as MonthlyAttendanceSummary;
      final pengajuan = results[1] as List<LeaveRequest>;

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
          }
          cur = cur.add(const Duration(days: 1));
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

  // ── Filter ────────────────────────────────────────────────────────────────
  void _applyFilter() {
    final statusFilter = _filterMap[_selectedFilter];
    setState(() {
      _filtered = (_all.where((r) {
        final mOk =
            r.date.month == _selectedMonth.month &&
            r.date.year == _selectedMonth.year;
        // Filter 'Lembur' cocok dengan status 'Lembur' ATAU 'Overtime'
        bool sOk;
        if (statusFilter == null) {
          sOk = true;
        } else if (statusFilter == 'Lembur') {
          sOk = r.status == 'Lembur' || r.status == 'Overtime';
        } else {
          sOk = r.status == statusFilter;
        }
        return mOk && sOk;
      }).toList())..sort((a, b) => b.date.compareTo(a.date));
    });
  }

  void _changeMonth(int delta) {
    setState(() {
      _selectedMonth = DateTime(
        _selectedMonth.year,
        _selectedMonth.month + delta,
        1,
      );
    });
    _loadData();
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
              IconButton(
                icon: const Icon(
                  Icons.chevron_left,
                  color: Colors.white,
                  size: 22,
                ),
                onPressed: () => _changeMonth(-1),
                padding: EdgeInsets.zero,
              ),
              Column(
                children: [
                  Text(
                    'Periode Saat Ini',
                    style: TextStyle(
                      color: Colors.white.withOpacity(0.8),
                      fontSize: 11,
                    ),
                  ),
                  Text(
                    _fmt(_selectedMonth, 'MMMM yyyy'),
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              IconButton(
                icon: Icon(
                  Icons.chevron_right,
                  color: isCurrent ? Colors.white30 : Colors.white,
                  size: 22,
                ),
                onPressed: isCurrent ? null : () => _changeMonth(1),
                padding: EdgeInsets.zero,
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
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
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
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 18,
                    vertical: 9,
                  ),
                  decoration: BoxDecoration(
                    color: sel ? const Color(0xFF135BEC) : Colors.white,
                    borderRadius: BorderRadius.circular(20),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.05),
                        blurRadius: 6,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Text(
                    f,
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: sel ? Colors.white : Colors.grey.shade600,
                    ),
                  ),
                ),
              ),
            );
          }).toList(),
        ),
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
              'Tidak ada data "$_selectedFilter"\nuntuk ${_fmt(_selectedMonth, 'MMMM yyyy')}',
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
          final r = _filtered[i];
          return r.isLeaveRecord ? _buildLeaveCard(r) : _buildAttendanceCard(r);
        },
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
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 3),
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
                    Row(
                      children: [
                        Icon(
                          Icons.calendar_today_rounded,
                          size: 12,
                          color: Colors.grey.shade400,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          _fmt(r.date, 'EEE, dd MMM yyyy'),
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade500,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 3),
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
                    color: sc.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: sc.withOpacity(0.3)),
                  ),
                  child: Text(
                    sl,
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                      color: sc,
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
        border: Border.all(color: sc.withOpacity(0.25), width: 1.5),
        boxShadow: [
          BoxShadow(
            color: sc.withOpacity(0.07),
            blurRadius: 10,
            offset: const Offset(0, 3),
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
                    color: sc.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(icon, color: sc, size: 22),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Icon(
                            Icons.calendar_today_rounded,
                            size: 11,
                            color: Colors.grey.shade400,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            _fmt(r.date, 'EEE, dd MMM yyyy'),
                            style: TextStyle(
                              fontSize: 11,
                              color: Colors.grey.shade400,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 4),
                      Text(
                        r.leaveType ?? r.status,
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFF0F172A),
                        ),
                      ),
                      if (r.leaveReason != null &&
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
                      const SizedBox(height: 8),
                      // Badge disetujui
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 3,
                        ),
                        decoration: BoxDecoration(
                          color: const Color(0xFF2ECC71).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: const Color(0xFF2ECC71).withOpacity(0.3),
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(
                              Icons.check_circle_outline_rounded,
                              size: 10,
                              color: Color(0xFF2ECC71),
                            ),
                            const SizedBox(width: 4),
                            const Text(
                              'PENGAJUAN DISETUJUI',
                              style: TextStyle(
                                fontSize: 9,
                                fontWeight: FontWeight.w700,
                                color: Color(0xFF2ECC71),
                              ),
                            ),
                          ],
                        ),
                      ),
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
                    color: sc.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(color: sc.withOpacity(0.3)),
                  ),
                  child: Text(
                    sl,
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                      color: sc,
                    ),
                  ),
                ),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 9),
            decoration: BoxDecoration(
              color: sc.withOpacity(0.04),
              borderRadius: const BorderRadius.vertical(
                bottom: Radius.circular(18),
              ),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.assignment_turned_in_outlined,
                  size: 13,
                  color: sc.withOpacity(0.7),
                ),
                const SizedBox(width: 6),
                Text(
                  'Tidak hadir · Pengajuan telah disetujui',
                  style: TextStyle(
                    fontSize: 11,
                    color: sc.withOpacity(0.8),
                    fontWeight: FontWeight.w500,
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
        color: color.withOpacity(0.07),
        borderRadius: BorderRadius.circular(12),
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
                  color: color,
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
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  Color _statusColor(String s) {
    switch (s) {
      case 'Tepat Waktu':
        return const Color(0xFF2ECC71);
      case 'Terlambat':
        return const Color(0xFFF59E0B);
      case 'Absent':
        return const Color(0xFF94A3B8);
      case 'Overtime':
        return const Color(0xFF8B5CF6);
      case 'Izin':
        return const Color(0xFF135BEC);
      case 'Cuti':
        return const Color(0xFF8B5CF6);
      case 'Lembur':
        return const Color(0xFFF59E0B);
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
}
