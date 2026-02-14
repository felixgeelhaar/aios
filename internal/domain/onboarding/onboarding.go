package onboarding

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const DefaultOAuthState = "aios"
const DefaultOAuthTimeout = 120 * time.Second

var ErrTokenRequired = fmt.Errorf("token is required")

type ConnectGoogleDriveCommand struct {
	TokenOverride string
	State         string
	Timeout       time.Duration
}

type ConnectGoogleDriveResult struct {
	CallbackURL string
}

type OAuthCodeResolver interface {
	ResolveCode(ctx context.Context, state string, timeout time.Duration) (callbackURL string, code string, err error)
}

type DriveConnector interface {
	ConnectGoogleDrive(ctx context.Context, token string) error
}

type TrayStatePort interface {
	SetGoogleDriveConnected(ctx context.Context, connected bool) error
}

func (c ConnectGoogleDriveCommand) Normalized() ConnectGoogleDriveCommand {
	state := strings.TrimSpace(c.State)
	if state == "" {
		state = DefaultOAuthState
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = DefaultOAuthTimeout
	}
	return ConnectGoogleDriveCommand{
		TokenOverride: strings.TrimSpace(c.TokenOverride),
		State:         state,
		Timeout:       timeout,
	}
}

