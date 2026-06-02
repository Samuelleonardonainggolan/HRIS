import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../models/overtime_request.dart';
import '../services/api_service.dart';
import '../utils/overtime_reward_calculator.dart';

class OvertimeRewardPicker extends StatefulWidget {
  final ScrollController scrollController;
  const OvertimeRewardPicker({super.key, required this.scrollController});

  @override
  State<OvertimeRewardPicker> createState() => _OvertimeRewardPickerState();
}

class _OvertimeRewardPickerState extends State<OvertimeRewardPicker> {
  bool _isLoading = true;
  List<OvertimeRequest> _overtimeRequests = [];
  int _basicSalary = 0;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    if (!mounted) return;
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final allOvertime = await ApiService.getMyOvertime();
      final userId = await ApiService.getUserId();
      final salaryResp = await ApiService.getActiveSalary(userId!);
      final basicSalary =
          int.tryParse(salaryResp['basic_salary']?.toString() ?? '0') ?? 0;

      if (mounted) {
        setState(() {
          _basicSalary = basicSalary;
          _overtimeRequests = allOvertime.where((o) {
            if (o.status != 'submitted' && o.status != 'published')
              return false;
            return o.employees.any((e) => e.userId == userId && e.isAgreed);
          }).toList();
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _errorMessage = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _claimReward(
    String id,
    String rewardType, {
    String? rewardDate,
    String? rewardOption,
  }) async {
    try {
      await ApiService.claimOvertimeReward(
        id,
        rewardType,
        rewardDate: rewardDate,
        rewardOption: rewardOption,
      );
      if (mounted) {
        String typeLabel = rewardType == 'money' ? 'Uang' : 'Potong Jam Kerja';
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Reward $typeLabel berhasil dipilih'),
            backgroundColor: Colors.green,
            behavior: SnackBarBehavior.floating,
          ),
        );
        _loadData(); // Refresh list
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal memilih reward: $e'),
            backgroundColor: Colors.red,
            behavior: SnackBarBehavior.floating,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Handle bar
        Center(
          child: Container(
            margin: const EdgeInsets.only(top: 12),
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: Colors.grey.shade300,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
        ),
        const SizedBox(height: 20),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20),
          child: Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: const Color(0xFF8B5CF6).withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.stars_rounded,
                  color: Color(0xFF8B5CF6),
                  size: 24,
                ),
              ),
              const SizedBox(width: 16),
              const Text(
                'Reward Lembur',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFF0F172A),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 8),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20),
          child: Text(
            'Pilih reward untuk lembur yang telah Anda selesaikan',
            style: TextStyle(color: Colors.grey.shade600, fontSize: 14),
          ),
        ),
        const SizedBox(height: 20),
        Expanded(
          child: _isLoading
              ? const Center(child: CircularProgressIndicator())
              : _errorMessage != null
              ? _buildError()
              : _overtimeRequests.isEmpty
              ? _buildEmpty()
              : ListView.builder(
                  controller: widget.scrollController,
                  padding: const EdgeInsets.fromLTRB(20, 0, 20, 20),
                  itemCount: _overtimeRequests.length,
                  itemBuilder: (context, index) {
                    return _buildRewardCard(_overtimeRequests[index]);
                  },
                ),
        ),
      ],
    );
  }

  Widget _buildEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.history_rounded, size: 64, color: Colors.grey.shade200),
          const SizedBox(height: 16),
          Text(
            'Tidak ada lembur pending reward',
            style: TextStyle(
              color: Colors.grey.shade500,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildError() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, color: Colors.redAccent, size: 48),
            const SizedBox(height: 16),
            Text(
              'Gagal memuat data lembur',
              style: TextStyle(
                color: Colors.grey.shade800,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              _errorMessage!,
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.grey.shade600, fontSize: 12),
            ),
            const SizedBox(height: 16),
            TextButton(onPressed: _loadData, child: const Text('Coba Lagi')),
          ],
        ),
      ),
    );
  }

  Widget _buildRewardCard(OvertimeRequest request) {
    final userId = ApiService.currentUser.value?.id;
    final myEntry = request.employees.firstWhere((e) => e.userId == userId);
    final hasReward =
        myEntry.reward != null &&
        myEntry.reward!.rewardType.isNotEmpty &&
        myEntry.reward!.rewardType != 'none';
    final double rewardAmount = myEntry.reward?.rewardNominal ?? (myEntry.reward?.rewardType == 'money'
        ? calculateOvertimeMoneyReward(_basicSalary, request.getDurationHours()).toDouble()
        : 0.0);

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.grey.shade100),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      DateFormat(
                        'EEEE, dd MMM yyyy',
                        'id',
                      ).format(request.date),
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 15,
                        color: Color(0xFF1E293B),
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFF8B5CF6).withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        '${request.getDurationHours().toStringAsFixed(1)} Jam',
                        style: const TextStyle(
                          color: Color(0xFF8B5CF6),
                          fontWeight: FontWeight.bold,
                          fontSize: 11,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  request.reason,
                  style: TextStyle(color: Colors.grey.shade600, fontSize: 13),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 16),
                if (hasReward)
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: const Color(0xFFF8FAFC),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.grey.shade100),
                    ),
                    child: Row(
                      children: [
                        Icon(
                          myEntry.reward!.rewardType == 'money'
                              ? Icons.payments_rounded
                              : Icons.timer_rounded,
                          color: const Color(0xFF8B5CF6),
                          size: 20,
                        ),
                        const SizedBox(width: 12),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Reward: ${myEntry.reward!.rewardTypeDisplay}',
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                fontSize: 13,
                                color: Color(0xFF1E293B),
                              ),
                            ),
                            if (myEntry.reward!.rewardType == 'money')
                              Text(
                                rewardAmount > 0
                                    ? 'Nominal: ${formatMoney(rewardAmount.toInt())}'
                                    : 'Nominal belum tersedia',
                                style: TextStyle(
                                  color: Colors.grey.shade600,
                                  fontSize: 11,
                                ),
                              ),
                            Text(
                              'Status: ${myEntry.reward!.statusDisplay}',
                              style: TextStyle(
                                color: _getStatusColor(myEntry.reward!.status),
                                fontSize: 11,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  )
                else
                  Row(
                    children: [
                      Expanded(
                        child: _buildChoiceBtn(
                          label: 'Uang',
                          icon: Icons.payments_rounded,
                          color: Colors.teal,
                          onTap: () => _confirmClaim(request, 'money'),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: _buildChoiceBtn(
                          label: 'Jam Kerja',
                          icon: Icons.timer_rounded,
                          color: Colors.blue,
                          onTap: () => _confirmClaim(request, 'time_off'),
                        ),
                      ),
                    ],
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChoiceBtn({
    required String label,
    required IconData icon,
    required Color color,
    required VoidCallback onTap,
  }) {
    return Material(
      color: color.withOpacity(0.08),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 10),
          decoration: BoxDecoration(
            border: Border.all(color: color.withOpacity(0.2)),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon, color: color, size: 18),
              const SizedBox(width: 8),
              Text(
                label,
                style: TextStyle(
                  color: color,
                  fontWeight: FontWeight.bold,
                  fontSize: 13,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'pending':
        return Colors.orange;
      case 'granted':
        return Colors.green;
      case 'used':
        return Colors.blue;
      default:
        return Colors.grey;
    }
  }

  void _confirmClaim(OvertimeRequest request, String type) async {
    if (type == 'time_off') {
      final picked = await showDatePicker(
        context: context,
        initialDate: DateTime.now(),
        firstDate: DateTime.now(),
        lastDate: DateTime.now().add(const Duration(days: 90)),
        helpText: 'Pilih Tanggal Reward',
      );
      if (picked != null) {
        final dateStr = DateFormat('yyyy-MM-dd').format(picked);
        _showFinalConfirmation(request, type, dateStr, 'early_out');
      }
      return;
    }

    _showFinalConfirmation(request, 'money', null, null);
  }

  void _showFinalConfirmation(
    OvertimeRequest request,
    String type,
    String? dateStr,
    String? option,
  ) {
    String rewardName = type == 'money' ? 'Uang Lembur' : 'Potong Jam Kerja';
    if (option == 'early_out') rewardName += ' (Pulang Cepat)';
    if (option == 'late_in') rewardName += ' (Masuk Terlambat)';

    final dateInfo = dateStr != null
        ? '\nTanggal: ${DateFormat('dd MMM yyyy', 'id').format(DateTime.parse(dateStr))}'
        : '';
    final double rewardAmount = type == 'money'
        ? calculateOvertimeMoneyReward(_basicSalary, request.getDurationHours()).toDouble()
        : 0.0;
    final amountInfo = type == 'money' && rewardAmount > 0
        ? '\nNominal estimasi: ${formatMoney(rewardAmount.toInt())}'
        : '';

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Konfirmasi Reward'),
        content: Text(
          'Anda akan mengklaim:\n\n$rewardName$amountInfo\nUntuk lembur tanggal ${DateFormat('dd MMM yyyy', 'id').format(request.date)}.$dateInfo\n\nLanjutkan?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _claimReward(
                request.id,
                type,
                rewardDate: dateStr,
                rewardOption: option,
              );
            },
            child: const Text('Klaim'),
          ),
        ],
      ),
    );
  }
}
