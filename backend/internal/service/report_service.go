package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
)

type ReportService interface {
	GetAttendanceActivityReport(ctx context.Context, period string, departmentID string, eventType string, approvalStatus string, search string) (*models.AttendanceActivityReportResponse, error)
}

type reportService struct {
	attendanceRepo      repository.AttendanceRepository
	leaveRequestRepo    repository.PengajuanIzinCutiRepository
	overtimeRequestRepo repository.OvertimeRequestRepository
	userRepo            repository.UserRepository
}

func NewReportService(
	attendanceRepo repository.AttendanceRepository,
	leaveRequestRepo repository.PengajuanIzinCutiRepository,
	overtimeRequestRepo repository.OvertimeRequestRepository,
	userRepo repository.UserRepository,
) ReportService {
	return &reportService{
		attendanceRepo:      attendanceRepo,
		leaveRequestRepo:    leaveRequestRepo,
		overtimeRequestRepo: overtimeRequestRepo,
		userRepo:            userRepo,
	}
}

func (s *reportService) GetAttendanceActivityReport(ctx context.Context, period string, departmentID string, eventType string, approvalStatus string, search string) (*models.AttendanceActivityReportResponse, error) {
	// 1. Parse period (YYYY-MM)
	var year, month int
	if period != "" {
		fmt.Sscanf(period, "%d-%d", &year, &month)
	} else {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}

	wib := time.FixedZone("WIB", 7*60*60)
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, wib)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	// Fetch all relevant data
	allRows := []models.AttendanceActivityReportRow{}

	// 2. Fetch Late & Missing Clock-outs
	// Note: We use FindManagerAttendanceExport to get all attendance records in range
	attendanceRecords, err := s.attendanceRepo.FindManagerAttendanceExport(ctx, startOfMonth, endOfMonth, departmentID, search)
	if err == nil {
		for _, rec := range attendanceRecords {
			// Check if Late
			if rec.Status == models.StatusLate {
				row := s.mapAttendanceToReportRow(rec, models.ReportTypeLate)
				allRows = append(allRows, row)
			}
			
			// Check if Missing Clock-out
			// A record is missing clock-out if clock_in exists but clock_out is nil 
			// and the date is before today (to avoid false positives for people still working)
			if rec.ClockInTime != nil && rec.ClockOutTime == nil && rec.Date.Before(time.Now().Truncate(24*time.Hour)) {
				row := s.mapAttendanceToReportRow(rec, models.ReportTypeMissingClockOut)
				allRows = append(allRows, row)
			}
		}
	}

	// 3. Fetch Leaves & Permissions
	leaveFilter := bson.M{
		"final_status": models.StatusApproved,
		"start_date":   bson.M{"$lt": endOfMonth.UTC()},
		"end_date":     bson.M{"$gte": startOfMonth.UTC()},
	}
	if departmentID != "" && departmentID != "all" {
		// We might need to filter by department after fetching users if leave_request doesn't have department_id
	}

	leaves, err := s.leaveRequestRepo.Find(ctx, leaveFilter)
	if err == nil {
		for _, leave := range leaves {
			// Fetch user details for each leave (can be optimized with a batch fetch)
			user, err := s.userRepo.FindByID(ctx, leave.UserID.Hex())
			if err != nil {
				continue
			}

			// Apply department and search filters
			if departmentID != "" && departmentID != "all" {
				if user.DepartmentID.Hex() != departmentID && user.DepartmentName != departmentID {
					continue
				}
			}
			if search != "" {
				sSearch := strings.ToLower(search)
				if !strings.Contains(strings.ToLower(user.FullName), sSearch) && !strings.Contains(strings.ToLower(user.PayrollNumber), sSearch) {
					continue
				}
			}

			row := s.mapLeaveToReportRow(leave, user)
			allRows = append(allRows, row)
		}
	}

	// 4. Fetch Overtime
	overtimeFilter := bson.M{
		"status": models.StatusPublished,
		"date":   bson.M{"$gte": startOfMonth.UTC(), "$lt": endOfMonth.UTC()},
	}
	
	overtimes, err := s.overtimeRequestRepo.Find(ctx, overtimeFilter)
	if err == nil {
		for _, ot := range overtimes {
			for _, emp := range ot.Employees {
				if emp.EmployeeStatus != models.EmployeeStatusAgreed {
					continue
				}

				user, err := s.userRepo.FindByID(ctx, emp.UserID.Hex())
				if err != nil {
					continue
				}

				// Apply department and search filters
				if departmentID != "" && departmentID != "all" {
					if user.DepartmentID.Hex() != departmentID && user.DepartmentName != departmentID {
						continue
					}
				}
				if search != "" {
					sSearch := strings.ToLower(search)
					if !strings.Contains(strings.ToLower(user.FullName), sSearch) && !strings.Contains(strings.ToLower(user.PayrollNumber), sSearch) {
						continue
					}
				}

				row := s.mapOvertimeToReportRow(ot, emp, user)
				allRows = append(allRows, row)
			}
		}
	}

	// 5. Final Filtering and Sorting
	filteredRows := []models.AttendanceActivityReportRow{}
	for _, row := range allRows {
		// Filter by event type
		if eventType != "" && eventType != "all" {
			if string(row.Type) != eventType {
				continue
			}
		}

		// Filter by approval status
		if approvalStatus != "" && approvalStatus != "all" {
			if strings.ToLower(row.ApprovalStatus) != strings.ToLower(approvalStatus) {
				continue
			}
		}

		filteredRows = append(filteredRows, row)
	}

	// Sort by date descending, then name ascending
	sort.Slice(filteredRows, func(i, j int) bool {
		if filteredRows[i].Date.Equal(filteredRows[j].Date) {
			return filteredRows[i].EmployeeName < filteredRows[j].EmployeeName
		}
		return filteredRows[i].Date.After(filteredRows[j].Date)
	})

	// 6. Calculate Summary
	summary := s.calculateSummary(filteredRows)

	return &models.AttendanceActivityReportResponse{
		Rows:    filteredRows,
		Summary: summary,
	}, nil
}

