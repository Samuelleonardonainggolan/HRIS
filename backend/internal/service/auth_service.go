// internal/service/auth_service.go
package service

import (
    "context"
    "errors"
    "time"

    "github.com/andikatampubolon10/hris-backend/pkg/auth"
    "github.com/andikatampubolon10/hris-backend/pkg/database/repository"
    "github.com/andikatampubolon10/hris-backend/pkg/models"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService interface {
    Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
    Register(ctx context.Context, req models.RegisterRequest) (*models.User, error)
    RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error)
    Logout(ctx context.Context, userID string) error // ← Add this
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
    // Find user by email
    user, err := s.userRepo.FindByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.New("invalid email or password")
    }

    // Verify password
    if !auth.CheckPasswordHash(req.Password, user.Password) {
        return nil, errors.New("invalid email or password")
    }

    // Check if user is active
    if !user.IsActive {
        return nil, errors.New("user account is inactive")
    }

    // Generate tokens
    accessToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.DepartmentID.Hex(),
        s.jwtSecret,
        s.jwtExpiry,
    )
    if err != nil {
        return nil, errors.New("failed to generate access token")
    }

    refreshToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.DepartmentID.Hex(),
        s.jwtSecret,
        "168h", // 7 days for refresh token
    )
    if err != nil {
        return nil, errors.New("failed to generate refresh token")
    }

    // Calculate expiry
    expiresIn := time.Now().Add(24 * time.Hour).Unix()

    return &models.LoginResponse{
        Success: true,
        Message: "Login successful",
        Data: models.LoginData{
            User:         user.ToResponse(),
            AccessToken:  accessToken,
            RefreshToken: refreshToken,
            ExpiresIn:    expiresIn,
        },
    }, nil
}

func (s *authService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
    // Check if email already exists
    existingUser, _ := s.userRepo.FindByEmail(ctx, req.Email)
    if existingUser != nil {
        return nil, errors.New("email already registered")
    }

    // Hash password
    hashedPassword, err := auth.HashPassword(req.Password)
    if err != nil {
        return nil, errors.New("failed to hash password")
    }

    // Convert department ID
    deptID, err := primitive.ObjectIDFromHex(req.DepartmentID)
    if err != nil {
        return nil, errors.New("invalid department ID")
    }

    // Convert position ID
    posID, err := primitive.ObjectIDFromHex(req.PositionID)
    if err != nil {
        return nil, errors.New("invalid position ID")
    }

    // Create user
    user := &models.User{
        NIK:          req.NIK,
        Email:        req.Email,
        Password:     hashedPassword,
        FullName:     req.FullName,
        Role:         req.Role,
        DepartmentID: deptID,
        PositionID:   posID,
        Phone:        req.Phone,
        Address:      req.Address,
        JoinDate:     time.Now(),
        IsActive:     true,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    return user, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
    // Validate refresh token
    claims, err := auth.ValidateToken(refreshToken, s.jwtSecret)
    if err != nil {
        return nil, errors.New("invalid refresh token")
    }

    // Get user
    user, err := s.userRepo.FindByID(ctx, claims.UserID)
    if err != nil {
        return nil, errors.New("user not found")
    }

    // Check if user is active
    if !user.IsActive {
        return nil, errors.New("user account is inactive")
    }

    // Generate new tokens
    newAccessToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.DepartmentID.Hex(),
        s.jwtSecret,
        s.jwtExpiry,
    )
    if err != nil {
        return nil, errors.New("failed to generate access token")
    }

    newRefreshToken, err := auth.GenerateToken(
        user.ID.Hex(),
        user.Role,
        user.DepartmentID.Hex(),
        s.jwtSecret,
        "168h",
    )
    if err != nil {
        return nil, errors.New("failed to generate refresh token")
    }

    expiresIn := time.Now().Add(24 * time.Hour).Unix()

    return &models.LoginResponse{
        Success: true,
        Message: "Token refreshed successfully",
        Data: models.LoginData{
            User:         user.ToResponse(),
            AccessToken:  newAccessToken,
            RefreshToken: newRefreshToken,
            ExpiresIn:    expiresIn,
        },
    }, nil
}

// Logout - Logout user
func (s *authService) Logout(ctx context.Context, userID string) error {
    // In JWT-based auth, we don't need to do anything server-side
    // The client should remove the token from their storage
    // 
    // If you want to implement token blacklisting:
    // 1. Create a "blacklisted_tokens" collection
    // 2. Store the token with expiry time
    // 3. Check blacklist in middleware before validating token
    //
    // For now, just return success
    return nil
}