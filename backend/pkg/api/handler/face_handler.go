package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type FaceHandler struct {
	faceService service.FaceService
}

func NewFaceHandler(svc service.FaceService) *FaceHandler {
	return &FaceHandler{faceService: svc}
}

func (h *FaceHandler) Health(c *gin.Context) {
	ok, err := h.faceService.Health(c.Request.Context())
	if err != nil || !ok {
		c.JSON(http.StatusBadGateway, models.ErrorResponse("Face service unhealthy", "unreachable"))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Face service healthy", gin.H{"ok": ok}))
}

func (h *FaceHandler) RegisterFace(c *gin.Context) {
	userID := c.PostForm("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Permintaan tidak valid", "user_id diperlukan"))
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Permintaan tidak valid", "foto diperlukan"))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Foto tidak valid", err.Error()))
		return
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Gagal membaca foto", err.Error()))
		return
	}

	filename := filepath.Base(file.Filename)
	if filename == "" {
		filename = "photo.jpg"
	}

	// Optional embedding
	faceEmbeddingStr := c.PostForm("face_embedding")
	if faceEmbeddingStr != "" {
		var faceEmbedding []float32
		if err := json.Unmarshal([]byte(faceEmbeddingStr), &faceEmbedding); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Format embedding tidak valid", err.Error()))
			return
		}
	}

	if err := h.faceService.ExtractAndSaveEmbedding(c.Request.Context(), userID, bytes, filename); err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse("Gagal mengekstrak embedding", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Embedding wajah berhasil didaftarkan", gin.H{
		"user_id": userID,
	}))
}

func (h *FaceHandler) ProcessAttendance(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Tidak terotorisasi", "user id tidak ditemukan. silahkan login ulang"))
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Galat", "Format user ID tidak valid"))
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Permintaan tidak valid", "foto diperlukan"))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membuka file", err.Error()))
		return
	}
	defer f.Close()

	photoBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membaca file", err.Error()))
		return
	}

	lat, _ := strconv.ParseFloat(c.PostForm("latitude"), 64)
	lng, _ := strconv.ParseFloat(c.PostForm("longitude"), 64)
	recordType := c.PostForm("record_type")

	result, err := h.faceService.ProcessAttendance(
		c.Request.Context(),
		userID,
		lat,
		lng,
		recordType,
		photoBytes,
		file.Filename,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal melakukan absensi", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Absensi berhasil", result))
}

func (h *FaceHandler) ExtractEmbedding(c *gin.Context) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Tidak terotorisasi", "user id tidak ditemukan"))
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Galat", "Format user ID tidak valid"))
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Permintaan tidak valid", "foto diperlukan"))
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membuka file", err.Error()))
		return
	}
	defer src.Close()

	photoBytes, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal membaca file", err.Error()))
		return
	}

	embedding, err := h.faceService.ExtractEmbeddingOnly(
		c.Request.Context(),
		userID,
		photoBytes,
		file.Filename,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Gagal mengekstrak embedding", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Embedding berhasil diekstrak", gin.H{
		"embedding": embedding,
		"dimension": len(embedding),
	}))
}
