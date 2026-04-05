package auth

import (
	"time"

	authcodex "cliro-go/internal/auth/codex"
	authkiro "cliro-go/internal/auth/kiro"
	"cliro-go/internal/config"
	syncauth "cliro-go/internal/sync/authtoken"
)

// Session status constants shared by codex and kiro auth sub-packages.
const (
	SessionPending = "pending"
	SessionSuccess = "success"
	SessionError   = "error"
)

// TokenExpired reports whether the account's access token has expired.
func TokenExpired(account config.Account, now time.Time) bool {
	if account.ExpiresAt <= 0 {
		return false
	}
	return now.Unix() >= account.ExpiresAt
}

// BlockedAccountMessage returns a human-readable blocked reason and true when
// the error string signals that an account has been deactivated/banned.
func BlockedAccountMessage(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	return config.BlockedAccountReason(err.Error())
}

// Re-exported types so app.go only needs to import the auth root package.
type CodexAuthStart = authcodex.AuthStart
type CodexAuthSessionView = authcodex.AuthSessionView

type KiroAuthStart = authkiro.AuthStart
type KiroAuthSessionView = authkiro.AuthSessionView

type KiloAuthSyncResult = syncauth.KiloAuthSyncResult
type OpencodeAuthSyncResult = syncauth.OpencodeAuthSyncResult
type CodexAuthSyncResult = syncauth.CodexAuthSyncResult
