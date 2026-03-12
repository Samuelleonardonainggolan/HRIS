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
	GetFaceRegistrationStatus(ctx context.Context, userID string) (bool, error) // Tambahkan interface method
}

type authService struct {
	userRepo          repository.UserRepository
	faceEmbeddingRepo repository.FaceEmbeddingRepository // TAMBAHKAN FIELD INI
	jwtSecret         string
	jwtExpiry         string
}

// UPDATE CONSTRUCTOR - Tambahkan faceEmbeddingRepo sebagai parameter
func NewAuthService(
	userRepo repository.UserRepository,
	faceEmbeddingRepo repository.FaceEmbeddingRepository, // TAMBAHKAN PARAMETER
	jwtSecret,
	jwtExpiry string,
) AuthService {
	return &authService{
		userRepo:          userRepo,
		faceEmbeddingRepo: faceEmbeddingRepo, // SIMPAN KE STRUCT
		jwtSecret:         jwtSecret,
		jwtExpiry:         jwtExpiry,
	}
}

// GetFaceRegistrationStatus - Method untuk cek status face
func (s *authService) GetFaceRegistrationStatus(ctx context.Context, userID string) (bool, error) {
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return false, err
	}
	return faceEmbedding != nil, nil
}

func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	log.Printf("🔐 Login attempt - Email: %s", req.Email)

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	if !auth.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// Generate tokens
	accessToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.DepartmentName, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := auth.GenerateToken(user.ID.Hex(), user.Role, user.DepartmentName, s.jwtSecret, "168h")
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	expiresIn := time.Now().Add(24 * time.Hour).Unix()

	// CEK APAKAH USER SUDAH PUNYA FACE EMBEDDING
	var requiresFaceRegistration bool = true
	faceEmbedding, err := s.faceEmbeddingRepo.FindByUserID(ctx, user.ID.Hex())
	if err == nil && faceEmbedding != nil {
		requiresFaceRegistration = false
	}

	log.Printf("✅ Login successful - User ID: %s, requiresFaceRegistration: %v", user.ID.Hex(), requiresFaceRegistration)

	// Buat response dengan field tambahan
	response := &models.LoginResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}

	// Simpan requiresFaceRegistration di context atau return value terpisah
	// Kita akan handle di handler dengan memanggil GetFaceRegistrationStatus lagi

	return response, nil
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
