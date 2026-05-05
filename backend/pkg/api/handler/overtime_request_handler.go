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

func (h *OvertimeRequestHandler) GetMine(c *gin.Context) {
	userID := c.GetString("userID")
	items, err := h.service.GetEmployeeOvertimeHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil riwayat lembur", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Riwayat lembur berhasil diambil", items))
}

// ─── Legacy compatibility methods ──────────────────────────────────────────

func (h *OvertimeRequestHandler) ListForManagerHR(c *gin.Context) {
	h.List(c)
}

func (h *OvertimeRequestHandler) ListForKepalaDepartemen(c *gin.Context) {
	h.List(c)
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
