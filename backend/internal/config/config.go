// internal/config/config.go
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort             string
	Environment            string
	MongoURI               string
	DatabaseName           string
	JWTSecret              string
	JWTExpiry              string // Keep as string for '24h' format support
	FaceServiceURL         string
	FaceAPIKey             string
	FaceHTTPTimeout        string
	PublicBaseURL          string
	FaceImageDir           string
	PengajuanDocDir        string
	SupabaseURL            string
	SupabaseAPIKey         string
	SupabaseServiceRoleKey string
	SupabaseBucket         string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Gunakan string langsung agar support format "24h", "60m", dll.
	jwtExpiryStr := getEnv("JWT_EXPIRY", "24h")

	return &Config{
		ServerPort:             getEnv("SERVER_PORT", "8080"),
		Environment:            getEnv("ENVIRONMENT", "development"),
		MongoURI:               getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:           getEnv("DATABASE_NAME", "hris_db"),
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiry:              jwtExpiryStr,
		FaceServiceURL:         getEnv("FACE_SERVICE_URL", "http://localhost:5000"),
		FaceAPIKey:             getEnv("FACE_API_KEY", ""),
		FaceHTTPTimeout:        getEnv("FACE_HTTP_TIMEOUT", "30s"),
		PublicBaseURL:          getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		FaceImageDir:           getEnv("FACE_IMAGE_DIR", "uploads/face"),
		PengajuanDocDir:        getEnv("PENGAJUAN_DOC_DIR", "uploads/pengajuan"),
		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseAPIKey:         getEnv("SUPABASE_API_KEY", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseBucket:         getEnv("SUPABASE_BUCKET", "hris-assets"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
