package auth

import (
	"time"

	"cliro-go/internal/config"
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

type AuthStart struct {
	SessionID       string `json:"sessionId"`
	AuthURL         string `json:"authUrl"`
	CallbackURL     string `json:"callbackUrl,omitempty"`
	VerificationURL string `json:"verificationUrl,omitempty"`
	UserCode        string `json:"userCode,omitempty"`
	ExpiresAt       int64  `json:"expiresAt,omitempty"`
	Status          string `json:"status"`
	AuthMethod      string `json:"authMethod,omitempty"`
	Provider        string `json:"provider,omitempty"`
}

type AuthSessionView struct {
	SessionID       string `json:"sessionId"`
	AuthURL         string `json:"authUrl"`
	CallbackURL     string `json:"callbackUrl,omitempty"`
	VerificationURL string `json:"verificationUrl,omitempty"`
	UserCode        string `json:"userCode,omitempty"`
	ExpiresAt       int64  `json:"expiresAt,omitempty"`
	Status          string `json:"status"`
	Error           string `json:"error,omitempty"`
	AccountID       string `json:"accountId,omitempty"`
	Email           string `json:"email,omitempty"`
	AuthMethod      string `json:"authMethod,omitempty"`
	Provider        string `json:"provider,omitempty"`
}

type AuthSyncTarget string

const (
	SyncTargetKilo     AuthSyncTarget = "kilo-cli"
	SyncTargetOpencode AuthSyncTarget = "opencode-cli"
	SyncTargetCodexCLI AuthSyncTarget = "codex-cli"
)

type AuthSyncResult struct {
	Target          string   `json:"target"`
	TargetPath      string   `json:"targetPath"`
	FileExisted     bool     `json:"fileExisted"`
	OpenAICreated   bool     `json:"openAICreated"`
	BackupPath      string   `json:"backupPath,omitempty"`
	BackupCreated   bool     `json:"backupCreated"`
	UpdatedFields   []string `json:"updatedFields"`
	AccountID       string   `json:"accountID"`
	Provider        string   `json:"provider"`
	SyncedExpires   int64    `json:"syncedExpires"`
	SyncedExpiresAt string   `json:"syncedExpiresAt,omitempty"`
	SyncedAt        string   `json:"syncedAt,omitempty"`
}
