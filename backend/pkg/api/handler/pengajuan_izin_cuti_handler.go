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
		status = models.StatusPending
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
	item, err := h.service.RejectByManagerHR(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to reject pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan rejected successfully", item))
}
