import 'dart:async';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:mobile_app/services/api_service.dart';
import 'package:mobile_app/services/sse_service.dart';
import 'package:mobile_app/widgets/in_app_notification.dart';

class NotificationsPage extends StatefulWidget {
  const NotificationsPage({super.key});

  @override
  State<NotificationsPage> createState() => _NotificationsPageState();
}

class _NotificationsPageState extends State<NotificationsPage> {
  late Future<List<Map<String, dynamic>>> _future;
  StreamSubscription? _sseSubscription;
  Timer? _autoReloadTimer;

  @override
  void initState() {
    super.initState();
    SSEService().hasNewNotification.value = false;
    SSEService().hasNewLeaveDecisionNotification.value = false;
    _future = _load();
    SSEService().refreshUnreadNotificationCount();
    _setupRealtime();
  }

  @override
  void dispose() {
    _sseSubscription?.cancel();
    _autoReloadTimer?.cancel();
    super.dispose();
  }

  void _setupRealtime() {
    _sseSubscription = SSEService().events.listen((event) {
      if (!mounted || event.type == 'ping') return;
      if (event.type == 'notification_created' ||
          event.type == 'leave_updated') {
        _scheduleAutoReload();
      }
    });
  }

  void _scheduleAutoReload() {
    _autoReloadTimer?.cancel();
    _autoReloadTimer = Timer(const Duration(milliseconds: 200), () {
      if (!mounted) return;
      setState(() {
        _future = _load();
      });
    });
  }

  Future<List<Map<String, dynamic>>> _load() {
    return ApiService.getNotifications(limit: 50);
  }

  Future<void> _refresh() async {
    setState(() {
      _future = _load();
    });
    await SSEService().refreshUnreadNotificationCount();
    await _future;
  }

  // Method untuk format date yang sudah diperbaiki sesuai real time
  String _formatDate(dynamic value) {
    DateTime? dt;
    
    if (value is String) {
      dt = DateTime.tryParse(value);
    } else if (value is DateTime) {
      dt = value;
    }
    
    if (dt == null) return '';
    
    // Konversi ke waktu lokal
    final localDt = dt.toLocal();
    
    // Format dengan tanggal dan waktu lengkap
    return DateFormat('dd MMM yyyy • HH:mm:ss', 'id_ID').format(localDt);
  }

  // Method untuk menampilkan waktu relatif (lebih user-friendly)
  String _getRelativeTime(dynamic value) {
    DateTime? dt;
    
    if (value is String) {
      dt = DateTime.tryParse(value);
    } else if (value is DateTime) {
      dt = value;
    }
    
    if (dt == null) return '';
    
    final localDt = dt.toLocal();
    final now = DateTime.now();
    final difference = now.difference(localDt);
    
    if (difference.inDays > 7) {
      return DateFormat('dd MMM yyyy', 'id_ID').format(localDt);
    } else if (difference.inDays > 0) {
      if (difference.inDays == 1) {
        return 'Kemarin';
      }
      return '${difference.inDays} hari yang lalu';
    } else if (difference.inHours > 0) {
      if (difference.inHours == 1) {
        return '1 jam yang lalu';
      }
      return '${difference.inHours} jam yang lalu';
    } else if (difference.inMinutes > 0) {
      if (difference.inMinutes == 1) {
        return '1 menit yang lalu';
      }
      return '${difference.inMinutes} menit yang lalu';
    } else {
      return 'Baru saja';
    }
  }

  Color _typeColor(String type) {
    final value = type.toLowerCase();
    if (value.contains('leave')) return const Color(0xFFF59E0B);
    if (value.contains('overtime')) return const Color(0xFF135BEC);
    if (value.contains('assignment')) return const Color(0xFF06B6D4);
    if (value.contains('attendance')) return const Color(0xFF10B981);
    return const Color(0xFF64748B);
  }

  IconData _typeIcon(String type) {
    final value = type.toLowerCase();
    if (value.contains('leave')) return Icons.event_note_rounded;
    if (value.contains('overtime')) return Icons.schedule_rounded;
    if (value.contains('assignment')) return Icons.assignment_rounded;
    if (value.contains('attendance')) return Icons.fingerprint_rounded;
    return Icons.notifications_rounded;
  }

  Future<void> _markAllAsRead() async {
    await ApiService.markAllNotificationsAsRead();
    await _refresh();
  }

