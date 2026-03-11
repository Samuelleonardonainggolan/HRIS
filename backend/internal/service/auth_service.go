// internal/service/auth_service.go
package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/andikatampubolon10/hris-backend/pkg/auth"
	"github.com/andikatampubolon10/hris-backend/pkg/database/repository"
	"github.com/andikatampubolon10/hris-backend/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService interface {
	Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	Register(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, error)
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
	log.Printf("🔐 Login attempt - Email: %s", req.Email)

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("❌ Error finding user: %v", err)
		return nil, errors.New("invalid email or password")
	}

	if user == nil {
		log.Printf("❌ User not found in database: %s", req.Email)
		return nil, errors.New("invalid email or password")
	}

	log.Printf("✅ User found - Email: %s, Role: %s, Active: %v", user.Email, user.Role, user.IsActive)
	log.Printf("📝 Password hash in DB: %s...", user.Password[:20]) // Show first 20 chars

	// Check password
	passwordMatch := auth.CheckPasswordHash(req.Password, user.Password)
	log.Printf("🔑 Password check - Input: %s, Match: %v", req.Password, passwordMatch)

	if !passwordMatch {
		log.Printf("❌ Password mismatch for: %s", req.Email)
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		log.Printf("❌ User account inactive: %s", req.Email)
		return nil, errors.New("user account is inactive")
	}

	// Generate access token
	accessToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.DepartmentName, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		log.Printf("❌ Failed to generate access token: %v", err)
		return nil, errors.New("failed to generate access token")
	}

	// Generate refresh token
	refreshToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.DepartmentName, s.jwtSecret, "168h")
	if err != nil {
		log.Printf("❌ Failed to generate refresh token: %v", err)
		return nil, errors.New("failed to generate refresh token")
	}

	expiresIn := time.Now().Add(24 * time.Hour).Unix()
	userResponse := user.ToResponse()

	log.Printf("✅ Login successful - User ID: %s, Role: %s", user.ID.Hex(), user.Role)

	return &models.LoginResponse{
		User:         userResponse,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) Register(ctx context.Context, req models.RegisterRequest) (*models.UserResponse, error) {
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

	// Create user with minimal info (for simple registration)
	user := &models.User{
		ID:               primitive.NewObjectID(),
		PayrollNumber:    "", // Will be set by HR later
		Email:            req.Email,
		Password:         hashedPassword,
		FullName:         req.FullName,
		BirthDate:        time.Time{},          // Will be set later
		Religion:         "",                   // Will be set later
		LastEducation:    "",                   // Will be set later
		YearEnrolled:     "",                   // Will be set later
		EmploymentStatus: "",                   // Will be set later
		DepartmentID:     primitive.ObjectID{}, // Will be set later
		DepartmentName:   "",                   // Will be set later
		PositionID:       primitive.ObjectID{}, // Will be set later
		PositionName:     "",                   // Will be set later
		Phone:            "",                   // Will be set later
		Address:          "",                   // Will be set later
		Role:             req.Role,
		IsActive:         true,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.New("failed to create user")
	}

	// Return user response
	response := user.ToResponse()
	return &response, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// Validate refresh token
	claims, err := auth.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Find user
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Generate new access token
	newAccessToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.DepartmentName, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Generate new refresh token
	newRefreshToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.DepartmentName, s.jwtSecret, "168h")
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	// Calculate expiry time
	expiresIn := time.Now().Add(24 * time.Hour).Unix()

	// Convert to response
	userResponse := user.ToResponse()

	return &models.LoginResponse{
		User:         userResponse,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID string) error {
	// In stateless JWT, logout is handled client-side by removing token
	// Optionally, implement token blacklist here if needed
	return nil
}
