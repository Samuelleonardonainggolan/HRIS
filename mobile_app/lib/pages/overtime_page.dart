import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:mobile_app/models/overtime_request.dart';
import 'package:mobile_app/models/assignment.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/services/sse_service.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:mobile_app/widgets/app_sidebar.dart';
import 'package:mobile_app/widgets/app_header.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';

class OvertimePage extends StatefulWidget {
  const OvertimePage({super.key});

  @override
  State<OvertimePage> createState() => _OvertimePageState();
}

class _OvertimePageState extends State<OvertimePage> {
  bool _isLoading = true;
  User? _user;
  List<OvertimeRequest> _items = [];
  List<Assignment> _assignments = [];
  List<dynamic> _combinedItems = []; // gabungan lembur + penugasan
  String _filterType = 'semua'; // semua | lembur | penugasan
  StreamSubscription? _sseSubscription;

  @override
  void initState() {
    super.initState();
    ApiService.currentUser.addListener(_syncProfile);
    _setupSSE();
    _loadData();
  }

  @override
  void dispose() {
    ApiService.currentUser.removeListener(_syncProfile);
    _sseSubscription?.cancel();
    super.dispose();
  }

  void _setupSSE() {
    _sseSubscription = SSEService().events.listen((event) {
      if (!mounted || event.type == 'ping') return;
      _loadData(silent: true);
    });
  }

  void _syncProfile() {
    if (!mounted) return;
    final currentUser = ApiService.currentUser.value;
    if (currentUser == null) return;
    setState(() => _user = currentUser);
  }

  bool _sameUserId(String a, String b) {
    return a.trim().toLowerCase() == b.trim().toLowerCase();
  }

  bool _isAssignedToUser(OvertimeRequest request, String userId) {
    return request.employees.any((e) => _sameUserId(e.userId, userId));
  }

  bool _isAssignmentAssignedToUser(Assignment assignment, String userId) {
    return assignment.employees.any((e) => _sameUserId(e.userId, userId));
  }

  bool _canUserSeeRequest(User user, OvertimeRequest request) {
    final assignedToMe = _isAssignedToUser(request, user.id);
    return assignedToMe;
  }

  bool _canUserSeeAssignment(User user, Assignment assignment) {
    final assignedToMe = _isAssignmentAssignedToUser(assignment, user.id);
    return assignedToMe;
  }

  Future<void> _loadData({bool silent = false}) async {
    if (!silent) {
      setState(() => _isLoading = true);
    }
    try {
      final user =
          ApiService.currentUser.value ?? await ApiService.getProfile();

      // Load overtime requests
      List<OvertimeRequest> overtimeData =
          await ApiService.getAssignedOvertimeRequests();
      final myOvertime = await ApiService.getMyOvertimeRequests();
      final shouldMergeMine = user.isManagerDept || user.isManagerHR;
      if (shouldMergeMine) {
        final merged = <String, OvertimeRequest>{
          for (final item in overtimeData) item.id: item,
          for (final item in myOvertime) item.id: item,
        };
        overtimeData = merged.values.toList()..sort((a, b) => b.date.compareTo(a.date));
      }

      overtimeData = overtimeData.where((item) => _canUserSeeRequest(user, item)).toList();

      // Load assignments
      List<Assignment> assignmentData =
          await ApiService.getAssignedAssignments();
      final myAssignments = await ApiService.getMyAssignments();
      if (shouldMergeMine) {
        final merged = <String, Assignment>{
          for (final item in assignmentData) item.id: item,
          for (final item in myAssignments) item.id: item,
        };
        assignmentData = merged.values.toList()..sort((a, b) => b.date.compareTo(a.date));
      }

      assignmentData = assignmentData.where((item) => _canUserSeeAssignment(user, item)).toList();

      // Combine & sort
      final combined = <dynamic>[...overtimeData, ...assignmentData]
        ..sort((a, b) {
          final dateA = (a is OvertimeRequest ? a.date : (a as Assignment).date);
          final dateB = (b is OvertimeRequest ? b.date : (b as Assignment).date);
          return dateB.compareTo(dateA);
        });

      if (!mounted) return;
      setState(() {
        _user = user;
        _items = overtimeData;
        _assignments = assignmentData;
        _combinedItems = combined;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _isLoading = false);
    }
  }

  Future<void> _agree(dynamic item) async {
    try {
      if (item is OvertimeRequest) {
        await ApiService.agreeOvertimeRequest(item.id);
      } else if (item is Assignment) {
        await ApiService.agreeAssignment(item.id);
      }
      if (!mounted) return;
      _showSnackBar('Anda menyetujui pengajuan', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal setuju: $e', isError: true);
    }
  }

  Future<void> _reject(dynamic item) async {
    final noteCtrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text(
          'Tolak Pengajuan',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        content: TextField(
          controller: noteCtrl,
          maxLines: 3,
          decoration: InputDecoration(
            hintText: 'Tulis alasan penolakan...',
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            filled: true,
            fillColor: Colors.grey.shade50,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text('Batal', style: TextStyle(color: Colors.grey.shade600)),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFFEF4444),
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: const Text('Tolak'),
          ),
        ],
      ),
    );

    if (ok != true) return;

    try {
      if (item is OvertimeRequest) {
        await ApiService.rejectOvertimeRequest(
          item.id,
          rejectionNote: noteCtrl.text.trim(),
        );
      } else if (item is Assignment) {
        await ApiService.rejectAssignment(
          item.id,
          rejectionNote: noteCtrl.text.trim(),
        );
      }
      if (!mounted) return;
      _showSnackBar('Pengajuan ditolak', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal menolak: $e', isError: true);
    }
  }

  String _greeting() {
    final hour = DateTime.now().hour;
    if (hour < 12) return 'Selamat Pagi';
    if (hour < 15) return 'Selamat Siang';
    if (hour < 18) return 'Selamat Sore';
    return 'Selamat Malam';
  }

  String _avatarUrl() {
    final avatar = (_user?.avatar ?? '').trim();
    if (avatar.isNotEmpty) return avatar;
    final name = Uri.encodeComponent(_user?.fullName ?? 'Employee');
    return 'https://ui-avatars.com/api/?name=$name&background=135BEC&color=fff&size=100';
  }

  Widget _avatarPreview({double size = 48}) {
    final avatar = (_user?.avatar ?? '').trim();
    if (avatar.isNotEmpty) {
      return Image.network(
        avatar,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) =>
            const Icon(Icons.person, color: Color(0xFF135BEC), size: 26),
      );
    }

    return Image.network(
      _avatarUrl(),
      fit: BoxFit.cover,
      errorBuilder: (_, __, ___) =>
          const Icon(Icons.person, color: Color(0xFF135BEC), size: 26),
    );
  }



