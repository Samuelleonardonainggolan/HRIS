// pkg/models/auth.go
package models

// LoginRequest - Request body for login
type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// LoginResponse - Response after successful login
type LoginResponse struct {
    Success bool         `json:"success"`
    Message string       `json:"message"`
    Data    LoginData    `json:"data"`
}

// LoginData - Login data containing user and tokens
type LoginData struct {
    User         UserResponse `json:"user"`
    AccessToken  string       `json:"access_token"`
    RefreshToken string       `json:"refresh_token"`
    ExpiresIn    int64        `json:"expires_in"`
}

// RegisterRequest - Request body for registration
type RegisterRequest struct {
    NIK          string `json:"nik" binding:"required"`
    Email        string `json:"email" binding:"required,email"`
    Password     string `json:"password" binding:"required,min=8"`
    FullName     string `json:"full_name" binding:"required"`
    Role         string `json:"role" binding:"required"`
    DepartmentID string `json:"department_id" binding:"required"`
    PositionID   string `json:"position_id" binding:"required"`
    Phone        string `json:"phone,omitempty"`
    Address      string `json:"address,omitempty"`
}

// RefreshTokenRequest - Request body for refresh token
type RefreshTokenRequest struct {
    RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse - Response after refresh token
type RefreshTokenResponse struct {
    Success bool              `json:"success"`
    Message string            `json:"message"`
    Data    RefreshTokenData  `json:"data"`
}

// RefreshTokenData - Refresh token data
type RefreshTokenData struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}