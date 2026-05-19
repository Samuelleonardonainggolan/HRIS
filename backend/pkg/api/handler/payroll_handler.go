// pkg/api/handler/payroll_handler.go
package handler

import (
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type PayrollHandler struct {
	payrollService service.PayrollService
}

func NewPayrollHandler(payrollService service.PayrollService) *PayrollHandler {
	return &PayrollHandler{payrollService: payrollService}
}

func (h *PayrollHandler) GetMyPayroll(c *gin.Context) {
	userIDRaw, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "missing user"))
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "invalid user"))
		return
	}

	now := time.Now()
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month := int(now.Month())
	year := now.Year()

	var err error
	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "invalid month"))
			return
		}
	}

	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil || year < 1970 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "invalid year"))
			return
		}
	}

	payroll, err := h.payrollService.GetPayrollForEmployee(c.Request.Context(), userID, month, year)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Not Found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Payroll retrieved successfully", payroll))
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PayrollHandler struct {
	repo           repository.PayrollRepository
	userRepo       repository.UserRepository
	salaryRepo     repository.EmployeeBasicSalaryRepository
	attendanceRepo repository.AttendanceRepository
	overtimeRepo   repository.OvertimeRequestRepository
	jamKerjaRepo   repository.JamKerjaRepository
}

func NewPayrollHandler(
	repo repository.PayrollRepository,
	userRepo repository.UserRepository,
	salaryRepo repository.EmployeeBasicSalaryRepository,
	attendanceRepo repository.AttendanceRepository,
	overtimeRepo repository.OvertimeRequestRepository,
	jamKerjaRepo repository.JamKerjaRepository,
) *PayrollHandler {
	return &PayrollHandler{
		repo:           repo,
		userRepo:       userRepo,
		salaryRepo:     salaryRepo,
		attendanceRepo: attendanceRepo,
		overtimeRepo:   overtimeRepo,
		jamKerjaRepo:   jamKerjaRepo,
	}
}

// GetPayrolls fetches payrolls with filters, including employees who don't have records yet
func (h *PayrollHandler) GetPayrolls(c *gin.Context) {
	monthStr := c.Query("month")
	yearStr := c.Query("year")
	deptID := c.Query("department_id")
	q := c.Query("q")

	var month, year int
	if monthStr != "" {
		month, _ = strconv.Atoi(monthStr)
	}
	if yearStr != "" {
		year, _ = strconv.Atoi(yearStr)
	}

	ctx := c.Request.Context()

	// 1. Fetch all active users
	users, err := h.userRepo.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	// 2. Fetch existing payroll records for the month/year
	filter := bson.M{}
	if month > 0 { filter["month"] = month }
	if year > 0 { filter["year"] = year }
	
	payrolls, _ := h.repo.FindAll(ctx, filter)
	payrollMap := make(map[string]models.Payroll)
	for _, p := range payrolls {
		payrollMap[p.UserID.Hex()] = p
	}

	response := []gin.H{}
	for _, u := range users {
		// Filter by department if requested
		if deptID != "" && deptID != "Semua Departemen" {
			if u.DepartmentID.Hex() != deptID && u.DepartmentName != deptID {
				continue
			}
		}

		// Filter by search query if requested
		if q != "" {
			// Sederhana: cek nama
			// ...
		}

		// Fetch basic salary
		salaryDoc, _ := h.salaryRepo.FindActiveByUserID(ctx, u.ID.Hex())
		if salaryDoc == nil {
			continue // Skip employees without basic salary
		}

		p, exists := payrollMap[u.ID.Hex()]
		
		initials := ""
		if len(u.FullName) > 0 {
			initials = string(u.FullName[0])
		}

		row := gin.H{
			"id":               "",
			"user_id":          u.ID.Hex(),
			"name":             u.FullName,
			"initials":         initials,
			"position":         u.PositionName,
			"department":       u.DepartmentName,
			"basicSalary":      salaryDoc.BasicSalary,
			"bonus10":          0,
			"overtime":         0,
			"deduction":        0,
			"netTotal":         salaryDoc.BasicSalary,
			"status":           "not_generated",
			"month":            month,
			"year":             year,
		}

		if exists {
			row["id"] = p.ID.Hex()
			row["basicSalary"] = p.BasicSalaryValue
			row["bonus10"] = p.OtherEarningsValue
			row["overtime"] = p.OvertimePayValue
			row["deduction"] = p.LateDeductionValue + p.AbsentDeductionValue
			row["netTotal"] = p.NetSalaryValue
			row["status"] = p.Status
		}

		response = append(response, row)
	}

	c.JSON(http.StatusOK, response)
}

