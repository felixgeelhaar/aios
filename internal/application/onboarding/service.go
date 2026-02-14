package onboarding

import (
	"context"
	"strings"

	domain "github.com/felixgeelhaar/aios/internal/domain/onboarding"
)

type Service struct {
	oauth     domain.OAuthCodeResolver
	connector domain.DriveConnector
	tray      domain.TrayStatePort
}

func NewService(oauth domain.OAuthCodeResolver, connector domain.DriveConnector, tray domain.TrayStatePort) Service {
	return Service{
		oauth:     oauth,
		connector: connector,
		tray:      tray,
	}
}

func (s Service) ConnectGoogleDrive(ctx context.Context, command domain.ConnectGoogleDriveCommand) (domain.ConnectGoogleDriveResult, error) {
	cmd := command.Normalized()

	token := strings.TrimSpace(cmd.TokenOverride)
	result := domain.ConnectGoogleDriveResult{}
	if token == "" {
		callbackURL, code, err := s.oauth.ResolveCode(ctx, cmd.State, cmd.Timeout)
		if err != nil {
			return domain.ConnectGoogleDriveResult{}, err
		}
		result.CallbackURL = callbackURL
		token = strings.TrimSpace(code)
	}
	if token == "" {
		return domain.ConnectGoogleDriveResult{}, domain.ErrTokenRequired
	}

	if err := s.connector.ConnectGoogleDrive(ctx, token); err != nil {
		return domain.ConnectGoogleDriveResult{}, err
	}
	if err := s.tray.SetGoogleDriveConnected(ctx, true); err != nil {
		return domain.ConnectGoogleDriveResult{}, err
	}
	return result, nil
}
