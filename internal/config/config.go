package config

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	LogLevel string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("POSTGRES_HOST", "postgres"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		DBUser:     getEnv("POSTGRES_USER", "yadro-test-2026-user"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "yadro-test-2026-pass"),
		DBName:     getEnv("POSTGRES_DB", "yadro-test-2026-db"),
		DBSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}

	cfg.Port = ":" + cfg.Port

	fmt.Printf("cfg.Port = %q\n", cfg.Port)
	fmt.Printf("cfg.DBHost = %q\n", cfg.DBHost)

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// func toInt(value string) int {
// 	integer, err := strconv.Atoi(value)
// 	if err != nil {
// 		panic(fmt.Sprintf("invalid integer config value %q: %v", value, err))
// 	}
// 	return integer
// }

// func secondsStringToDuration(value string) time.Duration {
// 	duration := toInt(value)
// 	return time.Duration(duration) * time.Second
// }

func (c *Config) PostgresDSN() string {
	ssl := c.DBSSLMode
	if ssl == "" {
		ssl = "disable"
	}
	host := c.DBHost
	if host == "" {
		host = "localhost"
	}
	port := c.DBPort
	if port == "" {
		port = "5432"
	}

	user := url.QueryEscape(c.DBUser)
	pass := url.QueryEscape(c.DBPassword)
	db := url.PathEscape(c.DBName)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, pass, host, port, db, url.QueryEscape(ssl),
	)
}
