import 'package:flutter/material.dart';

class PayrollCalculatorPage extends StatelessWidget {
  const PayrollCalculatorPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Kalkulator Payroll'),
      ),
      body: const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Text(
            'Fitur kalkulator payroll sedang disiapkan.',
            textAlign: TextAlign.center,
          ),
        ),
      ),
    );
  }
}
