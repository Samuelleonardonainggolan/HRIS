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

// ──── Employee ───────────────────────────────────────────────

func (h *OvertimeRequestHandler) Create(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID not found"))
		return
	}

	var req models.CreateOvertimeRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data pengajuan tidak valid", err.Error()))
		return
	}
	req.UserID = userID

	item, err := h.service.CreateOvertimeRequest(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membuat pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil dibuat", item))
}

func (h *OvertimeRequestHandler) GetMine(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID not found"))
		return
	}

	items, err := h.service.GetMyOvertimeRequests(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengambil data pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil diambil", items))
}

func (h *OvertimeRequestHandler) GetMineByID(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID not found"))
		return
	}

	id := c.Param("id")
	item, err := h.service.GetMyOvertimeRequestByID(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Pengajuan lembur tidak ditemukan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil diambil", item))
}

func (h *OvertimeRequestHandler) UpdateMine(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID not found"))
		return
	}

	id := c.Param("id")
	var req models.UpdateOvertimeRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Data pembaruan tidak valid", err.Error()))
		return
	}

	item, err := h.service.UpdateMyOvertimeRequest(c.Request.Context(), userID, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal memperbarui pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil diperbarui", item))
}

func (h *OvertimeRequestHandler) DeleteMine(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID not found"))
		return
	}

	id := c.Param("id")
	err := h.service.DeleteMyOvertimeRequest(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membatalkan pengajuan lembur", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan lembur berhasil dibatalkan", nil))
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
