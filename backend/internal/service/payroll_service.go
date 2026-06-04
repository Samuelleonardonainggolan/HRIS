// internal/service/payroll_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// wibPayroll adalah timezone WIB (UTC+7) untuk kalkulasi tanggal payroll.
var wibPayroll = time.FixedZone("WIB", 7*60*60)

// =============================================================================
// Interface
// =============================================================================

type PayrollService interface {
	// Untuk karyawan (staff) — melihat slip gaji sendiri
	GetPayrollForEmployee(ctx context.Context, userID string, month, year int) (*models.PayrollResponse, error)

	// Untuk akuntan — generate payroll seluruh karyawan
	GeneratePayrolls(ctx context.Context, month, year int) ([]models.PayrollRecord, error)

	// Untuk akuntan — list semua payroll
	ListPayrolls(ctx context.Context, month, year int) ([]models.PayrollRecord, error)

	// Untuk akuntan — detail satu payroll (dengan info karyawan)
	GetPayrollDetail(ctx context.Context, id string) (*models.PayrollDetailResponse, error)

	// Untuk akuntan — update status (approve / mark paid)
	UpdatePayrollStatus(ctx context.Context, id string, status string) error
}

// =============================================================================
// Implementation
// =============================================================================

type payrollService struct {
	payrollRepo    repository.PayrollRepository
	attendanceRepo repository.AttendanceRepository
	overtimeRepo   repository.OvertimeRequestRepository
	salaryRepo     repository.EmployeeBasicSalaryRepository
	userRepo       repository.UserRepository
}

func NewPayrollService(
	payrollRepo repository.PayrollRepository,
	attendanceRepo repository.AttendanceRepository,
	overtimeRepo repository.OvertimeRequestRepository,
	salaryRepo repository.EmployeeBasicSalaryRepository,
	userRepo repository.UserRepository,
) PayrollService {
	return &payrollService{
		payrollRepo:    payrollRepo,
		attendanceRepo: attendanceRepo,
		overtimeRepo:   overtimeRepo,
		salaryRepo:     salaryRepo,
		userRepo:       userRepo,
	}
}

// =============================================================================
// GetPayrollForEmployee — staff melihat slip sendiri
// =============================================================================

func (s *payrollService) GetPayrollForEmployee(ctx context.Context, userID string, month, year int) (*models.PayrollResponse, error) {
	payroll, err := s.payrollRepo.FindByUserAndMonthYear(ctx, userID, month, year)
	if err != nil {
		return nil, err
	}
	if payroll == nil {
		return nil, errors.New("slip gaji tidak ditemukan untuk periode ini")
	}

	resp := payroll.ToResponse()
	return &resp, nil
}

// =============================================================================
// GeneratePayrolls — akuntan generate payroll sebulan penuh
// =============================================================================

