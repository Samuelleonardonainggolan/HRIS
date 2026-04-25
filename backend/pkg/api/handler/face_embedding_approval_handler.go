// pkg/api/handler/face_embedding_approval_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type FaceEmbeddingApprovalHandler struct {
	svc service.FaceEmbeddingApprovalService
}

func NewFaceEmbeddingApprovalHandler(svc service.FaceEmbeddingApprovalService) *FaceEmbeddingApprovalHandler {
	return &FaceEmbeddingApprovalHandler{svc: svc}
}

// GET /face-embeddings/approval?q=&department=&active=true|false
func (h *FaceEmbeddingApprovalHandler) List(c *gin.Context) {
	q := c.Query("q")
	dept := c.Query("department")

	var activePtr *bool
	if v := c.Query("active"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "active harus boolean (true/false)"))
			return
		}
		activePtr = &b
	}

	items, err := h.svc.List(c.Request.Context(), q, dept, activePtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to list face embeddings", err.Error()))
			return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Face embedding approval list retrieved successfully", items))
}

// GET /face-embeddings/approval/:id
func (h *FaceEmbeddingApprovalHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	item, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Face embedding not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Face embedding retrieved successfully", item))
}