import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:mobile_app/models/overtime_request.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/api_service.dart';

class OvertimePage extends StatefulWidget {
  const OvertimePage({super.key});

  @override
  State<OvertimePage> createState() => _OvertimePageState();
}

class _OvertimePageState extends State<OvertimePage> {
  bool _isLoading = true;
  User? _user;
  List<OvertimeRequest> _items = [];

  bool get _isKadep => _user?.isManagerDept == true;
  bool get _isHR => _user?.isManagerHR == true;
  bool get _isEmployee => !_isKadep && !_isHR;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);
    try {
      final user = await ApiService.getProfile();

      List<OvertimeRequest> data;
      if (user.isManagerDept) {
        data = await ApiService.getMyOvertimeRequests();
      } else if (user.isManagerHR) {
        data = await ApiService.getOvertimeRequestsForHR();
      } else {
        data = await ApiService.getAssignedOvertimeRequests();
      }

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

  Future<void> _submitDraft(OvertimeRequest r) async {
    try {
      await ApiService.submitOvertimeRequest(r.id);
      if (!mounted) return;
      _showSnackBar('Pengajuan dikirim ke karyawan', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal submit: $e', isError: true);
    }
  }

  Future<void> _deleteDraft(OvertimeRequest r) async {
    try {
      await ApiService.deleteDraftOvertimeRequest(r.id);
      if (!mounted) return;
      _showSnackBar('Draft lembur dihapus', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal hapus draft: $e', isError: true);
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
        title: const Text('Tolak Pengajuan Lembur', style: TextStyle(fontWeight: FontWeight.bold)),
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
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
            ),
            child: const Text('Tolak'),
          ),
        ],
      ),
    );

    if (ok != true) return;

    try {
      await ApiService.rejectOvertimeRequest(r.id, rejectionNote: noteCtrl.text.trim());
      if (!mounted) return;
      _showSnackBar('Pengajuan lembur ditolak', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal menolak: $e', isError: true);
    }
  }

  Future<void> _publish(OvertimeRequest r) async {
    final urlCtrl = TextEditingController();
    final noteCtrl = TextEditingController();

    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text('Publish Surat Lembur', style: TextStyle(fontWeight: FontWeight.bold)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: urlCtrl,
              decoration: InputDecoration(
                labelText: 'URL surat (SPKL)',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                prefixIcon: const Icon(Icons.link),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: noteCtrl,
              maxLines: 3,
              decoration: InputDecoration(
                labelText: 'Catatan HR',
                border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text('Batal', style: TextStyle(color: Colors.grey.shade600)),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF135BEC),
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
            ),
            child: const Text('Publish'),
          ),
        ],
      ),
    );

    if (ok != true) return;

    try {
      await ApiService.publishOvertimeLetter(
        requestId: r.id,
        letterUrl: urlCtrl.text.trim(),
        notes: noteCtrl.text.trim(),
      );
      if (!mounted) return;
      _showSnackBar('Surat lembur dipublikasikan', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal publish: $e', isError: true);
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
        errorBuilder: (_, __, ___) => const Icon(Icons.person, color: Color(0xFF135BEC), size: 26),
      );
    }

    return Image.network(
      _avatarUrl(),
      fit: BoxFit.cover,
      errorBuilder: (_, __, ___) => const Icon(Icons.person, color: Color(0xFF135BEC), size: 26),
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
                      child: ClipOval(child: _avatarPreview(size: 48)),
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
                  _user?.fullName ?? 'Pengajuan Lembur',
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
        backgroundColor: isError ? const Color(0xFFEF4444) : const Color(0xFF135BEC),
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
                  : RefreshIndicator(
                      onRefresh: _loadData,
                      color: const Color(0xFF135BEC),
                      child: _items.isEmpty ? _buildEmptyState() : _buildList(),
                    ),
            ),
          ],
        ),
      ),
      floatingActionButton: _isKadep
          ? FloatingActionButton.extended(
              onPressed: _showCreateDialog,
              backgroundColor: const Color(0xFF135BEC),
              icon: const Icon(Icons.add, color: Colors.white),
              label: const Text('Tambah Lembur', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
            )
          : null,
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
                child: const Icon(Icons.history_toggle_off, size: 80, color: Color(0xFF94A3B8)),
              ),
              const SizedBox(height: 24),
              const Text(
                'Belum Ada Data Lembur',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF475569)),
              ),
              const SizedBox(height: 8),
              Text(
                'Semua pengajuan lembur akan tampil di sini.',
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
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 100),
      itemCount: _items.length,
      itemBuilder: (context, index) {
        final item = _items[index];
        return _OvertimeCard(
          item: item,
          user: _user!,
          onTap: () => _showDetail(item),
          onAction: (action) {
            if (action == 'submit') _submitDraft(item);
            if (action == 'delete') _deleteDraft(item);
            if (action == 'agree') _agree(item);
            if (action == 'reject') _reject(item);
            if (action == 'publish') _publish(item);
          },
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
        onAction: (action) {
          Navigator.pop(context);
          if (action == 'submit') _submitDraft(r);
          if (action == 'delete') _deleteDraft(r);
          if (action == 'agree') _agree(r);
          if (action == 'reject') _reject(r);
          if (action == 'publish') _publish(r);
        },
      ),
    );
  }

  Future<void> _showCreateDialog() async {
    final selectedEmployeeIds = <String>{};
    final employees = await ApiService.getEmployeesMyDepartment();
    final reasonCtrl = TextEditingController();
    DateTime date = DateTime.now();
    TimeOfDay start = const TimeOfDay(hour: 17, minute: 0);
    TimeOfDay end = const TimeOfDay(hour: 20, minute: 0);

    if (!mounted) return;

    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocal) => AlertDialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          title: const Text('Buat Pengajuan Lembur', style: TextStyle(fontWeight: FontWeight.bold)),
          content: SizedBox(
            width: double.maxFinite,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Tanggal & Waktu', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF64748B))),
                  const SizedBox(height: 8),
                  InkWell(
                    onTap: () async {
                      final picked = await showDatePicker(
                        context: ctx,
                        initialDate: date,
                        firstDate: DateTime.now().subtract(const Duration(days: 1)),
                        lastDate: DateTime.now().add(const Duration(days: 365)),
                      );
                      if (picked != null) setLocal(() => date = picked);
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
                      decoration: BoxDecoration(
                        color: Colors.grey.shade100,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        children: [
                          const Icon(Icons.calendar_today, size: 18, color: Color(0xFF135BEC)),
                          const SizedBox(width: 12),
                          Text(DateFormat('EEEE, dd MMM yyyy', 'id').format(date)),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: InkWell(
                          onTap: () async {
                            final t = await showTimePicker(context: ctx, initialTime: start);
                            if (t != null) setLocal(() => start = t);
                          },
                          child: Container(
                            padding: const EdgeInsets.symmetric(vertical: 12),
                            decoration: BoxDecoration(color: Colors.grey.shade100, borderRadius: BorderRadius.circular(12)),
                            child: Column(
                              children: [
                                const Text('Mulai', style: TextStyle(fontSize: 10, color: Color(0xFF64748B))),
                                Text(start.format(ctx), style: const TextStyle(fontWeight: FontWeight.bold)),
                              ],
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: InkWell(
                          onTap: () async {
                            final t = await showTimePicker(context: ctx, initialTime: end);
                            if (t != null) setLocal(() => end = t);
                          },
                          child: Container(
                            padding: const EdgeInsets.symmetric(vertical: 12),
                            decoration: BoxDecoration(color: Colors.grey.shade100, borderRadius: BorderRadius.circular(12)),
                            child: Column(
                              children: [
                                const Text('Selesai', style: TextStyle(fontSize: 10, color: Color(0xFF64748B))),
                                Text(end.format(ctx), style: const TextStyle(fontWeight: FontWeight.bold)),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  const Text('Alasan Penugasan', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF64748B))),
                  const SizedBox(height: 8),
                  TextField(
                    controller: reasonCtrl,
                    maxLines: 3,
                    decoration: InputDecoration(
                      hintText: 'Tulis penugasan di sini...',
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                      filled: true,
                      fillColor: Colors.grey.shade50,
                    ),
                  ),
                  const SizedBox(height: 20),
                  const Text('Pilih Karyawan', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF64748B))),
                  const SizedBox(height: 8),
                  ...employees.map(
                    (e) => CheckboxListTile(
                      value: selectedEmployeeIds.contains(e.id),
                      onChanged: (v) {
                        setLocal(() {
                          if (v == true) {
                            selectedEmployeeIds.add(e.id);
                          } else {
                            selectedEmployeeIds.remove(e.id);
                          }
                        });
                      },
                      title: Text(e.fullName, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
                      subtitle: Text(e.position, style: const TextStyle(fontSize: 12)),
                      controlAffinity: ListTileControlAffinity.leading,
                      contentPadding: EdgeInsets.zero,
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                  ),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text('Batal', style: TextStyle(color: Colors.grey.shade600)),
            ),
            ElevatedButton(
              onPressed: () => Navigator.pop(ctx, true),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF135BEC),
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
              ),
              child: const Text('Simpan Draft'),
            ),
          ],
        ),
      ),
    );

    if (ok != true) {
      reasonCtrl.dispose();
      return;
    }

    if (selectedEmployeeIds.isEmpty || reasonCtrl.text.trim().isEmpty) {
      _showSnackBar('Lengkapi alasan dan pilih minimal 1 karyawan', isError: true);
      reasonCtrl.dispose();
      return;
    }

    final startStr = '${start.hour.toString().padLeft(2, '0')}:${start.minute.toString().padLeft(2, '0')}';
    final endStr = '${end.hour.toString().padLeft(2, '0')}:${end.minute.toString().padLeft(2, '0')}';

    try {
      await ApiService.createOvertimeRequest(
        employeeIds: selectedEmployeeIds.toList(),
        date: date,
        startTime: startStr,
        endTime: endStr,
        reason: reasonCtrl.text.trim(),
      );
      if (!mounted) return;
      _showSnackBar('Draft lembur berhasil dibuat', isError: false);
      await _loadData();
    } catch (e) {
      _showSnackBar('Gagal membuat draft: $e', isError: true);
    } finally {
      reasonCtrl.dispose();
    }
  }
}

class _OvertimeCard extends StatelessWidget {
  final OvertimeRequest item;
  final User user;
  final VoidCallback onTap;
  final Function(String) onAction;

  const _OvertimeCard({
    required this.item,
    required this.user,
    required this.onTap,
    required this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    final statusColor = _getStatusColor(item.status);
    final isKadep = user.isManagerDept;
    final isHR = user.isManagerHR;
    final isEmployee = !isKadep && !isHR;

    OvertimeEmployee? myEntry;
    if (isEmployee) {
      try {
        myEntry = item.employees.firstWhere((e) => e.userId == user.id);
      } catch (_) {
        myEntry = null;
      }
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(20),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      item.statusDisplay.toUpperCase(),
                      style: TextStyle(
                        color: statusColor,
                        fontSize: 10,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                  if (isEmployee && myEntry != null)
                    _MyStatusBadge(status: myEntry.employeeStatus),
                  if (!isEmployee || myEntry == null)
                    _ActionMenu(item: item, user: user, onAction: onAction),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Container(
                    height: 56,
                    width: 56,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          DateFormat('dd').format(item.date),
                          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
                        ),
                        Text(
                          DateFormat('MMM').format(item.date),
                          style: TextStyle(fontSize: 12, color: Color(0xFF64748B), fontWeight: FontWeight.w500),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          item.reason.isNotEmpty ? item.reason : 'Tanpa Alasan',
                          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            const Icon(Icons.access_time, size: 14, color: Color(0xFF64748B)),
                            const SizedBox(width: 4),
                            Text(
                              '${item.startTime} - ${item.endTime}',
                              style: const TextStyle(fontSize: 13, color: Color(0xFF64748B)),
                            ),
                            const SizedBox(width: 12),
                            const Icon(Icons.timer_outlined, size: 14, color: Color(0xFF64748B)),
                            const SizedBox(width: 4),
                            Text(
                              '${item.getDurationHours().toStringAsFixed(1)}h',
                              style: const TextStyle(fontSize: 13, color: Color(0xFF64748B)),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              if (item.employees.isNotEmpty && !isEmployee) ...[
                const SizedBox(height: 16),
                const Divider(height: 1),
                const SizedBox(height: 12),
                Row(
                  children: [
                    SizedBox(
                      height: 24,
                      child: Stack(
                        children: List.generate(
                          item.employees.length > 3 ? 4 : item.employees.length,
                          (idx) {
                            if (idx == 3) {
                              return Positioned(
                                left: idx * 16.0,
                                child: Container(
                                  height: 24,
                                  width: 24,
                                  decoration: BoxDecoration(
                                    color: const Color(0xFFE2E8F0),
                                    shape: BoxShape.circle,
                                    border: Border.all(color: Colors.white, width: 2),
                                  ),
                                  child: Center(
                                    child: Text(
                                      '+${item.employees.length - 3}',
                                      style: const TextStyle(fontSize: 8, fontWeight: FontWeight.bold),
                                    ),
                                  ),
                                ),
                              );
                            }
                            return Positioned(
                              left: idx * 16.0,
                              child: Container(
                                height: 24,
                                width: 24,
                                decoration: BoxDecoration(
                                  color: const Color(0xFF135BEC),
                                  shape: BoxShape.circle,
                                  border: Border.all(color: Colors.white, width: 2),
                                ),
                                child: const Icon(Icons.person, size: 12, color: Colors.white),
                              ),
                            );
                          },
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      '${item.employees.length} Karyawan ditugaskan',
                      style: TextStyle(fontSize: 12, color: Color(0xFF64748B)),
                    ),
                  ],
                ),
              ],
            ],
          ),
        ),
      ),
    );
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

class _MyStatusBadge extends StatelessWidget {
  final String status;
  const _MyStatusBadge({required this.status});

  @override
  Widget build(BuildContext context) {
    Color color;
    IconData icon;
    String label;

    switch (status.toLowerCase()) {
      case 'pending':
        color = const Color(0xFFF59E0B);
        icon = Icons.hourglass_empty;
        label = 'Menunggu';
      case 'agreed':
        color = const Color(0xFF10B981);
        icon = Icons.check_circle_outline;
        label = 'Disetujui';
      case 'rejected':
        color = const Color(0xFFEF4444);
        icon = Icons.cancel_outlined;
        label = 'Ditolak';
      default:
        color = Colors.grey;
        icon = Icons.help_outline;
        label = 'Unknown';
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: color),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.bold),
          ),
        ],
      ),
    );
  }
}

