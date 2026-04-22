// pkg/api/handler/user_handler.go
package handler

import (
	"net/http"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateEmployee(c *gin.Context) {
	var req models.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	employee, tempPassword, err := h.userService.CreateEmployee(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	if tempPassword != nil {
		c.JSON(http.StatusCreated, models.SuccessResponse("Employee created successfully", gin.H{
			"employee":           employee,
			"temporary_password": *tempPassword,
		}))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse("Employee created successfully", employee))
}

// GetAllEmployees - Get all employees
func (h *UserHandler) GetAllEmployees(c *gin.Context) {
	userRoleRaw, roleExists := c.Get("userRole")
	userIDRaw, userExists := c.Get("userID")
	if roleExists && userExists {
		userRole, roleOk := userRoleRaw.(string)
		userID, userOk := userIDRaw.(string)
		if roleOk && userOk && userID != "" {
			if userRole == models.RoleManagerDepartemen || userRole == models.RoleAdminDepartemen {
				employees, err := h.userService.GetEmployeesMyDepartment(c.Request.Context(), userID)
				if err != nil {
					c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
					return
				}

				c.JSON(http.StatusOK, models.SuccessResponse("Employees retrieved successfully", employees))
				return
			}
		}
	}

	employees, err := h.userService.GetAllEmployees(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employees retrieved successfully", employees))
}

func (h *UserHandler) GetEmployeesMyDepartment(c *gin.Context) {
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

	employees, err := h.userService.GetEmployeesMyDepartment(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employees retrieved successfully", employees))
}

// GetEmployeeByID - Get employee by ID
func (h *UserHandler) GetEmployeeByID(c *gin.Context) {
	id := c.Param("id")
	employee, err := h.userService.GetEmployeeByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("Not Found", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee retrieved successfully", employee))
}

// UpdateEmployee - Update employee data
func (h *UserHandler) UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	userRoleRaw, roleExists := c.Get("userRole")
	userIDRaw, userExists := c.Get("userID")

	var employee *models.UserResponse
	var err error

	if roleExists && userExists {
		role, _ := userRoleRaw.(string)
		userID, _ := userIDRaw.(string)
		if role == models.RoleManagerDepartemen || role == models.RoleAdminDepartemen {
			employee, err = h.userService.UpdateEmployeeByManagerDepartemen(c.Request.Context(), userID, id, &req)
		} else {
			employee, err = h.userService.UpdateEmployee(c.Request.Context(), id, &req)
		}
	} else {
		employee, err = h.userService.UpdateEmployee(c.Request.Context(), id, &req)
	}

	if err != nil {
		if err.Error() == "akses ditolak" {
			c.JSON(http.StatusForbidden, models.ErrorResponse("Forbidden", err.Error()))
			return
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee updated successfully", employee))
}

// DeleteEmployee - Delete employee
func (h *UserHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	err := h.userService.DeleteEmployee(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employee deleted successfully", nil))
}

func (h *UserHandler) ImportEmployees(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "File is required"))
		return
	}
	defer file.Close()

	f, err := excelize.OpenReader(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "Invalid Excel file"))
		return
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		// Try to get the first sheet name if "Sheet1" doesn't exist
		sheetList := f.GetSheetList()
		if len(sheetList) > 0 {
			rows, err = f.GetRows(sheetList[0])
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "Failed to read rows"))
			return
		}
	}

	var employees []models.CreateEmployeeRequest
	// Skip header row
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 2 { // Skip empty rows
			continue
		}

		// Helper to safe get
		get := func(idx int) string {
			if idx < len(row) {
				return row[idx]
			}
			return ""
		}

		emp := models.CreateEmployeeRequest{
			PayrollNumber:    get(0),
			Email:            get(1),
			Password:         get(2),
			FullName:         get(3),
			BirthDate:        get(4),
			Religion:         get(5),
			LastEducation:    get(6),
			YearEnrolled:     get(7),
			EmploymentStatus: get(8),
			DepartmentID:     get(9),
			PositionID:       get(10),
			Phone:            get(11),
			Address:          get(12),
			Role:             get(13),
		}
		employees = append(employees, emp)
	}

	created, failures, err := h.userService.ImportEmployees(c.Request.Context(), employees)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Employees imported successfully", gin.H{
		"created": created,
		"failed":  len(failures),
		"errors":  failures,
		"total":   len(employees),
	}))
}

func (h *UserHandler) DownloadEmployeeTemplate(c *gin.Context) {
	f := excelize.NewFile()
	sheetName := "Sheet1"
	index, _ := f.NewSheet(sheetName)
	f.SetActiveSheet(index)

	headers := []string{
		"payroll_number",
		"email",
		"password",
		"full_name",
		"birth_date",
		"religion",
		"last_education",
		"year_enrolled",
		"employment_status",
		"department_id",
		"position_id",
		"phone",
		"address",
		"role",
	}

	// Set headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Set example data
	example := []string{
		"PAY001",
		"user@company.com",
		"password123",
		"Nama Pegawai",
		"2000-01-31",
		"Islam",
		"S1",
		"2023",
		"Tetap",
		"<department_id>",
		"<position_id>",
		"+6281234567890",
		"Alamat lengkap",
		models.RoleStaf,
	}

	for i, val := range example {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheetName, cell, val)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=employee_template.xlsx")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", "Failed to generate template"))
	}
}
