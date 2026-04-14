package config

import (
	"strings"
	"testing"
)

func TestNormalizeAuthCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"account_deactivated", "account_deactivated"},
		{"ACCOUNT_DEACTIVATED", "account_deactivated"},
		{"Account-Deactivated", "account_deactivated"},
		{"Account Deactivated", "account_deactivated"},
		{"account  deactivated", "account_deactivated"},
		{"  spaced  ", "spaced"},
		{"", ""},
	}

	for _, tc := range tests {
		got := normalizeAuthCode(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeAuthCode(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestIsKnownAuthCode(t *testing.T) {
	// Blocked codes
	if !isKnownAuthCode("account_deactivated") {
		t.Error("account_deactivated should be known")
	}
	if !isKnownAuthCode("ACCOUNT_SUSPENDED") {
		t.Error("ACCOUNT_SUSPENDED should be known (case insensitive)")
	}

	// Refreshable codes
	if !isKnownAuthCode("refresh_token_reused") {
		t.Error("refresh_token_reused should be known")
	}
	if !isKnownAuthCode("invalid_grant") {
		t.Error("invalid_grant should be known")
	}

	// Unknown codes
	if isKnownAuthCode("unknown_code") {
		t.Error("unknown_code should not be known")
	}
}

func TestIsBlockedAuthCode(t *testing.T) {
	blocked := []string{
		"account_deactivated",
		"account_disabled",
		"account_suspended",
		"user_deactivated",
		"org_deactivated",
	}
	for _, code := range blocked {
		if !isBlockedAuthCode(code) {
			t.Errorf("%q should be a blocked code", code)
		}
	}

	if isBlockedAuthCode("refresh_token_reused") {
		t.Error("refresh_token_reused should not be a blocked code")
	}
}

func TestIsRefreshableAuthCode(t *testing.T) {
	refreshable := []string{
		"refresh_token_reused",
		"refresh_token_invalid",
		"invalid_grant",
		"expired_token",
		"invalid_api_key",
		"authentication_required",
	}
	for _, code := range refreshable {
		if !isRefreshableAuthCode(code) {
			t.Errorf("%q should be a refreshable code", code)
		}
	}

	if isRefreshableAuthCode("account_deactivated") {
		t.Error("account_deactivated should not be a refreshable code")
	}
}

func TestExtractStatusCode(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"quota request failed (401): ...", 401},
		{"refresh token failed (403): ...", 403},
		{"error (500): ...", 500},
		{"no status code here", 0},
		{"(invalid)", 0},
		{"(1234)", 0},
		{"", 0},
	}

	for _, tc := range tests {
		got := extractStatusCode(tc.input)
		if got != tc.expected {
			t.Errorf("extractStatusCode(%q) = %d, want %d", tc.input, got, tc.expected)
		}
	}
}

func TestExtractInlineJSON(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"prefix {\"code\": \"test\"} suffix", "{\"code\": \"test\"}"},
		{"{\"error\": {}}", "{\"error\": {}}"},
		{"no json here", ""},
		{"{\"a\": 1}{\"b\": 2}", "{\"a\": 1}{\"b\": 2}"},
	}

	for _, tc := range tests {
		got := extractInlineJSON(tc.input)
		if got != tc.expected {
			t.Errorf("extractInlineJSON(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestExtractInlineMessage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"code: message text", "message text"},
		{"plain message", "plain message"},
		{"  spaced  ", "spaced"},
	}

	for _, tc := range tests {
		got := extractInlineMessage(tc.input)
		if got != tc.expected {
			t.Errorf("extractInlineMessage(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestStripLeadingAuthCode(t *testing.T) {
	tests := []struct {
		message string
		code    string
		want    string
	}{
		{"account_deactivated: Deactivated", "account_deactivated", "Deactivated"},
		{"ACCOUNT_DEACTIVATED: Deactivated", "account_deactivated", "Deactivated"},
		{"no colon here", "account_deactivated", "no colon here"},
	}

	for _, tc := range tests {
		got := stripLeadingAuthCode(tc.message, tc.code)
		if got != tc.want {
			t.Errorf("stripLeadingAuthCode(%q, %q) = %q, want %q", tc.message, tc.code, got, tc.want)
		}
	}
}

func TestHumanizeAuthCode(t *testing.T) {
	if humanizeAuthCode("account_deactivated") != "account deactivated" {
		t.Error("humanizeAuthCode failed for account_deactivated")
	}
	if humanizeAuthCode("REFRESH_TOKEN_REUSED") != "refresh token reused" {
		t.Error("humanizeAuthCode should be case insensitive")
	}
}

func TestCompactAuthMessage(t *testing.T) {
	short := "short message"
	if compactAuthMessage(short) != short {
		t.Error("short message should not be compacted")
	}

	long := strings.Repeat("a", 200)
	compact := compactAuthMessage(long)
	if len(compact) > 183 {
		t.Errorf("compacted message too long: %d chars", len(compact))
	}
	if !strings.HasSuffix(compact, "...") {
		t.Error("compacted message should have ... suffix")
	}
}

func TestAuthReasonMessage_Fallback(t *testing.T) {
	signal := authErrorSignal{}
	msg := authReasonMessage(signal, "Custom Fallback")
	if msg != "Custom Fallback" {
		t.Fatalf("message = %q, want %q", msg, "Custom Fallback")
	}
}

func TestAuthReasonMessage_Default(t *testing.T) {
	signal := authErrorSignal{}
	msg := authReasonMessage(signal, "  ")
	if msg != "Authentication required" {
		t.Fatalf("message = %q, want %q", msg, "Authentication required")
	}
}
