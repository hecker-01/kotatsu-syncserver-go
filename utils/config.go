package utils

import (
	"errors"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// Database Configuration
	DatabaseHost        string
	DatabasePort        int
	DatabaseName        string
	DatabaseUser        string
	DatabasePassword    string
	DatabaseRootPassword string // Optional: for auto database creation

	// JWT Configuration
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string

	// Server Configuration
	Port int

	// Application Configuration
	AllowNewRegister bool
	BaseURL          string

	// Mail Configuration
	MailProvider string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
}

// LoadConfig loads configuration from environment variables with defaults.
// Returns an error if required fields are missing.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Database defaults
		DatabaseHost:         getEnvOrDefault("DATABASE_HOST", "localhost"),
		DatabasePort:         getEnvAsIntOrDefault("DATABASE_PORT", 3306),
		DatabaseName:         getEnvOrDefault("DATABASE_NAME", "kotatsu_db"),
		DatabaseUser:         os.Getenv("DATABASE_USER"),
		DatabasePassword:     os.Getenv("DATABASE_PASSWORD"),
		DatabaseRootPassword: os.Getenv("DATABASE_ROOT_PASSWORD"),

		// JWT defaults
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTIssuer:   getEnvOrDefault("JWT_ISSUER", "http://0.0.0.0:9292/"),
		JWTAudience: getEnvOrDefault("JWT_AUDIENCE", "http://0.0.0.0:9292/resource"),

		// Server defaults
		Port: getEnvAsIntOrDefault("PORT", 9292),

		// Application defaults
		AllowNewRegister: getEnvAsBoolOrDefault("ALLOW_NEW_REGISTER", true),
		BaseURL:          getEnvOrDefault("BASE_URL", "http://localhost:9292"),

		// Mail defaults
		MailProvider: getEnvOrDefault("MAIL_PROVIDER", "console"),
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     getEnvAsIntOrDefault("SMTP_PORT", 587),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required configuration fields are set.
func (c *Config) validate() error {
	var missing []string

	if c.DatabaseUser == "" {
		missing = append(missing, "DATABASE_USER")
	}
	if c.DatabasePassword == "" {
		missing = append(missing, "DATABASE_PASSWORD")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	// If SMTP mail provider is configured, validate SMTP settings
	if c.MailProvider == "smtp" {
		if c.SMTPHost == "" {
			missing = append(missing, "SMTP_HOST")
		}
		if c.SMTPFrom == "" {
			missing = append(missing, "SMTP_FROM")
		}
	}

	if len(missing) > 0 {
		return errors.New("missing required environment variables: " + strings.Join(missing, ", "))
	}

	return nil
}

// getEnvOrDefault returns the value of the environment variable or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault returns the value of the environment variable as int or a default.
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvAsBoolOrDefault returns the value of the environment variable as bool or a default.
func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		lowerValue := strings.ToLower(value)
		return lowerValue == "true" || lowerValue == "1" || lowerValue == "yes"
	}
	return defaultValue
}

// AppInfo contains application metadata.
type AppInfo struct {
	Name    string
	Version string
	Env     string
}

// GetAppVersion returns the application version from environment or build info.
func GetAppVersion() string {
	return "v0.0.1"
}

// GetAppEnv returns the current environment from environment variable.
func GetAppEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return "production"
	}
	return env
}

// GetGoroutineCount returns the current number of goroutines.
func GetGoroutineCount() int {
	return runtime.NumGoroutine()
}

// GetAppInfo creates the application info struct.
func GetAppInfo() AppInfo {
	return AppInfo{
		Name:    "Kotatsu Sync Server",
		Version: GetAppVersion(),
		Env:     GetAppEnv(),
	}
}

// GetHealthData returns health check data for the /api/health endpoint.
func GetHealthData() map[string]interface{} {
	info := GetAppInfo()
	return map[string]interface{}{
		"api_name":    info.Name,
		"api_version": info.Version,
		"environment": info.Env,
		"goroutines":  GetGoroutineCount(),
		"git_commit":  getGitCommit(),
		"git_date":    getGitDate(),
		"go_version":  runtime.Version(),
		"os":          runtime.GOOS,
		"arch":        runtime.GOARCH,
	}
}

// getGitCommit returns the last commit hash from build info or empty string.
func getGitCommit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return "unknown"
}

// getGitDate returns the build time from build info or empty string.
func getGitDate() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				return setting.Value
			}
		}
	}
	return "unknown"
}
