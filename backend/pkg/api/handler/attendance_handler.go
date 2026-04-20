// pkg/api/handler/attendance_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/gin-gonic/gin"
)

var wib = time.FixedZone("WIB", 7*60*60)

type AttendanceHandler struct {
	attendanceService service.AttendanceService
	faceService       service.FaceService
}

func NewAttendanceHandler(attendanceService service.AttendanceService, faceService service.FaceService) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
		faceService:       faceService,
	}
}

// ✅ BARU: Get work schedule info untuk dashboard
// func (h *AttendanceHandler) GetWorkScheduleInfo(c *gin.Context) {
// 	userID, exists := c.Get("userID")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"status":  "error",
// 			"message": "Unauthorized",
// 		})
// 		return
// 	}

// 	info, err := h.attendanceService.GetWorkScheduleInfo(c.Request.Context(), userID.(string))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"status":  "error",
// 			"message": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status": "success",
// 		"data":   info,
// 	})
// }

func (h *AttendanceHandler) ProcessAttendance(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized - userID not found in context",
		})
		return
	}

	recordType := c.PostForm("record_type")
	latitudeStr := c.PostForm("latitude")
	longitudeStr := c.PostForm("longitude")
	verifyOnlyStr := c.PostForm("verify_only") // ✅ BARU: parameter untuk verifikasi saja atau submit

	if recordType == "" || latitudeStr == "" || longitudeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "record_type, latitude, and longitude are required",
		})
		return
	}

	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid latitude format",
		})
		return
	}

	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid longitude format",
		})
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Photo is required: " + err.Error(),
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to open file: " + err.Error(),
		})
		return
	}
	defer src.Close()

	photoBytes := make([]byte, file.Size)
	_, err = src.Read(photoBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to read file: " + err.Error(),
		})
		return
	}

	// ✅ PERBAIKAN: Extract verify_only parameter
	verifyOnly := verifyOnlyStr == "true"

	result, err := h.attendanceService.ProcessAttendanceWithFace(
		c.Request.Context(),
		userID.(string),
		photoBytes,
		file.Filename,
		latitude,
		longitude,
		recordType,
		verifyOnly, // ✅ Kirim flag untuk kontrol verifikasi vs submit
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": result.Message,
			"data": gin.H{
				"face_similarity":      result.FaceSimilarity,
				"location_valid":       result.LocationValid,
				"distance_m":           result.Distance,
				"is_clock_in_allowed":  result.IsClockInAllowed,
				"is_clock_out_allowed": result.IsClockOutAllowed,
				"clock_in_window":      result.ClockInWindow,
				"clock_out_window":     result.ClockOutWindow,
				"work_schedule_found":  result.WorkScheduleFound,
				"next_window_open":     result.NextWindowOpen,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

func (h *AttendanceHandler) GetTodayAttendance(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		userID, exists = c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
			})
			return
		}
	}

	attendance, err := h.attendanceService.GetTodayAttendance(c.Request.Context(), userID.(string))
	if err != nil || attendance == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"data":    nil,
			"message": "No attendance record for today",
		})
		return
	}

	clockIn := "--:--"
	if attendance.ClockInTime != nil {
		clockIn = attendance.ClockInTime.In(wib).Format("15:04")
	}

	response := map[string]interface{}{
		"id":              attendance.ID.Hex(),
		"date":            attendance.Date.In(wib).Format("2006-01-02"),
		"clock_in_time":   clockIn,
		"status":          string(attendance.Status),
		"work_hours":      attendance.WorkHours,
		"face_similarity": attendance.FaceSimilarity,
	}

	if attendance.ClockOutTime != nil {
		response["clock_out_time"] = attendance.ClockOutTime.In(wib).Format("15:04")
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

func (h *AttendanceHandler) GetMonthlyAttendance(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		userID, exists = c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Unauthorized",
			})
			return
		}
	}

	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month, _ := strconv.Atoi(monthStr)
	year, _ := strconv.Atoi(yearStr)

	if month == 0 {
		month = int(time.Now().Month())
	}
	if year == 0 {
		year = time.Now().Year()
	}

	summary, err := h.attendanceService.GetMonthlyAttendance(c.Request.Context(), userID.(string), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   summary,
	})
}
func (h *AttendanceHandler) GetScheduleInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	info, err := h.attendanceService.GetScheduleInfo(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   info,
	})
}

