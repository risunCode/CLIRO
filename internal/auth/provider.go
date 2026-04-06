package auth

import (
	"context"

	"cliro/internal/config"
)

type authProvider interface {
	StartAuth() (*AuthStart, error)
	StartSocialAuth(socialProvider string) (*AuthStart, error)
	GetSession(sessionID string) AuthSessionView
	CancelSession(sessionID string)
	SubmitCode(sessionID, code string) error
	RefreshAccount(account config.Account, force bool) (config.Account, error)
	Shutdown(ctx context.Context) error
}

type allSessionsProvider interface {
	AllSessions() []AuthSessionView
}