func (s *reportService) mapAttendanceToReportRow(rec models.ManagerAttendanceAggRow, reportType models.AttendanceActivityReportType) models.AttendanceActivityReportRow {
	wib := time.FixedZone("WIB", 7*60*60)
	
	row := models.AttendanceActivityReportRow{
		ID:               fmt.Sprintf("att-%s-%s", rec.UserID.Hex(), rec.Date.Format("20060102")),
		Date:             rec.Date.In(wib),
		DayLabel:         s.getDayLabel(rec.Date.In(wib)),
		EmployeeName:     rec.User.FullName,
		EmployeeNIK:      rec.User.PayrollNumber,
		EmployeeInitials: s.getInitials(rec.User.FullName),
		DepartmentName:   rec.User.DepartmentName,
		PositionName:     rec.User.PositionName,
		Type:             reportType,
		ApprovalStatus:   "approved", // Attendance records are generally "approved" unless disputed
		Impact:           "deduction",
	}

	if rec.ClockInTime != nil {
		row.ActualIn = rec.ClockInTime.In(wib).Format("15:04")
	}
	if rec.ClockOutTime != nil {
		row.ActualOut = rec.ClockOutTime.In(wib).Format("15:04")
	}

	if reportType == models.ReportTypeLate {
		row.Impact = "deduction"
		// Mock scheduled in if not available in rec
		row.ScheduledIn = "09:00" 
		if rec.ClockInTime != nil {
			// Calculate late minutes (simplified)
			schedIn, _ := time.ParseInLocation("15:04", "09:00", wib)
			schedIn = time.Date(rec.Date.Year(), rec.Date.Month(), rec.Date.Day(), schedIn.Hour(), schedIn.Minute(), 0, 0, wib)
			diff := rec.ClockInTime.In(wib).Sub(schedIn)
			if diff > 0 {
				row.LateMinutes = int(diff.Minutes())
			}
		}
	} else if reportType == models.ReportTypeMissingClockOut {
		row.Impact = "needs_review"
		row.ScheduledOut = "18:00"
	}

	return row
}

