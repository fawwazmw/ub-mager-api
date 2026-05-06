package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	ServerPort      string
	ServerEnv       string
	ReadTimeoutSec  int
	WriteTimeoutSec int
	IdleTimeoutSec  int

	// PostgreSQL
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisPoolSize int

	// JWT
	JWTSecret          string
	JWTAccessTokenTTL  string
	JWTRefreshTokenTTL string

	// CORS
	CORSOrigins string

	// OSRM
	OSRMBaseURL string

	// ML Service
	MLServiceURL string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	readTimeout, _ := strconv.Atoi(getEnv("READ_TIMEOUT_SEC", "15"))
	writeTimeout, _ := strconv.Atoi(getEnv("WRITE_TIMEOUT_SEC", "15"))
	idleTimeout, _ := strconv.Atoi(getEnv("IDLE_TIMEOUT_SEC", "60"))
	redisPoolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))

	cfg := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8081"),
		ServerEnv:          getEnv("SERVER_ENV", "development"),
		ReadTimeoutSec:     readTimeout,
		WriteTimeoutSec:    writeTimeout,
		IdleTimeoutSec:     idleTimeout,
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "ubmager"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		RedisHost:          getEnv("REDIS_HOST", "localhost"),
		RedisPort:          getEnv("REDIS_PORT", "6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisPoolSize:      redisPoolSize,
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTAccessTokenTTL:  getEnv("JWT_ACCESS_TTL", "900"),
		JWTRefreshTokenTTL: getEnv("JWT_REFRESH_TTL", "604800"),
		CORSOrigins:        getEnv("CORS_ORIGINS", "*"),
		OSRMBaseURL:        getEnv("OSRM_BASE_URL", "https://osrm.wardaya.my.id"),
		MLServiceURL:       getEnv("ML_SERVICE_URL", "http://localhost:8000"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
