import 'dart:async';
import 'package:flutter/material.dart';
import 'package:mobile_app/main.dart';

enum InAppNotificationType {
  attendance,
  leave,
  overtime,
  assignment,
}

class InAppNotification extends StatefulWidget {
  final String title;
  final String message;
  final InAppNotificationType type;
  final VoidCallback onDismiss;
  final VoidCallback? onTap;

  const InAppNotification({
    super.key,
    required this.title,
    required this.message,
    required this.type,
    required this.onDismiss,
    this.onTap,
  });

  /// Displays the custom animated sliding card at the top of the viewport globally.
  static void show({
    required String title,
    required String message,
    required InAppNotificationType type,
    VoidCallback? onTap,
  }) {
    final context = MyApp.navigatorKey.currentContext;
    if (context == null) {
      debugPrint('[Notification] Context is unavailable, skipping overlay');
      return;
    }

    final overlayState = Overlay.of(context);
    late OverlayEntry overlayEntry;

    overlayEntry = OverlayEntry(
      builder: (context) => InAppNotification(
        title: title,
        message: message,
        type: type,
        onTap: onTap,
        onDismiss: () {
          try {
            overlayEntry.remove();
          } catch (_) {}
        },
      ),
    );

    overlayState.insert(overlayEntry);
  }

  @override
  State<InAppNotification> createState() => _InAppNotificationState();
}

class _InAppNotificationState extends State<InAppNotification>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<Offset> _offsetAnimation;
  late Animation<double> _opacityAnimation;
  Timer? _dismissTimer;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 450),
      vsync: this,
    );

    _offsetAnimation = Tween<Offset>(
      begin: const Offset(0.0, -1.3),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _controller,
      curve: Curves.easeOutBack,
    ));

    _opacityAnimation = Tween<double>(
      begin: 0.0,
      end: 1.0,
    ).animate(CurvedAnimation(
      parent: _controller,
      curve: Curves.easeIn,
    ));

    _controller.forward();

    // Auto dismiss after 4.5 seconds
    _dismissTimer = Timer(const Duration(seconds: 4, milliseconds: 500), () {
      _dismiss();
    });
  }

  void _dismiss() {
    if (!mounted) return;
    _controller.reverse().then((_) {
      widget.onDismiss();
    });
  }

  @override
  void dispose() {
    _dismissTimer?.cancel();
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    Color primaryColor;
    Color bgColor;
    Color borderColor;
    IconData icon;

    switch (widget.type) {
      case InAppNotificationType.attendance:
        primaryColor = const Color(0xFF10B981); // Emerald Green
        bgColor = const Color(0xFFECFDF5);
        borderColor = const Color(0xFFA7F3D0);
        icon = Icons.check_circle_rounded;
        break;
      case InAppNotificationType.leave:
        primaryColor = const Color(0xFFF59E0B); // Golden Amber
        bgColor = const Color(0xFFFEF3C7);
        borderColor = const Color(0xFFFDE68A);
        icon = Icons.coffee_rounded;
        break;
      case InAppNotificationType.overtime:
        primaryColor = const Color(0xFF135BEC); // Indigo Blue
        bgColor = const Color(0xFFEEF2FF);
        borderColor = const Color(0xFFC7D2FE);
        icon = Icons.schedule_rounded;
        break;
      case InAppNotificationType.assignment:
        primaryColor = const Color(0xFF06B6D4); // Cyan/Teal
        bgColor = const Color(0xFFECFEFF);
        borderColor = const Color(0xFFCFFAFE);
        icon = Icons.assignment_rounded;
        break;
    }

    return SafeArea(
      child: Align(
        alignment: Alignment.topCenter,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 12.0),
          child: SlideTransition(
            position: _offsetAnimation,
            child: FadeTransition(
              opacity: _opacityAnimation,
              child: Dismissible(
                key: UniqueKey(),
                direction: DismissDirection.up,
                onDismissed: (_) => widget.onDismiss(),
                child: Material(
                  color: Colors.transparent,
                  child: GestureDetector(
                    onTap: () {
                      _dismiss();
                      widget.onTap?.call();
                    },
                    child: Container(
                      width: double.infinity,
                    constraints: const BoxConstraints(maxWidth: 480),
                    decoration: BoxDecoration(
                      color: bgColor,
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: borderColor, width: 1.5),
                      boxShadow: [
                        BoxShadow(
                          color: primaryColor.withOpacity(0.08),
                          blurRadius: 16,
                          offset: const Offset(0, 8),
                        ),
                        BoxShadow(
                          color: Colors.black.withOpacity(0.04),
                          blurRadius: 6,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: Padding(
                      padding: const EdgeInsets.all(14.0),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // Left Icon with White Circle Base and shadow
                          Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              shape: BoxShape.circle,
                              boxShadow: [
                                BoxShadow(
                                  color: primaryColor.withOpacity(0.12),
                                  blurRadius: 6,
                                  offset: const Offset(0, 2),
                                ),
                              ],
                            ),
                            child: Icon(
                              icon,
                              color: primaryColor,
                              size: 22,
                            ),
                          ),
                          const SizedBox(width: 12),
                          // Content Text
                          Expanded(
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  widget.title,
                                  style: TextStyle(
                                    color: primaryColor,
                                    fontSize: 13,
                                    fontWeight: FontWeight.w800,
                                    letterSpacing: 0.2,
                                  ),
                                ),
                                const SizedBox(height: 3),
                                Text(
                                  widget.message,
                                  style: const TextStyle(
                                    color: Color(0xFF334155), // Slate-700
                                    fontSize: 12.5,
                                    fontWeight: FontWeight.w600,
                                    height: 1.35,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(width: 8),
                          // Mini Close Button
                          IconButton(
                            onPressed: _dismiss,
                            icon: const Icon(
                              Icons.close_rounded,
                              color: Color(0xFF64748B), // Slate-500
                              size: 16,
                            ),
                            padding: EdgeInsets.zero,
                            constraints: const BoxConstraints(),
                            splashRadius: 16,
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
