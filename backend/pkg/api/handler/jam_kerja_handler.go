// pkg/api/handler/jam_kerja_handler.go
package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type JamKerjaHandler struct {
	jamKerjaService service.JamKerjaService
}

func NewJamKerjaHandler(jamKerjaService service.JamKerjaService) *JamKerjaHandler {
	return &JamKerjaHandler{
		jamKerjaService: jamKerjaService,
	}
}

// GET /api/v1/jam-kerja
func (h *JamKerjaHandler) GetAllJamKerja(c *gin.Context) {
	data, err := h.jamKerjaService.ListJamKerja(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to get jam kerja", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Jam kerja retrieved successfully", data))
}

// GET /api/v1/jam-kerja/user/:userId
func (h *JamKerjaHandler) GetJamKerjaByUserID(c *gin.Context) {
	userID := c.Param("userId")

	data, err := h.jamKerjaService.GetJamKerjaByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to get jam kerja", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Jam kerja retrieved successfully", data))
}

// PUT /api/v1/jam-kerja/user/:userId
func (h *JamKerjaHandler) UpdateJamKerjaByUserID(c *gin.Context) {
	userID := c.Param("userId")

	var req models.UpdateJamKerjaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	normalizeJamKerjaPayloadUpdate(&req)

	// ✅ validasi manual (required)
	if len(req.DayOfWeek) == 0 || req.StartTime == "" || req.EndTime == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "day_of_week, start_time, end_time wajib diisi"))
		return
	}

	data, err := h.jamKerjaService.UpdateJamKerjaByUserID(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to update jam kerja", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Jam kerja updated successfully", data))
}

// POST /api/v1/jam-kerja
func (h *JamKerjaHandler) CreateJamKerja(c *gin.Context) {
	var req models.CreateJamKerjaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	normalizeJamKerjaPayload(&req)

	// ✅ validasi manual (required)
	if len(req.DayOfWeek) == 0 || req.StartTime == "" || req.EndTime == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "day_of_week, start_time, end_time wajib diisi"))
		return
	}

	data, err := h.jamKerjaService.CreateJamKerja(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to create jam kerja", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Jam kerja created successfully", data))
}

func (h *JamKerjaHandler) GetAvailableEmployees(c *gin.Context) {
	q := c.Query("q")

	data, err := h.jamKerjaService.SearchAvailableEmployees(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to get employees", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employees retrieved successfully", data))
}


func normalizeJamKerjaPayload(req *models.CreateJamKerjaRequest) {
	// create
	if len(req.DayOfWeek) == 0 && len(req.HariKerja) > 0 {
		req.DayOfWeek = req.HariKerja
	}
	if req.StartTime == "" && req.WaktuMulai != "" {
		req.StartTime = req.WaktuMulai
	}
	if req.EndTime == "" && req.WaktuSelesai != "" {
		req.EndTime = req.WaktuSelesai
	}
	if req.IsActive == nil && req.Aktif != nil {
		req.IsActive = req.Aktif
	}
}

func normalizeJamKerjaPayloadUpdate(req *models.UpdateJamKerjaRequest) {
	// update
	if len(req.DayOfWeek) == 0 && len(req.HariKerja) > 0 {
		req.DayOfWeek = req.HariKerja
	}
	if req.StartTime == "" && req.WaktuMulai != "" {
		req.StartTime = req.WaktuMulai
	}
	if req.EndTime == "" && req.WaktuSelesai != "" {
		req.EndTime = req.WaktuSelesai
	}
	if req.IsActive == nil && req.Aktif != nil {
		req.IsActive = req.Aktif
	}
}
