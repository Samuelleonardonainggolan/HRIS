// pkg/api/handler/auth_handler.go
package handler

import (
    "net/http"

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

// Login godoc
// @Summary      Login user
// @Description  Login with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Login Request"
// @Success      200 {object} models.Response{data=models.LoginResponse}
// @Failure      400 {object} models.Response
// @Failure      401 {object} models.Response
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    var req models.LoginRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", err.Error()))
        return
    }

    result, err := h.authService.Login(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse("Login failed", err.Error()))
        return
    }

    c.JSON(http.StatusOK, models.SuccessResponse("Login successful", result))
}

// Register godoc
// @Summary      Register new user
// @Description  Register new user (Manager HR only)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RegisterRequest true "Register Request"
// @Success      201 {object} models.Response{data=models.User}
// @Failure      400 {object} models.Response
// @Security     BearerAuth
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
    var req models.RegisterRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", err.Error()))
        return
    }

    user, err := h.authService.Register(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse("Registration failed", err.Error()))
        return
    }

    c.JSON(http.StatusCreated, models.SuccessResponse("Registration successful", user))
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get new access token using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RefreshTokenRequest true "Refresh Token Request"
// @Success      200 {object} models.Response{data=models.LoginResponse}
// @Failure      401 {object} models.Response
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    var req models.RefreshTokenRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request", err.Error()))
        return
    }

    result, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, models.ErrorResponse("Token refresh failed", err.Error()))
        return
    }

    c.JSON(http.StatusOK, models.SuccessResponse("Token refreshed successfully", result))
}

// Logout godoc
// @Summary      Logout user
// @Description  Logout current user
// @Tags         auth
// @Produce      json
// @Success      200 {object} models.Response
// @Security     BearerAuth
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
    userID, _ := c.Get("user_id")

    err := h.authService.Logout(c.Request.Context(), userID.(string))
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse("Logout failed", err.Error()))
        return
    }

    c.JSON(http.StatusOK, models.SuccessResponse("Logout successful", nil))
}