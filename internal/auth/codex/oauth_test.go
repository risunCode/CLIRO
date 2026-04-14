package codex

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"
	"time"

	"cliro/internal/auth/shared"
	"cliro/internal/config"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("GenerateCodeVerifier: %v", err)
	}
	if verifier == "" {
		t.Fatal("verifier is empty")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(verifier)
	if err != nil {
		t.Fatalf("verifier is not valid base64url: %v", err)
	}
	if len(decoded) != 48 {
		t.Fatalf("decoded verifier length = %d, want 48", len(decoded))
	}
}

func TestGenerateCodeVerifier_Unique(t *testing.T) {
	a, _ := GenerateCodeVerifier()
	b, _ := GenerateCodeVerifier()
	if a == b {
		t.Fatal("two consecutive verifiers should not be equal")
	}
}

func TestCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := CodeChallenge(verifier)
	decoded, err := base64.RawURLEncoding.DecodeString(challenge)
	if err != nil {
		t.Fatalf("challenge is not valid base64url: %v", err)
	}
	if len(decoded) != 32 {
		t.Fatalf("challenge decoded length = %d, want 32 (SHA-256)", len(decoded))
	}
}

func TestCodeChallenge_Deterministic(t *testing.T) {
	verifier := "test-verifier-value"
	a := CodeChallenge(verifier)
	b := CodeChallenge(verifier)
	if a != b {
		t.Fatal("CodeChallenge should be deterministic for the same verifier")
	}
}

func TestBuildAuthURL(t *testing.T) {
	authURL := BuildAuthURL("state-123", "challenge-abc")
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth URL: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "auth.openai.com" {
		t.Fatalf("unexpected base URL: %s", authURL)
	}
	query := parsed.Query()
	if query.Get("client_id") != ClientID {
		t.Fatalf("client_id = %q", query.Get("client_id"))
	}
	if query.Get("state") != "state-123" {
		t.Fatalf("state = %q", query.Get("state"))
	}
	if query.Get("code_challenge") != "challenge-abc" {
		t.Fatalf("code_challenge = %q", query.Get("code_challenge"))
	}
	if query.Get("code_challenge_method") != "S256" {
		t.Fatalf("code_challenge_method = %q", query.Get("code_challenge_method"))
	}
	if query.Get("response_type") != "code" {
		t.Fatalf("response_type = %q", query.Get("response_type"))
	}
	if query.Get("originator") != "opencode" {
		t.Fatalf("originator = %q", query.Get("originator"))
	}
}

func TestParseIDToken_Valid(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{
		"email": "user@example.com",
		"exp": 1735689600,
		"https://api.openai.com/auth": {
			"chatgpt_account_id": "acct_123",
			"chatgpt_plan_type": "plus"
		}
	}`))
	token := header + "." + payload + ".signature"

	claims, err := ParseIDToken(token)
	if err != nil {
		t.Fatalf("ParseIDToken: %v", err)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("email = %q", claims.Email)
	}
	if claims.Exp != 1735689600 {
		t.Fatalf("exp = %d", claims.Exp)
	}
	if claims.CodexAuthInfo.ChatgptAccountID != "acct_123" {
		t.Fatalf("account_id = %q", claims.CodexAuthInfo.ChatgptAccountID)
	}
	if claims.CodexAuthInfo.ChatgptPlanType != "plus" {
		t.Fatalf("plan_type = %q", claims.CodexAuthInfo.ChatgptPlanType)
	}
}

func TestParseIDToken_InvalidFormat(t *testing.T) {
	_, err := ParseIDToken("not-a-jwt")
	if err == nil {
		t.Fatal("expected error for invalid JWT format")
	}
	if !strings.Contains(err.Error(), "invalid id token") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseIDToken_InvalidBase64(t *testing.T) {
	_, err := ParseIDToken("header.!!!.signature")
	if err == nil {
		t.Fatal("expected error for invalid base64 payload")
	}
}

func TestParseIDToken_InvalidJSON(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`not-json`))
	_, err := ParseIDToken(header + "." + payload + ".sig")
	if err == nil {
		t.Fatal("expected error for invalid JSON payload")
	}
}

func TestRenderCallbackPage(t *testing.T) {
	page := RenderCallbackPage("Test Title", "Test Message")
	if !strings.Contains(page, "Test Title") {
		t.Fatal("page should contain title")
	}
	if !strings.Contains(page, "Test Message") {
		t.Fatal("page should contain message")
	}
	if !strings.Contains(page, "<!doctype html>") {
		t.Fatal("page should be valid HTML")
	}
	if !strings.Contains(page, "CLIRO") {
		t.Fatal("page should contain CLIRO branding")
	}
}

func TestTokenExpired_Expired(t *testing.T) {
	account := config.Account{ExpiresAt: 1}
	if !shared.TokenExpired(account, time.Now()) {
		t.Fatal("expected token to be expired")
	}
}

func TestTokenExpired_NotExpired(t *testing.T) {
	account := config.Account{ExpiresAt: time.Now().Add(1 * time.Hour).Unix()}
	if shared.TokenExpired(account, time.Now()) {
		t.Fatal("expected token to not be expired")
	}
}

func TestTokenExpired_ZeroExpiresAt(t *testing.T) {
	account := config.Account{ExpiresAt: 0}
	if shared.TokenExpired(account, time.Now()) {
		t.Fatal("zero ExpiresAt should not be considered expired")
	}
}
