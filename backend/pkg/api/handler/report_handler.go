package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService service.ReportService
}

func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) GetAttendanceActivityReport(c *gin.Context) {
	period := c.Query("period")
	departmentID := c.Query("department_id")
	eventType := c.Query("type")
	approvalStatus := c.Query("status")
	search := c.Query("search")

	report, err := h.reportService.GetAttendanceActivityReport(c.Request.Context(), period, departmentID, eventType, approvalStatus, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Attendance activity report retrieved successfully", report))
}
