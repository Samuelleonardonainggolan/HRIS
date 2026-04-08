package handler

import (
	"encoding/json"
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type PositionHandler struct {
	positionService service.PositionService
}

func NewPositionHandler(positionService service.PositionService) *PositionHandler {
	return &PositionHandler{positionService: positionService}
}

func (h *PositionHandler) GetAllPositions(c *gin.Context) {
	departmentID := c.Query("department_id")

	positions, err := h.positionService.GetAllPositions(c.Request.Context(), departmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Positions retrieved successfully", positions))
}

func (h *PositionHandler) GetPositionByID(c *gin.Context) {
	id := c.Param("id")

	position, err := h.positionService.GetPositionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Not Found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Position retrieved successfully", position))
}

func (h *PositionHandler) UpdatePosition(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdatePositionRequest
	raw, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	if req.IsActive == nil {
		var alt struct {
			IsActive *bool `json:"isActive"`
		}
		if err := json.Unmarshal(raw, &alt); err == nil && alt.IsActive != nil {
			req.IsActive = alt.IsActive
		}
	}

	position, err := h.positionService.UpdatePosition(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to update position", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Position updated successfully", position))
}
