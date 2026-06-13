// lib/pages/new_request_page.dart
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:file_picker/file_picker.dart';
import 'package:mobile_app/models/leave_request.dart';
import 'package:mobile_app/services/api_service.dart';

class NewRequestPage extends StatefulWidget {
  final LeaveRequest? requestToEdit;
  const NewRequestPage({super.key, this.requestToEdit});

  @override
  State<NewRequestPage> createState() => _NewRequestPageState();
}

class _NewRequestPageState extends State<NewRequestPage> {
  // ─── State tipe pengajuan dari API ───────────────────────────────────────
  List<TipePengajuan> _allTipes = [];
  TipePengajuan? _selectedTipe;
  bool _isLoadingTipes = true;
  String? _tipeError;

  // 'Izin' | 'Cuti'
  // Kategori ini dipakai untuk filter tampilan form.
  String _category = 'Izin';

  // Common
  DateTime _startDate = DateTime.now();
  DateTime _endDate = DateTime.now();
  final _reasonCtrl = TextEditingController();
  PlatformFile? _file;
  bool _submitting = false;
  bool _isLoadingLeaveQuota = true;
  int _remainingLeaveQuota = 0;

  final _fmt = DateFormat('MM/dd/yyyy');

  int get _days => _endDate.difference(_startDate).inDays + 1;

  // ─── Filtered tipe berdasarkan kategori ──────────────────────────────────
  List<TipePengajuan> get _filteredTipes {
    return _allTipes.where((t) => t.namaKategori == _category).toList();
  }

  @override
  void initState() {
    super.initState();
    if (widget.requestToEdit != null) {
      final r = widget.requestToEdit!;
      _category = r.namaKategori;
      _startDate = r.startDate;
      _endDate = r.endDate;
      _reasonCtrl.text = r.reason;
      // Note: _selectedTipe will be set after _loadTipes
    }
    _loadTipes();
    _loadLeaveQuota();
  }

  Future<void> _loadLeaveQuota() async {
    setState(() => _isLoadingLeaveQuota = true);
    try {
      final remaining = await ApiService.getLeaveBalance();
      if (mounted) {
        setState(() {
          _remainingLeaveQuota = remaining;
          _isLoadingLeaveQuota = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoadingLeaveQuota = false);
      }
    }
  }

  // ✅ Load tipe pengajuan dari backend (GET /api/v1/pengajuan/tipe)
  Future<void> _loadTipes() async {
    setState(() {
      _isLoadingTipes = true;
      _tipeError = null;
    });
    try {
      final tipes = await ApiService.getTipePengajuan();
      if (mounted) {
        setState(() {
          _allTipes = tipes;
          
          if (widget.requestToEdit != null) {
            // Jika dalam mode edit, pilih tipe yang sesuai dengan pengajuan lama
            try {
              _selectedTipe = tipes.firstWhere(
                (t) => t.namaTipe == widget.requestToEdit!.type,
              );
              // Pastikan kategori juga sinkron dengan tipe yang dipilih
              _category = _selectedTipe!.namaKategori;
            } catch (_) {
              final firstMatch = tipes.where((t) => t.namaKategori == _category).toList();
              _selectedTipe = firstMatch.isNotEmpty ? firstMatch.first : null;
            }
          } else {
            // Default pilih tipe pertama yang kategorinya sesuai
            final firstMatch = tipes.where((t) => t.namaKategori == _category).toList();
            _selectedTipe = firstMatch.isNotEmpty ? firstMatch.first : null;
          }
          _isLoadingTipes = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _tipeError = 'Gagal memuat tipe pengajuan';
          _isLoadingTipes = false;
        });
      }
    }
  }

  Future<void> _pickDate({required bool isStart}) async {
    final minStartDate = _minimumStartDateForCurrentType();
    final firstAllowedDate = isStart ? minStartDate : _startDate;
    final normalizedFirst = DateTime(
      firstAllowedDate.year,
      firstAllowedDate.month,
      firstAllowedDate.day,
    );
    final requestedInitial = isStart ? _startDate : _endDate;
    final initialDate = requestedInitial.isBefore(normalizedFirst)
        ? normalizedFirst
        : requestedInitial;

    final picked = await showDatePicker(
      context: context,
      initialDate: initialDate,
      firstDate: normalizedFirst,
      lastDate: DateTime.now().add(const Duration(days: 365)),
      builder: (ctx, child) => Theme(
        data: Theme.of(ctx).copyWith(
          colorScheme: const ColorScheme.light(primary: Color(0xFF135BEC)),
        ),
        child: child!,
      ),
    );
    if (picked == null) return;
    setState(() {
      if (isStart) {
        _startDate = picked;
        if (_endDate.isBefore(_startDate)) {
          _endDate = _startDate;
        }
      } else {
        _endDate = picked.isBefore(_startDate) ? _startDate : picked;
      }
    });
  }

  Future<void> _pickFile() async {
    final res = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['pdf', 'png', 'jpg', 'jpeg'],
    );
    if (res != null) setState(() => _file = res.files.first);
  }

