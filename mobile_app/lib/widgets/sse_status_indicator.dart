import 'package:flutter/material.dart';
import 'package:mobile_app/services/sse_service.dart';

class SSEStatusIndicator extends StatefulWidget {
  final bool showText;
  
  const SSEStatusIndicator({
    super.key,
    this.showText = true,
  });

  @override
  State<SSEStatusIndicator> createState() => _SSEStatusIndicatorState();
}

class _SSEStatusIndicatorState extends State<SSEStatusIndicator>
    with SingleTickerProviderStateMixin {
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      duration: const Duration(seconds: 2),
      vsync: this,
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 0.4, end: 1.0).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<SSEConnectionState>(
      valueListenable: SSEService().connectionState,
      builder: (context, state, _) {
        Color dotColor;
        String statusText;
        Widget statusIcon;

        switch (state) {
          case SSEConnectionState.connected:
            dotColor = const Color(0xFF10B981); // Emerald Green
            statusText = 'Live Sync';
            statusIcon = FadeTransition(
              opacity: _pulseAnimation,
              child: Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(
                  color: dotColor,
                  shape: BoxShape.circle,
                  boxShadow: [
                    BoxShadow(
                      color: dotColor.withOpacity(0.6),
                      blurRadius: 4,
                      spreadRadius: 1.5,
                    ),
                  ],
                ),
              ),
            );
            break;

          case SSEConnectionState.connecting:
            dotColor = const Color(0xFFF59E0B); // Amber
            statusText = 'Connecting';
            statusIcon = const SizedBox(
              width: 8,
              height: 8,
              child: CircularProgressIndicator(
                strokeWidth: 1.5,
                valueColor: AlwaysStoppedAnimation<Color>(Color(0xFFF59E0B)),
              ),
            );
            break;

          case SSEConnectionState.disconnected:
          case SSEConnectionState.error:
            dotColor = const Color(0xFFEF4444); // Red
            statusText = 'Offline';
            statusIcon = Container(
              width: 8,
              height: 8,
              decoration: BoxDecoration(
                color: dotColor,
                shape: BoxShape.circle,
              ),
            );
            break;
        }

        return InkWell(
          onTap: state == SSEConnectionState.error || state == SSEConnectionState.disconnected
              ? () {
                  debugPrint('[SSE Indicator] User manual reconnect trigger');
                  SSEService().connect();
                }
              : null,
          borderRadius: BorderRadius.circular(20),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              border: Border.all(
                color: const Color(0xFFE2E8F0), // Slate-200
                width: 1.2,
              ),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.02),
                  blurRadius: 4,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                statusIcon,
                if (widget.showText) ...[
                  const SizedBox(width: 6),
                  Text(
                    statusText,
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w800,
                      color: state == SSEConnectionState.connected
                          ? const Color(0xFF1E293B) // Slate-800
                          : dotColor,
                      letterSpacing: 0.2,
                    ),
                  ),
                ],
              ],
            ),
          ),
        );
      },
    );
  }
}
