package core

import (
	"context"
	"time"

	domain "github.com/felixgeelhaar/aios/internal/domain/onboarding"
	"github.com/felixgeelhaar/aios/internal/runtime"
)

type oauthCodeResolverAdapter struct{}

func (oauthCodeResolverAdapter) ResolveCode(ctx context.Context, state string, timeout time.Duration) (string, string, error) {
	oauthCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	callbackURL, resultCh, stop, err := runtime.StartOAuthCallbackServer(oauthCtx, "127.0.0.1:0", state)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = stop() }()

	code, err := runtime.WaitForOAuthCode(oauthCtx, resultCh)
	if err != nil {
		return callbackURL, "", err
	}
	return callbackURL, code, nil
}

type driveConnectorAdapter struct {
	cfg Config
}

func (a driveConnectorAdapter) ConnectGoogleDrive(ctx context.Context, token string) error {
	rt, err := runtime.NewProductionRuntime(a.cfg.WorkspaceDir, a.cfg.TokenService)
	if err != nil {
		return err
	}
	return rt.ConnectGoogleDrive(ctx, token)
}

type trayStatePortAdapter struct {
	cfg Config
}

func (a trayStatePortAdapter) SetGoogleDriveConnected(_ context.Context, connected bool) error {
	_, err := RefreshTrayState(a.cfg, &connected)
	return err
}

var _ domain.OAuthCodeResolver = oauthCodeResolverAdapter{}
var _ domain.DriveConnector = driveConnectorAdapter{}
var _ domain.TrayStatePort = trayStatePortAdapter{}