func (s *payrollService) GeneratePayrolls(ctx context.Context, month, year int) ([]models.PayrollRecord, error) {
	// Ambil semua user aktif
	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal ambil daftar karyawan: %w", err)
	}
	activeUsers := make([]models.User, 0, len(users))
	for _, u := range users {
		if u.IsActive {
			activeUsers = append(activeUsers, u)
		}
	}

	// Batas hari yang sudah berlalu dalam bulan ini (WIB)
	nowWIB := time.Now().In(wibPayroll)
	lastPassedDay := s.lastPassedWorkday(nowWIB, month, year)

	var records []models.PayrollRecord

	for _, u := range activeUsers {
		userID := u.ID.Hex()

		// 1. Ambil gaji pokok
		salaryDoc, err := s.salaryRepo.FindActiveByUserID(ctx, userID)
		if err != nil || salaryDoc == nil {
			// Skip jika belum ada konfigurasi gaji
			continue
		}
		basicSalary := salaryDoc.BasicSalary // int64

		// 2. Ambil record absensi bulan ini
		uid, _ := primitive.ObjectIDFromHex(userID)
		attendances, err := s.attendanceRepo.FindByUserIDAndMonth(ctx, uid, year, month)
		if err != nil {
			attendances = []models.Attendance{}
		}

		// 3. Hitung keterlambatan & mangkir (hanya hari yang sudah berlalu)
		lateMinutes, absentDays, presentDays := s.calcAttendanceSummary(
			attendances, month, year, lastPassedDay,
		)

		// 4. Ambil overtime requests yang reward_type = "money" & status granted
		overtimePay, overtimeHours, overtimeDesc := s.calcOvertimePay(
			ctx, userID, month, year, basicSalary,
		)

		// 5. Hitung potongan
		const workdaysDivisor = 26
		const minutesPerWorkday = 480 // 8 jam × 60 menit

		// Potongan terlambat: (gaji_pokok / 26) / 480 × total_menit_terlambat
		lateDeductionValue := int64(0)
		if lateMinutes > 0 {
			ratePerMinute := float64(basicSalary) / workdaysDivisor / minutesPerWorkday
			lateDeductionValue = int64(ratePerMinute * float64(lateMinutes))
		}

		// Potongan mangkir: (gaji_pokok / 26) × hari_mangkir
		absentDeductionValue := int64(0)
		if absentDays > 0 {
			dailyRate := float64(basicSalary) / workdaysDivisor
			absentDeductionValue = int64(dailyRate * float64(absentDays))
		}

		totalDeduction := lateDeductionValue + absentDeductionValue
		netSalary := basicSalary + overtimePay - totalDeduction
		if netSalary < 0 {
			netSalary = 0
		}

		// 6. Cek apakah payroll sudah ada (upsert)
		existing, _ := s.payrollRepo.FindByUserAndMonthYear(ctx, userID, month, year)

		payroll := &models.Payroll{
			UserID: uid,
			Month:  month,
			Year:   year,

			BasicSalary:      fmt.Sprintf("Rp %d", basicSalary),
			BasicSalaryValue: basicSalary,
			NetSalary:        fmt.Sprintf("Rp %d", netSalary),
			NetSalaryValue:   netSalary,

			TotalDaysPresent: fmt.Sprintf("%d hari", presentDays),
			LateMinutesTotal: lateMinutes,
			AbsentDays:       absentDays,

			OvertimeHoursPaid: overtimeHours,
			OvertimePay:       overtimeDesc,
			OvertimePayValue:  overtimePay,

			LateDeduction:        fmt.Sprintf("Rp %d", lateDeductionValue),
			LateDeductionValue:   lateDeductionValue,
			AbsentDeduction:      fmt.Sprintf("Rp %d", absentDeductionValue),
			AbsentDeductionValue: absentDeductionValue,

			MonthlyHoursDivisor: 173,
			WorkdaysDivisor:     workdaysDivisor,
			MinutesPerWorkday:   minutesPerWorkday,

			Status: "draft",
		}

		if existing != nil {
			// Update — pertahankan status jika sudah approved/paid
			if existing.Status == "approved" || existing.Status == "paid" {
				payroll.Status = existing.Status
			}
			payroll.ID = existing.ID
			payroll.CreatedAt = existing.CreatedAt
			_ = s.payrollRepo.Update(ctx, existing.ID, payroll)
		} else {
			_ = s.payrollRepo.Create(ctx, payroll)
		}

		// Build record untuk response
		initials := ""
		if len(u.FullName) > 0 {
			parts := splitName(u.FullName)
			for i, p := range parts {
				if i < 2 && len(p) > 0 {
					initials += string(p[0])
				}
			}
		}

		records = append(records, models.PayrollRecord{
			ID:          payroll.ID.Hex(),
			UserID:      userID,
			Name:        u.FullName,
			Initials:    initials,
			Position:    u.PositionName,
			Department:  u.DepartmentName,
			BasicSalary: basicSalary,
			Overtime:    overtimePay,
			Deduction:   totalDeduction,
			NetTotal:    netSalary,
			Status:      payroll.Status,
			Month:       month,
			Year:        year,
		})
	}

	return records, nil
}

// =============================================================================
// ListPayrolls — akuntan melihat daftar payroll
// =============================================================================

