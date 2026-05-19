// pkg/api/handler/payroll_handler.go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type PayrollHandler struct {
	payrollService service.PayrollService
}

func NewPayrollHandler(payrollService service.PayrollService) *PayrollHandler {
	return &PayrollHandler{payrollService: payrollService}
}

func (h *PayrollHandler) GetMyPayroll(c *gin.Context) {
	userIDRaw, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "missing user"))
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "invalid user"))
		return
	}

	now := time.Now()
	monthStr := c.Query("month")
	yearStr := c.Query("year")

	month := int(now.Month())
	year := now.Year()

	var err error
	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "invalid month"))
			return
		}
	}

	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil || year < 1970 {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "invalid year"))
			return
		}
	}

	payroll, err := h.payrollService.GetPayrollForEmployee(c.Request.Context(), userID, month, year)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Not Found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Payroll retrieved successfully", payroll))
}
