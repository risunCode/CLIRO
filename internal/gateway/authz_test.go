package gateway

import (
	"net/http"
	"strings"
	"testing"

	"cliro/internal/config"
	"cliro/internal/logger"
)

func newAuthzTestServer(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	store, err := config.NewManager(dir)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	log := logger.New(100)
	return NewServer(store, nil, nil, log)
}

func TestInvalidRequest(t *testing.T) {
	err := InvalidRequest("test error")
	if err.Status != http.StatusBadRequest {
		t.Fatalf("status = %d", err.Status)
	}
	if err.Type != "invalid_request_error" {
		t.Fatalf("type = %q", err.Type)
	}
	if err.Message != "test error" {
		t.Fatalf("message = %q", err.Message)
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("unauthorized")
	if err.Status != http.StatusUnauthorized {
		t.Fatalf("status = %d", err.Status)
	}
	if err.Type != "authentication_error" {
		t.Fatalf("type = %q", err.Type)
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden("forbidden")
	if err.Status != http.StatusForbidden {
		t.Fatalf("status = %d", err.Status)
	}
	if err.Type != "permission_error" {
		t.Fatalf("type = %q", err.Type)
	}
}

func TestServerError(t *testing.T) {
	err := ServerError("server error")
	if err.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d", err.Status)
	}
	if err.Type != "server_error" {
		t.Fatalf("type = %q", err.Type)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := APIError{Message: "test message"}
	if err.Error() != "test message" {
		t.Fatalf("error() = %q", err.Error())
	}
}

func TestResolveProxyCredential_Empty(t *testing.T) {
	key, err := resolveProxyCredential(nil)
	if err != nil {
		t.Fatalf("nil request: %v", err)
	}
	if key != "" {
		t.Fatalf("key = %q", key)
	}
}

func TestResolveProxyCredential_Authorization(t *testing.T) {
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Authorization", "Bearer valid-token")

	key, err := resolveProxyCredential(req)
	if err != nil {
		t.Fatalf("valid bearer: %v", err)
	}
	if key != "valid-token" {
		t.Fatalf("key = %q, want %q", key, "valid-token")
	}
}

func TestResolveProxyCredential_XAPIKey(t *testing.T) {
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("X-API-Key", "x-api-key")

	key, err := resolveProxyCredential(req)
	if err != nil {
		t.Fatalf("x-api-key: %v", err)
	}
	if key != "x-api-key" {
		t.Fatalf("key = %q", key)
	}
}

func TestResolveProxyCredential_MalformedAuthorization(t *testing.T) {
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Authorization", "not-bearer-format")

	_, err := resolveProxyCredential(req)
	if err == nil {
		t.Fatal("expected error for malformed authorization")
	}
	if !strings.Contains(err.Error(), "malformed") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestResolveProxyCredential_ConflictingHeaders(t *testing.T) {
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Authorization", "Bearer token-a")
	req.Header.Set("X-API-Key", "token-b")

	_, err := resolveProxyCredential(req)
	if err == nil {
		t.Fatal("expected error for conflicting headers")
	}
	if !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestResolveProxyCredential_MatchingHeaders(t *testing.T) {
	req := &http.Request{
		Header: http.Header{},
	}
	req.Header.Set("Authorization", "Bearer same-token")
	req.Header.Set("X-API-Key", "same-token")

	key, err := resolveProxyCredential(req)
	if err != nil {
		t.Fatalf("matching headers: %v", err)
	}
	if key != "same-token" {
		t.Fatalf("key = %q", key)
	}
}

func TestValidateSecurityHeaders_NilRequest(t *testing.T) {
	srv := newAuthzTestServer(t)
	err := srv.validateSecurityHeaders(nil)
	if err.Status != http.StatusBadRequest {
		t.Fatalf("status = %d", err.Status)
	}
}

func TestValidateSecurityHeaders_NilStore(t *testing.T) {
	srv := newAuthzTestServer(t)
	srv.store = nil
	req, _ := http.NewRequest("GET", "/", nil)
	err := srv.validateSecurityHeaders(req)
	if err.Status != http.StatusInternalServerError {
		t.Fatalf("status = %d", err.Status)
	}
}

func TestValidateSecurityHeaders_NoAuthMode(t *testing.T) {
	srv := newAuthzTestServer(t)
	req, _ := http.NewRequest("GET", "/", nil)
	err := srv.validateSecurityHeaders(req)
	if err.Status != 0 {
		t.Fatalf("should pass when auth mode is disabled, got status %d", err.Status)
	}
}

func TestValidateSecurityHeaders_MissingKey(t *testing.T) {
	srv := newAuthzTestServer(t)
	if err := srv.store.SetProxyAPIKey("secret-key"); err != nil {
		t.Fatalf("set key: %v", err)
	}
	if err := srv.store.SetAuthorizationMode(true); err != nil {
		t.Fatalf("set auth mode: %v", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	err := srv.validateSecurityHeaders(req)
	if err.Status != http.StatusUnauthorized {
		t.Fatalf("status = %d", err.Status)
	}
}

func TestValidateSecurityHeaders_InvalidKey(t *testing.T) {
	srv := newAuthzTestServer(t)
	if err := srv.store.SetProxyAPIKey("secret-key"); err != nil {
		t.Fatalf("set key: %v", err)
	}
	if err := srv.store.SetAuthorizationMode(true); err != nil {
		t.Fatalf("set auth mode: %v", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer wrong-key")
	err := srv.validateSecurityHeaders(req)
	if err.Status != http.StatusUnauthorized {
		t.Fatalf("status = %d", err.Status)
	}
}

func TestValidateSecurityHeaders_ValidKey(t *testing.T) {
	srv := newAuthzTestServer(t)
	if err := srv.store.SetProxyAPIKey("secret-key"); err != nil {
		t.Fatalf("set key: %v", err)
	}
	if err := srv.store.SetAuthorizationMode(true); err != nil {
		t.Fatalf("set auth mode: %v", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer secret-key")
	err := srv.validateSecurityHeaders(req)
	if err.Status != 0 {
		t.Fatalf("should pass with valid key, got status %d", err.Status)
	}
}

func TestValidateSecurityHeaders_KeyNotConfigured(t *testing.T) {
	srv := newAuthzTestServer(t)
	if err := srv.store.SetAuthorizationMode(true); err != nil {
		t.Fatalf("set auth mode: %v", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer some-key")
	err := srv.validateSecurityHeaders(req)
	if err.Status != http.StatusForbidden {
		t.Fatalf("status = %d", err.Status)
	}
}
