import 'package:flutter/material.dart';
import 'package:mobile_app/splash.dart';
import 'package:mobile_app/theme/app_theme.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Labersa Absensi',
      theme: AppTheme.lightTheme,
      home: const SplashScreen(),
    );
  }
}