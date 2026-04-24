import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:mobile_app/theme/app_theme.dart';
import 'package:mobile_app/login.dart';
import 'package:mobile_app/splash.dart';
import 'package:mobile_app/main_navigation.dart';
import 'package:mobile_app/pages/face_registration.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'HRIS Mobile',
      theme: AppTheme.lightTheme,
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('id', 'ID'),
        Locale('en', 'US'),
      ],
      locale: const Locale('id', 'ID'),
      initialRoute: '/',
      routes: {
        '/': (context) => const SplashScreen(),
        '/login': (context) => const EmployeeLoginPage(),
        '/home': (context) => const MainNavigationPage(),
        '/face-registration': (context) => FaceRegistrationPage(
              userId: ModalRoute.of(context)!.settings.arguments as String,
            ),
      },
    );
  }
}