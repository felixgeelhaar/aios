package core

import (
	"os"
	"path/filepath"
)

type Config struct {
	WorkspaceDir string
	LogLevel     string
	TokenService string
	ProjectDir   string
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func DefaultConfig() Config {
	root := envOrDefault("AIOS_WORKSPACE_DIR", filepath.Join(".", ".aios"))
	return Config{
		WorkspaceDir: root,
		LogLevel:     envOrDefault("AIOS_LOG_LEVEL", "info"),
		TokenService: envOrDefault("AIOS_TOKEN_SERVICE", "aios"),
		ProjectDir:   envOrDefault("AIOS_PROJECT_DIR", "."),
	}
}
