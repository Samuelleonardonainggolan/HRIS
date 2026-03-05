// internal/config/config.go
package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    MongoURI      string
    DatabaseName  string
    JWTSecret     string
    JWTExpiry     string
    RefreshExpiry string
    ServerPort    string
    Environment   string
    FaceServiceURL string
    FaceAPIKey      string
    FaceHTTPTimeout string
}

func LoadConfig() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using environment variables")
    }

    return &Config{
        MongoURI:      getEnv("MONGO_URI", ""),
        DatabaseName:  getEnv("DATABASE_NAME", "hris_db"),
        JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
        JWTExpiry:     getEnv("JWT_EXPIRY", "24h"),
        RefreshExpiry: getEnv("REFRESH_EXPIRY", "168h"),
        ServerPort:    getEnv("SERVER_PORT", "8080"),
        Environment:   getEnv("ENVIRONMENT", "development"),
        FaceServiceURL: getEnv("FACE_SERVICE_URL", "http://localhost:8001"),
        FaceAPIKey:      getEnv("FACE_API_KEY", "labersa-internal-api-key-2026"),
        FaceHTTPTimeout: getEnv("FACE_HTTP_TIMEOUT", "30s"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
