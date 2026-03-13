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
	// Ambil userID dari form field, bukan dari path parameter
	userID := c.PostForm("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", "user_id is required"))
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", "photo is required"))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid photo", err.Error()))
		return
	}
	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid photo", err.Error()))
		return
	}

	filename := filepath.Base(file.Filename)
	if filename == "" {
		filename = "photo.jpg"
	}

	// Get face_embedding from form (optional)
	faceEmbeddingStr := c.PostForm("face_embedding")
	var faceEmbedding []float32
	if faceEmbeddingStr != "" {
		json.Unmarshal([]byte(faceEmbeddingStr), &faceEmbedding)
	}

	// Panggil service untuk extract dan save embedding
	if err := h.faceService.ExtractAndSaveEmbedding(c.Request.Context(), userID, bytes, filename); err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse("Failed to extract embedding", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Face embedding registered", gin.H{"user_id": userID}))
}

// pkg/api/handler/face_handler.go

func (h *FaceHandler) ProcessAttendance(c *gin.Context) {

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "user id tidak di temukan. silahkan login ulang"))
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Error", "Format user ID tidak valid"))
		return
	}

	// Ambil file foto
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", "photo is required"))
		return
	}

	f, _ := file.Open()
	defer f.Close()
	photoBytes, _ := io.ReadAll(f)

	// Ambil data form lainnya
	latStr := c.PostForm("latitude")
	lngStr := c.PostForm("longitude")
	recordType := c.PostForm("record_type") // 'checkin' atau 'checkout'

	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	// Panggil Service
	// Di dalam service ini, pastikan memanggil faceEmbeddingRepo.FindByUserID(userID)
	// Agar verifikasi wajah hanya membandingkan dengan wajah milik user tersebut
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("gagal melakukan absensi", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Absensi berhasil", result))
}

func (h *FaceHandler) ExtractEmbedding(c *gin.Context) {
	// Get user ID from context
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "user id tidak ditemukan"))
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Error", "Format user ID tidak valid"))
		return
	}

	// Get photo from request
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", "photo is required"))
		return
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to open file", err.Error()))
		return
	}
	defer src.Close()

	// Read file bytes
	photoBytes := make([]byte, file.Size)
	_, err = src.Read(photoBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to read file", err.Error()))
		return
	}

	// Extract embedding using face client
	embedding, err := h.faceService.ExtractEmbeddingOnly(
		c.Request.Context(),
		userID,
		photoBytes,
		file.Filename,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to extract embedding", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Embedding extracted successfully", gin.H{
		"embedding": embedding,
		"dimension": len(embedding),
	}))
}
