// lib/pages/new_request_page.dart
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:file_picker/file_picker.dart';
import 'package:mobile_app/services/api_service.dart';

class NewRequestPage extends StatefulWidget {
  const NewRequestPage({super.key});
  @override
  State<NewRequestPage> createState() => _NewRequestPageState();
}

class _NewRequestPageState extends State<NewRequestPage> {
  // ─── State tipe pengajuan dari API ─────────────────────────────────────────
  List<TipePengajuan> _allTipes = [];
  TipePengajuan? _selectedTipe;
  bool _isLoadingTipes = true;
  String? _tipeError;

  // 'Izin' | 'Cuti' | 'Lembur'
  // Kategori ini dipakai untuk filter tampilan form.
  // Saat tipe dipilih dari API, kategori akan mengikuti tipe tersebut.
  String _category = 'Izin';

  // Common
  DateTime _startDate = DateTime.now();
  DateTime _endDate = DateTime.now();
  final _reasonCtrl = TextEditingController();
  PlatformFile? _file;
  bool _submitting = false;
  bool _isLoadingLeaveQuota = true;
  int _remainingLeaveQuota = 0;

  // Lembur-only
  TimeOfDay _startTime = const TimeOfDay(hour: 17, minute: 0);
  TimeOfDay _endTime = const TimeOfDay(hour: 20, minute: 0);
  String _compensation = 'Upah';

  final _fmt = DateFormat('MM/dd/yyyy');

  int get _days => _endDate.difference(_startDate).inDays + 1;

  int get _lemburMins {
    final s = _startTime.hour * 60 + _startTime.minute;
    final e = _endTime.hour * 60 + _endTime.minute;
    return e > s ? e - s : 0;
  }

  String get _lemburStr {
    final h = _lemburMins ~/ 60;
    final m = _lemburMins % 60;
    return m == 0 ? '$h Jam 00 Menit' : '$h Jam $m Menit';
  }

  String get _kompensasiStr {
    final rp = _lemburMins * 10000 ~/ 60;
    if (_compensation == 'Upah')
      return 'Rp ${NumberFormat('#,###', 'id').format(rp)}';
    return '${(_lemburMins / 60).toStringAsFixed(1)} Jam Cuti';
  }

  // ─── Filtered tipe berdasarkan kategori ────────────────────────────────────
  List<TipePengajuan> get _filteredTipes {
    if (_category == 'Lembur') return [];
    return _allTipes.where((t) => t.namaKategori == _category).toList();
  }

