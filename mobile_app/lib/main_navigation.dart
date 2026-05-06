// lib/pages/main_navigation.dart
import 'package:flutter/material.dart';
import 'package:mobile_app/pages/dashboard_page.dart';
import 'package:mobile_app/pages/history_page.dart';
import 'package:mobile_app/pages/overtime_page.dart';
import 'package:mobile_app/pages/request_page.dart';
import 'package:mobile_app/pages/profile_page.dart';
import 'package:mobile_app/services/sse_service.dart';

class MainNavigationPage extends StatefulWidget {
  const MainNavigationPage({super.key});

  @override
  State<MainNavigationPage> createState() => _MainNavigationPageState();
}


class _MainNavigationPageState extends State<MainNavigationPage>
    with TickerProviderStateMixin {
  int _selectedIndex = 0;

  @override
  void initState() {
    super.initState();
    // Connect to real-time events when entering main app
    SSEService().connect();
  }

  @override
  void dispose() {
    // Disconnect when exiting main app (e.g. logout)
    SSEService().disconnect();
    super.dispose();
  }

  final List<Widget> _pages = [
    EmployeeDashboardPage(),
    HistoryPage(),
    RequestPage(),
    OvertimePage(),
    ProfilePage(),
  ];

  // Nav items config
  static const _navItems = [
    _NavItem(icon: Icons.home_rounded,          activeIcon: Icons.home_rounded,          label: 'Beranda'),
    _NavItem(icon: Icons.history_rounded,        activeIcon: Icons.history_rounded,        label: 'Riwayat'),
    _NavItem(icon: Icons.assignment_rounded,     activeIcon: Icons.assignment_rounded,     label: 'Pengajuan'),
    _NavItem(icon: Icons.schedule_rounded,       activeIcon: Icons.schedule_rounded,       label: 'Lembur'),
    _NavItem(icon: Icons.person_outline_rounded, activeIcon: Icons.person_rounded,         label: 'Profil'),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(index: _selectedIndex, children: _pages),
      bottomNavigationBar: _buildBottomNav(),
    );
  }

  Widget _buildBottomNav() {
    return Container(
      height: 76,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(28)),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF135BEC).withOpacity(0.08),
            blurRadius: 24,
            offset: const Offset(0, -4),
          ),
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, -1),
          ),
        ],
      ),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: List.generate(_navItems.length, (i) => _buildNavItem(i)),
          ),
        ),
      ),
    );
  }

  Widget _buildNavItem(int index) {
    final item = _navItems[index];
    final isSelected = _selectedIndex == index;

    return Expanded(
      child: GestureDetector(
        onTap: () => setState(() => _selectedIndex = index),
        behavior: HitTestBehavior.opaque,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 250),
          curve: Curves.easeInOut,
          padding: const EdgeInsets.symmetric(vertical: 6),
          decoration: BoxDecoration(
            color: isSelected
                ? const Color(0xFF135BEC).withOpacity(0.09)
                : Colors.transparent,
            borderRadius: BorderRadius.circular(16),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              AnimatedSwitcher(
                duration: const Duration(milliseconds: 200),
                child: Icon(
                  isSelected ? item.activeIcon : item.icon,
                  key: ValueKey(isSelected),
                  color: isSelected
                      ? const Color(0xFF135BEC)
                      : const Color(0xFF94A3B8),
                  size: isSelected ? 24 : 22,
                ),
              ),
              const SizedBox(height: 3),
              AnimatedDefaultTextStyle(
                duration: const Duration(milliseconds: 200),
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                  color: isSelected
                      ? const Color(0xFF135BEC)
                      : const Color(0xFF94A3B8),
                ),
                child: Text(item.label),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _NavItem {
  final IconData icon;
  final IconData activeIcon;
  final String label;
  const _NavItem({required this.icon, required this.activeIcon, required this.label});
}