func (s *reportService) mapLeaveToReportRow(leave models.LeaveRequest, user *models.User) models.AttendanceActivityReportRow {
	wib := time.FixedZone("WIB", 7*60*60)
	
	reportType := models.ReportTypeLeave
	if strings.Contains(strings.ToLower(leave.TypeName), "izin") {
		reportType = models.ReportTypePermission
	}

	row := models.AttendanceActivityReportRow{
		ID:               fmt.Sprintf("leave-%s", leave.ID.Hex()),
		Date:             leave.StartDate.In(wib),
		DayLabel:         s.getDayLabel(leave.StartDate.In(wib)),
		EmployeeName:     user.FullName,
		EmployeeNIK:      user.PayrollNumber,
		EmployeeInitials: s.getInitials(user.FullName),
		DepartmentName:   user.DepartmentName,
		PositionName:     user.PositionName,
		Type:             reportType,
		DateRange:        fmt.Sprintf("%s – %s (%d hari)", 
			leave.StartDate.In(wib).Format("02 Jan"), 
			leave.EndDate.In(wib).Format("02 Jan"), 
			leave.DaysTotal),
		ApprovalStatus:   strings.ToLower(leave.FinalStatus),
		Impact:           "no_impact",
		Note:             leave.Reason,
	}

	return row
}

func (s *reportService) mapOvertimeToReportRow(ot models.OvertimeRequest, emp models.OvertimeEmployee, user *models.User) models.AttendanceActivityReportRow {
	wib := time.FixedZone("WIB", 7*60*60)
	
	// Calculate hours
	startTime, _ := time.Parse("15:04", ot.StartTime)
	endTime, _ := time.Parse("15:04", ot.EndTime)
	duration := endTime.Sub(startTime)
	if duration < 0 {
		duration += 24 * time.Hour
	}

	row := models.AttendanceActivityReportRow{
		ID:               fmt.Sprintf("ot-%s-%s", ot.ID.Hex(), emp.UserID.Hex()),
		Date:             ot.Date.In(wib),
		DayLabel:         s.getDayLabel(ot.Date.In(wib)),
		EmployeeName:     user.FullName,
		EmployeeNIK:      user.PayrollNumber,
		EmployeeInitials: s.getInitials(user.FullName),
		DepartmentName:   user.DepartmentName,
		PositionName:     user.PositionName,
		Type:             models.ReportTypeOvertime,
		OvertimeStart:    ot.StartTime,
		OvertimeEnd:      ot.EndTime,
		OvertimeHours:    duration.Hours(),
		ApprovalStatus:   "approved",
		Impact:           "addition",
		Note:             ot.Reason,
	}

	return row
}

func (s *reportService) calculateSummary(rows []models.AttendanceActivityReportRow) models.AttendanceActivityReportSummary {
	summary := models.AttendanceActivityReportSummary{}
	
	uniqueLate := make(map[string]bool)
	uniqueMissing := make(map[string]bool)
	uniqueLeave := make(map[string]bool)
	uniquePermission := make(map[string]bool)
	uniqueOvertime := make(map[string]bool)

	for _, r := range rows {
		switch r.Type {
		case models.ReportTypeLate:
			summary.Late.Events++
			uniqueLate[r.EmployeeNIK] = true
		case models.ReportTypeMissingClockOut:
			summary.Missing.Events++
			uniqueMissing[r.EmployeeNIK] = true
		case models.ReportTypeLeave:
			summary.Leave.Events++
			uniqueLeave[r.EmployeeNIK] = true
		case models.ReportTypePermission:
			summary.Permission.Events++
			uniquePermission[r.EmployeeNIK] = true
		case models.ReportTypeOvertime:
			summary.Overtime.Events++
			summary.Overtime.Hours += r.OvertimeHours
			uniqueOvertime[r.EmployeeNIK] = true
		}
	}

	summary.Late.Unique = len(uniqueLate)
	summary.Missing.Unique = len(uniqueMissing)
	summary.Leave.Unique = len(uniqueLeave)
	summary.Permission.Unique = len(uniquePermission)
	summary.Overtime.Unique = len(uniqueOvertime)
	summary.Total = len(rows)

	return summary
}

func (s *reportService) getDayLabel(t time.Time) string {
	days := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	return days[t.Weekday()]
}

func (s *reportService) getInitials(name string) string {
	parts := strings.Split(name, " ")
	initials := ""
	for i := 0; i < len(parts) && i < 2; i++ {
		if len(parts[i]) > 0 {
			initials += string(parts[i][0])
		}
	}
	return strings.ToUpper(initials)
}
