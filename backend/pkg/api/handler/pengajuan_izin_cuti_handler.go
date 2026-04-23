package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type PengajuanIzinCutiHandler struct {
	service service.PengajuanIzinCutiService
}

func NewPengajuanIzinCutiHandler(service service.PengajuanIzinCutiService) *PengajuanIzinCutiHandler {
	return &PengajuanIzinCutiHandler{service: service}
}

func (h *PengajuanIzinCutiHandler) ListForManagerHR(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	if status == "" {
		status = "ALL"
	}

	items, err := h.service.ListForManagerHR(c.Request.Context(), status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to get pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan retrieved successfully", items))
}

func (h *PengajuanIzinCutiHandler) GetForManagerHR(c *gin.Context) {
	id := c.Param("id")
	item, err := h.service.GetForManagerHR(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Pengajuan not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan retrieved successfully", item))
}

func (h *PengajuanIzinCutiHandler) ApproveByManagerHR(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	item, err := h.service.ApproveByManagerHR(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to approve pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan approved successfully", item))
}

func (h *PengajuanIzinCutiHandler) RejectByManagerHR(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req models.RejectLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Alasan penolakan wajib diisi", err.Error()))
		return
	}

	item, err := h.service.RejectByManagerHR(c.Request.Context(), id, userID, req.RejectionReason)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to reject pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan rejected successfully", item))
}

func (h *PengajuanIzinCutiHandler) ListForKepalaDepartemen(c *gin.Context) {
	status := c.Query("status")
	search := c.Query("search")
	if status == "" {
		status = "ALL"
	}

	userID := c.GetString("userID")
	items, err := h.service.ListForKepalaDepartemen(c.Request.Context(), status, search, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to get pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan retrieved successfully", items))
}

func (h *PengajuanIzinCutiHandler) GetForKepalaDepartemen(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	item, err := h.service.GetForKepalaDepartemen(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Pengajuan not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan retrieved successfully", item))
}

func (h *PengajuanIzinCutiHandler) ApproveByKepalaDepartemen(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")
	item, err := h.service.ApproveByKepalaDepartemen(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to approve pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan approved successfully", item))
}

func (h *PengajuanIzinCutiHandler) RejectByKepalaDepartemen(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("userID")

	var req models.RejectLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Alasan penolakan wajib diisi", err.Error()))
		return
	}

	item, err := h.service.RejectByKepalaDepartemen(c.Request.Context(), id, userID, req.RejectionReason)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to reject pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan rejected successfully", item))
}
