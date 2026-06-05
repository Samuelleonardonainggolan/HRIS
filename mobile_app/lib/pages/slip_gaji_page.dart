import 'dart:async';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/models/user_model.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:mobile_app/services/sse_service.dart';

// ─── Model ───────────────────────────────────────────────────────────────────

class SlipGajiData {
  final String periode;
  final int gajiPokok;
  final int lembur;
  final int bonus;
  final int potonganTerlambat;
  final int potonganMangkir;
  final int potonganLainnya;
  final String status;
  final double overtimeHoursPaid;
  final int lateMinutes;
  final int absentDays;
  final int? customTotalBersih;

  SlipGajiData({
    required this.periode,
    required this.gajiPokok,
    required this.lembur,
    required this.bonus,
    required this.potonganTerlambat,
    required this.potonganMangkir,
    required this.potonganLainnya,
    required this.status,
    required this.overtimeHoursPaid,
    required this.lateMinutes,
    required this.absentDays,
    this.customTotalBersih,
  });

  int get totalPotongan => potonganTerlambat + potonganMangkir + potonganLainnya;
  int get totalPendapatan => gajiPokok + lembur + bonus;
  int get totalBersih => customTotalBersih ?? (totalPendapatan - totalPotongan);
}

// ─── Page ─────────────────────────────────────────────────────────────────────

class SlipGajiPage extends StatefulWidget {
  const SlipGajiPage({super.key});

  @override
  State<SlipGajiPage> createState() => _SlipGajiPageState();
}

