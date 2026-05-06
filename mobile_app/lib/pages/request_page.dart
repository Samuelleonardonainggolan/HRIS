// lib/pages/request_page.dart
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:mobile_app/models/leave_request.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:url_launcher/url_launcher.dart';

import 'new_request_page.dart';

class RequestPage extends StatefulWidget {
  const RequestPage({super.key});

  @override
  State<RequestPage> createState() => _RequestPageState();
}

class _RequestPageState extends State<RequestPage> {
  int _selectedTab = 0;
  final List<String> _tabs = ['Pengajuan Terbaru', 'Izin', 'Cuti'];
  bool _isLoading = true;
  User? _user;
  List<LeaveRequest> _requests = [];

  @override
  void initState() {
    super.initState();
    ApiService.currentUser.addListener(_syncProfile);
    _loadUser();
    _loadRequests();
  }

  @override
  void dispose() {
    ApiService.currentUser.removeListener(_syncProfile);
    super.dispose();
  }

  void _syncProfile() {
    if (!mounted) return;
    setState(() => _user = ApiService.currentUser.value);
  }

  Future<void> _loadUser() async {
    try {
      final user = await ApiService.getProfile();
      if (mounted) {
        setState(() => _user = user);
      }
    } catch (_) {}
  }

  Future<void> _loadRequests() async {
    setState(() => _isLoading = true);
    try {
      final data = await ApiService.getMyPengajuan();
      if (mounted) {
        setState(() {
          _requests = data;
          _isLoading = false;
        });
      }
    } catch (e) {
      debugPrint('[Request] load error: $e');
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _cancelRequest(LeaveRequest request) async {
    // Restrict cancellation if Kadep has already approved
    if (request.statusKepala == 'APPROVED') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Pengajuan tidak dapat dibatalkan karena sudah disetujui Kepala Departemen'),
          backgroundColor: Color(0xFFF59E0B),
        ),
      );
      return;
    }

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
      await ApiService.cancelPengajuan(request.id);
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

