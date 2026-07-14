package app

import (
	"crypto/sha256"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	RedisURL              string
	JWTSecret             string
	PublicURL             string
	FrontendOrigins       []string
	PaymentKey            [32]byte
	PaymentMock           bool
	AutoMigrate           bool
	AllowUserRegistration bool
}

func LoadConfig() Config {
	c := Config{
		Port:                  env("PORT", "8080"),
		DatabaseURL:           env("DATABASE_URL", "postgres://saaskit:saaskit@localhost:5432/saaskit?sslmode=disable"),
		RedisURL:              env("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:             env("JWT_SECRET", "change-me-in-production"),
		PublicURL:             strings.TrimRight(env("PUBLIC_URL", "http://localhost:8080"), "/"),
		FrontendOrigins:       strings.Split(env("FRONTEND_ORIGINS", "http://localhost:5666"), ","),
		PaymentMock:           envBool("PAYMENT_MOCK", true),
		AutoMigrate:           envBool("AUTO_MIGRATE", true),
		AllowUserRegistration: envBool("ALLOW_USER_REGISTRATION", true),
	}
	c.PaymentKey = sha256.Sum256([]byte(env("PAYMENT_CONFIG_KEY", c.JWTSecret)))
	return c
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