class _SlipGajiPageState extends State<SlipGajiPage>
    with SingleTickerProviderStateMixin {
  bool _isLoading = true;
  bool _isDownloading = false;
  String? _error;
  SlipGajiData? _slip;
  User? _user;
  StreamSubscription? _sseSubscription;

  late AnimationController _animCtrl;
  late Animation<double> _fadeAnim;

  // Warna utama
  static const _blue = Color(0xFF2E6FF2);
  static const _lightBlue = Color(0xFFE9F0FF);
  static const _green = Color(0xFF22C55E);
  static const _red = Color(0xFFEF4444);
  static const _gray = Color(0xFF8E94A3);

  int _parseMoney(dynamic value) {
    if (value == null) return 0;

    if (value is num) return value.round();

    final normalized = value
        .toString()
        .replaceAll(RegExp(r'[^0-9\-]'), '');
    if (normalized.isEmpty || normalized == '-') return 0;

    return int.tryParse(normalized) ?? 0;
  }

  @override
  void initState() {
    super.initState();
    _animCtrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
    );
    _fadeAnim = CurvedAnimation(parent: _animCtrl, curve: Curves.easeOut);
    _setupSSE();
    _loadData();
  }

  @override
  void dispose() {
    _sseSubscription?.cancel();
    _animCtrl.dispose();
    super.dispose();
  }

  void _setupSSE() {
    _sseSubscription = SSEService().events.listen((event) {
      if (!mounted || event.type == 'ping') return;
      _loadData(silent: true);
    });
  }

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
      final user = await ApiService.getProfile();
      final now = DateTime.now();

      // Ambil data payroll dari backend
      final payrollResp = await ApiService.getMyPayroll(now.month, now.year);
      
      final gajiPokok = _parseMoney(payrollResp['basic_salary_value'] ?? payrollResp['basic_salary']);
      final netSalary = _parseMoney(payrollResp['net_salary_value'] ?? payrollResp['net_salary']);
      
      final lembur = _parseMoney(payrollResp['overtime_pay_value']);
      final bonus = _parseMoney(payrollResp['other_earnings_value']);
      
      final potonganTerlambat = _parseMoney(payrollResp['late_deduction_value']);
      final potonganMangkir = _parseMoney(payrollResp['absent_deduction_value']);
      final potonganLainnya = _parseMoney(payrollResp['other_deductions_value']);
      
      final overtimeHoursPaid = (payrollResp['overtime_hours_paid'] ?? 0.0) is num 
          ? (payrollResp['overtime_hours_paid'] as num).toDouble() 
          : double.tryParse(payrollResp['overtime_hours_paid'].toString()) ?? 0.0;
          
      final lateMinutes = _parseMoney(payrollResp['late_minutes_total']);
      final absentDays = _parseMoney(payrollResp['absent_days']);

      final periode = DateFormat('MMMM yyyy', 'id_ID').format(now);

      if (mounted) {
        setState(() {
          _user = user;
          _slip = SlipGajiData(
            periode: periode,
            gajiPokok: gajiPokok,
            lembur: lembur,
            bonus: bonus,
            potonganTerlambat: potonganTerlambat,
            potonganMangkir: potonganMangkir,
            potonganLainnya: potonganLainnya,
            status: payrollResp['status'] ?? payrollResp['payment_status'] ?? 'Belum Terbayar',
            overtimeHoursPaid: overtimeHoursPaid,
            lateMinutes: lateMinutes,
            absentDays: absentDays,
            customTotalBersih: netSalary,
          );
          _isLoading = false;
        });
        _animCtrl.forward();
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString().replaceFirst('Exception: ', '');
          _isLoading = false;
        });
      }
    }
  }

  String _formatOvertimeHours(double hours) {
    return hours % 1 == 0 ? hours.toStringAsFixed(0) : hours.toStringAsFixed(1);
  }

  // ─── PDF generation ────────────────────────────────────────────────────────

  Future<void> _downloadPdf() async {
    if (_slip == null || _user == null) return;
    setState(() => _isDownloading = true);
    try {
      final pdf = pw.Document();
      final slip = _slip!;
      final user = _user!;
      final fmt = NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);

      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          margin: const pw.EdgeInsets.all(40),
          build: (ctx) => pw.Column(
            crossAxisAlignment: pw.CrossAxisAlignment.start,
            children: [
              // Header
              pw.Container(
                padding: const pw.EdgeInsets.all(16),
                decoration: pw.BoxDecoration(
                  color: PdfColor.fromHex('#2E6FF2'),
                  borderRadius: pw.BorderRadius.circular(8),
                ),
                child: pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  children: [
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        pw.Text('SLIP GAJI',
                            style: pw.TextStyle(
                                color: PdfColors.white,
                                fontSize: 20,
                                fontWeight: pw.FontWeight.bold)),
                        pw.SizedBox(height: 4),
                        pw.Text('Periode: ${slip.periode}',
                            style: const pw.TextStyle(
                                color: PdfColors.white, fontSize: 11)),
                      ],
                    ),
                    pw.Text(slip.status,
                        style: pw.TextStyle(
                            color: PdfColors.white,
                            fontSize: 12,
                            fontWeight: pw.FontWeight.bold)),
                  ],
                ),
              ),
              pw.SizedBox(height: 20),
              // Info karyawan
              pw.Text('Informasi Karyawan',
                  style: pw.TextStyle(
                      fontSize: 13, fontWeight: pw.FontWeight.bold)),
              pw.Divider(),
              _pdfRow('Nama', user.fullName),
              _pdfRow('NIK', user.nik ?? '-'),
              _pdfRow('Departemen', user.department ?? '-'),
              _pdfRow('Jabatan', user.position ?? '-'),
              pw.SizedBox(height: 16),
              // Rincian
              pw.Text('Rincian Gaji',
                  style: pw.TextStyle(
                      fontSize: 13, fontWeight: pw.FontWeight.bold)),
              pw.Divider(),
              _pdfRow('Gaji Pokok', fmt.format(slip.gajiPokok)),
              if (slip.lembur > 0)
                _pdfRow(
                    'Lembur (${_formatOvertimeHours(slip.overtimeHoursPaid)} jam)',
                    fmt.format(slip.lembur)),
              if (slip.bonus > 0)
                _pdfRow('Pendapatan Lain/Bonus', fmt.format(slip.bonus)),
              pw.SizedBox(height: 8),
              pw.Text('Potongan',
                  style: pw.TextStyle(
                      fontSize: 13, fontWeight: pw.FontWeight.bold, color: PdfColors.red)),
              pw.Divider(),
              if (slip.potonganTerlambat > 0)
                _pdfRow('Terlambat (${slip.lateMinutes} mnt)', '- ${fmt.format(slip.potonganTerlambat)}', isDeduction: true),
              if (slip.potonganMangkir > 0)
                _pdfRow('Mangkir (${slip.absentDays} hari)', '- ${fmt.format(slip.potonganMangkir)}', isDeduction: true),
              if (slip.potonganLainnya > 0 || (slip.potonganTerlambat == 0 && slip.potonganMangkir == 0))
                _pdfRow('Potongan Lain (BPJS dll)', '- ${fmt.format(slip.potonganLainnya)}', isDeduction: true),
              pw.Divider(),
              pw.Row(
                mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                children: [
                  pw.Text('Total Gaji Bersih',
                      style: pw.TextStyle(
                          fontSize: 14, fontWeight: pw.FontWeight.bold)),
                  pw.Text(fmt.format(slip.totalBersih),
                      style: pw.TextStyle(
                          fontSize: 14,
                          fontWeight: pw.FontWeight.bold,
                          color: PdfColor.fromHex('#2E6FF2'))),
                ],
              ),
              pw.SizedBox(height: 40),
              pw.Text(
                'Dokumen ini digenerate secara otomatis oleh sistem HRIS.',
                style: const pw.TextStyle(
                    fontSize: 9, color: PdfColors.grey),
              ),
            ],
          ),
        ),
      );

      final dir = await getTemporaryDirectory();
      final file = File(
          '${dir.path}/slip_gaji_${slip.periode.replaceAll(' ', '_')}.pdf');
      await file.writeAsBytes(await pdf.save());

      await Share.shareXFiles(
        [XFile(file.path)],
        subject: 'Slip Gaji ${slip.periode}',
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal membuat PDF: $e'),
            backgroundColor: _red,
            behavior: SnackBarBehavior.floating,
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isDownloading = false);
    }
  }

  pw.Widget _pdfRow(String label, String value, {bool isDeduction = false}) => pw.Padding(
        padding: const pw.EdgeInsets.symmetric(vertical: 4),
        child: pw.Row(
          mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
          children: [
            pw.Text(label,
                style: const pw.TextStyle(fontSize: 11, color: PdfColors.grey700)),
            pw.Text(value,
                style: pw.TextStyle(
                    fontSize: 11, 
                    fontWeight: pw.FontWeight.bold,
                    color: isDeduction ? PdfColors.red : PdfColors.black)),
          ],
        ),
      );

  // ─── Build ─────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      body: SafeArea(
        child: Column(
          children: [
            _buildAppBar(),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _error != null
                      ? _buildError()
                      : FadeTransition(
                          opacity: _fadeAnim,
                          child: _buildContent(),
                        ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAppBar() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 12),
      color: Colors.white,
      child: Row(
        children: [
          IconButton(
            icon: const Icon(Icons.arrow_back_ios_new_rounded, size: 20),
            onPressed: () => Navigator.pop(context),
          ),
          const Expanded(
            child: Text(
              'Rincian Gaji',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 17,
                fontWeight: FontWeight.bold,
                color: Color(0xFF1A1C1E),
              ),
            ),
          ),
          const SizedBox(width: 48), // spacer
        ],
      ),
    );
  }

  Widget _buildError() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.error_outline, size: 48, color: _red),
            const SizedBox(height: 16),
            Text(
              _error!,
              textAlign: TextAlign.center,
              style: const TextStyle(color: _gray),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: _loadData,
              icon: const Icon(Icons.refresh),
              label: const Text('Coba Lagi'),
              style: ElevatedButton.styleFrom(
                backgroundColor: _blue,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12)),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildContent() {
    final slip = _slip!;
    final fmt = NumberFormat.currency(
        locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);

    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 32),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Summary card ──
          _buildSummaryCard(slip, fmt),
          const SizedBox(height: 28),

          // ── Section label ──
          const Padding(
            padding: EdgeInsets.only(left: 4, bottom: 14),
            child: Text(
              'RINCIAN KOMPONEN',
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w700,
                letterSpacing: 1.2,
                color: _gray,
              ),
            ),
          ),

          // ── Component items ──
          _buildComponentItem(
            icon: Icons.receipt_long_rounded,
            iconBg: _lightBlue,
            iconColor: _blue,
            title: 'Gaji Pokok',
            subtitle: 'Gaji bulanan standar',
            amount: fmt.format(slip.gajiPokok),
            amountColor: const Color(0xFF1A1C1E),
          ),
          if (slip.lembur > 0) ...[
            const SizedBox(height: 12),
            _buildComponentItem(
              icon: Icons.wb_sunny_rounded,
              iconBg: const Color(0xFFFFF7ED),
              iconColor: Colors.orange,
              title: 'Lembur',
              subtitle: '${_formatOvertimeHours(slip.overtimeHoursPaid)} jam',
              amount: fmt.format(slip.lembur),
              amountColor: const Color(0xFF1A1C1E),
            ),
          ],
          if (slip.bonus > 0) ...[
            const SizedBox(height: 12),
            _buildComponentItem(
              icon: Icons.star_rounded,
              iconBg: const Color(0xFFFFFBEB),
              iconColor: const Color(0xFFF59E0B),
              title: 'Pendapatan Lain',
              subtitle: 'Bonus & penyesuaian positif',
              amount: fmt.format(slip.bonus),
              amountColor: const Color(0xFFF59E0B),
            ),
          ],
          const SizedBox(height: 28),
          const Padding(
            padding: EdgeInsets.only(left: 4, bottom: 14),
            child: Text(
              'RINCIAN POTONGAN',
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w700,
                letterSpacing: 1.2,
                color: _red,
              ),
            ),
          ),
          if (slip.potonganTerlambat > 0) ...[
            _buildComponentItem(
              icon: Icons.timer_off_rounded,
              iconBg: const Color(0xFFFEF2F2),
              iconColor: _red,
              title: 'Keterlambatan',
              subtitle: 'Total telat ${slip.lateMinutes} menit',
              amount: '- ${fmt.format(slip.potonganTerlambat)}',
              amountColor: _red,
            ),
            const SizedBox(height: 12),
          ],
          if (slip.potonganMangkir > 0) ...[
            _buildComponentItem(
              icon: Icons.event_busy_rounded,
              iconBg: const Color(0xFFFEF2F2),
              iconColor: _red,
              title: 'Mangkir / Alpha',
              subtitle: 'Tidak hadir ${slip.absentDays} hari',
              amount: '- ${fmt.format(slip.potonganMangkir)}',
              amountColor: _red,
            ),
            const SizedBox(height: 12),
          ],
          if (slip.potonganLainnya > 0 || (slip.potonganTerlambat == 0 && slip.potonganMangkir == 0)) ...[
            _buildComponentItem(
              icon: Icons.remove_circle_outline_rounded,
              iconBg: const Color(0xFFFEF2F2),
              iconColor: _red,
              title: 'Potongan Lainnya',
              subtitle: 'BPJS / Penyesuaian negatif',
              amount: '- ${fmt.format(slip.potonganLainnya)}',
              amountColor: _red,
            ),
          ],

          const SizedBox(height: 32),

          // ── Download button ──
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: _isDownloading ? null : _downloadPdf,
              icon: _isDownloading
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                          strokeWidth: 2, color: Colors.white),
                    )
                  : const Icon(Icons.download_rounded, size: 22),
              label: Text(
                _isDownloading ? 'Menyiapkan PDF...' : 'Unduh Slip Gaji (PDF)',
                style: const TextStyle(
                    fontSize: 15, fontWeight: FontWeight.bold),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: _blue,
                foregroundColor: Colors.white,
                disabledBackgroundColor: _blue.withOpacity(0.6),
                disabledForegroundColor: Colors.white70,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(18)),
                elevation: 4,
                shadowColor: _blue.withOpacity(0.4),
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ── Summary card ────────────────────────────────────────────────────────────

  Widget _buildSummaryCard(SlipGajiData slip, NumberFormat fmt) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFF2E6FF2), Color(0xFF5B8FF9)],
        ),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: _blue.withOpacity(0.35),
            blurRadius: 24,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Top row
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Total Gaji Bersih',
                    style: TextStyle(
                        color: Colors.white,
                        fontSize: 13,
                        fontWeight: FontWeight.w500),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    'Periode: ${slip.periode}',
                    style: TextStyle(
                        color: Colors.white.withOpacity(0.75), fontSize: 12),
                  ),
                ],
              ),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 5),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  slip.status.toUpperCase(),
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 10,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 0.8,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          // Total amount
          Text(
            fmt.format(slip.totalBersih),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 34,
              fontWeight: FontWeight.w800,
              letterSpacing: -0.5,
            ),
          ),
          const SizedBox(height: 20),
          // Chips row
          Row(
            children: [
              Expanded(
                child: _buildSummaryChip(
                  icon: Icons.trending_up_rounded,
                  label: 'Bonus',
                  value: slip.bonus > 0
                      ? '+${fmt.format(slip.bonus)}'
                      : '+Rp 0',
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildSummaryChip(
                  icon: Icons.trending_down_rounded,
                  label: 'Potongan',
                  value: slip.totalPotongan > 0
                      ? '- ${fmt.format(slip.totalPotongan)}'
                      : 'Rp 0',
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryChip({
    required IconData icon,
    required String label,
    required String value,
  }) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.15),
        borderRadius: BorderRadius.circular(14),
      ),
      child: Row(
        children: [
          Container(
            width: 32,
            height: 32,
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.2),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, color: Colors.white, size: 16),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label,
                    style: TextStyle(
                        color: Colors.white.withOpacity(0.8),
                        fontSize: 10)),
                const SizedBox(height: 2),
                Text(value,
                    style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.bold),
                    overflow: TextOverflow.ellipsis),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ── Component item ──────────────────────────────────────────────────────────

  Widget _buildComponentItem({
    required IconData icon,
    required Color iconBg,
    required Color iconColor,
    required String title,
    required String subtitle,
    required String amount,
    required Color amountColor,
  }) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: iconBg,
              borderRadius: BorderRadius.circular(14),
            ),
            child: Icon(icon, color: iconColor, size: 22),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title,
                    style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF1A1C1E))),
                const SizedBox(height: 2),
                Text(subtitle,
                    style: const TextStyle(fontSize: 11, color: _gray)),
              ],
            ),
          ),
          Text(
            amount,
            style: TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.bold,
              color: amountColor,
            ),
          ),
        ],
      ),
    );
  }
}