  void _showSnackBar(String message, {bool isError = false}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: isError
            ? const Color(0xFFEF4444)
            : const Color(0xFF135BEC),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      endDrawer: const AppSidebar(),
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: Column(
          children: [
            const AppHeader(),
            _buildFilterBar(),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _user == null
                  ? _buildLoadFailedState()
                  : RefreshIndicator(
                      onRefresh: _loadData,
                      color: const Color(0xFF135BEC),
                      child: _getFilteredItems().isEmpty ? _buildEmptyState() : _buildList(),
                    ),
            ),
          ],
        ),
      ),
    );
  }

  List<dynamic> _getFilteredItems() {
    switch (_filterType) {
      case 'lembur':
        return _items;
      case 'penugasan':
        return _assignments;
      default:
        return _combinedItems;
    }
  }

  Widget _buildFilterBar() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
      child: Row(
        children: [
          Expanded(
            child: _filterChip('Semua', 'semua'),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _filterChip('Lembur', 'lembur'),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _filterChip('Penugasan', 'penugasan'),
          ),
        ],
      ),
    );
  }

  Widget _filterChip(String label, String value) {
    final isSelected = _filterType == value;
    return GestureDetector(
      onTap: () {
        setState(() => _filterType = value);
        // Clear badge on tap
        if (value == 'lembur') {
          SSEService().hasNewOvertime.value = false;
        } else if (value == 'penugasan') {
          SSEService().hasNewAssignment.value = false;
        }
      },
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 9),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFF135BEC) : Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: isSelected ? const Color(0xFF135BEC) : const Color(0xFFE2E8F0),
          ),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              label,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: isSelected ? Colors.white : const Color(0xFF475569),
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
            // Show dot if new data
            if (value == 'lembur' || value == 'penugasan')
              ValueListenableBuilder<bool>(
                valueListenable: value == 'lembur' 
                  ? SSEService().hasNewOvertime 
                  : SSEService().hasNewAssignment,
                builder: (context, hasNew, _) {
                  if (!hasNew) return const SizedBox.shrink();
                  return Container(
                    margin: const EdgeInsets.only(left: 6),
                    width: 6,
                    height: 6,
                    decoration: BoxDecoration(
                      color: isSelected ? Colors.white : const Color(0xFFEF4444),
                      shape: BoxShape.circle,
                    ),
                  );
                },
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      children: [
        const SizedBox(height: 100),
        Center(
          child: Column(
            children: [
              Container(
                height: 200,
                width: 200,
                decoration: const BoxDecoration(
                  color: Color(0xFFF1F5F9),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.history_toggle_off,
                  size: 80,
                  color: Color(0xFF94A3B8),
                ),
              ),
              const SizedBox(height: 24),
              const Text(
                'Belum Ada Pengajuan Lembur',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF475569),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Pengajuan lembur untuk Anda akan tampil di sini.',
                style: TextStyle(color: Colors.grey.shade500),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildLoadFailedState() {
    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      children: [
        const SizedBox(height: 100),
        Center(
          child: Column(
            children: [
              const Icon(
                Icons.error_outline,
                size: 72,
                color: Color(0xFF94A3B8),
              ),
              const SizedBox(height: 16),
              const Text(
                'Gagal Memuat Data Lembur',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF475569),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Tarik ke bawah untuk mencoba lagi.',
                style: TextStyle(color: Colors.grey.shade500),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildList() {
    final items = _getFilteredItems();
    return ListView.builder(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 90),
      itemCount: items.length,
      itemBuilder: (context, index) {
        final item = items[index];
        if (item is OvertimeRequest) {
          return _OvertimeCard(
            item: item,
            user: _user!,
            onTap: () => _showDetail(item),
            onAgree: () => _agree(item),
            onReject: () => _reject(item),
          );
        } else if (item is Assignment) {
          return _AssignmentCard(
            item: item,
            user: _user!,
            onTap: () => _showDetailAssignment(item),
            onAgree: () => _agree(item),
            onReject: () => _reject(item),
            onUseDayOff: () => _useDayOff(item),
          );
        }
        return const SizedBox.shrink();
      },
    );
  }

  Future<void> _claimOvertimeReward(OvertimeRequest r, String rewardType, String? rewardDate, String? rewardOption) async {
    setState(() => _isLoading = true);
    try {
      await ApiService.claimOvertimeReward(r.id, rewardType, rewardDate: rewardDate, rewardOption: rewardOption);
      String typeLabel = rewardType == 'money' ? 'Uang' : 'Potong Jam Kerja';
      _showSnackBar('Reward $typeLabel berhasil dipilih');
      await _loadData();
    } catch (e) {
      _showSnackBar(e.toString(), isError: true);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _showDetail(OvertimeRequest r) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => _OvertimeDetailSheet(
        request: r,
        user: _user!,
        onAgree: () {
          Navigator.pop(context);
          _agree(r);
        },
        onReject: () {
          Navigator.pop(context);
          _reject(r);
        },
        onClaimReward: (rewardType, rewardDate, rewardOption) {
          Navigator.pop(context);
          _claimOvertimeReward(r, rewardType, rewardDate, rewardOption);
        },
      ),
    );
  }

  void _showDetailAssignment(Assignment a) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => _AssignmentDetailSheet(
        assignment: a,
        user: _user!,
        onAgree: () {
          Navigator.pop(context);
          _agree(a);
        },
        onReject: () {
          Navigator.pop(context);
          _reject(a);
        },
        onUseDayOff: () {
          Navigator.pop(context);
          _useDayOff(a);
        },
      ),
    );
  }

  Future<void> _useDayOff(Assignment a) async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: DateTime.now().add(const Duration(days: 1)),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
      locale: const Locale('id', 'ID'),
      helpText: 'Pilih Hari Libur Pengganti',
      cancelText: 'Batal',
      confirmText: 'Pilih',
    );

    if (picked == null) return;

    setState(() => _isLoading = true);
    try {
      final dateStr = DateFormat('yyyy-MM-dd').format(picked);
      await ApiService.useReplacementDayOff(a.id, dateStr);
      _showSnackBar('Hari libur pengganti berhasil ditetapkan');
      await _loadData();
    } catch (e) {
      _showSnackBar(e.toString(), isError: true);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }
}

class _OvertimeCard extends StatelessWidget {
  final OvertimeRequest item;
  final User user;
  final VoidCallback onTap;
  final VoidCallback onAgree;
  final VoidCallback onReject;

  const _OvertimeCard({
    required this.item,
    required this.user,
    required this.onTap,
    required this.onAgree,
    required this.onReject,
  });

  @override
  Widget build(BuildContext context) {
    final statusColor = _getStatusColor(item.status);
    final dateText = DateFormat('dd MMM yyyy', 'id').format(item.date);
    final durationText = '${item.getDurationHours().toStringAsFixed(1)} Jam';
    final participantSummary = _buildParticipantSummary();

    return GestureDetector(
      onTap: onTap,
      child: Container(
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
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Konten tengah
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Tanggal diajukan
                    Row(
                      children: [
                        Icon(
                          Icons.calendar_today_rounded,
                          size: 11,
                          color: Colors.grey.shade400,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          dateText,
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade400,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    // Judul / alasan lembur
                    Text(
                      item.reason.isNotEmpty ? item.reason : 'Penugasan Lembur',
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 6),
                    // Waktu lembur
                    Row(
                      children: [
                        const Icon(
                          Icons.access_time_filled_rounded,
                          size: 13,
                          color: Color(0xFF135BEC),
                        ),
                        const SizedBox(width: 5),
                        Text(
                          '${item.startTime} - ${item.endTime}',
                          style: const TextStyle(
                            fontSize: 12,
                            color: Color(0xFF135BEC),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    // Approval chain peserta
                    _buildParticipantPills(),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              // Kolom kanan: status + durasi
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      'Lembur',
                      style: const TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF8B5CF6),
                      ),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      durationText,
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildParticipantPills() {
    final myEntryList = item.employees.where(
      (e) => e.userId.trim().toLowerCase() == user.id.trim().toLowerCase(),
    );
    final myEntry = myEntryList.isNotEmpty ? myEntryList.first : null;

    return Wrap(
      spacing: 6,
      runSpacing: 4,
      children: [
        _miniPill(
          Icons.people_alt_rounded,
          _buildParticipantSummary(),
          const Color(0xFF10B981),
        ),
        _miniPill(
          Icons.shield_rounded,
          myEntry != null ? myEntry.statusDisplay : '-',
          _getEmployeeStatusColor(myEntry?.employeeStatus ?? ''),
        ),
      ],
    );
  }

  Widget _miniPill(IconData icon, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withOpacity(0.28)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 11, color: color),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              color: color,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }

  Color _getEmployeeStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return const Color(0xFFF59E0B);
      case 'agreed':
        return const Color(0xFF10B981);
      case 'rejected':
        return const Color(0xFFEF4444);
      default:
        return const Color(0xFF64748B);
    }
  }

  String _buildParticipantSummary() {
    if (item.employees.isEmpty) return 'Belum ada peserta';

    final names = item.employees.map((e) => e.displayName).toList();
    if (names.length <= 2) return names.join(', ');

    return '${names.take(2).join(', ')} +${names.length - 2}';
  }

  Color _getStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'draft':
        return const Color(0xFF94A3B8);
      case 'submitted':
        return const Color(0xFFF59E0B);
      case 'published':
        return const Color(0xFF10B981);
      default:
        return const Color(0xFF64748B);
    }
  }
}

class _OvertimeDetailSheet extends StatelessWidget {
  final OvertimeRequest request;
  final User user;
  final VoidCallback onAgree;
  final VoidCallback onReject;
  final void Function(String rewardType, String? rewardDate, String? rewardOption) onClaimReward;

  const _OvertimeDetailSheet({
    required this.request,
    required this.user,
    required this.onAgree,
    required this.onReject,
    required this.onClaimReward,
  });

  @override
  Widget build(BuildContext context) {
    final myEntry = _findMyEntry();
    final canRespond =
        request.isSubmitted && myEntry != null && myEntry.isPending;
    final participantCount = request.employees.length;

    return DraggableScrollableSheet(
      initialChildSize: 0.6,
      minChildSize: 0.4,
      maxChildSize: 0.9,
      builder: (_, controller) => Container(
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: ListView(
          controller: controller,
          padding: const EdgeInsets.all(20),
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey.shade300,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 20),
            Row(
              children: [
                Container(
                  height: 46,
                  width: 46,
                  decoration: BoxDecoration(
                    color: const Color(0xFF135BEC).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(
                    Icons.schedule_outlined,
                    color: Color(0xFF135BEC),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Detail Penugasan Lembur',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFF0F172A),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        DateFormat(
                          'EEEE, dd MMMM yyyy',
                          'id',
                        ).format(request.date),
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade600,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                _statusPill(
                  'Status ${request.statusDisplay}',
                  _getRequestStatusColor(request.status),
                ),
                _statusPill(
                  'Durasi ${request.getDurationHours().toStringAsFixed(1)} jam',
                  const Color(0xFF0F766E),
                ),
                _statusPill(
                  'Peserta $participantCount orang',
                  const Color(0xFF135BEC),
                ),
              ],
            ),
            const SizedBox(height: 20),
            _buildInfoRow(
              Icons.calendar_today_rounded,
              'Tanggal',
              DateFormat('dd MMM yyyy', 'id').format(request.date),
            ),
            const SizedBox(height: 12),
            _buildInfoRow(
              Icons.access_time_filled,
              'Waktu',
              '${request.startTime} - ${request.endTime} (${request.getDurationHours().toStringAsFixed(1)} Jam)',
            ),
            const SizedBox(height: 12),
            _buildInfoRow(
              Icons.business_center,
              'Departemen',
              request.departmentName,
            ),
            const SizedBox(height: 12),
            _buildInfoRow(Icons.person, 'Pengaju', request.requestedByName),
            const SizedBox(height: 16),
            const Text(
              'Alasan Penugasan',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Color(0xFF0F172A),
              ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: const Color(0xFFE2E8F0)),
              ),
              child: Text(
                request.reason.isNotEmpty
                    ? request.reason
                    : 'Tidak ada alasan yang disertakan.',
                style: const TextStyle(
                  fontSize: 12,
                  color: Color(0xFF334155),
                  height: 1.4,
                ),
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Peserta Lembur',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Color(0xFF0F172A),
              ),
            ),
            const SizedBox(height: 10),
            if (request.employees.isEmpty)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF8FAFC),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFE2E8F0)),
                ),
                child: const Text(
                  'Belum ada peserta yang tercatat.',
                  style: TextStyle(fontSize: 12, color: Color(0xFF475569)),
                ),
              )
            else
              ...request.employees.map(_buildParticipantRow),
            if (myEntry != null &&
                myEntry.letterUrl != null &&
                myEntry.letterUrl!.isNotEmpty) ...[
              const SizedBox(height: 20),
              const Text(
                'Surat Perintah Kerja Lembur (SPKL)',
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF0F172A),
                ),
              ),
              const SizedBox(height: 10),
              _buildDocumentTile(context, myEntry.letterUrl!),
            ],
            const SizedBox(height: 20),
            _buildRewardSection(context),
            const SizedBox(height: 24),
            if (canRespond)
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: onReject,
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        side: const BorderSide(color: Color(0xFFEF4444)),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(10),
                        ),
                      ),
                      child: const Text(
                        'Tolak',
                        style: TextStyle(
                          color: Color(0xFFEF4444),
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: onAgree,
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        backgroundColor: const Color(0xFF10B981),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(10),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        'Setujui',
                        style: TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  OvertimeEmployee? _findMyEntry() {
    for (final employee in request.employees) {
      if (employee.userId.trim().toLowerCase() ==
          user.id.trim().toLowerCase()) {
        return employee;
      }
    }
    return null;
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 18, color: const Color(0xFF64748B)),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: const TextStyle(fontSize: 11, color: Color(0xFF94A3B8)),
              ),
              Text(
                value,
                style: const TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF1E293B),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildParticipantRow(OvertimeEmployee employee) {
    final isMe =
        employee.userId.trim().toLowerCase() == user.id.trim().toLowerCase();
    final statusColor = _getEmployeeStatusColor(employee.employeeStatus);

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isMe ? const Color(0xFFF0F9FF) : const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isMe ? const Color(0xFFBAE6FD) : const Color(0xFFE2E8F0),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            height: 34,
            width: 34,
            decoration: BoxDecoration(
              color: statusColor.withOpacity(0.12),
              shape: BoxShape.circle,
            ),
            child: Icon(
              isMe ? Icons.person_rounded : Icons.groups_rounded,
              color: statusColor,
              size: 18,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        isMe
                            ? '${employee.displayName} (Anda)'
                            : employee.displayName,
                        style: const TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF0F172A),
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    const SizedBox(width: 8),
                    _statusPill(employee.statusDisplay, statusColor),
                  ],
                ),
                if (employee.rejectionNote != null &&
                    employee.rejectionNote!.trim().isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Text(
                    employee.rejectionNote!.trim(),
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFF475569),
                      height: 1.35,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDocumentTile(BuildContext context, String url) {
    return GestureDetector(
      onTap: () => _launchURL(context, url),
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: const Color(0xFFF0F9FF),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xFFBAE6FD)),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xFF0284C7).withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.picture_as_pdf_rounded,
                color: Color(0xFF0284C7),
                size: 18,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Surat Perintah Kerja Lembur',
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12),
                  ),
                  const Text(
                    'Ketuk untuk membuka dokumen',
                    style: TextStyle(fontSize: 11, color: Color(0xFF0369A1)),
                  ),
                ],
              ),
            ),
            IconButton(
              icon: const Icon(
                Icons.open_in_new_rounded,
                color: Color(0xFF0284C7),
                size: 18,
              ),
              onPressed: () => _launchURL(context, url),
              padding: EdgeInsets.zero,
            ),
          ],
        ),
      ),
    );
  }

  Widget _statusPill(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.25)),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.w700,
          color: color,
        ),
      ),
    );
  }

  Color _getRequestStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'draft':
        return const Color(0xFF94A3B8);
      case 'submitted':
        return const Color(0xFFF59E0B);
      case 'published':
        return const Color(0xFF10B981);
      default:
        return const Color(0xFF64748B);
    }
  }

  Color _getEmployeeStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return const Color(0xFFF59E0B);
      case 'agreed':
        return const Color(0xFF10B981);
      case 'rejected':
        return const Color(0xFFEF4444);
      default:
        return const Color(0xFF64748B);
    }
  }

  Widget _buildRewardSection(BuildContext context) {
    final myEntry = _findMyEntry();
    if (myEntry == null) return const SizedBox.shrink();

    final reward = myEntry.reward;
    final hasReward = reward != null && reward.rewardType.isNotEmpty && reward.rewardType != 'none';

    if (!hasReward) {
      final canChoose = (request.status == 'submitted' || request.status == 'published') && myEntry.isAgreed;

      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Reward Lembur',
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.bold,
              color: Color(0xFF0F172A),
            ),
          ),
          const SizedBox(height: 10),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.grey.shade100,
                        shape: BoxShape.circle,
                      ),
                      child: Icon(
                        Icons.card_giftcard_rounded,
                        color: Colors.grey.shade500,
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Reward belum dipilih',
                            style: TextStyle(
                              fontSize: 13,
                              fontWeight: FontWeight.bold,
                              color: Color(0xFF1E293B),
                            ),
                          ),
                          Text(
                            canChoose ? 'Silakan pilih jenis reward di bawah' : 'Belum dapat memilih reward',
                            style: TextStyle(
                              fontSize: 12,
                              color: Color(0xFF64748B),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                if (canChoose) ...[
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: () => _confirmClaim(context, 'money'),
                          icon: const Icon(Icons.payments_rounded, size: 16),
                          label: const Text('Uang'),
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.teal,
                            side: const BorderSide(color: Colors.teal),
                            padding: const EdgeInsets.symmetric(vertical: 10),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(10),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: ElevatedButton.icon(
                          onPressed: () => _confirmClaim(context, 'time_off'),
                          icon: const Icon(Icons.timer_rounded, size: 16),
                          label: const Text('Jam Kerja'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blue,
                            foregroundColor: Colors.white,
                            elevation: 0,
                            padding: const EdgeInsets.symmetric(vertical: 10),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(10),
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
        ],
      );
    }

    final status = reward.status.toLowerCase();
    Color statusColor;
    String statusText;

    switch (status) {
      case 'granted':
        statusColor = const Color(0xFF10B981);
        statusText = 'Disetujui';
        break;
      case 'used':
        statusColor = const Color(0xFF135BEC);
        statusText = 'Sudah Digunakan';
        break;
      case 'pending':
        statusColor = const Color(0xFFF59E0B);
        statusText = 'Menunggu Persetujuan';
        break;
      default:
        statusColor = const Color(0xFF64748B);
        statusText = 'Belum Diproses';
    }

    final isTimeOff = reward.rewardType == 'time_off';
    String optionText = '';
    if (isTimeOff) {
      if (reward.rewardOption == 'early_out') {
        optionText = ' (Pulang Cepat)';
      } else if (reward.rewardOption == 'late_in') {
        optionText = ' (Masuk Terlambat)';
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Reward Lembur',
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: Color(0xFF0F172A),
          ),
        ),
        const SizedBox(height: 10),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: statusColor.withOpacity(0.05),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: statusColor.withOpacity(0.2)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      reward.rewardType == 'money' ? Icons.payments_rounded : Icons.timer_rounded,
                      color: statusColor,
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${reward.rewardTypeDisplay}$optionText',
                          style: const TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF1E293B),
                          ),
                        ),
                        Text(
                          statusText,
                          style: TextStyle(
                            fontSize: 12,
                            color: statusColor,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (isTimeOff && reward.rewardDate != null) ...[
                const SizedBox(height: 12),
                const Divider(),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Tanggal Pengurangan:',
                      style: TextStyle(fontSize: 12, color: Color(0xFF64748B)),
                    ),
                    Text(
                      DateFormat('EEEE, dd MMM yyyy', 'id').format(reward.rewardDate!),
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                  ],
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  void _confirmClaim(BuildContext context, String type) async {
    if (type == 'time_off') {
      final picked = await showDatePicker(
        context: context,
        initialDate: DateTime.now().add(const Duration(days: 1)),
        firstDate: DateTime.now(),
        lastDate: DateTime.now().add(const Duration(days: 90)),
        locale: const Locale('id', 'ID'),
        helpText: 'Pilih Tanggal Reward',
        cancelText: 'Batal',
        confirmText: 'Pilih',
      );
      if (picked != null) {
        final dateStr = DateFormat('yyyy-MM-dd').format(picked);
        if (!context.mounted) return;
        showDialog(
          context: context,
          builder: (ctx) => AlertDialog(
            title: const Text('Pilih Opsi Pengurangan'),
            content: const Text('Pilih bagaimana Anda ingin menggunakan pengurangan jam kerja ini:'),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.pop(ctx);
                  _showFinalConfirmation(context, type, dateStr, 'late_in');
                },
                child: const Text('Masuk Terlambat'),
              ),
              ElevatedButton(
                onPressed: () {
                  Navigator.pop(ctx);
                  _showFinalConfirmation(context, type, dateStr, 'early_out');
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF135BEC),
                  foregroundColor: Colors.white,
                ),
                child: const Text('Pulang Cepat'),
              ),
            ],
          ),
        );
      }
      return;
    }

    _showFinalConfirmation(context, 'money', null, null);
  }

  void _showFinalConfirmation(BuildContext context, String type, String? dateStr, String? option) {
    String rewardName = type == 'money' ? 'Uang Lembur' : 'Potong Jam Kerja';
    if (option == 'early_out') rewardName += ' (Pulang Cepat)';
    if (option == 'late_in') rewardName += ' (Masuk Terlambat)';

    final dateInfo = dateStr != null ? '\nTanggal Klaim: ${DateFormat('dd MMM yyyy', 'id').format(DateTime.parse(dateStr))}' : '';

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Konfirmasi Reward'),
        content: Text('Anda akan mengklaim:\n\n$rewardName\nUntuk lembur tanggal ${DateFormat('dd MMM yyyy', 'id').format(request.date)}.$dateInfo\n\nLanjutkan?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(ctx);
              onClaimReward(type, dateStr, option);
            },
            child: const Text('Klaim'),
          ),
        ],
      ),
    );
  }

  Future<void> _launchURL(BuildContext context, String url) async {
    try {
      final uri = Uri.parse(url);
      // Buka di browser eksternal agar tidak langsung download
      final launched = await launchUrl(
        uri,
        mode: LaunchMode.externalApplication,
      );
      if (!launched) {
        // Fallback ke in-app web view jika browser tidak tersedia
        await launchUrl(uri, mode: LaunchMode.inAppWebView);
      }
    } catch (e) {
      debugPrint('Error launching URL: $e');
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Gagal membuka dokumen')),
        );
      }
    }
  }
}

