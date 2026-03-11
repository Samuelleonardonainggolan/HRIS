// pkg/auth/jwt.go
package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID     string `json:"user_id"`
	Role       string `json:"role"`
	Department string `json:"department"`
	jwt.RegisteredClaims
}

// GenerateToken - Generate JWT token dengan expiry yang fleksibel
func GenerateToken(userID, role, department, secret, expiry string) (string, error) {
	var expiryDuration time.Duration
	var err error

	// Coba parse sebagai duration (format seperti "60m", "1h")
	expiryDuration, err = time.ParseDuration(expiry)
	if err != nil {
		// Jika gagal, coba parse sebagai angka (asumsi dalam menit)
		if minutes, parseIntErr := strconv.Atoi(expiry); parseIntErr == nil {
			expiryDuration = time.Duration(minutes) * time.Minute
		} else {
			return "", errors.New("invalid expiry format: use format like '60m' or number for minutes")
		}
	}

	claims := Claims{
		UserID:     userID,
		Role:       role,
		Department: department,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken - Validate JWT token
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
