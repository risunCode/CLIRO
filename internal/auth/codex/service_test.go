package codex

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"cliro/internal/config"
	"cliro/internal/logger"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	dir := t.TempDir()
	store, err := config.NewManager(dir)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	log := logger.New(100)
	return NewService(store, log, nil, nil)
}

func TestGetAuthSession_NotFound(t *testing.T) {
	svc := newTestService(t)
	view := svc.GetAuthSession("nonexistent")
	if view.Status != sessionError {
		t.Fatalf("status = %q, want %q", view.Status, sessionError)
	}
	if view.Error != "session not found" {
		t.Fatalf("error = %q", view.Error)
	}
}

func TestCancelAuth_NotFound(t *testing.T) {
	svc := newTestService(t)
	svc.CancelAuth("nonexistent")
}

func TestSubmitAuthCode_EmptyCode(t *testing.T) {
	svc := newTestService(t)
	err := svc.SubmitAuthCode("some-session", "")
	if err != errEmptyAuthorizationCode {
		t.Fatalf("error = %v, want %v", err, errEmptyAuthorizationCode)
	}
}

func TestSubmitAuthCode_SessionNotFound(t *testing.T) {
	svc := newTestService(t)
	err := svc.SubmitAuthCode("nonexistent", "auth-code")
	if err != errSessionNotFound {
		t.Fatalf("error = %v, want %v", err, errSessionNotFound)
	}
}

func TestRefreshAccount_NoRefreshToken(t *testing.T) {
	svc := newTestService(t)
	account := config.Account{ID: "test-id", Email: "test@example.com", Provider: "codex"}
	_, err := svc.RefreshAccount(account, true)
	if err == nil {
		t.Fatal("expected error for missing refresh token")
	}
	if !strings.Contains(err.Error(), "no refresh token") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRefreshAccount_SkipsWhenNotExpired(t *testing.T) {
	svc := newTestService(t)
	account := config.Account{
		ID:           "test-id",
		Email:        "test@example.com",
		Provider:     "codex",
		RefreshToken: "some-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}
	result, err := svc.RefreshAccount(account, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Email != account.Email {
		t.Fatalf("email = %q, want %q", result.Email, account.Email)
	}
}

func TestShutdown_NoServer(t *testing.T) {
	svc := newTestService(t)
	if err := svc.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown without server: %v", err)
	}
}

func TestClient_Fallback(t *testing.T) {
	svc := newTestService(t)
	client := svc.client()
	if client == nil {
		t.Fatal("client() returned nil")
	}
}

func TestClient_WithProvider(t *testing.T) {
	svc := newTestService(t)
	called := false
	svc.httpClient = func() *http.Client {
		called = true
		return nil
	}
	client := svc.client()
	if !called {
		t.Fatal("httpClient function was not called")
	}
	if client == nil {
		t.Fatal("client() should fall back to default when provider returns nil")
	}
}
