package auth

// Session status constants shared by codex and kiro auth sub-packages.
const (
	SessionPending = "pending"
	SessionSuccess = "success"
	SessionError   = "error"
)

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
