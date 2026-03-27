package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv                  string
	HTTPAddr                string
	HTTPReadTimeout         time.Duration
	HTTPWriteTimeout        time.Duration
	HTTPIdleTimeout         time.Duration
	HTTPShutdownTimeout     time.Duration
	DatabaseURL             string
	DatabaseMaxConns        int32
	DatabaseMinConns        int32
	DatabaseMaxConnLifetime time.Duration
	DatabaseMaxConnIdleTime time.Duration
}

func Load() (Config, error) {
	httpReadTimeout, err := durationFromEnv("HTTP_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpWriteTimeout, err := durationFromEnv("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpIdleTimeout, err := durationFromEnv("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return Config{}, err
	}

	httpShutdownTimeout, err := durationFromEnv("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	databaseMaxConns, err := int32FromEnv("DATABASE_MAX_CONNS", 10)
	if err != nil {
		return Config{}, err
	}

	databaseMinConns, err := int32FromEnv("DATABASE_MIN_CONNS", 0)
	if err != nil {
		return Config{}, err
	}

	databaseMaxConnLifetime, err := durationFromEnv("DATABASE_MAX_CONN_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	databaseMaxConnIdleTime, err := durationFromEnv("DATABASE_MAX_CONN_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnv:                  stringFromEnv("APP_ENV", "development"),
		HTTPAddr:                stringFromEnv("HTTP_ADDR", ":8080"),
		HTTPReadTimeout:         httpReadTimeout,
		HTTPWriteTimeout:        httpWriteTimeout,
		HTTPIdleTimeout:         httpIdleTimeout,
		HTTPShutdownTimeout:     httpShutdownTimeout,
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		DatabaseMaxConns:        databaseMaxConns,
		DatabaseMinConns:        databaseMinConns,
		DatabaseMaxConnLifetime: databaseMaxConnLifetime,
		DatabaseMaxConnIdleTime: databaseMaxConnIdleTime,
	}, nil
}

func stringFromEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return duration, nil
}

func int32FromEnv(key string, fallback int32) (int32, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	return int32(parsed), nil
}
