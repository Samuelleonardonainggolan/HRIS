package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type EmployeeBasicSalaryHandler struct {
	svc service.EmployeeBasicSalaryService
}

func NewEmployeeBasicSalaryHandler(svc service.EmployeeBasicSalaryService) *EmployeeBasicSalaryHandler {
	return &EmployeeBasicSalaryHandler{svc: svc}
}

// GET /employee-basic-salaries?q=&department=&active=true|false
func (h *EmployeeBasicSalaryHandler) List(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to list salaries", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee basic salaries retrieved successfully", items))
}

// POST /employee-basic-salaries
func (h *EmployeeBasicSalaryHandler) Create(c *gin.Context) {
	var req models.CreateEmployeeBasicSalaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	created, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to create salary", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Employee basic salary created successfully", created))
}

// GET /employee-basic-salaries/users/:userId/latest
func (h *EmployeeBasicSalaryHandler) GetLatestByUser(c *gin.Context) {
	userID := c.Param("userId")

	salary, err := h.svc.GetLatestByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Salary not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee basic salary retrieved successfully", salary))
}

// PATCH /employee-basic-salaries/:id
func (h *EmployeeBasicSalaryHandler) UpdateBySalaryID(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateEmployeeBasicSalaryRequest
	raw, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	updated, err := h.svc.UpdateBySalaryID(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to update salary", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee basic salary updated successfully", updated))
}

// GET /employee-basic-salaries/users/:userId/active
func (h *EmployeeBasicSalaryHandler) GetActiveByUser(c *gin.Context) {
	userID := c.Param("userId")

	salary, err := h.svc.GetActiveByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Salary not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee basic salary retrieved successfully", salary))
}

// PATCH /employee-basic-salaries/users/:userId/active
func (h *EmployeeBasicSalaryHandler) UpdateActiveByUser(c *gin.Context) {
	userID := c.Param("userId")

	var req models.UpdateEmployeeBasicSalaryRequest
	raw, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	updated, err := h.svc.UpdateActiveByUserID(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to update salary", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee basic salary updated successfully", updated))
}

// POST /employee-basic-salaries/users/:userId/deactivate
func (h *EmployeeBasicSalaryHandler) DeactivateActiveByUser(c *gin.Context) {
	userID := c.Param("userId")

	if err := h.svc.DeactivateActiveByUserID(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to deactivate salary", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee basic salary deactivated successfully", nil))
}

// pkg/api/handler/employee_basic_salary_handler.go
func (h *EmployeeBasicSalaryHandler) AvailableEmployees(c *gin.Context) {
  q := c.Query("q")
  items, err := h.svc.ListAvailableEmployees(c.Request.Context(), q)
  if err != nil {
    c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to get employees", err.Error()))
    return
  }
  c.JSON(http.StatusOK, models.SuccessResponse("Available employees retrieved successfully", items))
}

