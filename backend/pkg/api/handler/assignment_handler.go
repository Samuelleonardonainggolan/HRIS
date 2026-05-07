package handler

import (
	"net/http"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type AssignmentHandler struct {
	service service.AssignmentService
}

func NewAssignmentHandler(service service.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{service: service}
}

func (h *AssignmentHandler) Create(c *gin.Context) {
	requestedByID := c.GetString("userID")

	var req models.CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data penugasan tidak valid", err.Error()))
		return
	}

	item, err := h.service.Create(c.Request.Context(), requestedByID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membuat penugasan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Penugasan berhasil dibuat", item))
}

func (h *AssignmentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Penugasan tidak ditemukan", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Data penugasan berhasil diambil", item))
}

func (h *AssignmentHandler) ListForManagerDepartemen(c *gin.Context) {
	deptID := c.Query("department_id")
	if deptID == "" {
		// Jika tidak ada di query, ambil dari user context (asumsi manager dept)
		// Namun biasanya manager dept sudah difilter di service atau route
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Department ID wajib diisi", ""))
		return
	}

	items, err := h.service.ListForManagerDepartemen(c.Request.Context(), deptID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil data penugasan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Data penugasan berhasil diambil", items))
}

func (h *AssignmentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data tidak valid", err.Error()))
		return
	}

	item, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal memperbarui penugasan", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Penugasan berhasil diperbarui", item))
}

func (h *AssignmentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal menghapus penugasan", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Penugasan berhasil dihapus", nil))
}

func (h *AssignmentHandler) PreviewOriginalSchedule(c *gin.Context) {
	userID := c.Query("user_id")
	dateStr := c.Query("date")

	if userID == "" || dateStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("User ID dan Date wajib diisi", ""))
		return
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Format tanggal tidak valid", err.Error()))
		return
	}

	orig, err := h.service.GetOriginalSchedule(c.Request.Context(), userID, parsedDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mendapatkan jadwal asli", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Jadwal asli berhasil diambil", orig))
}
