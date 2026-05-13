package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type OvertimeRequestHandler struct {
	service service.OvertimeRequestService
}

func NewOvertimeRequestHandler(service service.OvertimeRequestService) *OvertimeRequestHandler {
	return &OvertimeRequestHandler{service: service}
}

// Create - Submit new overtime (usually by Manager Dept)
func (h *OvertimeRequestHandler) Create(c *gin.Context) {
	requestedByID := c.GetString("userID")

	var req models.CreateOvertimeRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data pengajuan tidak valid", err.Error()))
		return
	}

	item, err := h.service.CreateOvertimeRequest(c.Request.Context(), requestedByID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membuat pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil dibuat", item))
}

// List - List overtime requests for Manager HR / Manager Dept
func (h *OvertimeRequestHandler) List(c *gin.Context) {
	role := c.GetString("role")
	deptID := c.Query("department_id")
	status := c.Query("status")

	filter := bson.M{}
	if deptID != "" {
		filter["department_id"] = deptID
	}
	if status != "" {
		filter["status"] = status
	}

	// For Manager Dept, restrict to their department (handled in service or here)
	// For simplicity, we trust the filter for now or add role checks.
	_ = role

	items, err := h.service.ListOvertimeRequests(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil data lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Data lembur berhasil diambil", items))
}

func (h *OvertimeRequestHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	item, err := h.service.GetOvertimeRequestByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Data lembur tidak ditemukan", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Data lembur berhasil diambil", item))
}

func (h *OvertimeRequestHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateOvertimeRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data tidak valid", err.Error()))
		return
	}

	item, err := h.service.UpdateOvertimeRequest(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal memperbarui data lembur", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Data lembur berhasil diperbarui", item))
}

func (h *OvertimeRequestHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteOvertimeRequest(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal menghapus data lembur", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Data lembur berhasil dihapus", nil))
}

// UpdateStatus - Employee confirms (agree/reject)
func (h *OvertimeRequestHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req models.UpdateEmployeeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Status tidak valid", err.Error()))
		return
	}

	if err := h.service.UpdateEmployeeStatus(c.Request.Context(), id, userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal memperbarui status", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Status berhasil diperbarui", nil))
}

func (h *OvertimeRequestHandler) ClaimReward(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req models.ClaimOvertimeRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body", "error": err.Error()})
		return
	}

	err := h.service.ClaimReward(c.Request.Context(), id, userID, req.RewardType, req.RewardDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengklaim reward", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reward berhasil diklaim"})
}

func (h *OvertimeRequestHandler) GetMine(c *gin.Context) {
	userID := c.GetString("userID")
	items, err := h.service.GetEmployeeOvertimeHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil riwayat lembur", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Riwayat lembur berhasil diambil", items))
}

// New route-compatible wrappers (route names expected elsewhere)
func (h *OvertimeRequestHandler) GetForMe(c *gin.Context) {
	h.GetMine(c)
}

func (h *OvertimeRequestHandler) GetForMeByID(c *gin.Context) {
	h.GetMineByID(c)
}

func (h *OvertimeRequestHandler) AgreeOvertimeRequest(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	req := models.UpdateEmployeeStatusRequest{Status: models.EmployeeStatusAgreed}
	if err := h.service.UpdateEmployeeStatus(c.Request.Context(), id, userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal memperbarui status", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Status berhasil diperbarui", nil))
}

func (h *OvertimeRequestHandler) RejectOvertimeRequest(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var body struct {
		RejectionNote string `json:"rejection_note"`
	}
	// it's okay if client doesn't send a body; default to empty note
	_ = c.ShouldBindJSON(&body)

	req := models.UpdateEmployeeStatusRequest{Status: models.EmployeeStatusRejected, RejectionNote: body.RejectionNote}
	if err := h.service.UpdateEmployeeStatus(c.Request.Context(), id, userID, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal memperbarui status", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Status berhasil diperbarui", nil))
}

func (h *OvertimeRequestHandler) Submit(c *gin.Context) {
	id := c.Param("id")
	status := models.StatusSubmitted
	req := models.UpdateOvertimeRequestRequest{Status: &status}
	item, err := h.service.UpdateOvertimeRequest(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengirim pengajuan lembur", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil dikirim", item))
}

func (h *OvertimeRequestHandler) Publish(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		LetterURL string `json:"letter_url"`
		Notes     string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data tidak valid", err.Error()))
		return
	}

	status := models.StatusPublished
	item, err := h.service.UpdateOvertimeRequest(c.Request.Context(), id, models.UpdateOvertimeRequestRequest{
		Status:    &status,
		LetterURL: &body.LetterURL,
		Notes:     &body.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal publish lembur", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Lembur berhasil dipublish", item))
}

func (h *OvertimeRequestHandler) PublishEmployee(c *gin.Context) {
	id := c.Param("id")
	userID := c.Param("user_id")

	var body struct {
		LetterURL string `json:"letter_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data tidak valid", err.Error()))
		return
	}

	if err := h.service.PublishEmployeeSPKL(c.Request.Context(), id, userID, body.LetterURL); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal publish SPKL karyawan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("SPKL karyawan berhasil dipublish", nil))
}

// ─── Legacy compatibility methods ──────────────────────────────────────────

func (h *OvertimeRequestHandler) ListForManagerHR(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")

	items, err := h.service.ListForManagerHR(c.Request.Context(), status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil data lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Data lembur berhasil diambil", items))
}

func (h *OvertimeRequestHandler) ListForKepalaDepartemen(c *gin.Context) {
	userID := c.GetString("userID")
	status := c.Query("status")
	search := c.Query("search")

	items, err := h.service.ListForKepalaDepartemen(c.Request.Context(), status, search, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil data lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Data lembur berhasil diambil", items))
}

func (h *OvertimeRequestHandler) GetForManagerHR(c *gin.Context) {
	h.GetByID(c)
}

func (h *OvertimeRequestHandler) GetForKepalaDepartemen(c *gin.Context) {
	h.GetByID(c)
}

func (h *OvertimeRequestHandler) ApproveByManagerHR(c *gin.Context) {
	// Not used in new flow, or map to 'Published'
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Please use Update endpoint to set status to published"})
}

func (h *OvertimeRequestHandler) RejectByManagerHR(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Please use Update endpoint"})
}

func (h *OvertimeRequestHandler) ApproveByKepalaDepartemen(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Use Create or Update"})
}

func (h *OvertimeRequestHandler) RejectByKepalaDepartemen(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Use Update"})
}

func (h *OvertimeRequestHandler) GetMineByID(c *gin.Context) {
	h.GetByID(c)
}

func (h *OvertimeRequestHandler) UpdateMine(c *gin.Context) {
	h.Update(c)
}

func (h *OvertimeRequestHandler) DeleteMine(c *gin.Context) {
	h.Delete(c)
}