func (s *payrollService) ListPayrolls(ctx context.Context, month, year int) ([]models.PayrollRecord, error) {
	filter := bson.M{}
	if month > 0 {
		filter["month"] = month
	}
	if year > 0 {
		filter["year"] = year
	}

	payrolls, err := s.payrollRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Ambil semua user_id yang terlibat agar bisa enrichment nama
	userIDs := make([]string, 0, len(payrolls))
	for _, p := range payrolls {
		userIDs = append(userIDs, p.UserID.Hex())
	}
	users, _ := s.userRepo.FindByIDs(ctx, userIDs)
	userMap := map[string]models.User{}
	for _, u := range users {
		userMap[u.ID.Hex()] = u
	}

	records := make([]models.PayrollRecord, 0, len(payrolls))
	for _, p := range payrolls {
		u := userMap[p.UserID.Hex()]

		initials := ""
		parts := splitName(u.FullName)
		for i, part := range parts {
			if i < 2 && len(part) > 0 {
				initials += string(part[0])
			}
		}

		records = append(records, models.PayrollRecord{
			ID:          p.ID.Hex(),
			UserID:      p.UserID.Hex(),
			Name:        u.FullName,
			Initials:    initials,
			Position:    u.PositionName,
			Department:  u.DepartmentName,
			BasicSalary: p.BasicSalaryValue,
			Overtime:    p.OvertimePayValue,
			Deduction:   p.LateDeductionValue + p.AbsentDeductionValue + p.OtherDeductionsValue,
			NetTotal:    p.NetSalaryValue,
			Status:      p.Status,
			Month:       p.Month,
			Year:        p.Year,
		})
	}

	return records, nil
}

// =============================================================================
// GetPayrollDetail — detail satu payroll lengkap
// =============================================================================

func (s *payrollService) GetPayrollDetail(ctx context.Context, id string) (*models.PayrollDetailResponse, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("payroll ID tidak valid")
	}

	p, err := s.payrollRepo.FindByID(ctx, oid)
	if err != nil || p == nil {
		return nil, errors.New("payroll tidak ditemukan")
	}

	u, _ := s.userRepo.FindByID(ctx, p.UserID.Hex())

	detail := &models.PayrollDetailResponse{
		PayrollResponse: p.ToResponse(),
	}
	if u != nil {
		detail.EmployeeName = u.FullName
		detail.EmployeePosition = u.PositionName
		detail.EmployeeDept = u.DepartmentName
		detail.PayrollNumber = u.PayrollNumber
	}

	return detail, nil
}

// =============================================================================
// UpdatePayrollStatus
// =============================================================================

func (s *payrollService) UpdatePayrollStatus(ctx context.Context, id string, status string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("payroll ID tidak valid")
	}

	p, err := s.payrollRepo.FindByID(ctx, oid)
	if err != nil || p == nil {
		return errors.New("payroll tidak ditemukan")
	}

	p.Status = status
	return s.payrollRepo.Update(ctx, oid, p)
}

// =============================================================================
// Internal Helpers
// =============================================================================

// lastPassedWorkday mengembalikan hari terakhir yang sudah berlalu (≤ hari ini)
// dalam bulan yang dimaksud. Jika bulan sudah lampau, kembalikan hari terakhir bulan itu.
func (s *payrollService) lastPassedWorkday(nowWIB time.Time, month, year int) int {
	// Hari ini dalam WIB
	todayYear := nowWIB.Year()
	todayMonth := int(nowWIB.Month())
	todayDay := nowWIB.Day()

	if year < todayYear || (year == todayYear && month < todayMonth) {
		// Bulan sudah lampau → ambil semua hari di bulan itu
		lastDay := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, wibPayroll).Day()
		return lastDay
	}

	if year == todayYear && month == todayMonth {
		// Bulan ini → sampai hari ini
		return todayDay
	}

	// Bulan depan → 0 hari yang sudah berlalu (jangan hitung apapun)
	return 0
}

// calcAttendanceSummary menghitung menit terlambat, hari mangkir, dan hari hadir.
// Hanya mempertimbangkan hari kerja yang sudah berlalu (≤ lastPassedDay).
func (s *payrollService) calcAttendanceSummary(
	attendances []models.Attendance,
	month, year, lastPassedDay int,
) (lateMinutes int, absentDays int, presentDays int) {
	if lastPassedDay <= 0 {
		return
	}

	// Bangun set tanggal hadir berdasarkan ClockInTime —
	// mangkir hanya jika karyawan tidak clock_in sama sekali pada hari itu.
	// Karyawan yang clock_in tapi tidak clock_out tetap dianggap HADIR.
	presentSet := map[int]bool{}
	for _, att := range attendances {
		// Hanya anggap hadir jika ada ClockInTime (sudah absen masuk)
		if att.ClockInTime == nil {
			continue
		}

		dayWIB := att.Date.In(wibPayroll).Day()
		presentSet[dayWIB] = true

		// Hitung keterlambatan: hanya untuk hari ≤ lastPassedDay
		if dayWIB > lastPassedDay {
			continue
		}

		presentDays++

		// Keterlambatan: jika ada ClockInTime dan status Late
		if att.Status == models.StatusLate {
			clockInWIB := att.ClockInTime.In(wibPayroll)
			scheduleStart := time.Date(
				clockInWIB.Year(), clockInWIB.Month(), clockInWIB.Day(),
				8, 0, 0, 0, wibPayroll,
			)
			if clockInWIB.After(scheduleStart) {
				diff := clockInWIB.Sub(scheduleStart)
				lateMinutes += int(diff.Minutes())
			}
		}
	}

	// Hitung mangkir: hari kerja yang sudah berlalu namun tidak ada record kehadiran
	// (weekday saja, tidak termasuk weekend)
	for day := 1; day <= lastPassedDay; day++ {
		d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, wibPayroll)
		if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue // skip weekend
		}
		if !presentSet[day] {
			absentDays++
		}
	}

	return
}

