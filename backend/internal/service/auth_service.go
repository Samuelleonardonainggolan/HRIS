// internal/service/auth_service.go
package service

import (
    "context"
    "errors"
    "time"

    "github.com/andikatampubolon10/hris-backend/pkg/auth"
    "github.com/andikatampubolon10/hris-backend/pkg/database/repository"
    "github.com/andikatampubolon10/hris-backend/pkg/models"
)

type AuthService interface {
    Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
    Register(ctx context.Context, req models.RegisterRequest) (*models.User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error)
    Logout(ctx context.Context, userID string) error
}

type authService struct {
    userRepo  repository.UserRepository
    jwtSecret string
    jwtExpiry string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret, jwtExpiry string) AuthService {
    return &authService{
        userRepo:  userRepo,
        jwtSecret: jwtSecret,
        jwtExpiry: jwtExpiry,
    }
}

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
    user, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err != nil || user == nil {
        return nil, errors.New("invalid email or password")
    }

    if !auth.CheckPasswordHash(req.Password, user.Password) {
        return nil, errors.New("invalid email or password")
    }

    if !user.IsActive {
        return nil, errors.New("user account is inactive")
    }

    accessToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.Department, s.jwtSecret, s.jwtExpiry)
    if err != nil {
        return nil, errors.New("failed to generate access token")
    }

    refreshToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.Department, s.jwtSecret, "168h")
    if err != nil {
        return nil, errors.New("failed to generate refresh token")
    }

    expiresIn := time.Now().Add(24 * time.Hour).Unix()

    user.Password = ""
    return &models.LoginResponse{
        User:         user,
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    expiresIn,
    }, nil
}

func (s *authService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
    existingUser, _ := s.userRepo.FindByEmail(ctx, req.Email)
    if existingUser != nil {
        return nil, errors.New("email already registered")
    }

    hashedPassword, err := auth.HashPassword(req.Password)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }

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
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
    claims, err := auth.ValidateToken(refreshToken, s.jwtSecret)
    if err != nil {
        return nil, errors.New("invalid refresh token")
    }

    user, err := s.userRepo.FindByID(ctx, claims.UserID)
    if err != nil || user == nil {
        return nil, errors.New("user not found")
    }

    if !user.IsActive {
        return nil, errors.New("user account is inactive")
    }

    newAccessToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.Department, s.jwtSecret, s.jwtExpiry)
    if err != nil {
        return nil, errors.New("failed to generate access token")
    }

    newRefreshToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.Department, s.jwtSecret, "168h")
    if err != nil {
        return nil, errors.New("failed to generate refresh token")
    }

    expiresIn := time.Now().Add(24 * time.Hour).Unix()

    user.Password = ""
    return &models.LoginResponse{
        User:         user,
        AccessToken:  newAccessToken,
        RefreshToken: newRefreshToken,
        ExpiresIn:    expiresIn,
    }, nil
}

func (s *authService) Logout(ctx context.Context, userID string) error {
    return nil
}
