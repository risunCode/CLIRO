package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"cliro/internal/config"
	"cliro/internal/logger"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	store, err := config.NewManager(dir)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	log := logger.New(100)
	return NewManager(store, log)
}

func TestNewManager_InitializesProviders(t *testing.T) {
	mgr := newTestManager(t)
	if mgr.providers == nil {
		t.Fatal("providers map not initialized")
	}
	if _, ok := mgr.providers["codex"]; !ok {
		t.Fatal("codex provider not registered")
	}
	if _, ok := mgr.providers["kiro"]; !ok {
		t.Fatal("kiro provider not registered")
	}
}

func TestHTTPClient_DefaultTimeout(t *testing.T) {
	mgr := newTestManager(t)
	client := mgr.httpClient()
	if client == nil {
		t.Fatal("httpClient returned nil")
	}
	if client.Timeout != 60*time.Second {
		t.Fatalf("timeout = %v, want 60s", client.Timeout)
	}
}

func TestHTTPClient_CustomClient(t *testing.T) {
	mgr := newTestManager(t)
	custom := &http.Client{Timeout: 30 * time.Second}
	mgr.SetHTTPClient(custom)
	client := mgr.httpClient()
	if client.Timeout != 30*time.Second {
		t.Fatalf("timeout = %v, want 30s", client.Timeout)
	}
}

func TestHTTPClient_NilGuard(t *testing.T) {
	mgr := newTestManager(t)
	mgr.SetHTTPClient(nil)
	client := mgr.httpClient()
	if client == nil {
		t.Fatal("httpClient should not return nil even after SetHTTPClient(nil)")
	}
}

func TestShutdown_WithoutError(t *testing.T) {
	mgr := newTestManager(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := mgr.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestSetQuotaRefresher(t *testing.T) {
	mgr := newTestManager(t)
	mgr.SetQuotaRefresher(&testQuotaRefresher{fn: func(id string) error {
		return nil
	}})
	if mgr.quotaRefresher == nil {
		t.Fatal("quotaRefresher not set")
	}
}

type testQuotaRefresher struct {
	fn func(string) error
}

func (t *testQuotaRefresher) RefreshQuotaOnly(id string) error {
	return t.fn(id)
}
