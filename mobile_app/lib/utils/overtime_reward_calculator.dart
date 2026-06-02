import 'package:intl/intl.dart';

double calculateOvertimeBasis(int basicSalary) {
  return basicSalary / 173.0;
}

double calculateOvertimeFactor(double overtimeHours) {
  if (overtimeHours <= 0) {
    return 0;
  }
  if (overtimeHours <= 1) {
    return overtimeHours * 1.5;
  }
  return 1.5 + (overtimeHours - 1) * 2.0;
}

int calculateOvertimeMoneyReward(int basicSalary, double overtimeHours) {
  if (basicSalary <= 0 || overtimeHours <= 0) {
    return 0;
  }
  final basis = calculateOvertimeBasis(basicSalary);
  final factor = calculateOvertimeFactor(overtimeHours);
  return (basis * factor).toInt();
}

String formatMoney(int amount) {
  return NumberFormat.currency(
    locale: 'id',
    symbol: 'Rp ',
    decimalDigits: 0,
  ).format(amount);
}
