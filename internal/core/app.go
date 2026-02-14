package core

import (
	"fmt"
	"os"

	"github.com/felixgeelhaar/aios/internal/runtime"
)

type App struct {
	cfg Config
	log *Logger
}

func NewApp(cfg Config) *App {
	return &App{cfg: cfg, log: NewLogger(cfg.LogLevel)}
}

func (a *App) Run(mode string) error {
	if mode != "tray" && mode != "cli" {
		return fmt.Errorf("unsupported mode %q", mode)
	}
	if err := os.MkdirAll(a.cfg.WorkspaceDir, 0o750); err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	if mode == "tray" {
		rt, err := runtime.NewProductionRuntime(a.cfg.WorkspaceDir, a.cfg.TokenService)
		if err != nil {
			return fmt.Errorf("initialize secure runtime: %w", err)
		}
		if !rt.UsesSecureTokenStore() {
			return fmt.Errorf("insecure token store configured")
		}
		if _, err := RefreshTrayState(a.cfg, nil); err != nil {
			return fmt.Errorf("initialize tray state: %w", err)
		}
	}
	a.log.Info("app started in %s mode", mode)
	return nil
}
