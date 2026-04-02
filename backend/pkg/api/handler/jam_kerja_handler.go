// pkg/api/handler/work_schedule_handler.go
package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type WorkScheduleHandler struct {
	workScheduleService service.WorkScheduleService
}

func NewWorkScheduleHandler(workScheduleService service.WorkScheduleService) *WorkScheduleHandler {
	return &WorkScheduleHandler{workScheduleService: workScheduleService}
}

func (h *WorkScheduleHandler) ListForManagerDepartment(c *gin.Context) {
	// sesuaikan dengan AuthMiddleware Anda:
	// kalau middleware menyimpan "user" (struct) di context.
	u, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "user tidak ditemukan"))
		return
	}
	user := u.(*models.User) // sesuaikan jika tipe berbeda

	data, err := h.workScheduleService.ListForManagerDepartment(c.Request.Context(), user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to get work schedules", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Work schedules retrieved successfully", data))
}

func (h *WorkScheduleHandler) UpsertForManagerDepartment(c *gin.Context) {
	targetUserID := c.Param("userId")

	var req models.UpsertWorkScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	u, ok := c.Get("user")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "user tidak ditemukan"))
		return
	}
	user := u.(*models.User)

	data, err := h.workScheduleService.UpsertForManagerDepartment(c.Request.Context(), user.ID.Hex(), targetUserID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to update work schedule", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Work schedule updated successfully", data))
}