// pkg/api/handler/department_handler.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type DepartmentHandler struct {
	departmentService service.DepartmentService
}

func NewDepartmentHandler(departmentService service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: departmentService,
	}
}

// CreateDepartment creates a new department
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req models.CreateDepartmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	department, err := h.departmentService.CreateDepartment(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to create department", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Department created successfully", department))
}

// GetDepartmentByID gets department by ID
func (h *DepartmentHandler) GetDepartmentByID(c *gin.Context) {
	id := c.Param("id")

	department, err := h.departmentService.GetDepartmentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Department not found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Department retrieved successfully", department))
}

// GetAllDepartments gets all departments
func (h *DepartmentHandler) GetAllDepartments(c *gin.Context) {
	departments, err := h.departmentService.GetAllDepartments(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to get departments", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Departments retrieved successfully", departments))
}

// UpdateDepartment updates department by ID
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateDepartmentRequest
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

	department, err := h.departmentService.UpdateDepartment(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to update department", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Department updated successfully", department))
}

// DeleteDepartment deletes department by ID
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id := c.Param("id")

	err := h.departmentService.DeleteDepartment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to delete department", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Department deleted successfully", nil))
}