  Future<void> _editRequest(LeaveRequest request) async {
    // Restrict editing if Kadep has already approved
    if (request.statusKepala == 'APPROVED') {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Pengajuan tidak dapat diubah karena sudah disetujui Kepala Departemen'),
          backgroundColor: Color(0xFFF59E0B),
        ),
      );
      return;
    }

    final result = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => NewRequestPage(requestToEdit: request),
      ),
    );

    if (result == true) {
      await _loadRequests();
    }
  }

  Future<void> _launchURL(String url) async {
    final uri = Uri.parse(url);
    if (!await launchUrl(uri, mode: LaunchMode.externalApplication)) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Gagal membuka dokumen')),
        );
      }
    }
  }

  void _showRequestDetail(LeaveRequest request) {
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
                      color: _kategoriColor(request.namaKategori).withValues(
                        alpha: 0.12,
                      ),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Icon(
                      _kategoriIcon(request.namaKategori),
                      color: _kategoriColor(request.namaKategori),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      request.type,
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
                    'Kategori ${request.namaKategori}',
                    _kategoriColor(request.namaKategori),
                  ),
                  _detailBadge(
                    'Status ${_statusLabel(request.statusFinal)}',
                    _statusColor(request.statusFinal),
                  ),
                  _detailBadge(
                    'Durasi ${request.days} hari',
                    const Color(0xFF0F766E),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                'Periode: ${dateFmt.format(request.startDate)} - ${dateFmt.format(request.endDate)}',
              ),
              const SizedBox(height: 12),
              const Text(
                'Alasan',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 4),
              Text(request.reason.isEmpty ? '-' : request.reason),
              const SizedBox(height: 14),
              const Text(
                'Tahap Persetujuan',
                style: TextStyle(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              _approvalRow(
                'Kepala Departemen',
                request.statusKepala,
                actorName: request.kepalaDepartemenName,
              ),
              const SizedBox(height: 6),
              _approvalRow(
                'Manager HR',
                request.statusManagerHr,
                actorName: request.managerHrName,
              ),
              if (request.statusFinal == 'REJECTED') ...[
                const SizedBox(height: 14),
                const Text(
                  'Alasan Penolakan',
                  style: TextStyle(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 8),
                if ((request.rejectionReasonKepalaDept ?? '').trim().isNotEmpty)
                  _rejectionReasonBox(
                    label: 'Kepala Departemen',
                    reason: request.rejectionReasonKepalaDept!.trim(),
                  ),
                if ((request.rejectionReasonKepalaDept ?? '').trim().isNotEmpty &&
                    (request.rejectionReasonManagerHr ?? '').trim().isNotEmpty)
                  const SizedBox(height: 8),
                if ((request.rejectionReasonManagerHr ?? '').trim().isNotEmpty)
                  _rejectionReasonBox(
                    label: 'Manager HR',
                    reason: request.rejectionReasonManagerHr!.trim(),
                  ),
                if ((request.rejectionReasonKepalaDept ?? '').trim().isEmpty &&
                    (request.rejectionReasonManagerHr ?? '').trim().isEmpty)
                  _rejectionReasonBox(
                    label: 'Keterangan',
                    reason: 'Pengajuan ditolak.',
                  ),
              ],
              if (request.dokumenUrl != null && request.dokumenUrl!.isNotEmpty) ...[
                const SizedBox(height: 20),
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF0F9FF),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: const Color(0xFFBAE6FD)),
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: const Color(0xFF0284C7).withOpacity(0.1),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(Icons.picture_as_pdf_rounded, color: Color(0xFF0284C7), size: 20),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              'Dokumen Lampiran',
                              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                            ),
                            const Text('Ketuk untuk melihat dokumen', style: TextStyle(fontSize: 12, color: Color(0xFF0369A1))),
                          ],
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.open_in_new_rounded, color: Color(0xFF0284C7)),
                        onPressed: () => _launchURL(request.dokumenUrl!),
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: 16),
              if (_canEditRequest(request))
                Row(
                  children: [
                    Expanded(
                      child: OutlinedButton(
                        onPressed: () {
                          Navigator.pop(context);
                          _editRequest(request);
                        },
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                          side: const BorderSide(color: Color(0xFF135BEC)),
                        ),
                        child: const Text('Edit', style: TextStyle(color: Color(0xFF135BEC), fontWeight: FontWeight.bold)),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: const Color(0xFFEF4444),
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                          elevation: 0,
                        ),
                        onPressed: () {
                          Navigator.pop(context);
                          _cancelRequest(request);
                        },
                        child: const Text(
                          'Batalkan',
                          style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                        ),
                      ),
                    ),
                  ],
                ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_filtered.isEmpty) {
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
    }

    return RefreshIndicator(
      onRefresh: _loadRequests,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 90),
        itemCount: _filtered.length,
        itemBuilder: (_, index) => _buildCard(_filtered[index]),
      ),
    );
  }

  Widget _buildCard(LeaveRequest request) {
    final statusColor = _statusColor(request.statusFinal);
    final statusLabel = _statusLabel(request.statusFinal);
    final dateText = request.days <= 1
        ? DateFormat('dd MMM yyyy', 'id').format(request.startDate)
        : '${DateFormat('dd MMM', 'id').format(request.startDate)} s/d ${DateFormat('dd MMM', 'id').format(request.endDate)}';

    return GestureDetector(
      onTap: () => _showRequestDetail(request),
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
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
                  color: _kategoriColor(request.namaKategori)
                      .withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  _kategoriIcon(request.namaKategori),
                  color: _kategoriColor(request.namaKategori),
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
                          'Diajukan: ${DateFormat('dd MMM yyyy', 'id').format(request.createdAt)}',
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
                      request.type,
                      style: const TextStyle(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF0F172A),
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        const Icon(
                          Icons.date_range_rounded,
                          size: 13,
                          color: Color(0xFF135BEC),
                        ),
                        const SizedBox(width: 5),
                        Text(
                          dateText,
                          style: const TextStyle(
                            fontSize: 12,
                            color: Color(0xFF135BEC),
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    _buildApprovalChain(request),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  if (request.statusFinal == 'PENDING')
                    PopupMenuButton<String>(
                      icon: const Icon(Icons.more_horiz, size: 18),
                      onSelected: (value) {
                        if (value == 'detail') _showRequestDetail(request);
                        if (value == 'edit') _editRequest(request);
                        if (value == 'cancel') _cancelRequest(request);
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
                      color: statusColor.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(
                        color: statusColor.withValues(alpha: 0.3),
                      ),
                    ),
                    child: Text(
                      statusLabel,
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
                      '${request.days} Hari',
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

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(28)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
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
                      color: const Color(0xFF135BEC).withValues(alpha: 0.3),
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

  Widget _buildTabs() {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.symmetric(vertical: 12),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Row(
          children: List.generate(_tabs.length, (i) {
            final selected = _selectedTab == i;
            return Padding(
              padding: const EdgeInsets.only(right: 10),
              child: InkWell(
                onTap: () => setState(() => _selectedTab = i),
                borderRadius: BorderRadius.circular(20),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 250),
                  padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                  decoration: BoxDecoration(
                    color: selected ? const Color(0xFF135BEC) : const Color(0xFFF1F5F9),
                    borderRadius: BorderRadius.circular(20),
                    boxShadow: selected
                        ? [
                            BoxShadow(
                              color: const Color(0xFF135BEC).withOpacity(0.3),
                              blurRadius: 10,
                              offset: const Offset(0, 4),
                            )
                          ]
                        : null,
                  ),
                  child: Text(
                    _tabs[i],
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: selected ? FontWeight.bold : FontWeight.w600,
                      color: selected ? Colors.white : const Color(0xFF64748B),
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

  List<LeaveRequest> get _filtered {
    if (_selectedTab == 0) return _requests;
    final category = _tabs[_selectedTab].toLowerCase();
    return _requests
        .where((request) => request.namaKategori.toLowerCase() == category)
        .toList();
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
          final result = await Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const NewRequestPage()),
          );
          if (result == true) {
            await _loadRequests();
          }
        },
        backgroundColor: const Color(0xFF135BEC),
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _buildApprovalChain(LeaveRequest request) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _approvalPill(request.statusKepala, 'Ka.Dept'),
        const SizedBox(height: 6),
        _approvalPill(request.statusManagerHr, 'Mgr HR'),
      ],
    );
  }

  Widget _approvalRow(String label, String status, {String? actorName}) {
    final color = _statusColor(status);
    final actor = (actorName ?? '').trim();
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(Icons.circle, size: 10, color: color),
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
          style: TextStyle(
            color: color,
            fontWeight: FontWeight.w700,
            fontSize: 12,
          ),
        ),
      ],
    );
  }

  Widget _approvalPill(String status, String label) {
    final color = _statusColor(status);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withValues(alpha: 0.28)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.shield_rounded, size: 11, color: color),
          const SizedBox(width: 4),
          Text(
            '$label: ${_statusLabel(status)}',
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

  Widget _detailBadge(String text, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
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

  bool _canEditRequest(LeaveRequest request) {
    // Cannot edit if final status is APPROVED
    if (request.statusFinal == 'APPROVED') return false;
    // Cannot edit if Kadep has already approved (as requested)
    if (request.statusKepala == 'APPROVED') return false;
    // Cannot edit if CANCELLED
    if (request.statusFinal == 'CANCELLED') return false;

    return true; // Pending or Rejected can be edited
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

  String _statusLabel(String status) {
    switch (status.toUpperCase()) {
      case 'APPROVED':
      case 'DISETUJUI':
        return 'DISETUJUI';
      case 'REJECTED':
      case 'DITOLAK':
        return 'DITOLAK';
      case 'CANCELLED':
      case 'DIBATALKAN':
        return 'DIBATALKAN';
      default:
        return 'MENUNGGU';
    }
  }

  Color _statusColor(String status) {
    switch (status.toUpperCase()) {
      case 'APPROVED':
      case 'DISETUJUI':
        return const Color(0xFF2ECC71);
      case 'PENDING':
      case 'MENUNGGU':
        return const Color(0xFFF59E0B);
      case 'REJECTED':
      case 'DITOLAK':
        return const Color(0xFFEF4444);
      case 'CANCELLED':
      case 'DIBATALKAN':
        return const Color(0xFF64748B);
      default:
        return Colors.grey;
    }
  }

  Color _kategoriColor(String kategori) {
    switch (kategori.toLowerCase()) {
      case 'izin':
        return const Color(0xFF135BEC);
      case 'cuti':
        return const Color(0xFF8B5CF6);
      default:
        return const Color(0xFF135BEC);
    }
  }

  IconData _kategoriIcon(String kategori) {
    switch (kategori.toLowerCase()) {
      case 'izin':
        return Icons.assignment_late_rounded;
      case 'cuti':
        return Icons.beach_access_rounded;
      default:
        return Icons.assignment_rounded;
    }
  }
}
