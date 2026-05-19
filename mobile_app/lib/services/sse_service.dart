import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../widgets/in_app_notification.dart';
import 'api_service.dart';

enum SSEConnectionState {
  disconnected,
  connecting,
  connected,
  error,
}

class SSEEvent {
  final String type;
  final Map<String, dynamic>? payload;

  SSEEvent({required this.type, this.payload});

  factory SSEEvent.fromJson(Map<String, dynamic> json) {
    return SSEEvent(
      type: json['type'] as String,
      payload: json['payload'] as Map<String, dynamic>?,
    );
  }
}

class SSEService {
  static final SSEService _instance = SSEService._internal();
  factory SSEService() => _instance;
  SSEService._internal();

  HttpClient? _ioClient;
  StreamSubscription? _streamSubscription;
  final _eventController = StreamController<SSEEvent>.broadcast();
  
  // ValueNotifier for real-time tracking of connection state
  final connectionState = ValueNotifier<SSEConnectionState>(SSEConnectionState.disconnected);
  bool _isManualDisconnect = false;
  int _reconnectAttempts = 0;
  Timer? _reconnectTimer;
  
  // ValueNotifier to trigger active widgets/tables to refetch data from API
  final refreshCounter = ValueNotifier<int>(0);
  
  // Flag for new notification dots (badges) on bottom nav bar
  final hasNewOvertime = ValueNotifier<bool>(false);
  final hasNewAssignment = ValueNotifier<bool>(false);
  final hasNewLeaveRequest = ValueNotifier<bool>(false);

  Stream<SSEEvent> get events => _eventController.stream;

  bool get isConnected => connectionState.value == SSEConnectionState.connected;

  Future<void> connect() async {
    // Prevent multiple parallel connection instances
    if (connectionState.value == SSEConnectionState.connected || 
        connectionState.value == SSEConnectionState.connecting) {
      print('[SSE] Already connecting or connected. Current state: ${connectionState.value}');
      return;
    }

    _isManualDisconnect = false;
    connectionState.value = SSEConnectionState.connecting;

    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString('access_token');
    
    if (token == null) {
      print('[SSE] Aborting connect: No active token found in SharedPreferences');
      connectionState.value = SSEConnectionState.disconnected;
      return;
    }

    try {
      final baseUri = Uri.parse('${ApiService.baseUrl}/realtime/connect');
      final uri = baseUri.replace(queryParameters: {'token': token});

      print('[SSE] Connecting to Server-Sent Events stream: ${uri.toString()}');
      
      _ioClient = HttpClient();
      _ioClient!.connectionTimeout = const Duration(seconds: 15);
      
      final request = await _ioClient!.getUrl(uri);
      
      // Set explicit SSE request headers for maximum compatibility
      request.headers.set('Accept', 'text/event-stream');
      request.headers.set('Cache-Control', 'no-cache');
      request.headers.set('Connection', 'keep-alive');
      
      final response = await request.close();

      if (response.statusCode == 200) {
        connectionState.value = SSEConnectionState.connected;
        _reconnectAttempts = 0; // Reset attempts upon successful connection
        print('[SSE] Server-Sent Events bridge established successfully (dart:io)');

        _streamSubscription = response
            .transform(utf8.decoder)
            .transform(const LineSplitter())
            .listen(
          (line) {
            print('[SSE] Raw stream line received: "$line"');
            if (line.startsWith('data: ')) {
              final dataStr = line.substring(6).trim();
              if (dataStr.isEmpty) return;

              try {
                final jsonMap = json.decode(dataStr);
                final event = SSEEvent.fromJson(jsonMap);
                
                print('[SSE] Event received -> Type: "${event.type}", Payload: ${event.payload}');
                
                // Dispatch event to local app listeners
                _eventController.add(event);
                
                // Increment dynamic UI counter to force widget updates
                refreshCounter.value++;
                
                final payload = event.payload;
                final message = payload?['message'] as String?;

                // Inspect event type, flag badges, and launch premium notifications
                if (event.type == 'overtime_updated') {
                  hasNewOvertime.value = true;
                  if (message != null && message.isNotEmpty) {
                    InAppNotification.show(
                      title: 'Pemberitahuan Lembur',
                      message: message,
                      type: InAppNotificationType.overtime,
                    );
                  }
                } else if (event.type == 'assignment_updated') {
                  hasNewAssignment.value = true;
                  if (message != null && message.isNotEmpty) {
                    InAppNotification.show(
                      title: 'Penugasan Baru',
                      message: message,
                      type: InAppNotificationType.assignment,
                    );
                  }
                } else if (event.type == 'leave_updated') {
                  final action = payload?['action'] as String?;
                  if (action == 'status_updated' || action == 'update') {
                    hasNewLeaveRequest.value = true;
                    if (message != null && message.isNotEmpty) {
                      InAppNotification.show(
                        title: 'Pengajuan Cuti/Izin',
                        message: message,
                        type: InAppNotificationType.leave,
                      );
                    }
                  }
                } else if (event.type == 'attendance_updated') {
                  final status = payload?['status'] as String?;
                  if (status != 'connected') {
                    if (message != null && message.isNotEmpty) {
                      InAppNotification.show(
                        title: 'Info Kehadiran',
                        message: message,
                        type: InAppNotificationType.attendance,
                      );
                    }
                  }
                }
              } catch (e) {
                print('[SSE] JSON decode failure: $e | Line data: $dataStr');
              }
            }
          },
          onDone: () {
            print('[SSE] Connection closed from host server side (onDone)');
            _handleDisconnectOrError(SSEConnectionState.disconnected);
          },
          onError: (e) {
            print('[SSE] Stream listening error encountered: $e');
            _handleDisconnectOrError(SSEConnectionState.error);
          },
          cancelOnError: true,
        );
      } else {
        print('[SSE] Handshake rejected, status: ${response.statusCode}');
        _handleDisconnectOrError(SSEConnectionState.error);
      }
    } catch (e) {
      print('[SSE] Exception raised during socket connect: $e');
      _handleDisconnectOrError(SSEConnectionState.error);
    }
  }

  void _handleDisconnectOrError(SSEConnectionState targetState) {
    if (connectionState.value == SSEConnectionState.disconnected && _isManualDisconnect) {
      return;
    }
    
    connectionState.value = targetState;
    _streamSubscription?.cancel();
    _streamSubscription = null;
    _ioClient?.close(force: true);
    _ioClient = null;
    
    if (!_isManualDisconnect) {
      _scheduleReconnect();
    }
  }

  void _scheduleReconnect() {
    _reconnectTimer?.cancel();
    
    // Exponential Backoff algorithm: 2s, 4s, 8s, 16s, up to 32s limit
    final delaySeconds = (1 << (_reconnectAttempts + 1)).clamp(2, 32);
    _reconnectAttempts++;
    
    print('[SSE] Reconnection scheduler engaged. Retry #$_reconnectAttempts in $delaySeconds seconds');
    
    _reconnectTimer = Timer(Duration(seconds: delaySeconds), () {
      if (connectionState.value != SSEConnectionState.connected && !_isManualDisconnect) {
        connect();
      }
    });
  }

  void disconnect() {
    _isManualDisconnect = true;
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempts = 0;
    
    connectionState.value = SSEConnectionState.disconnected;
    _streamSubscription?.cancel();
    _streamSubscription = null;
    _ioClient?.close(force: true);
    _ioClient = null;
    
    hasNewOvertime.value = false;
    hasNewAssignment.value = false;
    hasNewLeaveRequest.value = false;
    print('[SSE] Manual disconnect completed. All notifier values cleaned.');
  }
}
