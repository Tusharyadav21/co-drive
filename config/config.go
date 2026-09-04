package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration for Co-Drive.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	Auth     AuthConfig     `yaml:"auth"`
}

// ServerConfig defines the HTTP server host and listening port.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// DatabaseConfig holds database storage connection parameters.
type DatabaseConfig struct {
	URL string `yaml:"url"`
}

// SMTPConfig holds mail server connection and sender credentials.
type SMTPConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	FromEmail string `yaml:"from_email"`
}

// AuthConfig defines authentication parameters including token TTLs.
type AuthConfig struct {
	OTPTTLMinutes           int `yaml:"otp_ttl_minutes"`
	PasswordResetTTLMinutes int `yaml:"password_reset_ttl_minutes"`
}

// Load reads configuration from a YAML file (if specified) and applies environment
// variable overrides with safe defaults. This function and file is the sole location
// in the application permitted to read environment variables.
func Load(path string) (*Config, error) {
	// Auto-load .env file into environment if present
	loadDotEnv(".env")

	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			URL: "postgres://localhost:5432/co-drive?sslmode=disable",
		},
		Auth: AuthConfig{
			OTPTTLMinutes:           10,
			PasswordResetTTLMinutes: 15,
		},
		SMTP: SMTPConfig{
			Port:      587,
			FromEmail: "no-reply@co-drive.app",
		},
	}

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Environment variable overrides (prioritizing SERVER_PORT then PORT)
	if portStr := getEnvFirst("SERVER_PORT", "PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.Server.Port = p
		}
	}
	if host := getEnvFirst("SERVER_HOST", "HOST"); host != "" {
		cfg.Server.Host = host
	}

	if dbURL := getEnvFirst("DATABASE_URL", "DB_URL", "POSTGRES_URL"); dbURL != "" {
		cfg.Database.URL = dbURL
	}

	if otpTTL := os.Getenv("OTP_TTL_MINUTES"); otpTTL != "" {
		if val, err := strconv.Atoi(otpTTL); err == nil && val > 0 {
			cfg.Auth.OTPTTLMinutes = val
		}
	}
	if resetTTL := os.Getenv("PASSWORD_RESET_TTL_MINUTES"); resetTTL != "" {
		if val, err := strconv.Atoi(resetTTL); err == nil && val > 0 {
			cfg.Auth.PasswordResetTTLMinutes = val
		}
	}

	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		cfg.SMTP.Host = smtpHost
	}
	if smtpPort := os.Getenv("SMTP_PORT"); smtpPort != "" {
		if p, err := strconv.Atoi(smtpPort); err == nil && p > 0 {
			cfg.SMTP.Port = p
		}
	}
	if smtpUser := os.Getenv("SMTP_USER"); smtpUser != "" {
		cfg.SMTP.Username = smtpUser
	}
	if smtpPass := os.Getenv("SMTP_PASS"); smtpPass != "" {
		cfg.SMTP.Password = smtpPass
	}
	if from := os.Getenv("SMTP_FROM"); from != "" {
		cfg.SMTP.FromEmail = from
	}

	return cfg, nil
}

// loadDotEnv parses a local .env file and sets environment variables that are not yet populated.
func loadDotEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}

// getEnvFirst checks multiple environment variable names in order and returns the first non-empty value.
func getEnvFirst(keys ...string) string {
	for _, k := range keys {
		if val := os.Getenv(k); val != "" {
			return val
		}
	}
	return ""
}