  @override
  void initState() {
    super.initState();
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
          // Default pilih tipe pertama yang kategorinya sesuai
          final firstMatch = tipes
              .where((t) => t.namaKategori == _category)
              .toList();
          _selectedTipe = firstMatch.isNotEmpty ? firstMatch.first : null;
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
    final firstAllowedDate = isStart
        ? (_category == 'Lembur' ? DateTime.now() : minStartDate)
        : _startDate;
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

  Future<void> _pickTime({required bool isStart}) async {
    final picked = await showTimePicker(
      context: context,
      initialTime: isStart ? _startTime : _endTime,
      builder: (ctx, child) => Theme(
        data: Theme.of(ctx).copyWith(
          colorScheme: const ColorScheme.light(primary: Color(0xFF135BEC)),
        ),
        child: child!,
      ),
    );
    if (picked == null) return;
    setState(() {
      if (isStart)
        _startTime = picked;
      else
        _endTime = picked;
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

    if (_category != 'Lembur' && _selectedTipe == null) {
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
      if (_category == 'Lembur') {
        final dateFormat = DateFormat('yyyy-MM-dd');
        final startTimeStr = '${_startTime.hour.toString().padLeft(2, '0')}:${_startTime.minute.toString().padLeft(2, '0')}';
        final endTimeStr = '${_endTime.hour.toString().padLeft(2, '0')}:${_endTime.minute.toString().padLeft(2, '0')}';

        await ApiService.submitOvertime(
          tanggal: dateFormat.format(_startDate),
          startTime: startTimeStr,
          endTime: endTimeStr,
          alasan: _reasonCtrl.text.trim(),
          total: _kompensasiStr,
        );

        if (mounted) {
          Navigator.pop(context, true);
          _showSnack('Pengajuan Lembur berhasil dikirim');
        }
        return;
      }

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

      await ApiService.submitPengajuan(
        tipePengajuanId: _selectedTipe!.id,
        tanggalMulai: dateFormat.format(_startDate),
        tanggalSelesai: dateFormat.format(_endDate),
        totalHari: _days,
        alasan: _reasonCtrl.text.trim(),
        dokumenUrl: _file?.path,
      );

      if (mounted) {
        Navigator.pop(context, true);
        _showSnack('Pengajuan ${_selectedTipe!.namaTipe} berhasil dikirim');
      }
    } catch (e) {
      if (mounted) {
        _showSnack('Gagal mengirimkan pengajuan: $e', isError: true);
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
            if (_category != 'Lembur') ...[
              _sectionLabel('Tipe Pengajuan'),
              _tipeSelector(),
              const SizedBox(height: 20),
            ],

            if (_requiresQuotaDeduction) ...[
              _quotaInfoCard(),
              const SizedBox(height: 20),
            ],

            if (_category == 'Lembur') _lemburFields(),
            if (_category != 'Lembur') _leaveFields(),
            const SizedBox(height: 20),

            _sectionLabel(
              _category == 'Lembur' ? 'Alasan / Deskripsi Pekerjaan' : 'Alasan',
            ),
            _reasonField(),

            if (_category != 'Lembur') ...[
              const SizedBox(height: 20),
              _attachmentField(),
            ],

            if (_category == 'Lembur') ...[
              const SizedBox(height: 20),
              _compensationSection(),
              const SizedBox(height: 16),
              _lemburSummary(),
            ],

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

  // ── Selector kategori (Izin / Cuti / Lembur) ──────────────────────────────
  Widget _categorySelector() {
    final categories = ['Izin', 'Cuti', 'Lembur'];
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
              margin: EdgeInsets.only(right: cat != 'Lembur' ? 8 : 0),
              padding: const EdgeInsets.symmetric(vertical: 12),
              decoration: BoxDecoration(
                color: sel ? const Color(0xFF135BEC) : Colors.white,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: sel ? const Color(0xFF135BEC) : Colors.grey.shade200,
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.04),
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

  // ── Selector tipe dari API ─────────────────────────────────────────────────
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
            color: Colors.black.withOpacity(0.03),
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
                      if (t.potongKuota || t.wajibLampiran)
                        Row(
                          children: [
                            if (t.potongKuota)
                              _badge('Potong Kuota', const Color(0xFFF59E0B)),
                            if (t.potongKuota && t.wajibLampiran)
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
        color: color.withOpacity(0.1),
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
                ? 'Izin sakit dapat diajukan mulai hari ini.'
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
              if (_selectedTipe?.potongKuota == true) ...[
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
    return _category != 'Lembur' && (_selectedTipe?.potongKuota ?? false);
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
    if (_category == 'Lembur' || _isCurrentTypeSick()) {
      return DateTime(now.year, now.month, now.day);
    }
    final min = now.add(const Duration(days: 2));
    return DateTime(min.year, min.month, min.day);
  }

  bool _isCurrentTypeSick() {
    final typeName = _selectedTipe?.namaTipe.toLowerCase() ?? '';
    return typeName.contains('sakit');
  }

  Widget _lemburFields() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _sectionLabel('Tanggal'),
        _dateBox(_fmt.format(_startDate), () => _pickDate(isStart: true)),
        const SizedBox(height: 16),
        Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _sectionLabel('Jam Mulai'),
                  _timeBox(
                    _startTime.format(context),
                    () => _pickTime(isStart: true),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _sectionLabel('Jam Selesai'),
                  _timeBox(
                    _endTime.format(context),
                    () => _pickTime(isStart: false),
                  ),
                ],
              ),
            ),
          ],
        ),
      ],
    );
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
            BoxShadow(color: Colors.black.withOpacity(0.03), blurRadius: 6),
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

  Widget _timeBox(String label, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Row(
          children: [
            const Icon(
              Icons.access_time_rounded,
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
          hintText: _category == 'Lembur'
              ? 'Jelaskan detail pekerjaan yang dilakukan saat lembur...'
              : 'Jelaskan secara singkat alasan pengajuan ${_category.toLowerCase()} Anda...',
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
                          ? const Color(0xFFEF4444).withOpacity(0.4)
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

  Widget _compensationSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.only(bottom: 8),
          child: Text(
            'OPSI KOMPENSASI',
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              color: Color(0xFF64748B),
              letterSpacing: 0.8,
            ),
          ),
        ),
        ...['Upah', 'Cuti'].map((opt) {
          final sel = _compensation == opt;
          return GestureDetector(
            onTap: () => setState(() => _compensation = opt),
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              margin: const EdgeInsets.only(bottom: 8),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(14),
                border: Border.all(
                  color: sel ? const Color(0xFF135BEC) : Colors.grey.shade200,
                  width: sel ? 1.5 : 1,
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.03),
                    blurRadius: 8,
                  ),
                ],
              ),
              child: Row(
                children: [
                  Container(
                    height: 20,
                    width: 20,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: sel
                            ? const Color(0xFF135BEC)
                            : Colors.grey.shade400,
                        width: 2,
                      ),
                      color: sel ? const Color(0xFF135BEC) : Colors.transparent,
                    ),
                    child: sel
                        ? const Icon(
                            Icons.check_rounded,
                            color: Colors.white,
                            size: 12,
                          )
                        : null,
                  ),
                  const SizedBox(width: 14),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        opt == 'Upah'
                            ? 'Upah Lembur (Uang)'
                            : 'Pertukaran Jam Kerja (Cuti/Off)',
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFF0F172A),
                        ),
                      ),
                      Text(
                        opt == 'Upah'
                            ? 'Kompensasi dibayarkan dalam bentuk uang'
                            : 'Ganti waktu lembur dengan pengurangan jam kerja',
                        style: TextStyle(
                          fontSize: 11,
                          color: Colors.grey.shade500,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          );
        }).toList(),
      ],
    );
  }

  Widget _lemburSummary() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 10),
        ],
      ),
      child: Column(
        children: [
          _summaryRow(
            Icons.timelapse_rounded,
            'Total Durasi',
            _lemburStr,
            const Color(0xFF135BEC),
          ),
          const SizedBox(height: 8),
          _summaryRow(
            Icons.payments_rounded,
            'Estimasi Kompensasi',
            _kompensasiStr,
            const Color(0xFF2ECC71),
          ),
        ],
      ),
    );
  }

  Widget _summaryRow(IconData icon, String label, String value, Color color) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Row(
          children: [
            Icon(icon, size: 16, color: color),
            const SizedBox(width: 8),
            Text(
              label,
              style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
            ),
          ],
        ),
        Text(
          value,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
      ],
    );
  }

  Widget _submitBtn() {
    final quotaBlocked =
        _requiresQuotaDeduction && _days > _remainingLeaveQuota;
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton(
        onPressed: (_submitting || quotaBlocked) ? null : _submit,
        style: ElevatedButton.styleFrom(
          backgroundColor: quotaBlocked
              ? const Color(0xFF94A3B8)
              : const Color(0xFF135BEC),
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(vertical: 16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(14),
          ),
          elevation: 0,
        ),
        child: _submitting
            ? const SizedBox(
                height: 20,
                width: 20,
                child: CircularProgressIndicator(
                  color: Colors.white,
                  strokeWidth: 2,
                ),
              )
            : Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Text(
                    'Kirim Pengajuan',
                    style: TextStyle(fontSize: 15, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(width: 8),
                  Container(
                    padding: const EdgeInsets.all(4),
                    decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.2),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.send_rounded, size: 14),
                  ),
                ],
              ),
      ),
    );
  }

  Widget _cancelBtn() {
    return SizedBox(
      width: double.infinity,
      child: TextButton(
        onPressed: () => Navigator.pop(context),
        style: TextButton.styleFrom(
          padding: const EdgeInsets.symmetric(vertical: 14),
        ),
        child: Text(
          'Batal',
          style: TextStyle(
            fontSize: 14,
            color: Colors.grey.shade500,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }
}
