// lib/pages/request_page.dart
import 'package:flutter/material.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/models/leave_request.dart';
import 'package:intl/intl.dart';
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

  @override
  void initState() {
    super.initState();
    _loadUser();
    _loadRequests();
  }

  Future<void> _loadUser() async {
    try {
      final u = await ApiService.getProfile();
      if (mounted) setState(() => _user = u);
    } catch (_) {}
  }

  Future<void> _loadRequests() async {
    setState(() => _isLoading = true);
    try {
      // ✅ Ambil data real dari backend
      final data = await ApiService.getMyPengajuan();
      if (mounted) setState(() { _requests = data; _isLoading = false; });
    } catch (e) {
      print('[Request] load error: $e');
      if (mounted) setState(() => _isLoading = false);
    }
  }

  List<LeaveRequest> get _filtered {
    if (_selectedTab == 0) return _requests;
    final cat = _tabs[_selectedTab]; // 'Izin' | 'Cuti' | 'Lembur'
    return _requests.where((r) =>
      r.namaKategori.toLowerCase() == cat.toLowerCase()
    ).toList();
  }

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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(child: Column(children: [
        _buildHeader(),
        _buildTabs(),
        Expanded(child: _buildBody()),
      ])),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          final res = await Navigator.push(context, MaterialPageRoute(builder: (_) => const NewRequestPage()));
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
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 16, offset: const Offset(0, 4))],
      ),
      child: Row(children: [
        Stack(children: [
          Container(
              height: 48, width: 48,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                gradient: const LinearGradient(colors: [Color(0xFF135BEC), Color(0xFF3B7BF6)]),
                boxShadow: [BoxShadow(color: const Color(0xFF135BEC).withOpacity(0.3), blurRadius: 8, offset: const Offset(0, 2))],
              ),
              child: Padding(padding: const EdgeInsets.all(2),
                child: Container(
                  decoration: const BoxDecoration(shape: BoxShape.circle, color: Colors.white),
                  child: ClipOval(child: Image.network(_avatarUrl(), fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => const Icon(Icons.person, color: Color(0xFF135BEC), size: 26))),
                )),
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

  Widget _buildTabs() {
    return Container(
      color: Colors.transparent,
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(children: List.generate(_tabs.length, (i) {
          final sel = _selectedTab == i;
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: GestureDetector(
              onTap: () => setState(() => _selectedTab = i),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 9),
                decoration: BoxDecoration(
                  color: sel ? const Color(0xFF135BEC) : Colors.white,
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.05), blurRadius: 6, offset: const Offset(0, 2))],
                ),
                child: Text(_tabs[i], style: TextStyle(
                  fontSize: 13, fontWeight: FontWeight.w600,
                  color: sel ? Colors.white : Colors.grey.shade600)),
              ),
            ),
          );
        })),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) return const Center(child: CircularProgressIndicator());
    if (_filtered.isEmpty) return Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
      Icon(Icons.assignment_outlined, size: 52, color: Colors.grey.shade300),
      const SizedBox(height: 12),
      Text('Belum ada pengajuan', style: TextStyle(color: Colors.grey.shade500)),
      const SizedBox(height: 6),
      Text('Ketuk + untuk buat pengajuan baru', style: TextStyle(color: Colors.grey.shade400, fontSize: 12)),
    ]));

    return RefreshIndicator(
      onRefresh: _loadRequests,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 90),
        itemCount: _filtered.length,
        itemBuilder: (_, i) => _buildCard(_filtered[i]),
      ),
    );
  }

  Widget _buildCard(LeaveRequest r) {
    final isLembur = r.namaKategori.toLowerCase() == 'lembur';
    final sc  = _statusColor(r.statusFinal);
    final sl  = _statusLabel(r.statusFinal);
    final fmt = DateFormat('EEE, dd MMM yyyy', 'id');

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white, borderRadius: BorderRadius.circular(18),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 10, offset: const Offset(0, 3))],
      ),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
        child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
          // Ikon kategori
          Container(
            height: 44, width: 44,
            decoration: BoxDecoration(
              color: _kategoriColor(r.namaKategori).withOpacity(0.1),
              borderRadius: BorderRadius.circular(12)),
            child: Icon(_kategoriIcon(r.namaKategori),
              color: _kategoriColor(r.namaKategori), size: 22)),
          const SizedBox(width: 14),
          // Info tengah
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Row(children: [
              Icon(Icons.calendar_today_rounded, size: 11, color: Colors.grey.shade400),
              const SizedBox(width: 4),
              Text(fmt.format(r.startDate),
                style: TextStyle(fontSize: 11, color: Colors.grey.shade400, fontWeight: FontWeight.w500)),
            ]),
            const SizedBox(height: 4),
            // ✅ Gunakan r.type (nama_tipe dari backend, misal "Izin Sakit")
            Text(r.type,
              style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: Color(0xFF0F172A))),
            const SizedBox(height: 6),
            Row(children: [
              Icon(isLembur ? Icons.access_time_rounded : Icons.date_range_rounded,
                size: 13, color: const Color(0xFF135BEC)),
              const SizedBox(width: 5),
              Text(
                r.days <= 1 && !isLembur
                    ? DateFormat('dd MMM yyyy').format(r.startDate)
                    : isLembur
                        ? DateFormat('dd MMM yyyy').format(r.startDate)
                        : '${DateFormat('dd MMM').format(r.startDate)} s/d ${DateFormat('dd MMM').format(r.endDate)}',
                style: const TextStyle(fontSize: 12, color: Color(0xFF135BEC), fontWeight: FontWeight.w500)),
            ]),
            const SizedBox(height: 8),
            // ✅ Approval chain bertahap
            _buildApprovalChain(r),
          ])),
          const SizedBox(width: 10),
          // Kanan: status + durasi
          Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
              decoration: BoxDecoration(color: sc.withOpacity(0.1), borderRadius: BorderRadius.circular(20),
                border: Border.all(color: sc.withOpacity(0.3))),
              child: Text(sl, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: sc))),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
              decoration: BoxDecoration(color: const Color(0xFFF1F5F9), borderRadius: BorderRadius.circular(12)),
              child: Text(
                isLembur ? '— Jam' : '${r.days} Hari',
                style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: Color(0xFF0F172A)))),
          ]),
        ]),
      ),
    );
  }

  Widget _buildApprovalChain(LeaveRequest r) {
    return Row(children: [
      _approvalDot(r.statusKepala, 'Ka.Dept'),
      Padding(padding: const EdgeInsets.symmetric(horizontal: 4),
        child: Container(width: 14, height: 1, color: Colors.grey.shade200)),
      _approvalDot(r.statusManagerHr, 'Mgr HR'),
    ]);
  }

  Widget _approvalDot(String status, String label) {
    final c = status == 'APPROVED'
        ? const Color(0xFF2ECC71)
        : status == 'REJECTED'
            ? const Color(0xFFEF4444)
            : const Color(0xFFF59E0B);
    return Row(mainAxisSize: MainAxisSize.min, children: [
      Container(width: 6, height: 6, decoration: BoxDecoration(shape: BoxShape.circle, color: c)),
      const SizedBox(width: 3),
      Text(label, style: TextStyle(fontSize: 9, color: Colors.grey.shade400, fontWeight: FontWeight.w500)),
    ]);
  }

  String _statusLabel(String s) {
    switch (s.toUpperCase()) {
      case 'APPROVED': return 'DISETUJUI';
      case 'REJECTED': return 'DITOLAK';
      default:         return 'MENUNGGU';
    }
  }

  IconData _kategoriIcon(String k) {
    switch (k.toLowerCase()) {
      case 'cuti':   return Icons.beach_access_rounded;
      case 'lembur': return Icons.timelapse_rounded;
      default:       return Icons.assignment_late_rounded;
    }
  }

  Color _kategoriColor(String k) {
    switch (k.toLowerCase()) {
      case 'cuti':   return const Color(0xFF8B5CF6);
      case 'lembur': return const Color(0xFFF59E0B);
      default:       return const Color(0xFF135BEC);
    }
  }

  String _calcHours(String? start, String? end) {
    if (start == null || end == null) return '0 Jam';
    try {
      final s = start.split(':');
      final e = end.split(':');
      final mins = (int.parse(e[0]) * 60 + int.parse(e[1])) - (int.parse(s[0]) * 60 + int.parse(s[1]));
      final h = mins ~/ 60; final m = mins % 60;
      return m == 0 ? '$h Jam' : '$h Jam $m Menit';
    } catch (_) { return '0 Jam'; }
  }

  Color _statusColor(String s) {
    switch (s.toUpperCase()) {
      case 'APPROVED': return const Color(0xFF2ECC71);
      case 'PENDING':  return const Color(0xFFF59E0B);
      case 'REJECTED': return const Color(0xFFEF4444);
      default: return Colors.grey;
    }
  }
}