func (h *AttendanceHandler) GetManagerAttendanceRecords(c *gin.Context) {
	from, toExclusive, ok := parseManagerAttendanceDateRange(c)
	if !ok {
		return
	}

	department := c.Query("department")
	q := c.Query("q")

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)
	if pageSize > 100 {
		pageSize = 100
	}

	resp, err := h.attendanceService.GetManagerAttendance(c.Request.Context(), from, toExclusive, department, q, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp})
}

func (h *AttendanceHandler) ExportManagerAttendanceRecords(c *gin.Context) {
	from, toExclusive, ok := parseManagerAttendanceDateRange(c)
	if !ok {
		return
	}

	department := c.Query("department")
	q := c.Query("q")

	reader, filename, err := h.attendanceService.ExportManagerAttendanceCSVStream(c.Request.Context(), from, toExclusive, department, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	defer reader.Close()

	if filename == "" {
		filename = "presensi.csv"
	}
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", filename)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", contentDisposition)
	c.DataFromReader(http.StatusOK, -1, "text/csv; charset=utf-8", reader, nil)
}

func (h *AttendanceHandler) GetManagerDeptAttendanceRecords(c *gin.Context) {
	from, toExclusive, ok := parseManagerAttendanceDateRange(c)
	if !ok {
		return
	}

	department, exists := c.Get("userDepartment")
	departmentName, okCast := department.(string)
	if !exists || !okCast || departmentName == "" {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "departemen user tidak ditemukan"})
		return
	}

	q := c.Query("q")
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64)
	if pageSize > 100 {
		pageSize = 100
	}

	resp, err := h.attendanceService.GetManagerAttendance(
		c.Request.Context(),
		from,
		toExclusive,
		departmentName,
		q,
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp})
}

func (h *AttendanceHandler) ExportManagerDeptAttendanceRecords(c *gin.Context) {
	from, toExclusive, ok := parseManagerAttendanceDateRange(c)
	if !ok {
		return
	}

	department, exists := c.Get("userDepartment")
	departmentName, okCast := department.(string)
	if !exists || !okCast || departmentName == "" {
		c.JSON(http.StatusForbidden, gin.H{"status": "error", "message": "departemen user tidak ditemukan"})
		return
	}

	q := c.Query("q")
	reader, filename, err := h.attendanceService.ExportManagerAttendanceCSVStream(
		c.Request.Context(),
		from,
		toExclusive,
		departmentName,
		q,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	defer reader.Close()

	if filename == "" {
		filename = "presensi.csv"
	}
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", filename)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", contentDisposition)
	c.DataFromReader(http.StatusOK, -1, "text/csv; charset=utf-8", reader, nil)
}

func parseManagerAttendanceDateRange(c *gin.Context) (time.Time, time.Time, bool) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "from dan to wajib diisi (YYYY-MM-DD)"})
		return time.Time{}, time.Time{}, false
	}

	fromDate, err := time.ParseInLocation("2006-01-02", fromStr, wib)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "format from tidak valid"})
		return time.Time{}, time.Time{}, false
	}
	toDate, err := time.ParseInLocation("2006-01-02", toStr, wib)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "format to tidak valid"})
		return time.Time{}, time.Time{}, false
	}
	if toDate.Before(fromDate) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "to tidak boleh sebelum from"})
		return time.Time{}, time.Time{}, false
	}

	from := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, wib)
	toExclusive := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, wib).Add(24 * time.Hour)

	return from, toExclusive, true
}
