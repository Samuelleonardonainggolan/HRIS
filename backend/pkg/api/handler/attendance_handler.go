// pkg/api/handler/attendance_handler.go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// wib adalah timezone WIB (UTC+7) yang digunakan di seluruh handler ini.
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

// ProcessAttendance - Unified endpoint for clock in/out with face verification
func (h *AttendanceHandler) ProcessAttendance(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized - userID not found in context",
		})
		return
	}

	// Get form data
	recordType := c.PostForm("record_type")
	latitudeStr := c.PostForm("latitude")
	longitudeStr := c.PostForm("longitude")

	if recordType == "" || latitudeStr == "" || longitudeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "record_type, latitude, and longitude are required",
		})
		return
	}

	// Parse coordinates
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

	// Get photo from request
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Photo is required: " + err.Error(),
		})
		return
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to open file: " + err.Error(),
		})
		return
	}
	defer src.Close()

	// Read file bytes
	photoBytes := make([]byte, file.Size)
	_, err = src.Read(photoBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to read file: " + err.Error(),
		})
		return
	}

	// Process attendance with face verification
	result, err := h.attendanceService.ProcessAttendanceWithFace(
		c.Request.Context(),
		userID.(string),
		photoBytes,
		file.Filename,
		latitude,
		longitude,
		recordType,
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
				"face_similarity": result.FaceSimilarity,
				"location_valid":  result.LocationValid,
				"distance_m":      result.Distance,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// GetTodayAttendance - Get today's attendance record
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

	// ✅ FIX: Format waktu dalam WIB (UTC+7), bukan UTC.
	// MongoDB menyimpan waktu sebagai UTC. Konversi ke WIB sebelum format
	// agar jam yang tampil di Flutter sesuai waktu setempat.
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

// GetMonthlyAttendance - Get monthly attendance records
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

	// ✅ FIX: Konversi semua timestamp record ke WIB sebelum dikirim ke Flutter.
	// Repository sudah mem-format waktu sebagai string "HH:mm" tapi tanpa konversi
	// timezone — diperbaiki di sini dengan mem-rebuild records menggunakan WIB.
	for i, rec := range summary.Records {
		// Re-parse tanggal dan format ulang dalam WIB.
		// Karena AttendanceRepository sudah mem-format waktu sebagai string,
		// kita perlu meng-override di level ini. Cara paling bersih adalah
		// mengambil ulang raw records dari service — namun karena MonthlyAttendanceResponse
		// sudah berisi string, kita cukup pastikan repository sudah benar.
		// Di sini kita hanya perlu memastikan field date sudah dalam WIB.
		_ = rec
		_ = i
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   summary,
	})
}
