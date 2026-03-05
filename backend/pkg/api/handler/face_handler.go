package handler

import (
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
	userID := c.Param("id")
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
	if err := h.faceService.ExtractAndSaveEmbedding(c.Request.Context(), userID, bytes, filename); err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse("Failed to extract embedding", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Face embedding registered", gin.H{"user_id": userID}))
}

func (h *FaceHandler) ProcessAttendance(c *gin.Context) {
	userIDRaw, _ := c.Get("user_id")
	userID, _ := userIDRaw.(string)

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

	latStr := c.PostForm("latitude")
	lngStr := c.PostForm("longitude")
	recordType := c.PostForm("record_type")
	if recordType == "" {
		recordType = "checkin"
	}
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid latitude", err.Error()))
		return
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid longitude", err.Error()))
		return
	}
	filename := filepath.Base(file.Filename)
	if filename == "" {
		filename = "selfie.jpg"
	}
	res, err := h.faceService.ProcessAttendance(c.Request.Context(), userID, lat, lng, recordType, bytes, filename)
	if err != nil {
		c.JSON(http.StatusBadGateway, models.ErrorResponse("Attendance processing failed", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Attendance processed", res))
}

