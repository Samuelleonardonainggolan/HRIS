// internal/service/auth_service.go
package service

import (
    "context"
    "errors"
    "time"

    "github.com/andikatampubolon10/hris-backend/internal/config"
    "github.com/andikatampubolon10/hris-backend/pkg/auth"
    "github.com/andikatampubolon10/hris-backend/pkg/database/repository"
    "github.com/andikatampubolon10/hris-backend/pkg/models"
)

type AuthService interface {
    Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
    Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error)
    Logout(ctx context.Context, userID string) error
}

type authService struct {
    userRepo repository.UserRepository
    config   *config.Config
}

func NewAuthService(userRepo repository.UserRepository, config *config.Config) AuthService {
    return &authService{
        userRepo: userRepo,
        config:   config,
    }
}

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
    // Find user by email
    user, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }

    if user == nil {
        return nil, errors.New("invalid email or password")
    }

    // Check if user is active
    if !user.IsActive {
        return nil, errors.New("account is inactive, please contact administrator")
    }

    // Verify password
    if !auth.CheckPasswordHash(req.Password, user.Password) {
        return nil, errors.New("invalid email or password")
    }

    // Generate JWT tokens
    accessToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.Department,
        s.config.JWTSecret,
        s.config.JWTExpiry,
    )
    if err != nil {
        return nil, errors.New("failed to generate access token")
    }

    refreshToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.Department,
        s.config.JWTSecret,
        s.config.RefreshExpiry,
    )
    if err != nil {
        return nil, errors.New("failed to generate refresh token")
    }

    // Calculate expiry
    duration, _ := time.ParseDuration(s.config.JWTExpiry)
    expiresIn := time.Now().Add(duration).Unix()

    // Remove password from response
    user.Password = ""

    return &models.LoginResponse{
        User:         user,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    expiresIn,
    }, nil
}

func (s *authService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
    // Check if user exists
    existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        return nil, err
    }

    if existingUser != nil {
        return nil, errors.New("email already registered")
    }

    // Check if NIK exists
    existingNIK, err := s.userRepo.FindByNIK(ctx, req.NIK)
    if err != nil {
        return nil, err
    }

    if existingNIK != nil {
        return nil, errors.New("NIK already registered")
    }

    // Hash password
    hashedPassword, err := auth.HashPassword(req.Password)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }

    // Create user
    user := &models.User{
        NIK:        req.NIK,
        Email:      req.Email,
        Password:   hashedPassword,
        FullName:   req.FullName,
        Role:       req.Role,
        Department: req.Department,
        Position:   req.Position,
        Phone:      req.Phone,
        Address:    req.Address,
        JoinDate:   time.Now(),
        IsActive:   true,
    }

    err = s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, err
    }

    // Remove password from response
    user.Password = ""

    return user, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
    // Validate refresh token
    claims, err := auth.ValidateToken(refreshToken, s.config.JWTSecret)
    if err != nil {
        return nil, errors.New("invalid refresh token")
    }

    // Get user
    user, err := s.userRepo.FindByID(ctx, claims.UserID)
    if err != nil || user == nil {
        return nil, errors.New("user not found")
    }

    // Check if user is active
    if !user.IsActive {
        return nil, errors.New("account is inactive")
    }

    // Generate new tokens
    accessToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.Department,
        s.config.JWTSecret,
        s.config.JWTExpiry,
    )
    if err != nil {
        return nil, errors.New("failed to generate access token")
    }

    newRefreshToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.Department,
        s.config.JWTSecret,
        s.config.RefreshExpiry,
    )
    if err != nil {
        return nil, errors.New("failed to generate refresh token")
    }

    duration, _ := time.ParseDuration(s.config.JWTExpiry)
    expiresIn := time.Now().Add(duration).Unix()

    // Remove password from response
    user.Password = ""

    return &models.LoginResponse{
        User:         user,
        AccessToken:  accessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    expiresIn,
    }, nil
}

func (s *authService) Logout(ctx context.Context, userID string) error {
    // In JWT, logout is typically handled client-side by removing the token
    // But you can implement token blacklist here if needed
    // For now, we just return success
    return nil
}