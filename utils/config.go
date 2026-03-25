package utils

import (
	"os"
	"runtime"
	"runtime/debug"
)

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