class _ActionMenu extends StatelessWidget {
  final OvertimeRequest item;
  final User user;
  final Function(String) onAction;

  const _ActionMenu({required this.item, required this.user, required this.onAction});

  @override
  Widget build(BuildContext context) {
    final isKadep = user.isManagerDept;
    final isHR = user.isManagerHR;

    List<PopupMenuItem<String>> items = [];

    if (isKadep && item.isDraft) {
      items.addAll([
        const PopupMenuItem(value: 'submit', child: Text('Kirim ke Karyawan')),
        const PopupMenuItem(value: 'delete', child: Text('Hapus Draft')),
      ]);
    }

    if (isHR && item.isSubmitted) {
      items.add(const PopupMenuItem(value: 'publish', child: Text('Publish SPKL')));
    }

    if (items.isEmpty) return const SizedBox.shrink();

    return PopupMenuButton<String>(
      onSelected: onAction,
      icon: const Icon(Icons.more_vert, color: Color(0xFF64748B)),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      itemBuilder: (_) => items,
    );
  }
}

class _OvertimeDetailSheet extends StatelessWidget {
  final OvertimeRequest request;
  final User user;
  final Function(String) onAction;

  const _OvertimeDetailSheet({
    required this.request,
    required this.user,
    required this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    final isEmployee = !user.isManagerDept && !user.isManagerHR;
    
    OvertimeEmployee? myEntry;
    if (isEmployee) {
      try {
        myEntry = request.employees.firstWhere((e) => e.userId == user.id);
      } catch (_) {
        myEntry = null;
      }
    }
    
    final canRespond = isEmployee && request.isSubmitted && (myEntry?.isPending ?? false);

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
          padding: const EdgeInsets.all(24),
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
            const SizedBox(height: 24),
            Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Detail Penugasan Lembur',
                        style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        DateFormat('EEEE, dd MMMM yyyy', 'id').format(request.date),
                        style: TextStyle(color: Color(0xFF64748B)),
                      ),
                    ],
                  ),
                ),
                _StatusBadge(status: request.status),
              ],
            ),
            const SizedBox(height: 32),
            _buildInfoRow(Icons.access_time_filled, 'Waktu', '${request.startTime} - ${request.endTime} (${request.getDurationHours().toStringAsFixed(1)} Jam)'),
            const SizedBox(height: 16),
            _buildInfoRow(Icons.business_center, 'Departemen', request.departmentName),
            const SizedBox(height: 16),
            _buildInfoRow(Icons.person, 'Pengaju', request.requestedByName),
            const SizedBox(height: 32),
            const Text('Alasan Penugasan', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFF0F172A))),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: const Color(0xFFE2E8F0)),
              ),
              child: Text(
                request.reason.isNotEmpty ? request.reason : 'Tidak ada alasan yang disertakan.',
                style: const TextStyle(fontSize: 14, color: Color(0xFF334155), height: 1.5),
              ),
            ),
            const SizedBox(height: 32),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Daftar Karyawan', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFF0F172A))),
                Text('${request.employees.length} Orang', style: TextStyle(fontSize: 14, color: Color(0xFF64748B))),
              ],
            ),
            const SizedBox(height: 16),
            ...request.employees.map((e) => _EmployeeListItem(employee: e)),
            if (request.isPublished && request.letterUrl != null) ...[
              const SizedBox(height: 32),
              const Text('Surat Perintah Kerja Lembur', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold, color: Color(0xFF0F172A))),
              const SizedBox(height: 12),
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(color: const Color(0xFFEFF6FF), borderRadius: BorderRadius.circular(12)),
                  child: const Icon(Icons.picture_as_pdf, color: Color(0xFF135BEC)),
                ),
                title: const Text('Download SPKL', style: TextStyle(fontWeight: FontWeight.bold)),
                subtitle: Text(request.letterUrl!),
                trailing: const Icon(Icons.download_rounded),
                onTap: () {
                  // Implement download logic
                },
              ),
            ],
            const SizedBox(height: 40),
            if (canRespond)
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => onAction('reject'),
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        side: const BorderSide(color: Color(0xFFEF4444)),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      child: const Text('Tolak', style: TextStyle(color: Color(0xFFEF4444), fontWeight: FontWeight.bold)),
                    ),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () => onAction('agree'),
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                        backgroundColor: const Color(0xFF10B981),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                        elevation: 0,
                      ),
                      child: const Text('Setujui', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                    ),
                  ),
                ],
              ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      children: [
        Icon(icon, size: 20, color: const Color(0xFF64748B)),
        const SizedBox(width: 12),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label, style: const TextStyle(fontSize: 12, color: Color(0xFF94A3B8))),
            Text(value, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF1E293B))),
          ],
        ),
      ],
    );
  }
}