// GenerateMonthlyPayrolls triggers generation for all active employees
// POST /payrolls/generate
func (h *PayrollHandler) GenerateMonthlyPayrolls(c *gin.Context) {
	var req struct {
		Month int `json:"month" binding:"required"`
		Year  int `json:"year" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 1. Fetch all active users
	users, err := h.userRepo.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	generatedCount := 0
	for _, u := range users {
		// 2. Fetch Basic Salary
		salaryDoc, err := h.salaryRepo.FindActiveByUserID(ctx, u.ID.Hex())
		if err != nil || salaryDoc == nil {
			continue // Skip if no basic salary set
		}

		// 3. Check if already exists
		existing, _ := h.repo.FindAll(ctx, bson.M{
			"user_id": u.ID,
			"month":   req.Month,
			"year":    req.Year,
		})
		if len(existing) > 0 {
			continue // Skip or Update? Let's skip for now
		}

		// 4. Calculate Attendance (Lateness & Absence)
		attendances, _ := h.attendanceRepo.FindByUserIDAndMonth(ctx, u.ID, req.Year, req.Month)
		
		daysPresent := len(attendances)
		totalLateMinutes := 0
		
		// Fetch Jam Kerja for the user
		jk, _ := h.jamKerjaRepo.FindByUserID(ctx, u.ID.Hex())

		for _, att := range attendances {
			if att.Status == models.StatusLate && att.ClockInTime != nil && jk != nil {
				// Compare ClockInTime with jk.StartTime
				// We only care about the time (HH:mm)
				
				// Standard start time for the day of the week
				// For now, assume jk.StartTime is the standard for all workdays
				// Convert both to minutes from midnight for comparison
				
				clockInMinutes := att.ClockInTime.Hour()*60 + att.ClockInTime.Minute()
				startMinutes := jk.StartTime.Hour()*60 + jk.StartTime.Minute()
				
				if clockInMinutes > startMinutes {
					totalLateMinutes += (clockInMinutes - startMinutes)
				}
			}
		}

		// Calculate how many workdays should have passed so far
		now := time.Now()
		daysInMonth := time.Date(req.Year, time.Month(req.Month+1), 0, 0, 0, 0, 0, time.UTC).Day()
		
		var referenceWorkdays int
		if req.Year < now.Year() || (req.Year == now.Year() && req.Month < int(now.Month())) {
			// Past month: assume full 26 workdays
			referenceWorkdays = 26
		} else if req.Year == now.Year() && req.Month == int(now.Month()) {
			// Current month: pro-rate the 26 days based on days passed so far
			referenceWorkdays = (now.Day() * 26) / daysInMonth
		} else {
			// Future month
			referenceWorkdays = 0
		}
		
		absentDays := referenceWorkdays - daysPresent
		if absentDays < 0 {
			absentDays = 0
		}

		// 5. Calculate Overtime (Money reward only)
		overtimes, _ := h.overtimeRepo.Find(ctx, bson.M{
			"employees.user_id": u.ID,
			"date": bson.M{
				"$gte": time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC),
				"$lt":  time.Date(req.Year, time.Month(req.Month+1), 1, 0, 0, 0, 0, time.UTC),
			},
			"status": models.StatusPublished,
		})

		totalOvertimeHours := 0.0
		for _, ot := range overtimes {
			for _, emp := range ot.Employees {
				if emp.UserID == u.ID && emp.Reward.RewardType == models.OvertimeRewardTypeMoney {
					totalOvertimeHours += ot.GetDurationHours()
				}
			}
		}

		// 6. Create Payroll Record
		payroll := &models.Payroll{
			UserID:                u.ID,
			Month:                 req.Month,
			Year:                  req.Year,
			BasicSalaryValue:      salaryDoc.BasicSalary,
			BasicSalary:           fmt.Sprintf("%d", salaryDoc.BasicSalary),
			MonthlyHoursDivisor:   173,
			WorkdaysDivisor:       26,
			MinutesPerWorkday:     480,
			Status:                "draft",
		}

		payroll.OvertimePayValue = payroll.CalculateOvertimePay(totalOvertimeHours)
		payroll.OvertimePay = fmt.Sprintf("%d", payroll.OvertimePayValue)
		
		payroll.LateDeductionValue = payroll.CalculateLateDeduction(totalLateMinutes)
		payroll.LateDeduction = fmt.Sprintf("%d", payroll.LateDeductionValue)
		
		payroll.AbsentDeductionValue = payroll.CalculateAbsentDeduction(absentDays)
		payroll.AbsentDeduction = fmt.Sprintf("%d", payroll.AbsentDeductionValue)

		payroll.RecalculateNetSalary()
		payroll.NetSalary = fmt.Sprintf("%d", payroll.NetSalaryValue)

		err = h.repo.Create(ctx, payroll)
		if err == nil {
			generatedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Successfully generated %d payroll records", generatedCount),
	})
}

// GetPayrollDetail fetches single payroll
// GET /payrolls/:id
func (h *PayrollHandler) GetPayrollDetail(c *gin.Context) {
	id := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	payroll, err := h.repo.FindByID(c.Request.Context(), oid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payroll not found"})
		return
	}

	user, _ := h.userRepo.FindByID(c.Request.Context(), payroll.UserID.Hex())
	
	resp := payroll.ToResponse()
	// You might want to wrap this in a larger object with user details
	c.JSON(http.StatusOK, gin.H{
		"payroll": resp,
		"user":    user,
	})
}
