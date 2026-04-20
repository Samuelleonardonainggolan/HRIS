// pkg/api/handler/pengajuan_handler.go
package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// PengajuanHandler menangani request yang berkaitan dengan pengajuan izin/cuti.
type PengajuanHandler struct {
	pengajuanService service.PengajuanService
}

func NewPengajuanHandler(pengajuanService service.PengajuanService) *PengajuanHandler {
	return &PengajuanHandler{
		pengajuanService: pengajuanService,
	}
}

// GetTipePengajuan - GET /api/v1/pengajuan/tipe
// Mengembalikan semua tipe pengajuan yang tersedia (Izin Sakit, Cuti Tahunan, dll.)
// ✅ Endpoint ini sebelumnya 404 karena belum didaftarkan di router.
func (h *PengajuanHandler) GetTipePengajuan(c *gin.Context) {
	tipes, err := h.pengajuanService.GetAllTipePengajuan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Gagal mengambil tipe pengajuan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   tipes,
	})
}

// CreatePengajuan - POST /api/v1/pengajuan
// Menerima pengajuan izin/cuti baru dari karyawan.
func (h *PengajuanHandler) CreatePengajuan(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	var req service.CreatePengajuanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Request tidak valid: " + err.Error(),
		})
		return
	}

	req.UserID = userID.(string)

	result, err := h.pengajuanService.CreatePengajuan(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Pengajuan berhasil dikirimkan",
		"data":    result,
	})
}

// GetMyPengajuan - GET /api/v1/pengajuan
// Mengembalikan daftar pengajuan milik user yang sedang login.
func (h *PengajuanHandler) GetMyPengajuan(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	list, err := h.pengajuanService.GetPengajuanByUser(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   list,
	})
}

// UpdatePengajuan - PUT /api/v1/pengajuan/:id
func (h *PengajuanHandler) UpdatePengajuan(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	id := c.Param("id")
	var req service.UpdatePengajuanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Request tidak valid: " + err.Error(),
		})
		return
	}

	result, err := h.pengajuanService.UpdatePengajuan(c.Request.Context(), userID.(string), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pengajuan berhasil diperbarui",
		"data":    result,
	})
}

// CancelPengajuan - DELETE /api/v1/pengajuan/:id
func (h *PengajuanHandler) CancelPengajuan(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Unauthorized",
		})
		return
	}

	id := c.Param("id")
	err := h.pengajuanService.CancelPengajuan(c.Request.Context(), userID.(string), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pengajuan berhasil dibatalkan",
	})
}
