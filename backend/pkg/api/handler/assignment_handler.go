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
	var (
		items []models.AssignmentResponse
		err   error
	)

	if deptID != "" {
		items, err = h.service.ListForManagerDepartemen(c.Request.Context(), deptID)
	} else {
		managerUserID := c.GetString("userID")
		items, err = h.service.ListForManagerByUser(c.Request.Context(), managerUserID)
	}

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
		// Pass the exact error message as `message` so the frontend can display it
		c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse(err.Error(), ""))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Jadwal asli berhasil diambil", orig))
}

func (h *AssignmentHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	requestedByID := c.GetString("userID")

	item, err := h.service.Submit(c.Request.Context(), id, requestedByID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal submit penugasan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Penugasan berhasil disubmit", item))
}

func (h *AssignmentHandler) GetForMe(c *gin.Context) {
	userID := c.GetString("userID")
	items, err := h.service.ListForEmployee(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil penugasan untuk Anda", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Data penugasan berhasil diambil", items))
}

func (h *AssignmentHandler) GetForMeByID(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	item, err := h.service.GetForEmployeeByID(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Penugasan tidak ditemukan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Data penugasan berhasil diambil", item))
}

func (h *AssignmentHandler) AgreeAssignment(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	item, err := h.service.UpdateEmployeeStatus(c.Request.Context(), id, userID, models.UpdateAssignmentEmployeeStatusRequest{
		Status: models.AssignmentEmployeeStatusAgreed,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal menyetujui penugasan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Penugasan berhasil disetujui", item))
}

func (h *AssignmentHandler) RejectAssignment(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req struct {
		RejectionNote string `json:"rejection_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data tidak valid", err.Error()))
		return
	}

	item, err := h.service.UpdateEmployeeStatus(c.Request.Context(), id, userID, models.UpdateAssignmentEmployeeStatusRequest{
		Status:        models.AssignmentEmployeeStatusRejected,
		RejectionNote: req.RejectionNote,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal menolak penugasan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Penugasan berhasil ditolak", item))
}

func (h *AssignmentHandler) UseReplacementDayOff(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req models.UseReplacementDayOffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data tidak valid", err.Error()))
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.ReplacementOffDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Format tanggal tidak valid", err.Error()))
		return
	}

	item, err := h.service.UseReplacementDayOff(c.Request.Context(), id, userID, parsedDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal menggunakan day off", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Berhasil menetapkan hari off pengganti", item))
}

func (h *AssignmentHandler) GrantDayOffReward(c *gin.Context) {
	id := c.Param("id")
	employeeUserID := c.Query("user_id")

	if employeeUserID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("user_id wajib diisi", ""))
		return
	}

	item, err := h.service.GrantDayOffReward(c.Request.Context(), id, employeeUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal memberikan day off reward", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Day off reward berhasil diberikan", item))
}
