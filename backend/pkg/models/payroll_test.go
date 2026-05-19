package models

import (
	"testing"
)

func TestPayrollCalculations(t *testing.T) {
	payroll := &Payroll{
		BasicSalaryValue:    17300000, // Easier calculation: 17.3M / 173 = 100k per hour basis
		MonthlyHoursDivisor: 173,
		WorkdaysDivisor:     26,
		MinutesPerWorkday:   480,
	}

	t.Run("CalculateOvertimePay", func(t *testing.T) {
		// 3 hours overtime: (1.5 + 2.0 + 2.0) = 5.5 * Basis
		// Basis = 17,300,000 / 173 = 100,000
		// Expect = 5.5 * 100,000 = 550,000
		expected := int64(550000)
		result := payroll.CalculateOvertimePay(3.0)
		if result != expected {
			t.Errorf("Expected %d, got %d", expected, result)
		}

		// 1 hour: 1.5 * Basis = 150,000
		expected1 := int64(150000)
		result1 := payroll.CalculateOvertimePay(1.0)
		if result1 != expected1 {
			t.Errorf("Expected %d, got %d", expected1, result1)
		}

		// 0.5 hour: 0.5 * 1.5 * Basis = 75,000
		expectedHalf := int64(75000)
		resultHalf := payroll.CalculateOvertimePay(0.5)
		if resultHalf != expectedHalf {
			t.Errorf("Expected %d, got %d", expectedHalf, resultHalf)
		}
	})

	t.Run("CalculateLateDeduction", func(t *testing.T) {
		// Basic Salary / 26 / 480
		// 17,300,000 / 26 = 665,384.615...
		// 665,384.615 / 480 = 1,386.217...
		// For 10 minutes: 13,862.17... -> 13,862 (int64)
		expected := int64(13862)
		result := payroll.CalculateLateDeduction(10)
		if result != expected {
			t.Errorf("Expected %d, got %d", expected, result)
		}
	})

	t.Run("CalculateAbsentDeduction", func(t *testing.T) {
		// Basic Salary / 26 * days
		// 17,300,000 / 26 = 665,384.615...
		// For 2 days: 1,330,769.23... -> 1,330,769
		expected := int64(1330769)
		result := payroll.CalculateAbsentDeduction(2)
		if result != expected {
			t.Errorf("Expected %d, got %d", expected, result)
		}
	})

	t.Run("RecalculateNetSalary", func(t *testing.T) {
		payroll.OvertimePayValue = 550000
		payroll.LateDeductionValue = 13862
		payroll.AbsentDeductionValue = 1330769
		
		payroll.RecalculateNetSalary()
		
		// Net = 17,300,000 + 550,000 - 13,862 - 1,330,769 = 16,505,369
		expected := int64(16505369)
		if payroll.NetSalaryValue != expected {
			t.Errorf("Expected %d, got %d", expected, payroll.NetSalaryValue)
		}
	})
}