  Future<void> _markAsRead(String id) async {
    await ApiService.markNotificationAsRead(id);
    await SSEService().refreshUnreadNotificationCount();
    if (mounted)
      setState(() {
        _future = _load();
      });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifikasi'),
        centerTitle: false,
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF0F172A),
        elevation: 0,
        actions: [
          TextButton.icon(
            onPressed: _markAllAsRead,
            icon: const Icon(Icons.done_all_rounded, size: 18),
            label: const Text('Tandai dibaca'),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: FutureBuilder<List<Map<String, dynamic>>>(
          future: _future,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Center(child: CircularProgressIndicator());
            }

            if (snapshot.hasError) {
              return Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(
                      Icons.error_outline,
                      size: 72,
                      color: Color(0xFFEF4444),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'Gagal memuat notifikasi',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: Colors.grey.shade700,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      snapshot.error.toString(),
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.grey.shade500,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: _refresh,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF135BEC),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: const Text('Coba Lagi'),
                    ),
                  ],
                ),
              );
            }

            final items = snapshot.data ?? [];
            if (items.isEmpty) {
              return ListView(
                physics: const AlwaysScrollableScrollPhysics(),
                children: const [
                  SizedBox(height: 120),
                  Icon(
                    Icons.notifications_off_rounded,
                    size: 72,
                    color: Color(0xFF94A3B8),
                  ),
                  SizedBox(height: 12),
                  Center(
                    child: Text(
                      'Belum ada notifikasi',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF334155),
                      ),
                    ),
                  ),
                  SizedBox(height: 8),
                  Center(
                    child: Text(
                      'Notifikasi akan muncul di sini',
                      style: TextStyle(
                        fontSize: 13,
                        color: Color(0xFF64748B),
                      ),
                    ),
                  ),
                ],
              );
            }

            return ListView.separated(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: 10),
              itemBuilder: (context, index) {
                final item = items[index];
                final id = (item['id'] ?? '').toString();
                final title = (item['title'] ?? 'Notifikasi').toString();
                final message = (item['message'] ?? '').toString();
                final type = (item['type'] ?? '').toString();
                final unread = item['is_read'] != true;
                final color = _typeColor(type);
                
                // Menggunakan relative time yang lebih user-friendly
                final relativeTime = _getRelativeTime(item['created_at']);
                // Full date untuk tooltip (opsional)
                final fullDate = _formatDate(item['created_at']);

                return InkWell(
                  borderRadius: BorderRadius.circular(18),
                  onTap: () async {
                    if (id.isNotEmpty && unread) {
                      await _markAsRead(id);
                    }
                    if (!context.mounted) return;
                    InAppNotification.show(
                      title: title,
                      message: message,
                      type: type.toLowerCase().contains('overtime')
                          ? InAppNotificationType.overtime
                          : type.toLowerCase().contains('assignment')
                          ? InAppNotificationType.assignment
                          : type.toLowerCase().contains('attendance')
                          ? InAppNotificationType.attendance
                          : InAppNotificationType.leave,
                    );
                  },
                  child: Tooltip(
                    message: fullDate.isNotEmpty ? fullDate : relativeTime,
                    child: Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: unread ? const Color(0xFFF8FAFC) : Colors.white,
                        borderRadius: BorderRadius.circular(18),
                        border: Border.all(
                          color: unread
                              ? color.withOpacity(0.16)
                              : const Color(0xFFE2E8F0),
                          width: unread ? 1.5 : 1,
                        ),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.04),
                            blurRadius: 12,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: color.withOpacity(0.12),
                              shape: BoxShape.circle,
                            ),
                            child: Icon(
                              _typeIcon(type), 
                              color: color, 
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
                                    Expanded(
                                      child: Text(
                                        title,
                                        style: TextStyle(
                                          fontSize: 15,
                                          fontWeight: unread
                                              ? FontWeight.w800
                                              : FontWeight.w700,
                                          color: const Color(0xFF0F172A),
                                          height: 1.2,
                                        ),
                                        maxLines: 2,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ),
                                    if (unread) ...[
                                      const SizedBox(width: 8),
                                      Container(
                                        width: 10,
                                        height: 10,
                                        decoration: const BoxDecoration(
                                          color: Color(0xFFEF4444),
                                          shape: BoxShape.circle,
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  message,
                                  style: const TextStyle(
                                    fontSize: 13,
                                    color: Color(0xFF334155),
                                    height: 1.4,
                                  ),
                                  maxLines: 3,
                                  overflow: TextOverflow.ellipsis,
                                ),
                                if (relativeTime.isNotEmpty) ...[
                                  const SizedBox(height: 10),
                                  Row(
                                    children: [
                                      Icon(
                                        Icons.access_time_rounded,
                                        size: 12,
                                        color: const Color(0xFF64748B),
                                      ),
                                      const SizedBox(width: 4),
                                      Text(
                                        relativeTime,
                                        style: const TextStyle(
                                          fontSize: 11,
                                          color: Color(0xFF64748B),
                                          fontWeight: FontWeight.w500,
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }
}