class _AssignmentCard extends StatelessWidget {
  final Assignment item;
  final User user;
  final VoidCallback onTap;
  final VoidCallback onAgree;
  final VoidCallback onReject;
  final VoidCallback onUseDayOff;

  const _AssignmentCard({
    required this.item,
    required this.user,
    required this.onTap,
    required this.onAgree,
    required this.onReject,
    required this.onUseDayOff,
  });

  @override
  Widget build(BuildContext context) {
    final statusColor = _getStatusColor(item.status);
    final dateText = DateFormat('dd MMM yyyy', 'id').format(item.date);
    final participantSummary = _buildParticipantSummary();
    
    final myEntryList = item.employees.where(
      (e) => e.userId.trim().toLowerCase() == user.id.trim().toLowerCase(),
    );
    final myEntry = myEntryList.isNotEmpty ? myEntryList.first : null;

    return GestureDetector(
      onTap: onTap,
      child: Container(
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
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Icon kategori penugasan (berbeda dari lembur)
              const SizedBox.shrink(),
              // Konten tengah
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Tanggal diajukan
                    Row(
                      children: [
                        Icon(
                          Icons.calendar_today_rounded,
                          size: 11,
                          color: Colors.grey.shade400,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          dateText,
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade400,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    // Judul / alasan penugasan
                    Text(
                      item.reason.isNotEmpty ? item.reason : 'Penugasan',
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 6),
                    // Waktu penugasan
                    Row(
                      children: [
                        const Icon(
                          Icons.access_time_filled_rounded,
                          size: 13,
                          color: Color(0xFF135BEC),
                        ),
                        const SizedBox(width: 5),
                        Text(
                          '${item.startTime} - ${item.endTime}',
                          style: const TextStyle(
                            fontSize: 12,
                            color: Color(0xFF135BEC),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    // Approval chain peserta
                    _buildParticipantPills(),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              // Kolom kanan: status
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      'Penugasan',
                      style: const TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF8B5CF6),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildParticipantPills() {
    final myEntryList = item.employees.where(
      (e) => e.userId.trim().toLowerCase() == user.id.trim().toLowerCase(),
    );
    final myEntry = myEntryList.isNotEmpty ? myEntryList.first : null;

    return Wrap(
      spacing: 6,
      runSpacing: 4,
      children: [
        _miniPill(
          Icons.people_alt_rounded,
          _buildParticipantSummary(),
          const Color(0xFF10B981),
        ),
        _miniPill(
          Icons.shield_rounded,
          myEntry != null ? _getEmployeeStatusDisplay(myEntry.employeeStatus) : '-',
          _getEmployeeStatusColor(myEntry?.employeeStatus ?? ''),
        ),
      ],
    );
  }

  Widget _miniPill(IconData icon, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withOpacity(0.28)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 11, color: color),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 10,
              color: color,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }

  Color _getEmployeeStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return const Color(0xFFF59E0B);
      case 'agreed':
        return const Color(0xFF10B981);
      case 'rejected':
        return const Color(0xFFEF4444);
      default:
        return const Color(0xFF64748B);
    }
  }

  String _getEmployeeStatusDisplay(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return 'Menunggu';
      case 'agreed':
        return 'Disetujui';
      case 'rejected':
        return 'Ditolak';
      default:
        return status;
    }
  }

  String _buildParticipantSummary() {
    if (item.employees.isEmpty) return 'Belum ada peserta';

    final names = item.employees.map((e) => e.fullName).toList();
    if (names.length <= 2) return names.join(', ');

    return '${names.take(2).join(', ')} +${names.length - 2}';
  }

  Color _getStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'draft':
        return const Color(0xFF94A3B8);
      case 'submitted':
        return const Color(0xFFF59E0B);
      case 'published':
        return const Color(0xFF10B981);
      default:
        return const Color(0xFF64748B);
    }
  }
}

class _AssignmentDetailSheet extends StatelessWidget {
  final Assignment assignment;
  final User user;
  final VoidCallback onAgree;
  final VoidCallback onReject;
  final VoidCallback onUseDayOff;

  const _AssignmentDetailSheet({
    required this.assignment,
    required this.user,
    required this.onAgree,
    required this.onReject,
    required this.onUseDayOff,
  });

  @override
  Widget build(BuildContext context) {
    final myEntry = _findMyEntry();
    final canRespond =
        assignment.isSubmitted && myEntry != null && myEntry.employeeStatus == 'pending';
    final participantCount = assignment.employees.length;

    return DraggableScrollableSheet(
      initialChildSize: 0.6,
      minChildSize: 0.4,
      maxChildSize: 0.9,
      builder: (_, controller) => Container(
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: ListView(
          controller: controller,
          padding: const EdgeInsets.all(20),
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey.shade300,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 20),
            Row(
              children: [
                Container(
                  height: 46,
                  width: 46,
                  decoration: BoxDecoration(
                    color: const Color(0xFF8B5CF6).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(
                    Icons.assignment_rounded,
                    color: Color(0xFF8B5CF6),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Detail Penugasan',
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Color(0xFF0F172A),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        DateFormat(
                          'EEEE, dd MMMM yyyy',
                          'id',
                        ).format(assignment.date),
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade600,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                _statusPill(
                  'Status ${assignment.statusDisplay}',
                  _getRequestStatusColor(assignment.status),
                ),
                _statusPill(
                  'Peserta $participantCount orang',
                  const Color(0xFF135BEC),
                ),
              ],
            ),
            const SizedBox(height: 20),
            _buildInfoRow(
              Icons.calendar_today_rounded,
              'Tanggal',
              DateFormat('dd MMM yyyy', 'id').format(assignment.date),
            ),
            const SizedBox(height: 12),
            _buildInfoRow(
              Icons.access_time_filled,
              'Waktu',
              '${assignment.startTime} - ${assignment.endTime}',
            ),
            const SizedBox(height: 12),
            _buildInfoRow(
              Icons.business_center,
              'Departemen',
              assignment.departmentName,
            ),
            const SizedBox(height: 12),
            _buildInfoRow(Icons.person, 'Pengaju', assignment.requestedByName),
            const SizedBox(height: 16),
            const Text(
              'Alasan Penugasan',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Color(0xFF0F172A),
              ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: const Color(0xFFE2E8F0)),
              ),
              child: Text(
                assignment.reason.isNotEmpty
                    ? assignment.reason
                    : 'Tidak ada alasan yang disertakan.',
                style: const TextStyle(
                  fontSize: 12,
                  color: Color(0xFF334155),
                  height: 1.4,
                ),
              ),
            ),
            const SizedBox(height: 20),
            const Text(
              'Peserta Penugasan',
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Color(0xFF0F172A),
              ),
            ),
            const SizedBox(height: 10),
            if (assignment.employees.isEmpty)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFF8FAFC),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFE2E8F0)),
                ),
                child: const Text(
                  'Belum ada peserta yang tercatat.',
                  style: TextStyle(fontSize: 12, color: Color(0xFF475569)),
                ),
              )
            else
              ...assignment.employees.map(_buildParticipantRow),
            const SizedBox(height: 24),
            _buildRewardSection(context),
            const SizedBox(height: 24),
            if (canRespond)
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: onReject,
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        side: const BorderSide(color: Color(0xFFEF4444)),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(10),
                        ),
                      ),
                      child: const Text(
                        'Tolak',
                        style: TextStyle(
                          color: Color(0xFFEF4444),
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: onAgree,
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 12),
                        backgroundColor: const Color(0xFF10B981),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(10),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        'Setujui',
                        style: TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  Widget _buildRewardSection(BuildContext context) {
    final myEntry = _findMyEntry();
    if (myEntry == null) return const SizedBox.shrink(); // Always show if participant

    final status = myEntry.dayOffStatus.toLowerCase();
    Color statusColor;
    String statusText;

    switch (status) {
      case 'granted':
        statusColor = const Color(0xFF10B981);
        statusText = 'Tersedia';
        break;
      case 'used':
        statusColor = const Color(0xFF135BEC);
        statusText = 'Sudah Digunakan';
        break;
      case 'pending':
        statusColor = const Color(0xFFF59E0B);
        statusText = 'Menunggu Konfirmasi';
        break;
      default:
        statusColor = const Color(0xFF64748B);
        statusText = 'Belum Diproses';
    }

    final isUsed = status == 'used';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Day Off Reward',
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: Color(0xFF0F172A),
          ),
        ),
        const SizedBox(height: 10),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: statusColor.withOpacity(0.05),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: statusColor.withOpacity(0.2)),
          ),
          child: Column(
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      Icons.card_giftcard_rounded,
                      color: statusColor,
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Reward Hari Libur Pengganti',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF1E293B),
                          ),
                        ),
                        Text(
                          statusText,
                          style: TextStyle(
                            fontSize: 12,
                            color: statusColor,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (isUsed && myEntry.replacementOffDate != null) ...[
                const SizedBox(height: 12),
                const Divider(),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Tanggal Pengganti:',
                      style: TextStyle(fontSize: 12, color: Color(0xFF64748B)),
                    ),
                    Text(
                      DateFormat('EEEE, dd MMM yyyy', 'id').format(myEntry.replacementOffDate!),
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton.icon(
                    onPressed: () => _exportToPDF(context),
                    icon: const Icon(Icons.picture_as_pdf_rounded, size: 16),
                    label: const Text('Download Surat Penugasan'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFF135BEC),
                      foregroundColor: Colors.white,
                      elevation: 0,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                  ),
                ),
              ],
              if (!isUsed) ...[
                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: myEntry.employeeStatus == 'agreed' ? onUseDayOff : null,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: myEntry.employeeStatus == 'agreed' 
                          ? (status == 'granted' ? statusColor : const Color(0xFF8B5CF6))
                          : Colors.grey.shade400,
                      foregroundColor: Colors.white,
                      elevation: 0,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                    child: Text(
                      myEntry.employeeStatus == 'agreed' ? 'Pilih Hari Libur Pengganti' : 'Setujui Penugasan Terlebih Dahulu',
                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  AssignmentEmployee? _findMyEntry() {
    for (final employee in assignment.employees) {
      if (employee.userId.trim().toLowerCase() ==
          user.id.trim().toLowerCase()) {
        return employee;
      }
    }
    return null;
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: 18, color: const Color(0xFF64748B)),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: const TextStyle(fontSize: 11, color: Color(0xFF94A3B8)),
              ),
              Text(
                value,
                style: const TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: Color(0xFF1E293B),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildParticipantRow(AssignmentEmployee employee) {
    final isMe =
        employee.userId.trim().toLowerCase() == user.id.trim().toLowerCase();
    final statusColor = _getEmployeeStatusColor(employee.employeeStatus);

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: isMe ? const Color(0xFFF0F9FF) : const Color(0xFFF8FAFC),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isMe ? const Color(0xFFBAE6FD) : const Color(0xFFE2E8F0),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            height: 34,
            width: 34,
            decoration: BoxDecoration(
              color: statusColor.withOpacity(0.12),
              shape: BoxShape.circle,
            ),
            child: Icon(
              isMe ? Icons.person_rounded : Icons.groups_rounded,
              color: statusColor,
              size: 18,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        isMe
                            ? '${employee.fullName} (Anda)'
                            : employee.fullName,
                        style: const TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w700,
                          color: Color(0xFF0F172A),
                        ),
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    const SizedBox(width: 8),
                    _statusPill(_getStatusDisplay(employee.employeeStatus), statusColor),
                  ],
                ),
                if (employee.rejectionNote != null &&
                    employee.rejectionNote!.trim().isNotEmpty) ...[
                  const SizedBox(height: 6),
                  Text(
                    employee.rejectionNote!.trim(),
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFF475569),
                      height: 1.35,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _statusPill(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: color.withOpacity(0.25)),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.w700,
          color: color,
        ),
      ),
    );
  }

  Color _getRequestStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'draft':
        return const Color(0xFF94A3B8);
      case 'submitted':
        return const Color(0xFFF59E0B);
      case 'published':
        return const Color(0xFF10B981);
      default:
        return const Color(0xFF64748B);
    }
  }

  Color _getEmployeeStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return const Color(0xFFF59E0B);
      case 'agreed':
        return const Color(0xFF10B981);
      case 'rejected':
        return const Color(0xFFEF4444);
      default:
        return const Color(0xFF64748B);
    }
  }

  String _getStatusDisplay(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return 'Menunggu';
      case 'agreed':
        return 'Disetujui';
      case 'rejected':
        return 'Ditolak';
      default:
        return status;
    }
  }

  Future<void> _exportToPDF(BuildContext context) async {
    try {
      final pdf = pw.Document();
      final times = pw.Font.times();
      final timesBold = pw.Font.timesBold();
      final myEntry = _findMyEntry();

      // Load logo
      final ByteData bytes = await rootBundle.load('assets/Picture1.jpg');
      final Uint8List list = bytes.buffer.asUint8List();
      final pw.MemoryImage logoImage = pw.MemoryImage(list);

      String rewardText = 'Hari Libur Pengganti';
      if (myEntry != null && myEntry.replacementOffDate != null) {
        final dateStr = DateFormat('EEEE, dd MMMM yyyy', 'id').format(myEntry.replacementOffDate!);
        rewardText += ' ($dateStr)';
      }

      pdf.addPage(
        pw.MultiPage(
          pageFormat: PdfPageFormat.a4,
          margin: const pw.EdgeInsets.all(32),
          header: (context) => pw.Column(
            children: [
              pw.Stack(
                alignment: pw.Alignment.center,
                children: [
                  pw.Align(
                    alignment: pw.Alignment.centerLeft,
                    child: pw.Image(logoImage, width: 60, height: 60),
                  ),
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.center,
                    children: [
                      pw.Text(
                        'PT. Labersa Hutahaean',
                        style: pw.TextStyle(
                          font: timesBold,
                          fontSize: 18,
                          color: PdfColor.fromInt(0xFF988300), // Golden
                        ),
                      ),
                      pw.SizedBox(height: 2),
                      pw.Text(
                        'HEAD OFFICE - WILAYAH TOBA',
                        style: pw.TextStyle(
                          font: timesBold,
                          fontSize: 14,
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
            ],
          ),
          build: (context) => [
            pw.Text(
              'SURAT PERINTAH PENUGASAN',
              style: pw.TextStyle(
                font: timesBold,
                fontSize: 14,
                color: PdfColors.black,
              ),
              textAlign: pw.TextAlign.center,
            ),
            pw.SizedBox(height: 15),
            pw.Text(
              'Kepada saudara yang namanya tersebut di bawah ini diperintahkan untuk melaksanakan tugas:',
              style: pw.TextStyle(font: times, fontSize: 11),
            ),
            pw.SizedBox(height: 10),
            
            // Employees (No Table)
            for (final emp in assignment.employees)
              pw.Padding(
                padding: const pw.EdgeInsets.only(left: 10, bottom: 5),
                child: pw.Row(
                  crossAxisAlignment: pw.CrossAxisAlignment.start,
                  children: [
                    pw.Text('- ', style: pw.TextStyle(font: times, fontSize: 11)),
                    pw.Expanded(
                      child: pw.Column(
                        crossAxisAlignment: pw.CrossAxisAlignment.start,
                        children: [
                          pw.Text(emp.fullName, style: pw.TextStyle(font: timesBold, fontSize: 11)),
                          pw.Text('Jabatan: ${emp.positionName.isNotEmpty ? emp.positionName : '-'}', style: pw.TextStyle(font: times, fontSize: 10, color: PdfColors.grey700)),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            
            pw.SizedBox(height: 15),
            
            // Details
            _pdfDetailRow('Untuk keperluan / tugas', assignment.reason.isNotEmpty ? assignment.reason : '-', times, timesBold),
            _pdfDetailRow('Pada hari / tanggal', DateFormat('EEEE, dd MMMM yyyy', 'id').format(assignment.date), times, timesBold),
            _pdfDetailRow('Waktu', '${assignment.startTime} - ${assignment.endTime}', times, timesBold),
            _pdfDetailRow('Pilihan Reward', rewardText, times, timesBold),
            
            pw.SizedBox(height: 30),
            
            // Signatures
            pw.Row(
              mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
              children: [
                _pdfSignBlock('Yang memberi perintah,', 'Departement Head', times, timesBold),
                _pdfSignBlock('Yang menerima perintah,', 'Karyawan', times, timesBold),
                _pdfSignBlock('Disetujui Oleh,', 'Office Manager / HRM / General Manager', times, timesBold),
              ],
            ),
          ],
        ),
      );

      await Printing.layoutPdf(
        onLayout: (PdfPageFormat format) async => pdf.save(),
        name: 'Surat_Penugasan_${assignment.id}.pdf',
      );
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Gagal export PDF: $e')));
    }
  }

  pw.Widget _pdfDetailRow(String label, String value, pw.Font font, pw.Font fontBold) {
    return pw.Padding(
      padding: const pw.EdgeInsets.symmetric(vertical: 2),
      child: pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.SizedBox(width: 120, child: pw.Text(label, style: pw.TextStyle(font: font, fontSize: 11))),
          pw.Text(': ', style: pw.TextStyle(font: font, fontSize: 11)),
          pw.Expanded(child: pw.Text(value, style: pw.TextStyle(font: fontBold, fontSize: 11))),
        ],
      ),
    );
  }

  pw.Widget _pdfSignBlock(String label, String subLabel, pw.Font font, pw.Font fontBold) {
    return pw.Column(
      children: [
        pw.Text(label, style: pw.TextStyle(font: font, fontSize: 10)),
        pw.SizedBox(height: 30),
        pw.Text('( ____________________ )', style: pw.TextStyle(font: font, fontSize: 10)),
        pw.SizedBox(height: 4),
        pw.Text(subLabel, style: pw.TextStyle(font: fontBold, fontSize: 8)),
      ],
    );
  }
}