class _StatusBadge extends StatelessWidget {
  final String status;
  const _StatusBadge({required this.status});

  @override
  Widget build(BuildContext context) {
    Color color;
    switch (status.toLowerCase()) {
      case 'draft': color = const Color(0xFF94A3B8); break;
      case 'submitted': color = const Color(0xFFF59E0B); break;
      case 'published': color = const Color(0xFF10B981); break;
      default: color = Colors.grey;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(12)),
      child: Text(
        status.toUpperCase(),
        style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.bold),
      ),
    );
  }
}

class _EmployeeListItem extends StatelessWidget {
  final OvertimeEmployee employee;
  const _EmployeeListItem({required this.employee});

  @override
  Widget build(BuildContext context) {
    Color statusColor;
    IconData statusIcon;

    switch (employee.employeeStatus.toLowerCase()) {
      case 'pending':
        statusColor = const Color(0xFFF59E0B);
        statusIcon = Icons.access_time;
        break;
      case 'agreed':
        statusColor = const Color(0xFF10B981);
        statusIcon = Icons.check_circle;
        break;
      case 'rejected':
        statusColor = const Color(0xFFEF4444);
        statusIcon = Icons.cancel;
        break;
      default:
        statusColor = Colors.grey;
        statusIcon = Icons.help;
    }

    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          const CircleAvatar(
            radius: 18,
            backgroundColor: Color(0xFFF1F5F9),
            child: Icon(Icons.person, size: 18, color: Color(0xFF64748B)),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(employee.displayName, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 14)),
                if (employee.employeeStatus == 'rejected' && employee.rejectionNote != null)
                  Text(
                    'Alasan: ${employee.rejectionNote}',
                    style: const TextStyle(fontSize: 12, color: Color(0xFFEF4444)),
                  ),
              ],
            ),
          ),
          Row(
            children: [
              Icon(statusIcon, size: 14, color: statusColor),
              const SizedBox(width: 4),
              Text(
                employee.statusDisplay,
                style: TextStyle(fontSize: 12, color: statusColor, fontWeight: FontWeight.bold),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
