// Package config handles application configuration loading and validation.
package config

import (
	"easyflow-oauth2-server/pkg/logger"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Environment represents the application environment.
type Environment string

// Define possible environments.
const (
	Development Environment = "development"
	Production  Environment = "production"
)

// Config holds the application configuration values.
type Config struct {
	// Application
	LogLevel          logger.LogLevel
	Port              string
	TrustedProxies    []string
	FrontendURL       string
	SaltRounds        int
	Domain            string
	Environment       Environment
	SessionCookieName string
	// Database
	DatabaseURL    string
	MigrationsPath string
	// Valkey
	ValkeyURL        string
	ValkeyUsername   string
	ValkeyPassword   string
	ValkeyClientName string
	// JWT
	JwtIssuer                  string
	JwtSessionTokenExpiryHours int    // in hours
	JwtSecret                  string // Needs to be 32 bytes long (32 characters)
}

// Get an environment variable or return a default value.
func getEnv(key, fallback string, validation func(value string) bool, log *logger.Logger) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		value = fallback
	}

	if !validation(value) {
		log.PrintfError("Invalid value for %s: %s\n", key, value)
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int, validation func(value int) bool, log *logger.Logger) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		value = strconv.Itoa(fallback)
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		log.PrintfError("Could not parse string to integer for %s: %s", key, value)
		return fallback
	}

	if !validation(i) {
		log.PrintfError("Invalid value for %s: %s\n", key, value)
		return fallback
	}
	return i
}

func getEnvSlice(
	key, fallback string,
	validation func(value string) bool,
	log *logger.Logger,
) []string {
	value, ok := os.LookupEnv(key)
	if !ok {
		value = fallback
	}

	// If empty, return empty slice
	if value == "" {
		return []string{}
	}

	// Split by comma and trim spaces
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	for _, part := range parts {
		if !validation(part) {
			log.PrintfError("Invalid value for %s: %s\n", key, part)
			return strings.Split(fallback, ",")
		}
	}
	return parts
}

// LoadDefaultConfig loads and validates the configuration from environment variables.
// Loads the .env file if present in the current working directory.
// If an environment variable is not set, it uses the provided default value.
func LoadDefaultConfig() (*Config, error) {
	log := logger.NewLogger(os.Stdout, "Config", logger.INFO, "System")

	logLevel := logger.LogLevel(getEnv("LOG_LEVEL", "DEBUG", func(value string) bool {
		switch value {
		case string(logger.DEBUG),
			string(logger.INFO),
			string(logger.WARNING),
			string(logger.ERROR):
			return true
		default:
			return false
		}
	}, log))

	pwd, err := os.Getwd()
	if err != nil {
		log.PrintfError("Error getting current working directory: %v", err)
		return nil, err
	}
	// Load .env file if present
	err = godotenv.Load(fmt.Sprintf("%s/.env", pwd))
	if err != nil {
		log.PrintfInfo("No env file present")
	}

	return &Config{
		// Application
		LogLevel: logLevel,
		Port: getEnv("PORT", "8080", func(value string) bool {
			if i, err := strconv.Atoi(value); err == nil && i > 0 && i < 65536 {
				return true
			}
			return false
		}, log),
		TrustedProxies: getEnvSlice("TRUSTED_PROXIES", "",
			func(value string) bool {
				// Simple validation: split by comma and check if each part is a valid IP or CIDR
				if _, err := url.Parse(value); err != nil {
					if net.ParseIP(value) == nil {
						if _, _, err := net.ParseCIDR(value); err != nil {
							return false
						}
					}
				}
				return true
			}, log),
		FrontendURL: getEnv("FRONTEND_URL", "",
			func(value string) bool {
				_, err := url.ParseRequestURI(value)
				return err == nil
			}, log),
		SaltRounds: getEnvInt("SALT_ROUNDS", 10, func(value int) bool {
			return value >= 4 && value <= 31
		}, log),
		// Database
		DatabaseURL: getEnv("DATABASE_URL", "", func(value string) bool {
			_, err := url.Parse(value)
			return err == nil
		}, log),
		MigrationsPath: getEnv(
			"MIGRATIONS_PATH",
			"pkg/database/sql/migrations",
			func(value string) bool {
				_, err := os.Stat(value)
				return err == nil
			},
			log,
		),
		Environment: Environment(getEnv("ENVIRONMENT", "development", func(value string) bool {
			return value == string(Development) || value == string(Production)
		}, log)),
		Domain: getEnv("DOMAIN", "localhost", func(value string) bool {
			_, err := url.Parse(value)
			return err == nil
		}, log),
		SessionCookieName: getEnv("SESSION_COOKIE_NAME", "session_token", func(value string) bool {
			return value != ""
		}, log),
		// Valkey
		ValkeyURL: getEnv("VALKEY_URL", "", func(value string) bool {
			_, err := url.Parse(value)
			if err == nil {
				return true
			}
			host, port, err := net.SplitHostPort(value)
			if err != nil {
				return false
			}
			if net.ParseIP(host) != nil && port != "" {
				return true
			}
			return false
		}, log),
		ValkeyUsername: getEnv(
			"VALKEY_USERNAME",
			"",
			func(value string) bool { return value != "" },
			log,
		),
		ValkeyPassword: getEnv(
			"VALKEY_PASSWORD",
			"",
			func(_ string) bool { return true },
			log,
		),
		ValkeyClientName: getEnv(
			"VALKEY_CLIENT_NAME",
			"",
			func(value string) bool { return value != "" },
			log,
		),
		// JWT
		JwtIssuer: getEnv(
			"JWT_ISSUER",
			"",
			func(value string) bool { return value != "" },
			log,
		),
		JwtSessionTokenExpiryHours: getEnvInt(
			"JWT_SESSION_TOKEN_EXPIRY_HOURS",
			1,
			func(value int) bool { return value > 0 },
			log,
		),
		JwtSecret: getEnv("JWT_SECRET", "", func(value string) bool {
			return len([]byte(value)) == 32
		}, log),
	}, nil
}