  Future<void> _submit() async {
    if (_reasonCtrl.text.trim().isEmpty) {
      _showSnack('Alasan wajib diisi', isError: true);
      return;
    }

    if (_selectedTipe == null) {
      _showSnack('Pilih tipe pengajuan terlebih dahulu', isError: true);
      return;
    }

    // Jika tipe wajib lampiran tapi belum upload
    if (_selectedTipe?.wajibLampiran == true && _file == null) {
      _showSnack('Lampiran dokumen wajib untuk tipe ini', isError: true);
      return;
    }

    if (_requiresQuotaDeduction && _days > _remainingLeaveQuota) {
      _showSnack(
        'Sisa cuti tidak cukup. Tersisa $_remainingLeaveQuota hari.',
        isError: true,
      );
      return;
    }

    setState(() => _submitting = true);

    try {
      final dateFormat = DateFormat('yyyy-MM-dd');
      if (_endDate.isBefore(_startDate)) {
        _showSnack(
          'Tanggal selesai tidak boleh sebelum tanggal mulai',
          isError: true,
        );
        return;
      }
      if (!_isCurrentTypeSick()) {
        final minStart = _minimumStartDateForCurrentType();
        final normalizedStart = DateTime(
          _startDate.year,
          _startDate.month,
          _startDate.day,
        );
        final normalizedMin = DateTime(
          minStart.year,
          minStart.month,
          minStart.day,
        );
        if (normalizedStart.isBefore(normalizedMin)) {
          _showSnack(
            'Pengajuan selain izin sakit hanya boleh diajukan minimal H-2',
            isError: true,
          );
          return;
        }
      }

      if (widget.requestToEdit != null) {
        await ApiService.updatePengajuan(
          pengajuanId: widget.requestToEdit!.id,
          tipePengajuanId: _selectedTipe!.id,
          tanggalMulai: dateFormat.format(_startDate),
          tanggalSelesai: dateFormat.format(_endDate),
          totalHari: _days,
          alasan: _reasonCtrl.text.trim(),
          dokumenUrl: _file?.path,
        );
      } else {
        await ApiService.submitPengajuan(
          tipePengajuanId: _selectedTipe!.id,
          tanggalMulai: dateFormat.format(_startDate),
          tanggalSelesai: dateFormat.format(_endDate),
          totalHari: _days,
          alasan: _reasonCtrl.text.trim(),
          dokumenUrl: _file?.path,
        );
      }

      if (mounted) {
        Navigator.pop(context, true);
        final actionText = widget.requestToEdit != null ? 'diperbarui' : 'dikirim';
        _showSnack('Pengajuan ${_selectedTipe!.namaTipe} berhasil $actionText');
      }
    } catch (e) {
      if (mounted) {
        _showSnack('Gagal mengirimkan pengajuan: ${e.toString().replaceAll("Exception: ", "")}', isError: true);
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  void _showSnack(String msg, {bool isError = false}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: isError
            ? const Color(0xFFEF4444)
            : const Color(0xFF2ECC71),
      ),
    );
  }

  @override
  void dispose() {
    _reasonCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: const Color(0xFFF1F5F9),
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(
              Icons.arrow_back_rounded,
              color: Color(0xFF0F172A),
              size: 18,
            ),
          ),
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          'Pengajuan $_category',
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: Color(0xFF0F172A),
          ),
        ),
        centerTitle: true,
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(1),
          child: Container(height: 1, color: Colors.grey.shade100),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(20, 20, 20, 0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _sectionLabel('Kategori Pengajuan'),
            _categorySelector(),
            const SizedBox(height: 20),

            // ── Tipe pengajuan dari API ──
            _sectionLabel('Tipe Pengajuan'),
            _tipeSelector(),
            const SizedBox(height: 20),

            if (_showQuotaInfoCard) ...[
              _quotaInfoCard(),
              const SizedBox(height: 20),
            ],

            _leaveFields(),
            const SizedBox(height: 20),

            _sectionLabel('Alasan'),
            _reasonField(),

            const SizedBox(height: 20),
            _attachmentField(),

            const SizedBox(height: 32),
            _submitBtn(),
            const SizedBox(height: 12),
            _cancelBtn(),
            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  Widget _sectionLabel(String t) => Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Text(
      t,
      style: const TextStyle(
        fontSize: 13,
        fontWeight: FontWeight.w600,
        color: Color(0xFF64748B),
        letterSpacing: 0.3,
      ),
    ),
  );

  // ── Selector kategori (Izin / Cuti) ────────────────────────────────────
  Widget _categorySelector() {
    final categories = ['Izin', 'Cuti'];
    return Row(
      children: categories.map((cat) {
        final sel = _category == cat;
        return Expanded(
          child: GestureDetector(
            onTap: () {
              setState(() {
                _category = cat;
                // Reset tipe saat ganti kategori
                final matching = _allTipes
                    .where((t) => t.namaKategori == cat)
                    .toList();
                _selectedTipe = matching.isNotEmpty ? matching.first : null;
                final minStart = _minimumStartDateForCurrentType();
                if (_startDate.isBefore(minStart)) {
                  _startDate = minStart;
                  if (_endDate.isBefore(_startDate)) {
                    _endDate = _startDate;
                  }
                }
              });
            },
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 180),
              margin: EdgeInsets.only(right: cat != categories.last ? 8 : 0),
              padding: const EdgeInsets.symmetric(vertical: 12),
              decoration: BoxDecoration(
                color: sel ? const Color(0xFF135BEC) : Colors.white,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: sel ? const Color(0xFF135BEC) : Colors.grey.shade200,
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.04),
                    blurRadius: 6,
                  ),
                ],
              ),
              child: Text(
                cat,
                textAlign: TextAlign.center,
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
    );
  }

  // ── Selector tipe dari API ──────────────────────────────────────────────
  Widget _tipeSelector() {
    if (_isLoadingTipes) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: const Row(
          children: [
            SizedBox(
              height: 18,
              width: 18,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            SizedBox(width: 12),
            Text(
              'Memuat tipe pengajuan...',
              style: TextStyle(fontSize: 13, color: Color(0xFF94A3B8)),
            ),
          ],
        ),
      );
    }

    if (_tipeError != null) {
      return GestureDetector(
        onTap: _loadTipes,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: const Color(0xFFFFF1F1),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFFFFCDD2)),
          ),
          child: Row(
            children: [
              const Icon(
                Icons.refresh_rounded,
                size: 16,
                color: Color(0xFFEF4444),
              ),
              const SizedBox(width: 8),
              Text(
                _tipeError!,
                style: const TextStyle(fontSize: 13, color: Color(0xFFEF4444)),
              ),
              const Spacer(),
              const Text(
                'Coba lagi',
                style: TextStyle(
                  fontSize: 12,
                  color: Color(0xFFEF4444),
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      );
    }

    final options = _filteredTipes;
    if (options.isEmpty) {
      return Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Text(
          'Tidak ada tipe untuk kategori $_category',
          style: TextStyle(fontSize: 13, color: Colors.grey.shade400),
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.03),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<TipePengajuan>(
          value: _selectedTipe,
          isExpanded: true,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          borderRadius: BorderRadius.circular(12),
          hint: Text(
            'Pilih tipe ${_category.toLowerCase()}',
            style: TextStyle(fontSize: 14, color: Colors.grey.shade400),
          ),
          items: options
              .map(
                (t) => DropdownMenuItem(
                  value: t,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        t.namaTipe,
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w500,
                          color: Color(0xFF0F172A),
                        ),
                      ),
                      if ((_category == 'Cuti' && t.potongKuota) || t.wajibLampiran)
                        Row(
                          children: [
                            if (_category == 'Cuti' && t.potongKuota)
                              _badge('Potong Kuota', const Color(0xFFF59E0B)),
                            if (_category == 'Cuti' && t.potongKuota && t.wajibLampiran)
                              const SizedBox(width: 4),
                            if (t.wajibLampiran)
                              _badge('Wajib Lampiran', const Color(0xFFEF4444)),
                          ],
                        ),
                    ],
                  ),
                ),
              )
              .toList(),
          onChanged: (v) {
            if (v == null) return;
            setState(() {
              _selectedTipe = v;
              final minStart = _minimumStartDateForCurrentType();
              if (_startDate.isBefore(minStart)) {
                _startDate = minStart;
                if (!_endDate.isAfter(_startDate)) {
                  _endDate = _startDate.add(const Duration(days: 1));
                }
              }
            });
          },
        ),
      ),
    );
  }

  Widget _badge(String text, Color color) {
    return Container(
      margin: const EdgeInsets.only(top: 2),
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 9,
          color: color,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  Widget _leaveFields() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: double.infinity,
          margin: const EdgeInsets.only(bottom: 10),
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: _isCurrentTypeSick()
                ? const Color(0xFFDCFCE7)
                : const Color(0xFFFEF3C7),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Text(
            _isCurrentTypeSick()
                ? 'Izin sakit dapat diajukan hari ini.'
                : 'Pengajuan untuk tipe ini hanya bisa diajukan minimal H-2.',
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: _isCurrentTypeSick()
                  ? const Color(0xFF166534)
                  : const Color(0xFF92400E),
            ),
          ),
        ),
        Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _sectionLabel('Tanggal Mulai'),
                  _dateBox(
                    _fmt.format(_startDate),
                    () => _pickDate(isStart: true),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _sectionLabel('Tanggal Selesai'),
                  _dateBox(
                    _fmt.format(_endDate),
                    () => _pickDate(isStart: false),
                  ),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: const Color(0xFFEFF6FF),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFFBFDBFE)),
          ),
          child: Row(
            children: [
              const Icon(
                Icons.calendar_month_rounded,
                color: Color(0xFF135BEC),
                size: 20,
              ),
              const SizedBox(width: 10),
              Text(
                '$_days Hari',
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF135BEC),
                ),
              ),
              if (_selectedTipe?.potongKuota == true && _category == 'Cuti') ...[
                const Spacer(),
                const Icon(
                  Icons.info_outline_rounded,
                  size: 14,
                  color: Color(0xFFF59E0B),
                ),
                const SizedBox(width: 4),
                const Text(
                  'Memotong kuota cuti',
                  style: TextStyle(fontSize: 11, color: Color(0xFFF59E0B)),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  bool get _requiresQuotaDeduction {
    return _category == 'Cuti' && (_selectedTipe?.potongKuota ?? false);
  }

  bool get _showQuotaInfoCard {
    return _category == 'Cuti' && _requiresQuotaDeduction;
  }

  Widget _quotaInfoCard() {
    final overLimit = _days > _remainingLeaveQuota;
    final consumesQuota = _selectedTipe?.potongKuota ?? false;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: overLimit
            ? const Color(0xFFFFF1F2)
            : consumesQuota
            ? const Color(0xFFF0FDF4)
            : const Color(0xFFEFF6FF),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: overLimit
              ? const Color(0xFFFCA5A5)
              : consumesQuota
              ? const Color(0xFF86EFAC)
              : const Color(0xFFBFDBFE),
        ),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: overLimit
                  ? const Color(0xFFFEE2E2)
                  : consumesQuota
                  ? const Color(0xFFDCFCE7)
                  : const Color(0xFFE0F2FE),
              shape: BoxShape.circle,
            ),
            child: Icon(
              overLimit
                  ? Icons.warning_amber_rounded
                  : consumesQuota
                  ? Icons.beach_access
                  : Icons.info_outline_rounded,
              color: overLimit
                  ? const Color(0xFFB91C1C)
                  : consumesQuota
                  ? const Color(0xFF15803D)
                  : const Color(0xFF135BEC),
              size: 20,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _isLoadingLeaveQuota
                      ? 'Memuat sisa cuti...'
                      : 'Sisa cuti tahun ini: $_remainingLeaveQuota hari',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                    color: overLimit
                        ? const Color(0xFFB91C1C)
                        : consumesQuota
                        ? const Color(0xFF166534)
                        : const Color(0xFF135BEC),
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  overLimit
                      ? 'Jumlah hari pengajuan melebihi sisa kuota cuti.'
                      : consumesQuota
                      ? 'Pengajuan ini akan mengurangi kuota cuti setelah approved.'
                      : 'Tipe yang dipilih tidak mengurangi kuota cuti.',
                  style: TextStyle(
                    fontSize: 12,
                    color: overLimit
                        ? const Color(0xFFB91C1C)
                        : consumesQuota
                        ? const Color(0xFF166534)
                        : const Color(0xFF135BEC),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  DateTime _minimumStartDateForCurrentType() {
    final now = DateTime.now();
    if (_isCurrentTypeSick()) {
      return DateTime(now.year, now.month, now.day);
    }
    final min = now.add(const Duration(days: 2));
    return DateTime(min.year, min.month, min.day);
  }

  bool _isCurrentTypeSick() {
    final typeName = _selectedTipe?.namaTipe.toLowerCase() ?? '';
    return typeName.contains('sakit');
  }

  Widget _dateBox(String label, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
          boxShadow: [
            BoxShadow(color: Colors.black.withValues(alpha: 0.03), blurRadius: 6),
          ],
        ),
        child: Row(
          children: [
            const Icon(
              Icons.calendar_today_rounded,
              size: 16,
              color: Color(0xFF135BEC),
            ),
            const SizedBox(width: 8),
            Text(
              label,
              style: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w500,
                color: Color(0xFF0F172A),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _reasonField() {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: TextField(
        controller: _reasonCtrl,
        maxLines: 4,
        style: const TextStyle(fontSize: 13, color: Color(0xFF0F172A)),
        decoration: InputDecoration(
          hintText: 'Jelaskan secara singkat alasan pengajuan ${_category.toLowerCase()} Anda...',
          hintStyle: TextStyle(color: Colors.grey.shade400, fontSize: 13),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.all(16),
        ),
      ),
    );
  }

  Widget _attachmentField() {
    final wajib = _selectedTipe?.wajibLampiran == true;
    return GestureDetector(
      onTap: _pickFile,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _sectionLabel('Unggah Lampiran'),
              Text(
                wajib ? 'Wajib' : 'Opsional',
                style: TextStyle(
                  fontSize: 11,
                  color: wajib ? const Color(0xFFEF4444) : Colors.grey.shade400,
                  fontStyle: FontStyle.italic,
                  fontWeight: wajib ? FontWeight.w600 : FontWeight.normal,
                ),
              ),
            ],
          ),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: _file != null ? const Color(0xFFEFF6FF) : Colors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: _file != null
                    ? const Color(0xFF135BEC)
                    : (wajib
                          ? const Color(0xFFEF4444).withValues(alpha: 0.4)
                          : Colors.grey.shade200),
                width: _file != null || wajib ? 1.5 : 1,
              ),
            ),
            child: _file != null
                ? Row(
                    children: [
                      const Icon(
                        Icons.insert_drive_file_rounded,
                        color: Color(0xFF135BEC),
                        size: 22,
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          _file!.name,
                          style: const TextStyle(
                            fontSize: 13,
                            color: Color(0xFF135BEC),
                            fontWeight: FontWeight.w500,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      GestureDetector(
                        onTap: () => setState(() => _file = null),
                        child: Icon(
                          Icons.close_rounded,
                          size: 18,
                          color: Colors.grey.shade400,
                        ),
                      ),
                    ],
                  )
                : Column(
                    children: [
                      Icon(
                        Icons.cloud_upload_outlined,
                        color: Colors.grey.shade400,
                        size: 36,
                      ),
                      const SizedBox(height: 8),
                      const Text(
                        'Surat Keterangan Medis / Dokumen Pendukung',
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF0F172A),
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'Ketuk untuk memilih file',
                        style: TextStyle(
                          fontSize: 11,
                          color: Colors.grey.shade400,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 10),
                      Wrap(
                        spacing: 6,
                        children: ['PDF', 'PNG', 'JPG']
                            .map(
                              (ext) => Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 10,
                                  vertical: 3,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.grey.shade100,
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Text(
                                  ext,
                                  style: TextStyle(
                                    fontSize: 10,
                                    color: Colors.grey.shade500,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                              ),
                            )
                            .toList(),
                      ),
                    ],
                  ),
          ),
        ],
      ),
    );
  }

  Widget _submitBtn() {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton(
        onPressed: _submitting ? null : _submit,
        style: ElevatedButton.styleFrom(
          backgroundColor: const Color(0xFF135BEC),
          disabledBackgroundColor: Colors.grey.shade300,
          padding: const EdgeInsets.symmetric(vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
        child: _submitting
            ? const SizedBox(
                height: 18,
                width: 18,
                child: CircularProgressIndicator(
                  strokeWidth: 2.5,
                  valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                ),
              )
            : const Text(
                'Kirim Pengajuan',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.white,
                ),
              ),
      ),
    );
  }

  Widget _cancelBtn() {
    return SizedBox(
      width: double.infinity,
      child: OutlinedButton(
        onPressed: () => Navigator.pop(context),
        style: OutlinedButton.styleFrom(
          side: BorderSide(color: Colors.grey.shade300),
          padding: const EdgeInsets.symmetric(vertical: 12),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
        child: Text(
          'Batal',
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w600,
            color: Colors.grey.shade600,
          ),
        ),
      ),
    );
  }
}
