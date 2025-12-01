package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL     string
	GRPCPort        int
	LogLevel        string
	ShutdownTimeout time.Duration
	Version         string
	JWTSecret       string
	JWTDuration     time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
		Version:         getEnv("VERSION", "dev"),
		JWTSecret:       getEnv("JWT_SECRET", "default-secret-change-in-production"),
		JWTDuration:     getDurationEnv("JWT_DURATION", 24*time.Hour),
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	portStr := os.Getenv("GRPC_PORT")
	if portStr == "" {
		return nil, fmt.Errorf("GRPC_PORT is required")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("GRPC_PORT must be a valid integer: %w", err)
	}
	cfg.GRPCPort = port

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return defaultValue
}
