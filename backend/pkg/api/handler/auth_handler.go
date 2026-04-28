// pkg/api/handler/auth_handler.go
package handler

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andikatampubolon10/hris-backend/internal/service"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login - User login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	response, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	// CEK STATUS FACE REGISTRATION
	requiresFaceRegistration := true
	faceStatus, _ := h.authService.GetFaceRegistrationStatus(c.Request.Context(), response.User.ID)
	if faceStatus {
		requiresFaceRegistration = false
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"user":                       response.User,
			"access_token":               response.AccessToken,
			"refresh_token":              response.RefreshToken,
			"expires_in":                 response.ExpiresIn,
			"requires_face_registration": requiresFaceRegistration,
		},
	})
}
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	// ✅ Register already returns *UserResponse, not *User
	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	// ✅ user is already UserResponse, no need to call ToResponse()
	c.JSON(http.StatusCreated, models.SuccessResponse("User registered successfully", user))
}

// RefreshToken - Refresh access token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout - User logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User not found"))
		return
	}
	err := h.authService.Logout(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Logout successful", gin.H{"user_id": userID}))
}

// GetProfile - GET /api/v1/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID tidak ditemukan"))
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "Format user ID tidak valid"))
		return
	}
	profile, err := h.authService.GetProfile(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Profile retrieved", profile))
}

// UpdateProfile - PUT /api/v1/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID tidak ditemukan"))
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "Format user ID tidak valid"))
		return
	}
	var req models.UpdateUserRequest
	if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
		req.Phone = strings.TrimSpace(c.PostForm("phone"))
		req.Address = strings.TrimSpace(c.PostForm("address"))
		req.FullName = strings.TrimSpace(c.PostForm("full_name"))
		req.PayrollNumber = strings.TrimSpace(c.PostForm("payroll_number"))
		if birthDate := strings.TrimSpace(c.PostForm("birth_date")); birthDate != "" {
			req.BirthDate = birthDate
		}
		if value := strings.TrimSpace(c.PostForm("is_active")); value != "" {
			if parsed, parseErr := strconv.ParseBool(value); parseErr == nil {
				req.IsActive = &parsed
			}
		}

		if fileHeader, err := c.FormFile("avatar"); err == nil {
			file, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "gagal membuka file avatar: "+err.Error()))
				return
			}
			defer file.Close()

			photoBytes, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "gagal membaca file avatar: "+err.Error()))
				return
			}

			avatarURL, err := h.authService.UploadProfilePhoto(c.Request.Context(), userIDStr, photoBytes, filepath.Base(fileHeader.Filename))
			if err != nil {
				c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", "gagal upload avatar: "+err.Error()))
				return
			}
			req.Avatar = avatarURL
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
			return
		}
	}
	profile, err := h.authService.UpdateProfile(c.Request.Context(), userIDStr, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Internal Server Error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Profile updated successfully", profile))
}

// ChangePassword - POST /api/v1/profile/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "User ID tidak ditemukan"))
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "Format user ID tidak valid"))
		return
	}
	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}
	if err := h.authService.ChangePassword(c.Request.Context(), userIDStr, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Bad Request", err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("Password berhasil diubah", nil))
}
