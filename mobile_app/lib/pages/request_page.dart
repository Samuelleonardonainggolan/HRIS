// lib/pages/request_page.dart
import 'package:flutter/material.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/models/leave_request.dart';
import 'package:mobile_app/models/overtime_request.dart';
import 'package:mobile_app/services/sse_service.dart';
import 'package:intl/intl.dart';
import 'dart:async';
import 'new_request_page.dart';

class RequestPage extends StatefulWidget {
  const RequestPage({super.key});
  @override
  State<RequestPage> createState() => _RequestPageState();
}

class _RequestPageState extends State<RequestPage> {
  int _selectedTab = 0; // 0=Semua, 1=Izin, 2=Cuti, 3=Lembur
  final _tabs = ['Pengajuan Terbaru', 'Izin', 'Cuti', 'Lembur'];
  bool _isLoading = true;
  User? _user;

  // Dummy data — ganti dengan API call
  List<LeaveRequest> _requests = [];
  List<OvertimeRequest> _overtimeRequests = [];

  StreamSubscription? _sseSubscription;

  @override
  void initState() {
    super.initState();
    ApiService.currentUser.addListener(_syncProfile);
    SSEService().refreshCounter.addListener(_onRemoteUpdate);
    _loadUser();
    _loadRequests();
  }

  void _onRemoteUpdate() {
    if (mounted) {
      _loadRequests(silent: true);
    }
  }

  void _syncProfile() {
    if (!mounted) return;
    setState(() => _user = ApiService.currentUser.value);
  }

  Future<void> _loadUser() async {
    try {
      final u = await ApiService.getProfile();
      if (mounted) setState(() => _user = u);
    } catch (_) {}
  }

  @override
  void dispose() {
    ApiService.currentUser.removeListener(_syncProfile);
    SSEService().refreshCounter.removeListener(_onRemoteUpdate);
    super.dispose();
  }

