import 'dart:async';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:mobile_app/models/overtime_request.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/services/sse_service.dart';
import 'package:url_launcher/url_launcher.dart';

class OvertimePage extends StatefulWidget {
  const OvertimePage({super.key});

  @override
  State<OvertimePage> createState() => _OvertimePageState();
}

class _OvertimePageState extends State<OvertimePage> {
  bool _isLoading = true;
  User? _user;
  List<OvertimeRequest> _items = [];
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
      _loadData();
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

  bool _canUserSeeRequest(User user, OvertimeRequest request) {
    final assignedToMe = _isAssignedToUser(request, user.id);
    return assignedToMe;
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);
    try {
      final user =
          ApiService.currentUser.value ?? await ApiService.getProfile();

      List<OvertimeRequest> data =
          await ApiService.getAssignedOvertimeRequests();
      final mine = await ApiService.getMyOvertimeRequests();
      final shouldMergeMine = user.isManagerDept || user.isManagerHR;
      if (shouldMergeMine) {
        final merged = <String, OvertimeRequest>{
          for (final item in data) item.id: item,
          for (final item in mine) item.id: item,
        };
        data = merged.values.toList()..sort((a, b) => b.date.compareTo(a.date));
      }

      data = data.where((item) => _canUserSeeRequest(user, item)).toList();

      if (!mounted) return;
      setState(() {
        _user = user;
        _items = data;
        _isLoading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _isLoading = false);
    }
  }

  Future<void> _agree(OvertimeRequest r) async {
    try {
      await ApiService.agreeOvertimeRequest(r.id);
      if (!mounted) return;
      _showSnackBar('Anda menyetujui lembur', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal setuju: $e', isError: true);
    }
  }

  Future<void> _reject(OvertimeRequest r) async {
    final noteCtrl = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text(
          'Tolak Pengajuan Lembur',
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
      await ApiService.rejectOvertimeRequest(
        r.id,
        rejectionNote: noteCtrl.text.trim(),
      );
      if (!mounted) return;
      _showSnackBar('Pengajuan lembur ditolak', isError: false);
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
              Container(
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
                      child: Image.network(
                        _avatarUrl(),
                        fit: BoxFit.cover,
                        errorBuilder: (_, _, _) => const Icon(
                          Icons.person,
                          color: Color(0xFF135BEC),
                          size: 26,
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
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: Column(
          children: [
            _buildHeader(),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _user == null
                  ? _buildLoadFailedState()
                  : RefreshIndicator(
                      onRefresh: _loadData,
                      color: const Color(0xFF135BEC),
                      child: _items.isEmpty ? _buildEmptyState() : _buildList(),
                    ),
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
    return ListView.builder(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 90),
      itemCount: _items.length,
      itemBuilder: (context, index) {
        final item = _items[index];
        return _OvertimeCard(
          item: item,
          user: _user!,
          onTap: () => _showDetail(item),
          onAgree: () => _agree(item),
          onReject: () => _reject(item),
        );
      },
    );
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
      ),
    );
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
              // Icon kategori lembur
              Container(
                height: 44,
                width: 44,
                decoration: BoxDecoration(
                  color: const Color(0xFF135BEC).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(
                  Icons.work_history_rounded,
                  color: Color(0xFF135BEC),
                  size: 22,
                ),
              ),
              const SizedBox(width: 14),
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
                  const SizedBox(height: 24),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 5,
                    ),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: statusColor.withOpacity(0.3),
                      ),
                    ),
                    child: Text(
                      item.statusDisplay,
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                        color: statusColor,
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

  const _OvertimeDetailSheet({
    required this.request,
    required this.user,
    required this.onAgree,
    required this.onReject,
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