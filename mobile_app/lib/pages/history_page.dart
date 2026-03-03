import 'dart:io';
import 'package:flutter/material.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:mobile_app/widgets/custom_app_bar.dart';
import 'package:mobile_app/models/attendance_model.dart';
import 'package:intl/intl.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:pdf/pdf.dart';
import 'package:printing/printing.dart';
import 'package:share_plus/share_plus.dart';
import 'package:path_provider/path_provider.dart';

class HistoryPage extends StatefulWidget {
  const HistoryPage({super.key});

  @override
  State<HistoryPage> createState() => _HistoryPageState();
}

class _HistoryPageState extends State<HistoryPage> with SingleTickerProviderStateMixin {
  DateTime _selectedMonth = DateTime.now();
  String _selectedFilter = 'All';
  List<AttendanceRecord> _attendanceRecords = [];
  List<AttendanceRecord> _filteredRecords = [];
  bool _isLoading = true;
  
  late TabController _tabController;
  
  final List<String> _filters = ['All', 'On Time', 'Late', 'Absent', 'Overtime'];

  final GlobalKey<ScaffoldState> _scaffoldKey = GlobalKey<ScaffoldState>();

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _loadDummyData();
  }

  Future<void> _loadDummyData() async {
    setState(() => _isLoading = true);
    
    // Simulasi loading data
    await Future.delayed(const Duration(seconds: 1));
    
    // Dummy data
    _attendanceRecords = [
      AttendanceRecord(
        id: '1',
        date: DateTime.now().subtract(const Duration(days: 1)),
        clockIn: '08:30',
        clockOut: '17:30',
        status: 'On Time',
        workHours: 9.0,
        overtimeHours: 1.0,
      ),
      AttendanceRecord(
        id: '2',
        date: DateTime.now().subtract(const Duration(days: 2)),
        clockIn: '08:45',
        clockOut: '17:45',
        status: 'On Time',
        workHours: 9.0,
        overtimeHours: 1.0,
      ),
      AttendanceRecord(
        id: '3',
        date: DateTime.now().subtract(const Duration(days: 3)),
        clockIn: '09:15',
        clockOut: '18:00',
        status: 'Late',
        workHours: 8.75,
        overtimeHours: 1.5,
      ),
      AttendanceRecord(
        id: '4',
        date: DateTime.now().subtract(const Duration(days: 4)),
        clockIn: '08:20',
        clockOut: '17:20',
        status: 'On Time',
        workHours: 9.0,
        overtimeHours: 1.0,
      ),
      AttendanceRecord(
        id: '5',
        date: DateTime.now().subtract(const Duration(days: 5)),
        clockIn: '--:--',
        clockOut: '--:--',
        status: 'Absent',
        workHours: 0.0,
        overtimeHours: 0.0,
      ),
      AttendanceRecord(
        id: '6',
        date: DateTime.now().subtract(const Duration(days: 6)),
        clockIn: '08:00',
        clockOut: '19:00',
        status: 'Overtime',
        workHours: 11.0,
        overtimeHours: 3.0,
      ),
      AttendanceRecord(
        id: '7',
        date: DateTime.now().subtract(const Duration(days: 7)),
        clockIn: '08:30',
        clockOut: '17:30',
        status: 'On Time',
        workHours: 9.0,
        overtimeHours: 1.0,
      ),
    ];
    
    _filterRecords();
    setState(() => _isLoading = false);
  }

  void _filterRecords() {
    setState(() {
      _filteredRecords = _attendanceRecords.where((record) {
        // Filter by month
        bool monthMatch = record.date.month == _selectedMonth.month && 
                          record.date.year == _selectedMonth.year;
        
        // Filter by status
        bool statusMatch = _selectedFilter == 'All' || 
                          record.status == _selectedFilter;
        
        return monthMatch && statusMatch;
      }).toList();
      
      // Sort by date descending
      _filteredRecords.sort((a, b) => b.date.compareTo(a.date));
    });
  }

  Future<void> _exportToPDF() async {
    try {
      final pdf = pw.Document();
      
      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          build: (pw.Context context) {
            return pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.start,
              children: [
                pw.Text(
                  'Attendance Report',
                  style: pw.TextStyle(
                    fontSize: 24,
                    fontWeight: pw.FontWeight.bold,
                  ),
                ),
                pw.SizedBox(height: 10),
                pw.Text(
                  'Month: ${DateFormat('MMMM yyyy').format(_selectedMonth)}',
                  style: const pw.TextStyle(fontSize: 14),
                ),
                pw.Text(
                  'Filter: $_selectedFilter',
                  style: const pw.TextStyle(fontSize: 14),
                ),
                pw.SizedBox(height: 20),
                pw.TableHelper.fromTextArray(
                  headers: ['Date', 'Clock In', 'Clock Out', 'Status', 'Work Hours'],
                  data: _filteredRecords.map((record) {
                    return [
                      DateFormat('dd MMM yyyy').format(record.date),
                      record.clockIn,
                      record.clockOut,
                      record.status,
                      '${record.workHours}h',
                    ];
                  }).toList(),
                ),
                pw.SizedBox(height: 20),
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  children: [
                    pw.Text('Total Days: ${_filteredRecords.length}'),
                    pw.Text(
                      'Total Hours: ${_filteredRecords.fold(0.0, (sum, record) => sum + record.workHours).toStringAsFixed(1)}h',
                    ),
                  ],
                ),
              ],
            );
          },
        ),
      );

      // Save PDF to temporary file
      final output = await getTemporaryDirectory();
      final file = File('${output.path}/attendance_report_${DateTime.now().millisecondsSinceEpoch}.pdf');
      await file.writeAsBytes(await pdf.save());

      // Share PDF
      await Share.shareXFiles(
        [XFile(file.path)],
        text: 'Attendance Report ${DateFormat('MMMM yyyy').format(_selectedMonth)}',
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('PDF exported successfully'),
            backgroundColor: AppTheme.successColor,
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.all(Radius.circular(12))),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error exporting PDF: $e'),
            backgroundColor: AppTheme.errorColor,
            behavior: SnackBarBehavior.floating,
            shape: const RoundedRectangleBorder(borderRadius: BorderRadius.all(Radius.circular(12))),
          ),
        );
      }
    }
  }

  void _showMonthPicker() {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          title: const Text('Select Month'),
          content: SizedBox(
            width: 300,
            height: 300,
            child: YearPicker(
              firstDate: DateTime(2020),
              lastDate: DateTime.now(),
              initialDate: _selectedMonth,
              selectedDate: _selectedMonth,
              onChanged: (DateTime dateTime) {
                setState(() {
                  _selectedMonth = dateTime;
                  _filterRecords();
                });
                Navigator.pop(context);
              },
            ),
          ),
        );
      },
    );
  }

  String _getGreeting() {
    final hour = DateTime.now().hour;
    if (hour < 12) return "Good Morning";
    if (hour < 15) return "Good Afternoon";
    if (hour < 18) return "Good Evening";
    return "Good Night";
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: true,
      child: Scaffold(
        key: _scaffoldKey,
        backgroundColor: const Color(0xFFF8FAFC),
        body: SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              double horizontalPadding = constraints.maxWidth > 600 ? 40 : 20;
              double maxWidth = constraints.maxWidth > 600 ? 600 : double.infinity;
              
              return Center(
                child: Container(
                  constraints: BoxConstraints(maxWidth: maxWidth),
                  child: Column(
                    children: [
                      // Header yang sama dengan dashboard
                      _buildHeader(horizontalPadding),
                      
                      Expanded(
                        child: SingleChildScrollView(
                          physics: const BouncingScrollPhysics(),
                          padding: EdgeInsets.symmetric(horizontal: horizontalPadding),
                          child: Column(
                            children: [
                              const SizedBox(height: 16),
                              
                              // Filter Section
                              _buildFilterSection(),
                              
                              const SizedBox(height: 16),
                              
                              // Summary Cards
                              _buildSummaryCards(),
                              
                              const SizedBox(height: 20),
                              
                              // Tab Bar
                              Container(
                                color: Colors.white,
                                child: TabBar(
                                  controller: _tabController,
                                  tabs: const [
                                    Tab(text: 'List View'),
                                    Tab(text: 'Calendar View'),
                                  ],
                                  labelColor: const Color(0xFF135BEC),
                                  unselectedLabelColor: Colors.grey,
                                  indicatorColor: const Color(0xFF135BEC),
                                ),
                              ),
                              
                              const SizedBox(height: 12),
                              
                              // Tab Content
                              _isLoading
                                  ? const SizedBox(
                                      height: 200,
                                      child: Center(child: CircularProgressIndicator())
                                    )
                                  : _filteredRecords.isEmpty
                                      ? _buildEmptyState()
                                      : SizedBox(
                                          height: constraints.maxHeight * 0.55,
                                          child: TabBarView(
                                            controller: _tabController,
                                            children: [
                                              ListView.builder(
                                                padding: const EdgeInsets.only(bottom: 16),
                                                itemCount: _filteredRecords.length,
                                                itemBuilder: (context, index) {
                                                  final record = _filteredRecords[index];
                                                  return _buildAttendanceCard(record);
                                                },
                                              ),
                                              _buildCalendarView(),
                                            ],
                                          ),
                                        ),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  // ================= HEADER (Sama dengan Dashboard) =================
  Widget _buildHeader(double horizontalPadding) {
    return Container(
      padding: EdgeInsets.symmetric(horizontal: horizontalPadding, vertical: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(30),
          bottomRight: Radius.circular(30),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 20,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              const SizedBox(width: 8),
              Stack(
                children: [
                  Hero(
                    tag: 'profile',
                    child: Container(
                      height: 52,
                      width: 52,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: const LinearGradient(
                          colors: [Color(0xFF135BEC), Color(0xFF3B7BF6)],
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: const Color(0xFF135BEC).withOpacity(0.3),
                            blurRadius: 10,
                            offset: const Offset(0, 3),
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
                              'https://ui-avatars.com/api/?name=Alex+Morgan&background=135BEC&color=fff&size=100',
                              fit: BoxFit.cover,
                              errorBuilder: (context, error, stackTrace) {
                                return Container(
                                  color: Colors.white,
                                  child: const Icon(
                                    Icons.person,
                                    color: Color(0xFF135BEC),
                                    size: 30,
                                  ),
                                );
                              },
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                  Positioned(
                    bottom: 2,
                    right: 2,
                    child: Container(
                      height: 14,
                      width: 14,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: const Color(0xFF2ECC71),
                        border: Border.all(
                          color: Colors.white,
                          width: 2.5,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(width: 14),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _getGreeting(),
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w500,
                      color: Colors.grey.shade600,
                    ),
                  ),
                  const SizedBox(height: 2),
                  const Text(
                    "Alex Morgan",
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Color(0xFF0F172A),
                    ),
                  ),
                ],
              ),
            ],
          ),
          
          Row(
            children: [
              Container(
                height: 48,
                width: 48,
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  shape: BoxShape.circle,
                ),
                child: IconButton(
                  icon: const Icon(
                    Icons.picture_as_pdf,
                    color: Color(0xFFEF4444),
                    size: 22,
                  ),
                  onPressed: _filteredRecords.isNotEmpty ? _exportToPDF : null,
                  padding: EdgeInsets.zero,
                ),
              ),
              const SizedBox(width: 8),
              Container(
                height: 48,
                width: 48,
                decoration: BoxDecoration(
                  color: const Color(0xFFF1F5F9),
                  shape: BoxShape.circle,
                ),
                child: IconButton(
                  icon: const Icon(
                    Icons.share,
                    color: Color(0xFF135BEC),
                    size: 22,
                  ),
                  onPressed: _filteredRecords.isNotEmpty ? _exportToPDF : null,
                  padding: EdgeInsets.zero,
                ),
              ),
              const SizedBox(width: 8),
              Stack(
                children: [
                  Container(
                    height: 48,
                    width: 48,
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
                    top: 10,
                    right: 10,
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
        ],
      ),
    );
  }

  Widget _buildFilterSection() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: InkWell(
              onTap: _showMonthPicker,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.grey.shade300),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.calendar_today, size: 16, color: Color(0xFF135BEC)),
                    const SizedBox(width: 8),
                    Text(
                      DateFormat('MMMM yyyy').format(_selectedMonth),
                      style: const TextStyle(
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              decoration: BoxDecoration(
                border: Border.all(color: Colors.grey.shade300),
                borderRadius: BorderRadius.circular(12),
              ),
              child: DropdownButtonHideUnderline(
                child: DropdownButton<String>(
                  value: _selectedFilter,
                  isExpanded: true,
                  icon: const Icon(Icons.arrow_drop_down),
                  items: _filters.map((String filter) {
                    return DropdownMenuItem<String>(
                      value: filter,
                      child: Text(filter),
                    );
                  }).toList(),
                  onChanged: (String? newValue) {
                    if (newValue != null) {
                      setState(() {
                        _selectedFilter = newValue;
                        _filterRecords();
                      });
                    }
                  },
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryCards() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          _buildStatItem(
            icon: Icons.today,
            value: '${_filteredRecords.length}',
            label: "Total Days",
            color: const Color(0xFF135BEC),
          ),
          Container(
            height: 30,
            width: 1,
            color: Colors.grey.shade200,
          ),
          _buildStatItem(
            icon: Icons.access_time,
            value: '${_filteredRecords.fold(0.0, (sum, record) => sum + record.workHours).toStringAsFixed(1)}h',
            label: "Work Hours",
            color: const Color(0xFFF59E0B),
          ),
          Container(
            height: 30,
            width: 1,
            color: Colors.grey.shade200,
          ),
          _buildStatItem(
            icon: Icons.timelapse,
            value: '${_filteredRecords.fold(0.0, (sum, record) => sum + record.overtimeHours).toStringAsFixed(1)}h',
            label: "Overtime",
            color: const Color(0xFF8B5CF6),
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem({
    required IconData icon,
    required String value,
    required String label,
    required Color color,
  }) {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            shape: BoxShape.circle,
          ),
          child: Icon(icon, color: color, size: 18),
        ),
        const SizedBox(height: 6),
        Text(
          value,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: Color(0xFF0F172A),
          ),
        ),
        Text(
          label,
          style: TextStyle(
            fontSize: 9,
            color: Colors.grey.shade600,
          ),
        ),
      ],
    );
  }

  Widget _buildAttendanceCard(AttendanceRecord record) {
    Color statusColor;
    switch (record.status) {
      case 'On Time':
        statusColor = AppTheme.successColor;
        break;
      case 'Late':
        statusColor = AppTheme.warningColor;
        break;
      case 'Absent':
        statusColor = AppTheme.errorColor;
        break;
      case 'Overtime':
        statusColor = const Color(0xFF8B5CF6);
        break;
      default:
        statusColor = Colors.grey;
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                DateFormat('EEEE, dd MMM yyyy').format(record.date),
                style: const TextStyle(
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                  color: Color(0xFF0F172A),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: statusColor.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  record.status,
                  style: TextStyle(
                    fontSize: 10,
                    color: statusColor,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildTimeInfo('Clock In', record.clockIn, Icons.login, AppTheme.successColor),
              _buildTimeInfo('Clock Out', record.clockOut, Icons.logout, AppTheme.errorColor),
              _buildTimeInfo('Hours', '${record.workHours}h', Icons.timer, AppTheme.primaryColor),
              _buildTimeInfo('OT', '${record.overtimeHours}h', Icons.timelapse, const Color(0xFF8B5CF6)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildTimeInfo(String label, String value, IconData icon, Color color) {
    return Column(
      children: [
        Icon(icon, color: color, size: 14),
        const SizedBox(height: 4),
        Text(
          value,
          style: const TextStyle(
            fontWeight: FontWeight.bold,
            fontSize: 12,
            color: Color(0xFF0F172A),
          ),
        ),
        Text(
          label,
          style: TextStyle(
            fontSize: 8,
            color: Colors.grey.shade600,
          ),
        ),
      ],
    );
  }

  Widget _buildCalendarView() {
    return Container(
      padding: const EdgeInsets.all(8),
      child: CalendarView(
        records: _filteredRecords,
        onDateSelected: (DateTime date) {
          final record = _filteredRecords.firstWhere(
            (r) => r.date.year == date.year && 
                   r.date.month == date.month && 
                   r.date.day == date.day,
            orElse: () => AttendanceRecord(
              id: '',
              date: date,
              clockIn: '--:--',
              clockOut: '--:--',
              status: 'No Data',
              workHours: 0,
              overtimeHours: 0,
            ),
          );
          _showDetailDialog(record);
        },
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.history_toggle_off,
            size: 80,
            color: Colors.grey.shade400,
          ),
          const SizedBox(height: 16),
          Text(
            'No attendance records found',
            style: TextStyle(
              fontSize: 16,
              color: Colors.grey.shade600,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Try selecting a different month or filter',
            style: TextStyle(
              fontSize: 14,
              color: Colors.grey.shade500,
            ),
          ),
        ],
      ),
    );
  }

  void _showDetailDialog(AttendanceRecord record) {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          title: Text(
            DateFormat('dd MMM yyyy').format(record.date),
            style: const TextStyle(fontSize: 16),
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.login, color: AppTheme.successColor, size: 20),
                title: const Text('Clock In', style: TextStyle(fontSize: 13)),
                trailing: Text(record.clockIn, style: const TextStyle(fontSize: 13)),
                dense: true,
              ),
              ListTile(
                leading: const Icon(Icons.logout, color: AppTheme.errorColor, size: 20),
                title: const Text('Clock Out', style: TextStyle(fontSize: 13)),
                trailing: Text(record.clockOut, style: const TextStyle(fontSize: 13)),
                dense: true,
              ),
              ListTile(
                leading: const Icon(Icons.timer, color: AppTheme.primaryColor, size: 20),
                title: const Text('Work Hours', style: TextStyle(fontSize: 13)),
                trailing: Text('${record.workHours}h', style: const TextStyle(fontSize: 13)),
                dense: true,
              ),
              ListTile(
                leading: const Icon(Icons.timelapse, color: Color(0xFF8B5CF6), size: 20),
                title: const Text('Overtime', style: TextStyle(fontSize: 13)),
                trailing: Text('${record.overtimeHours}h', style: const TextStyle(fontSize: 13)),
                dense: true,
              ),
              const Divider(),
              ListTile(
                title: const Text('Status', style: TextStyle(fontSize: 13)),
                trailing: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: record.status == 'On Time' 
                        ? AppTheme.successColor.withOpacity(0.1)
                        : record.status == 'Late'
                            ? AppTheme.warningColor.withOpacity(0.1)
                            : record.status == 'Absent'
                                ? AppTheme.errorColor.withOpacity(0.1)
                                : Colors.grey.shade100,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    record.status,
                    style: TextStyle(
                      fontSize: 11,
                      color: record.status == 'On Time'
                          ? AppTheme.successColor
                          : record.status == 'Late'
                              ? AppTheme.warningColor
                              : record.status == 'Absent'
                                  ? AppTheme.errorColor
                                  : Colors.grey.shade700,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Close'),
            ),
          ],
        );
      },
    );
  }
}

// Simple Calendar View Widget
class CalendarView extends StatelessWidget {
  final List<AttendanceRecord> records;
  final Function(DateTime) onDateSelected;

  const CalendarView({
    super.key,
    required this.records,
    required this.onDateSelected,
  });

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final firstDay = DateTime(now.year, now.month, 1);
    
    return Column(
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
              .map((day) => Expanded(
                    child: Text(
                      day,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: Colors.grey.shade700,
                        fontSize: 11,
                      ),
                    ),
                  ))
              .toList(),
        ),
        const SizedBox(height: 8),
        
        GridView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 7,
            childAspectRatio: 1,
          ),
          itemCount: 35,
          itemBuilder: (context, index) {
            final date = firstDay.add(Duration(days: index - firstDay.weekday + 1));
            final isCurrentMonth = date.month == now.month;
            
            final record = records.firstWhere(
              (r) => r.date.year == date.year && 
                     r.date.month == date.month && 
                     r.date.day == date.day,
              orElse: () => AttendanceRecord(
                id: '',
                date: date,
                clockIn: '',
                clockOut: '',
                status: '',
                workHours: 0,
                overtimeHours: 0,
              ),
            );
            
            Color? backgroundColor;
            if (record.status.isNotEmpty) {
              if (record.status == 'On Time') backgroundColor = AppTheme.successColor.withOpacity(0.2);
              else if (record.status == 'Late') backgroundColor = AppTheme.warningColor.withOpacity(0.2);
              else if (record.status == 'Absent') backgroundColor = AppTheme.errorColor.withOpacity(0.2);
              else if (record.status == 'Overtime') backgroundColor = const Color(0xFF8B5CF6).withOpacity(0.2);
            }
            
            return GestureDetector(
              onTap: () => onDateSelected(date),
              child: Container(
                margin: const EdgeInsets.all(2),
                decoration: BoxDecoration(
                  color: backgroundColor,
                  shape: BoxShape.circle,
                ),
                child: Center(
                  child: Text(
                    date.day.toString(),
                    style: TextStyle(
                      color: isCurrentMonth ? const Color(0xFF0F172A) : Colors.grey.shade400,
                      fontWeight: record.status.isNotEmpty ? FontWeight.bold : FontWeight.normal,
                      fontSize: 12,
                    ),
                  ),
                ),
              ),
            );
          },
        ),
      ],
    );
  }
}