  Future<void> _loadRequests({bool silent = false}) async {
    if (!silent) setState(() => _isLoading = true);
    try {
      // ✅ Ambil data real dari backend
      final results = await Future.wait([
        ApiService.getMyPengajuan(),
        ApiService.getMyOvertime(),
      ]);
      
      if (mounted) {
        setState(() {
          _requests = results[0] as List<LeaveRequest>;
          _overtimeRequests = results[1] as List<OvertimeRequest>;
          _isLoading = false;
        });
      }
    } catch (e) {
      print('[Request] load error: $e');
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _cancelRequest(LeaveRequest r) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Batalkan Pengajuan'),
        content: const Text(
          'Pengajuan yang dibatalkan tidak bisa dikembalikan. Lanjutkan?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Tidak'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Ya, Batalkan'),
          ),
        ],
      ),
    );

    if (ok != true) return;

    try {
      await ApiService.cancelPengajuan(r.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Pengajuan berhasil dibatalkan')),
      );
      await _loadRequests();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Gagal membatalkan pengajuan: $e'),
          backgroundColor: const Color(0xFFEF4444),
        ),
      );
    }
  }

  Future<void> _editRequest(LeaveRequest r) async {
    DateTime startDate = r.startDate;
    DateTime endDate = r.endDate;
    final reasonCtrl = TextEditingController(text: r.reason);
    final isSakit = _isSickType(r.type);
    final minStartDate = isSakit
        ? DateTime.now()
        : DateTime.now().add(const Duration(days: 2));

    final edited = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocal) => AlertDialog(
          title: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFFDBEAFE),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: const Icon(
                  Icons.edit_note_rounded,
                  color: Color(0xFF135BEC),
                  size: 20,
                ),
              ),
              const SizedBox(width: 10),
              const Text('Edit Pengajuan'),
            ],
          ),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: isSakit
                        ? const Color(0xFFDCFCE7)
                        : const Color(0xFFFEF3C7),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    isSakit
                        ? 'Izin sakit dapat diajukan mulai hari ini.'
                        : 'Untuk tipe ini, tanggal mulai minimal H-2.',
                    style: TextStyle(
                      fontSize: 12,
                      color: isSakit
                          ? const Color(0xFF166534)
                          : const Color(0xFF92400E),
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                _buildEditableDateRow(
                  label: 'Tanggal Mulai',
                  value: DateFormat('yyyy-MM-dd').format(startDate),
                  icon: Icons.calendar_today,
                  onTap: () async {
                    final picked = await showDatePicker(
                      context: ctx,
                      initialDate: startDate.isBefore(minStartDate)
                          ? minStartDate
                          : startDate,
                      firstDate: minStartDate,
                      lastDate: DateTime.now().add(const Duration(days: 365)),
                    );
                    if (picked != null) {
                      setLocal(() {
                        startDate = picked;
                        if (endDate.isBefore(startDate)) endDate = startDate;
                      });
                    }
                  },
                ),
                const SizedBox(height: 8),
                _buildEditableDateRow(
                  label: 'Tanggal Selesai',
                  value: DateFormat('yyyy-MM-dd').format(endDate),
                  icon: Icons.calendar_month,
                  onTap: () async {
                    final picked = await showDatePicker(
                      context: ctx,
                      initialDate: endDate,
                      firstDate: startDate,
                      lastDate: DateTime.now().add(const Duration(days: 365)),
                    );
                    if (picked != null) {
                      setLocal(() => endDate = picked);
                    }
                  },
                ),
                const SizedBox(height: 10),
                TextField(
                  controller: reasonCtrl,
                  maxLines: 3,
                  decoration: const InputDecoration(
                    labelText: 'Alasan',
                    border: OutlineInputBorder(),
                  ),
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: const Text('Batal'),
            ),
            ElevatedButton(
              onPressed: () => Navigator.pop(ctx, true),
              child: const Text('Simpan'),
            ),
          ],
        ),
      ),
    );

    if (edited != true) {
      reasonCtrl.dispose();
      return;
    }

    try {
      final totalHari = endDate.difference(startDate).inDays + 1;
      await ApiService.updatePengajuan(
        pengajuanId: r.id,
        tanggalMulai: DateFormat('yyyy-MM-dd').format(startDate),
        tanggalSelesai: DateFormat('yyyy-MM-dd').format(endDate),
        totalHari: totalHari,
        alasan: reasonCtrl.text.trim(),
      );

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Pengajuan berhasil diperbarui')),
      );
      await _loadRequests();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Gagal mengedit pengajuan: $e'),
          backgroundColor: const Color(0xFFEF4444),
        ),
      );
    } finally {
      reasonCtrl.dispose();
    }
  }

  void _showRequestDetail(LeaveRequest r) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) {
        final dateFmt = DateFormat('dd MMM yyyy', 'id');
        return Padding(
          padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(99),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Container(
                    height: 46,
                    width: 46,
                    decoration: BoxDecoration(
                      color: _kategoriColor(r.namaKategori).withOpacity(0.12),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(
                      _kategoriIcon(r.namaKategori),
                      color: _kategoriColor(r.namaKategori),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      r.type,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                runSpacing: 8,
                children: [
                  _detailBadge(
                    'Kategori ${r.namaKategori}',
                    _kategoriColor(r.namaKategori),
                  ),
                  _detailBadge(
                    'Status ${_statusLabel(r.statusFinal)}',
                    _statusColor(r.statusFinal),
                  ),
                  _detailBadge(
                    'Durasi ${r.days} hari',
                    const Color(0xFF0F766E),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                'Periode: ${dateFmt.format(r.startDate)} - ${dateFmt.format(r.endDate)}',
              ),
              const SizedBox(height: 12),
              const Text(
                'Alasan',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 4),
              Text(r.reason.isEmpty ? '-' : r.reason),
              const SizedBox(height: 14),
              const Text(
                'Tahap Persetujuan',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              _approvalRow(
                'Kepala Departemen',
                r.statusKepala,
                actorName: r.kepalaDepartemenName,
              ),
              const SizedBox(height: 6),
              _approvalRow(
                'Manager HR',
                r.statusManagerHr,
                actorName: r.managerHrName,
              ),
              if (r.statusFinal == 'REJECTED') ...[
                const SizedBox(height: 14),
                const Text(
                  'Alasan Penolakan',
                  style: TextStyle(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 8),
                if ((r.rejectionReasonKepalaDept ?? '').trim().isNotEmpty)
                  _rejectionReasonBox(
                    label: 'Kepala Departemen',
                    reason: r.rejectionReasonKepalaDept!.trim(),
                  ),
                if ((r.rejectionReasonKepalaDept ?? '').trim().isNotEmpty &&
                    (r.rejectionReasonManagerHr ?? '').trim().isNotEmpty)
                  const SizedBox(height: 8),
                if ((r.rejectionReasonManagerHr ?? '').trim().isNotEmpty)
                  _rejectionReasonBox(
                    label: 'Manager HR',
                    reason: r.rejectionReasonManagerHr!.trim(),
                  ),
                if ((r.rejectionReasonKepalaDept ?? '').trim().isEmpty &&
                    (r.rejectionReasonManagerHr ?? '').trim().isEmpty)
                  _rejectionReasonBox(
                    label: 'Keterangan',
                    reason: 'Pengajuan ditolak.',
                  ),
              ],
            ],
          ),
        );
      },
    );
  }

  Widget _rejectionReasonBox({required String label, required String reason}) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFEF2F2),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFFECACA)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              color: Color(0xFFB91C1C),
            ),
          ),
          const SizedBox(height: 4),
          Text(
            reason,
            style: const TextStyle(
              fontSize: 12,
              color: Color(0xFF7F1D1D),
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  Widget _approvalRow(String label, String status, {String? actorName}) {
    final c = _statusColor(status);
    final actor = (actorName ?? '').trim();
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(Icons.circle, size: 10, color: c),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label),
              if (actor.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.only(top: 2),
                  child: Text(
                    actor,
                    style: TextStyle(
                      fontSize: 11,
                      color: Colors.grey.shade600,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
            ],
          ),
        ),
        Text(
          _statusLabel(status),
          style: TextStyle(color: c, fontWeight: FontWeight.w700, fontSize: 12),
        ),
      ],
    );
  }

  List<dynamic> get _filtered {
    if (_selectedTab == 0) {
      final all = <dynamic>[..._requests, ..._overtimeRequests];
      all.sort((a, b) {
        final d1 = a is LeaveRequest ? a.startDate : (a as OvertimeRequest).date;
        final d2 = b is LeaveRequest ? b.startDate : (b as OvertimeRequest).date;
        return d2.compareTo(d1);
      });
      return all;
    }
    final cat = _tabs[_selectedTab];
    if (cat == 'Lembur') return _overtimeRequests;
    return _requests
        .where((r) => r.namaKategori.toLowerCase() == cat.toLowerCase())
        .toList();
  }

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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: Column(
          children: [
            _buildHeader(),
            _buildTabs(),
            Expanded(child: _buildBody()),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          final res = await Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const NewRequestPage()),
          );
          if (res == true) _loadRequests();
        },
        backgroundColor: const Color(0xFF135BEC),
        elevation: 4,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        child: const Icon(Icons.add_rounded, color: Colors.white, size: 28),
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

  Widget _buildTabs() {
    return Container(
      color: Colors.transparent,
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: List.generate(_tabs.length, (i) {
            final sel = _selectedTab == i;
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: GestureDetector(
                onTap: () => setState(() => _selectedTab = i),
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
                    _tabs[i],
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: sel ? Colors.white : Colors.grey.shade600,
                    ),
                  ),
                ),
              ),
            );
          }),
        ),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) return const Center(child: CircularProgressIndicator());
    if (_filtered.isEmpty)
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.assignment_outlined,
              size: 52,
              color: Colors.grey.shade300,
            ),
            const SizedBox(height: 12),
            Text(
              'Belum ada pengajuan',
              style: TextStyle(color: Colors.grey.shade500),
            ),
            const SizedBox(height: 6),
            Text(
              'Ketuk + untuk buat pengajuan baru',
              style: TextStyle(color: Colors.grey.shade400, fontSize: 12),
            ),
          ],
        ),
      );

    return RefreshIndicator(
      onRefresh: _loadRequests,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 90),
        itemCount: _filtered.length,
        itemBuilder: (_, i) => _buildCard(_filtered[i]),
      ),
    );
  }

  Widget _buildCard(dynamic r) {
    if (r is OvertimeRequest) return _buildOvertimeCard(r);
    final LeaveRequest lr = r as LeaveRequest;
    final isLembur = lr.namaKategori.toLowerCase() == 'lembur';
    final sc = _statusColor(lr.statusFinal);
    final sl = _statusLabel(lr.statusFinal);
    final fmt = DateFormat('EEE, dd MMM yyyy', 'id');

    return GestureDetector(
      onTap: () => _showRequestDetail(lr),
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
              // Ikon kategori
              Container(
                height: 44,
                width: 44,
                decoration: BoxDecoration(
                  color: _kategoriColor(lr.namaKategori).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  _kategoriIcon(lr.namaKategori),
                  color: _kategoriColor(lr.namaKategori),
                  size: 22,
                ),
              ),
              const SizedBox(width: 14),
              // Info tengah
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
                          'Diajukan: ${fmt.format(lr.createdAt)}',
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade400,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    // ✅ Gunakan r.type (nama_tipe dari backend, misal "Izin Sakit")
                    Text(
                      lr.type,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        Icon(
                          isLembur
                              ? Icons.access_time_rounded
                              : Icons.date_range_rounded,
                          size: 13,
                          color: const Color(0xFF135BEC),
                        ),
                        const SizedBox(width: 2),
                        Text(
                          lr.days <= 1 && !isLembur
                              ? DateFormat('dd MMM yyyy').format(lr.startDate)
                              : isLembur
                              ? DateFormat('dd MMM yyyy').format(lr.startDate)
                              : '${DateFormat('dd MMM').format(lr.startDate)} s/d ${DateFormat('dd MMM').format(lr.endDate)}',
                          style: const TextStyle(
                            fontSize: 11,
                            color: Color(0xFF135BEC),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    // ✅ Approval chain bertahap
                    _buildApprovalChain(lr),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              // Kanan: status + durasi
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  if (lr.statusFinal == 'PENDING')
                    PopupMenuButton<String>(
                      icon: const Icon(Icons.more_horiz, size: 18),
                      onSelected: (v) {
                        if (v == 'detail') _showRequestDetail(lr);
                        if (v == 'edit') _editRequest(lr);
                        if (v == 'cancel') _cancelRequest(lr);
                      },
                      itemBuilder: (_) => const [
                        PopupMenuItem(
                          value: 'detail',
                          child: Text('Lihat Detail'),
                        ),
                        PopupMenuItem(
                          value: 'edit',
                          child: Text('Edit Pengajuan'),
                        ),
                        PopupMenuItem(
                          value: 'cancel',
                          child: Text('Batalkan Pengajuan'),
                        ),
                      ],
                    )
                  else
                    const SizedBox(height: 24),
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
                      isLembur ? '— Jam' : '${lr.days} Hari',
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

  Widget _buildOvertimeCard(OvertimeRequest r) {
    final sc = _statusColor(r.finalStatus);
    final sl = _statusLabel(r.finalStatus);
    final fmt = DateFormat('EEE, dd MMM yyyy', 'id');
    final timeFmt = DateFormat('HH:mm');

    return GestureDetector(
      onTap: () => _showOvertimeDetail(r),
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
              Container(
                height: 44,
                width: 44,
                decoration: BoxDecoration(
                  color: _kategoriColor('Lembur').withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  _kategoriIcon('Lembur'),
                  color: _kategoriColor('Lembur'),
                  size: 22,
                ),
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
                          'Diajukan: ${fmt.format(r.createdAt)}',
                          style: TextStyle(
                            fontSize: 11,
                            color: Colors.grey.shade400,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      'Lembur',
                      style: TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        const Icon(
                          Icons.access_time_rounded,
                          size: 13,
                          color: Color(0xFF135BEC),
                        ),
                        const SizedBox(width: 2),
                        Text(
                          '${DateFormat('dd MMM yyyy').format(r.date)} (${timeFmt.format(r.startTime)} - ${timeFmt.format(r.endTime)})',
                          style: const TextStyle(
                            fontSize: 11,
                            color: Color(0xFF135BEC),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    _buildApprovalChain(
                      LeaveRequest(
                        id: r.id, type: 'Lembur', namaKategori: 'Lembur', startDate: r.date, endDate: r.date, reason: r.reason, status: r.finalStatus, statusFinal: r.finalStatus, statusKepala: r.statusKepalaDepartemen, statusManagerHr: r.statusManagerHr, days: 1, createdAt: r.createdAt
                      )
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  if (r.finalStatus == 'PENDING')
                    PopupMenuButton<String>(
                      icon: const Icon(Icons.more_horiz, size: 18),
                      onSelected: (v) {
                        if (v == 'detail') _showOvertimeDetail(r);
                        if (v == 'edit') _editOvertimeRequest(r);
                        if (v == 'cancel') _cancelOvertimeRequest(r);
                      },
                      itemBuilder: (_) => const [
                        PopupMenuItem(value: 'detail', child: Text('Lihat Detail')),
                        PopupMenuItem(value: 'edit', child: Text('Edit Pengajuan')),
                        PopupMenuItem(value: 'cancel', child: Text('Batalkan Pengajuan')),
                      ],
                    )
                  else
                    const SizedBox(height: 24),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                    decoration: BoxDecoration(
                      color: sc.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: sc.withOpacity(0.3)),
                    ),
                    child: Text(
                      sl,
                      style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: sc),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      r.total.isEmpty ? '—' : r.total,
                      style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)),
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

  void _showOvertimeDetail(OvertimeRequest r) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) {
        final dateFmt = DateFormat('dd MMM yyyy', 'id');
        final timeFmt = DateFormat('HH:mm');
        return Padding(
          padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(width: 40, height: 4, decoration: BoxDecoration(color: Colors.grey.shade300, borderRadius: BorderRadius.circular(99))),
              ),
              const SizedBox(height: 16),
              Row(
                children: [
                  Container(
                    height: 46, width: 46,
                    decoration: BoxDecoration(color: _kategoriColor('Lembur').withOpacity(0.12), borderRadius: BorderRadius.circular(12)),
                    child: Icon(_kategoriIcon('Lembur'), color: _kategoriColor('Lembur')),
                  ),
                  const SizedBox(width: 12),
                  const Expanded(child: Text('Lembur', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold))),
                ],
              ),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8, runSpacing: 8,
                children: [
                  _detailBadge('Kategori Lembur', _kategoriColor('Lembur')),
                  _detailBadge('Status ${_statusLabel(r.finalStatus)}', _statusColor(r.finalStatus)),
                ],
              ),
              const SizedBox(height: 12),
              Text('Tanggal: ${dateFmt.format(r.date)}'),
              Text('Jam: ${timeFmt.format(r.startTime)} - ${timeFmt.format(r.endTime)}'),
              Text('Total: ${r.total}'),
              const SizedBox(height: 12),
              const Text('Alasan / Deskripsi Pekerjaan', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 4),
              Text(r.reason.isEmpty ? '-' : r.reason),
              const SizedBox(height: 14),
              const Text('Tahap Persetujuan', style: TextStyle(fontWeight: FontWeight.w700)),
              const SizedBox(height: 8),
              _approvalRow('Kepala Departemen', r.statusKepalaDepartemen, actorName: r.kepalaDepartemenId),
              const SizedBox(height: 6),
              _approvalRow('Manager HR', r.statusManagerHr, actorName: r.managerHrId),
              if (r.finalStatus == 'REJECTED') ...[
                const SizedBox(height: 14),
                const Text('Alasan Penolakan', style: TextStyle(fontWeight: FontWeight.w700)),
                const SizedBox(height: 8),
                if ((r.rejectionReasonKepalaDept ?? '').trim().isNotEmpty)
                  _rejectionReasonBox(label: 'Kepala Departemen', reason: r.rejectionReasonKepalaDept!.trim()),
                if ((r.rejectionReasonManagerHr ?? '').trim().isNotEmpty)
                  _rejectionReasonBox(label: 'Manager HR', reason: r.rejectionReasonManagerHr!.trim()),
              ],
            ],
          ),
        );
      },
    );
  }

  Future<void> _cancelOvertimeRequest(OvertimeRequest r) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text('Batalkan Lembur'),
        content: const Text('Pengajuan lembur yang dibatalkan tidak bisa dikembalikan. Lanjutkan?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Tidak')),
          ElevatedButton(onPressed: () => Navigator.pop(context, true), child: const Text('Ya, Batalkan')),
        ],
      ),
    );

    if (ok != true) return;

    try {
      await ApiService.cancelOvertime(r.id);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Lembur berhasil dibatalkan')));
      await _loadRequests();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Gagal membatalkan lembur: $e'), backgroundColor: const Color(0xFFEF4444)));
    }
  }

  Future<void> _editOvertimeRequest(OvertimeRequest r) async {
    DateTime date = r.date;
    TimeOfDay start = TimeOfDay.fromDateTime(r.startTime);
    TimeOfDay end = TimeOfDay.fromDateTime(r.endTime);
    final reasonCtrl = TextEditingController(text: r.reason);

    final edited = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocal) => AlertDialog(
          title: const Text('Edit Lembur'),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildEditableDateRow(
                  label: 'Tanggal',
                  value: DateFormat('yyyy-MM-dd').format(date),
                  icon: Icons.calendar_today,
                  onTap: () async {
                    final picked = await showDatePicker(
                      context: ctx,
                      initialDate: date,
                      firstDate: DateTime.now().subtract(const Duration(days: 7)),
                      lastDate: DateTime.now().add(const Duration(days: 30)),
                    );
                    if (picked != null) setLocal(() => date = picked);
                  },
                ),
                const SizedBox(height: 8),
                _buildEditableDateRow(
                  label: 'Jam Mulai',
                  value: '${start.hour.toString().padLeft(2, '0')}:${start.minute.toString().padLeft(2, '0')}',
                  icon: Icons.access_time,
                  onTap: () async {
                    final picked = await showTimePicker(context: ctx, initialTime: start);
                    if (picked != null) setLocal(() => start = picked);
                  },
                ),
                const SizedBox(height: 8),
                _buildEditableDateRow(
                  label: 'Jam Selesai',
                  value: '${end.hour.toString().padLeft(2, '0')}:${end.minute.toString().padLeft(2, '0')}',
                  icon: Icons.access_time_filled,
                  onTap: () async {
                    final picked = await showTimePicker(context: ctx, initialTime: end);
                    if (picked != null) setLocal(() => end = picked);
                  },
                ),
                const SizedBox(height: 10),
                TextField(controller: reasonCtrl, maxLines: 3, decoration: const InputDecoration(labelText: 'Deskripsi Pekerjaan', border: OutlineInputBorder())),
              ],
            ),
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Batal')),
            ElevatedButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Simpan')),
          ],
        ),
      ),
    );

    if (edited != true) return;

    try {
      final s = '${start.hour.toString().padLeft(2, '0')}:${start.minute.toString().padLeft(2, '0')}';
      final e = '${end.hour.toString().padLeft(2, '0')}:${end.minute.toString().padLeft(2, '0')}';
      await ApiService.updateOvertime(
        id: r.id,
        tanggal: DateFormat('yyyy-MM-dd').format(date),
        startTime: s,
        endTime: e,
        alasan: reasonCtrl.text.trim(),
        total: _calcHours(s, e),
      );

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Lembur berhasil diperbarui')));
      await _loadRequests();
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('Gagal mengedit lembur: $e'), backgroundColor: const Color(0xFFEF4444)));
    } finally {
      reasonCtrl.dispose();
    }
  }

  Widget _buildApprovalChain(LeaveRequest r) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _approvalPill(r.statusKepala, 'Ka.Dept'),
        const SizedBox(height: 6),
        _approvalPill(r.statusManagerHr, 'Mgr HR'),
      ],
    );
  }

  Widget _approvalPill(String status, String label) {
    final c = status == 'APPROVED'
        ? const Color(0xFF2ECC71)
        : status == 'REJECTED'
        ? const Color(0xFFEF4444)
        : status == 'CANCELLED'
        ? const Color(0xFF64748B)
        : const Color(0xFFF59E0B);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: c.withOpacity(0.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: c.withOpacity(0.28)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.shield_rounded, size: 11, color: c),
          const SizedBox(width: 4),
          Text(
            '$label: ${_statusLabel(status)}',
            style: TextStyle(
              fontSize: 10,
              color: c,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }

  Widget _detailBadge(String text, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: color,
          fontWeight: FontWeight.w700,
          fontSize: 11,
        ),
      ),
    );
  }

  Widget _buildEditableDateRow({
    required String label,
    required String value,
    required IconData icon,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade300),
        ),
        child: Row(
          children: [
            Icon(icon, size: 18, color: const Color(0xFF135BEC)),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    label,
                    style: const TextStyle(
                      fontSize: 11,
                      color: Color(0xFF64748B),
                    ),
                  ),
                  Text(
                    value,
                    style: const TextStyle(fontWeight: FontWeight.w700),
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right_rounded, color: Color(0xFF94A3B8)),
          ],
        ),
      ),
    );
  }

  bool _isSickType(String typeName) {
    return typeName.toLowerCase().contains('sakit');
  }

  String _statusLabel(String s) {
    switch (s.toUpperCase()) {
      case 'APPROVED':
        return 'DISETUJUI';
      case 'REJECTED':
        return 'DITOLAK';
      case 'CANCELLED':
        return 'DIBATALKAN';
      default:
        return 'MENUNGGU';
    }
  }

  IconData _kategoriIcon(String k) {
    switch (k.toLowerCase()) {
      case 'cuti':
        return Icons.beach_access_rounded;
      case 'lembur':
        return Icons.timelapse_rounded;
      default:
        return Icons.assignment_late_rounded;
    }
  }

  Color _kategoriColor(String k) {
    switch (k.toLowerCase()) {
      case 'cuti':
        return const Color(0xFF8B5CF6);
      case 'lembur':
        return const Color(0xFFF59E0B);
      default:
        return const Color(0xFF135BEC);
    }
  }

  String _calcHours(String? start, String? end) {
    if (start == null || end == null) return '0 Jam';
    try {
      final s = start.split(':');
      final e = end.split(':');
      final mins =
          (int.parse(e[0]) * 60 + int.parse(e[1])) -
          (int.parse(s[0]) * 60 + int.parse(s[1]));
      final h = mins ~/ 60;
      final m = mins % 60;
      return m == 0 ? '$h Jam' : '$h Jam $m Menit';
    } catch (_) {
      return '0 Jam';
    }
  }

  Color _statusColor(String s) {
    switch (s.toUpperCase()) {
      case 'APPROVED':
        return const Color(0xFF2ECC71);
      case 'PENDING':
        return const Color(0xFFF59E0B);
      case 'REJECTED':
        return const Color(0xFFEF4444);
      case 'CANCELLED':
        return const Color(0xFF64748B);
      default:
        return Colors.grey;
    }
  }
}
