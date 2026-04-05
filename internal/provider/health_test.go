package provider

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"cliro-go/internal/config"
)

func TestClassifyHTTPFailure(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		message string
		class   FailureClass
		resp    int
	}{
		{
			name:    "refreshable auth",
			status:  http.StatusUnauthorized,
			message: `{"message":"The bearer token included in the request is invalid.","reason":null}`,
			class:   FailureAuthRefreshable,
			resp:    http.StatusUnauthorized,
		},
		{
			name:    "request shape",
			status:  http.StatusBadRequest,
			message: "Invalid value: 'input_text'. Supported values are: 'output_text' and 'refusal'.",
			class:   FailureRequestShape,
			resp:    http.StatusBadGateway,
		},
		{
			name:    "quota cooldown",
			status:  http.StatusTooManyRequests,
			message: "rate limit reached",
			class:   FailureQuotaCooldown,
			resp:    http.StatusTooManyRequests,
		},
		{
			name:    "retryable upstream",
			status:  http.StatusServiceUnavailable,
			message: "upstream unavailable",
			class:   FailureRetryableTransport,
			resp:    http.StatusBadGateway,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decision := ClassifyHTTPFailure(tc.status, tc.message)
			if decision.Class != tc.class {
				t.Fatalf("class = %q", decision.Class)
			}
			if decision.Status != tc.resp {
				t.Fatalf("status = %d", decision.Status)
			}
		})
	}
}

func TestClassifyTransportFailure(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		class FailureClass
		resp  int
	}{
		{name: "nil error", class: FailureRetryableTransport, resp: http.StatusBadGateway},
		{name: "transport timeout", err: errors.New("timeout contacting upstream"), class: FailureRetryableTransport, resp: http.StatusBadGateway},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decision := ClassifyTransportFailure(tc.err)
			if decision.Class != tc.class {
				t.Fatalf("class = %q", decision.Class)
			}
			if decision.Status != tc.resp {
				t.Fatalf("status = %d", decision.Status)
			}
		})
	}
}

func TestTransientCooldown_UsesFixedPolicy(t *testing.T) {
	if got := TransientCooldown(1); got != 10*time.Second {
		t.Fatalf("first cooldown = %s", got)
	}
	if got := TransientCooldown(2); got != 30*time.Second {
		t.Fatalf("second cooldown = %s", got)
	}
	if got := TransientCooldown(5); got != 60*time.Second {
		t.Fatalf("overflow cooldown = %s", got)
	}
}

func TestSynthesizeQuota(t *testing.T) {
	now := time.Now().Unix()

	tests := []struct {
		name    string
		account config.Account
		err     error
		status  string
	}{
		{name: "healthy default", account: config.Account{}, status: "healthy"},
		{name: "unknown with error", account: config.Account{}, err: errors.New("boom"), status: "unknown"},
		{name: "cooldown exhausted", account: config.Account{CooldownUntil: now + 60, LastError: "cooldown active"}, status: "exhausted"},
		{name: "degraded from last error", account: config.Account{LastError: "stale quota"}, status: "degraded"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			quota := SynthesizeQuota(tc.account, tc.err)
			if quota.Status != tc.status {
				t.Fatalf("status = %q", quota.Status)
			}
		})
	}
}

func TestNormalizeQuotaStatus(t *testing.T) {
	tests := map[string]string{
		"ready":              "healthy",
		"warning":            "low",
		"insufficient_quota": "exhausted",
		"custom":             "custom",
		"":                   "",
	}

	for input, want := range tests {
		if got := NormalizeQuotaStatus(input); got != want {
			t.Fatalf("NormalizeQuotaStatus(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBucketStatus(t *testing.T) {
	tests := []struct {
		name   string
		bucket config.QuotaBucket
		want   string
	}{
		{name: "normalized explicit status", bucket: config.QuotaBucket{Status: "warning"}, want: "low"},
		{name: "derived exhausted", bucket: config.QuotaBucket{Total: 10, Used: 10}, want: "exhausted"},
		{name: "derived healthy", bucket: config.QuotaBucket{Total: 10, Remaining: 8}, want: "healthy"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BucketStatus(tc.bucket); got != tc.want {
				t.Fatalf("BucketStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCompactHTTPBody(t *testing.T) {
	if got := CompactHTTPBody([]byte("   ")); got != "empty response" {
		t.Fatalf("empty body = %q", got)
	}
	if got := CompactHTTPBody([]byte(" body ")); got != "body" {
		t.Fatalf("trimmed body = %q", got)
	}
	if got := CompactHTTPBody([]byte(`{"error":{"message":"Your refresh token has already been used to generate a new access token.","code":"refresh_token_reused"}}`)); got != "refresh_token_reused: Your refresh token has already been used to generate a new access token." {
		t.Fatalf("json body = %q", got)
	}
}
