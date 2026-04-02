package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type PengajuanHandler struct {
	service service.PengajuanService
}

func NewPengajuanHandler(service service.PengajuanService) *PengajuanHandler {
	return &PengajuanHandler{service: service}
}

func (h *PengajuanHandler) GetTipePengajuan(c *gin.Context) {
	items, err := h.service.GetTipePengajuan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to get tipe pengajuan", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Tipe pengajuan retrieved successfully", items))
}

func (h *PengajuanHandler) GetMyPengajuan(c *gin.Context) {
	userID := c.GetString("userID")
	items, err := h.service.GetMyPengajuan(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to get pengajuan", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Pengajuan retrieved successfully", items))
}

type createPengajuanRequest struct {
	TipePengajuanID string `json:"tipe_pengajuan_id" binding:"required"`
	TanggalMulai    string `json:"tanggal_mulai" binding:"required"`
	TanggalSelesai  string `json:"tanggal_selesai" binding:"required"`
	TotalHari       int    `json:"total_hari" binding:"required"`
	Alasan          string `json:"alasan" binding:"required"`
	DokumenURL      string `json:"dokumen_url,omitempty"`
}

func (h *PengajuanHandler) CreatePengajuan(c *gin.Context) {
	var req createPengajuanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", err.Error()))
		return
	}

	userID := c.GetString("userID")
	created, err := h.service.CreatePengajuan(c.Request.Context(), userID, service.CreatePengajuanRequest{
		TipePengajuanID: req.TipePengajuanID,
		TanggalMulai:    req.TanggalMulai,
		TanggalSelesai:  req.TanggalSelesai,
		TotalHari:       req.TotalHari,
		Alasan:          req.Alasan,
		DokumenURL:      req.DokumenURL,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to create pengajuan", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Pengajuan created successfully", created))
}