// calcOvertimePay menghitung total bonus lembur dari overtime requests
// yang reward_type = "money" dan reward status = "granted" dalam bulan/tahun tsb.
// Formula:
//
//	basis = basicSalary / 173
//	jam 1  → 1.5 × basis
//	jam 2+ → 2.0 × basis
//	total 3 jam = (1.5 + 2 + 2) × basis = 5.5 × basis
func (s *payrollService) calcOvertimePay(
	ctx context.Context,
	userID string, month, year int,
	basicSalary int64,
) (totalPay int64, totalHours float64, description string) {
	const divisor = 173.0
	basis := float64(basicSalary) / divisor

	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return
	}

	// Cari overtime requests yang melibatkan user ini dalam bulan terkait
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, wibPayroll).UTC()
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	overtimes, err := s.overtimeRepo.Find(ctx, bson.M{
		"employees.user_id": uid,
		"date": bson.M{
			"$gte": startOfMonth,
			"$lt":  endOfMonth,
		},
		// Sertakan "submitted" dan "published" — reward bisa sudah
		// "granted" meskipun status OT belum berubah ke published.
		"status": bson.M{"$in": []string{models.StatusSubmitted, models.StatusPublished}},
	})
	if err != nil {
		return
	}

	sessionCount := 0
	for _, ot := range overtimes {
		for _, emp := range ot.Employees {
			if emp.UserID != uid {
				continue
			}
			// Syarat: reward_type = "money" dan status = "granted"
			if emp.Reward.RewardType != models.OvertimeRewardTypeMoney {
				continue
			}
			if emp.Reward.Status != models.OvertimeRewardStatusGranted &&
				emp.Reward.Status != models.OvertimeRewardStatusUsed {
				continue
			}

			// Hitung durasi lembur sesi ini
			hours := ot.GetDurationHours()
			if hours <= 0 {
				continue
			}

			sessionPay := int64(0)
			{
				rv := reflect.ValueOf(emp.Reward)
				if rv.IsValid() {
					if rv.Kind() == reflect.Ptr {
						rv = rv.Elem()
					}
					if rv.Kind() == reflect.Struct {
						f := rv.FieldByName("RewardNominal")
						if f.IsValid() {
							if f.CanFloat() {
								sessionPay = int64(f.Float())
							} else if f.CanInt() {
								sessionPay = f.Int()
							}
						}
					}
				}
			}
			if sessionPay <= 0 {
				// Fallback untuk data lama yang belum menyimpan nominal reward
				factor := overtimeFactor(hours)
				sessionPay = int64(basis * factor)
			}

			totalPay += sessionPay
			totalHours += hours
			sessionCount++
		}
	}

	if sessionCount > 0 {
		description = fmt.Sprintf("%.1f jam × basis (Rp %.0f)", totalHours, basis)
	}

	return
}

// overtimeFactor menghitung total faktor perkalian untuk n jam lembur.
// Jam ke-1: 1.5×, jam ke-2 dst: 2.0×
// Contoh: 3 jam → 1.5 + 2.0 + 2.0 = 5.5
func overtimeFactor(hours float64) float64 {
	if hours <= 0 {
		return 0
	}
	if hours <= 1 {
		return hours * 1.5
	}
	return 1.5 + (hours-1)*2.0
}

// splitName memisahkan nama penuh menjadi kata-kata.
func splitName(name string) []string {
	parts := []string{}
	current := ""
	for _, r := range name {
		if r == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
