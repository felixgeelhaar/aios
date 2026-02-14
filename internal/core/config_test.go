package core

import "testing"

func TestDefaultConfigFromEnv(t *testing.T) {
	t.Setenv("AIOS_WORKSPACE_DIR", "/tmp/aios")
	t.Setenv("AIOS_LOG_LEVEL", "debug")
	t.Setenv("AIOS_TOKEN_SERVICE", "aios-test")
	cfg := DefaultConfig()
	if cfg.WorkspaceDir != "/tmp/aios" {
		t.Fatalf("workspace mismatch: %s", cfg.WorkspaceDir)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("log level mismatch: %s", cfg.LogLevel)
	}
	if cfg.TokenService != "aios-test" {
		t.Fatalf("token service mismatch: %s", cfg.TokenService)
	}
}

// AC4: Must support environment variable override for project directory.
func TestDefaultConfigProjectDirOverride(t *testing.T) {
	t.Setenv("AIOS_PROJECT_DIR", "/custom/project")
	cfg := DefaultConfig()
	if cfg.ProjectDir != "/custom/project" {
		t.Fatalf("project dir mismatch: %s", cfg.ProjectDir)
	}
}
