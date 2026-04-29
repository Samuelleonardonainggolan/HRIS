package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type OvertimeRequestHandler struct {
	service service.OvertimeRequestService
}

func NewOvertimeRequestHandler(service service.OvertimeRequestService) *OvertimeRequestHandler {
	return &OvertimeRequestHandler{service: service}
}

// ──── Manager HR ─────────────────────────────────────────────

func (h *OvertimeRequestHandler) ListForManagerHR(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	if status == "" {
		status = "ALL"
	}

	items, err := h.service.ListForManagerHR(c.Request.Context(), status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil data pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil diambil", items))
}

func (h *OvertimeRequestHandler) GetForManagerHR(c *gin.Context) {
	id := c.Param("id")
	item, err := h.service.GetForManagerHR(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Pengajuan lembur tidak ditemukan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil diambil", item))
}

func (h *OvertimeRequestHandler) ApproveByManagerHR(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	item, err := h.service.ApproveByManagerHR(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal menyetujui pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur disetujui", item))
}

func (h *OvertimeRequestHandler) RejectByManagerHR(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req models.RejectOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Alasan penolakan wajib diisi", err.Error()))
		return
	}

	item, err := h.service.RejectByManagerHR(c.Request.Context(), id, userID, req.RejectionReason)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal menolak pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur ditolak", item))
}

// ──── Kepala Departemen ──────────────────────────────────────

func (h *OvertimeRequestHandler) ListForKepalaDepartemen(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	if status == "" {
		status = "ALL"
	}

	userID := c.GetString("userID")
	items, err := h.service.ListForKepalaDepartemen(c.Request.Context(), status, search, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil data pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil diambil", items))
}

func (h *OvertimeRequestHandler) GetForKepalaDepartemen(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	item, err := h.service.GetForKepalaDepartemen(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Pengajuan lembur tidak ditemukan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil diambil", item))
}

func (h *OvertimeRequestHandler) ApproveByKepalaDepartemen(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	item, err := h.service.ApproveByKepalaDepartemen(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal menyetujui pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur disetujui", item))
}

func (h *OvertimeRequestHandler) RejectByKepalaDepartemen(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req models.RejectOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Alasan penolakan wajib diisi", err.Error()))
		return
	}

	item, err := h.service.RejectByKepalaDepartemen(c.Request.Context(), id, userID, req.RejectionReason)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal menolak pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur ditolak", item))
}
