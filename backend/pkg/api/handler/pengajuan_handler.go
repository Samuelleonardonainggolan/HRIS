// pkg/api/handler/pengajuan_handler.go
package handler

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

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
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		totalHari, err := strconv.Atoi(strings.TrimSpace(formValue(c, "total_hari", "days_total")))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "total_hari wajib berupa angka",
			})
			return
		}

		req = service.CreatePengajuanRequest{
			UserID:          userID.(string),
			TipePengajuanID: formValue(c, "tipe_pengajuan_id", "request_type_id"),
			TanggalMulai:    formValue(c, "tanggal_mulai", "start_date"),
			TanggalSelesai:  formValue(c, "tanggal_selesai", "end_date"),
			TotalHari:       totalHari,
			Alasan:          formValue(c, "alasan", "reason"),
			DokumenURL:      formValue(c, "dokumen_url", "document_url"),
		}

		if fileHeader, err := c.FormFile("document"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "gagal membuka file dokumen: " + err.Error(),
				})
				return
			}
			defer file.Close()

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "gagal membaca file dokumen: " + err.Error(),
				})
				return
			}

			docURL, err := h.pengajuanService.UploadDocument(c.Request.Context(), fileBytes, filepath.Base(fileHeader.Filename))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "gagal upload dokumen: " + err.Error(),
				})
				return
			}
			req.DokumenURL = docURL
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Request tidak valid: " + err.Error(),
			})
			return
		}
		req.UserID = userID.(string)
	}

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
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		if value := strings.TrimSpace(formValue(c, "tipe_pengajuan_id", "request_type_id")); value != "" {
			req.TipePengajuanID = &value
		}
		if value := strings.TrimSpace(formValue(c, "tanggal_mulai", "start_date")); value != "" {
			req.TanggalMulai = &value
		}
		if value := strings.TrimSpace(formValue(c, "tanggal_selesai", "end_date")); value != "" {
			req.TanggalSelesai = &value
		}
		if value := strings.TrimSpace(formValue(c, "alasan", "reason")); value != "" {
			req.Alasan = &value
		}
		if value := strings.TrimSpace(formValue(c, "dokumen_url", "document_url")); value != "" {
			req.DokumenURL = &value
		}
		if value := strings.TrimSpace(formValue(c, "total_hari", "days_total")); value != "" {
			totalHari, err := strconv.Atoi(value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "total_hari wajib berupa angka",
				})
				return
			}
			req.TotalHari = &totalHari
		}

		if fileHeader, err := c.FormFile("document"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "gagal membuka file dokumen: " + err.Error(),
				})
				return
			}
			defer file.Close()

			fileBytes, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "gagal membaca file dokumen: " + err.Error(),
				})
				return
			}

			docURL, err := h.pengajuanService.UploadDocument(c.Request.Context(), fileBytes, filepath.Base(fileHeader.Filename))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "error",
					"message": "gagal upload dokumen: " + err.Error(),
				})
				return
			}
			req.DokumenURL = &docURL
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Request tidak valid: " + err.Error(),
			})
			return
		}
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

func formValue(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.PostForm(key)); value != "" {
			return value
		}
	}

	return ""
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
