import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'api_service.dart';

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

  http.Client? _client;
  final _eventController = StreamController<SSEEvent>.broadcast();
  bool _isConnected = false;
  
  // ValueNotifier untuk memicu refresh di UI secara global
  final refreshCounter = ValueNotifier<int>(0);
  
  // Flag untuk indikator data baru (badge)
  final hasNewOvertime = ValueNotifier<bool>(false);
  final hasNewAssignment = ValueNotifier<bool>(false);
  final hasNewLeaveRequest = ValueNotifier<bool>(false);

  Stream<SSEEvent> get events => _eventController.stream;

  Future<void> connect() async {
    if (_isConnected) {
      print('[SSE] Already connected');
      return;
    }

    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString('access_token');
    
    if (token == null) {
      print('[SSE] Cannot connect: No token found');
      return;
    }

    try {
      _client = http.Client();
      final baseUri = Uri.parse('${ApiService.baseUrl}/realtime/connect');
      final uri = baseUri.replace(queryParameters: {'token': token});

      print('[SSE] Connecting to: ${uri.toString()}');
      final request = http.Request('GET', uri);

      final response = await _client!.send(request);

      if (response.statusCode == 200) {
        _isConnected = true;
        print('[SSE] Connected to realtime stream successfully');

        response.stream
            .transform(utf8.decoder)
            .transform(const LineSplitter())
            .listen(
          (line) {
            if (line.startsWith('data: ')) {
              final dataStr = line.substring(6).trim();
              if (dataStr.isEmpty) return;

              try {
                final jsonMap = json.decode(dataStr);
                final event = SSEEvent.fromJson(jsonMap);
                _eventController.add(event);
                
                // Trigger global refresh
                refreshCounter.value++;
                
                // Set badge flags
                if (event.type == 'overtime_updated') {
                  hasNewOvertime.value = true;
                } else if (event.type == 'assignment_updated') {
                  hasNewAssignment.value = true;
                } else if (event.type == 'leave_updated') {
                  hasNewLeaveRequest.value = true;
                }
                
                print('[SSE] Received event: ${event.type}');
              } catch (e) {
                print('[SSE] JSON parse error: $e | Data: $dataStr');
              }
            }
          },
          onDone: () {
            print('[SSE] Connection closed by server');
            _isConnected = false;
            _reconnect();
          },
          onError: (e) {
            print('[SSE] Error on stream: $e');
            _isConnected = false;
            _reconnect();
          },
        );
      } else {
        print('[SSE] Failed to connect: Status ${response.statusCode}');
        _isConnected = false;
        _reconnect();
      }
    } catch (e) {
      print('[SSE] Exception during connect: $e');
      _isConnected = false;
      _reconnect();
    }
  }

  void _reconnect() {
    if (!_isConnected) {
      print('[SSE] Reconnect scheduled in 5 seconds...');
      Future.delayed(const Duration(seconds: 5), () {
        if (!_isConnected) connect();
      });
    }
  }

  void disconnect() {
    _isConnected = false;
    _client?.close();
    _client = null;
    hasNewOvertime.value = false;
    hasNewAssignment.value = false;
    hasNewLeaveRequest.value = false;
    print('[SSE] Disconnected and flags cleared');
  }
}
