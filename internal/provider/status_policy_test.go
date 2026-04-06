package provider

import (
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"cliro/internal/account"
	"cliro/internal/config"
)

func TestRetryPlanner_DecideRefreshOnceForAuthFailure(t *testing.T) {
	planner := NewRetryPlanner(nil, "codex", nil)
	decision := planner.Decide(AttemptResult{
		Status:      401,
		Message:     "token expired",
		Failure:     FailureDecision{Class: FailureAuthRefreshable, Status: 401, Message: "token expired", RetryAllowed: true},
		Attempt:     1,
		Success:     false,
		Final:       false,
		EmptyStream: false,
	})
	if !decision.RefreshAuth || !decision.Retry {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestRetryPlanner_DecideTransportRetryBeforeClientBytes(t *testing.T) {
	planner := NewRetryPlanner(nil, "kiro", nil)
	decision := planner.Decide(AttemptResult{
		Status:          502,
		Failure:         ClassifyTransportFailure(errors.New("boom")),
		ClientBytesSent: false,
		RetryCause:      "transport",
	})
	if !decision.Retry || !decision.ExcludeAccount {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestRetryPlanner_DecideNoRetryForRequestShape(t *testing.T) {
	planner := NewRetryPlanner(nil, "kiro", nil)
	decision := planner.Decide(AttemptResult{
		Status:  400,
		Failure: FailureDecision{Class: FailureRequestShape, Status: 400, Message: "bad request"},
	})
	if decision.Retry || decision.RefreshAuth {
		t.Fatalf("decision = %#v", decision)
	}
}

func TestRetryPlanner_DecideEmptyStreamRetryOnlyBeforeClientBytes(t *testing.T) {
	planner := NewRetryPlanner(nil, "kiro", nil)
	first := planner.Decide(AttemptResult{
		Status:          502,
		Failure:         FailureDecision{Class: FailureEmptyStream, Status: 502, Message: "empty stream", RetryAllowed: true},
		EmptyStream:     true,
		ClientBytesSent: false,
	})
	if !first.Retry {
		t.Fatalf("expected retry for empty stream without client bytes: %#v", first)
	}
	second := planner.Decide(AttemptResult{
		Status:          502,
		Failure:         FailureDecision{Class: FailureEmptyStream, Status: 502, Message: "empty stream", RetryAllowed: true},
		EmptyStream:     true,
		ClientBytesSent: true,
	})
	if second.Retry {
		t.Fatalf("expected no retry after client bytes: %#v", second)
	}
}

func TestRetryPlanner_NextAccountSkipsExcludedAccounts(t *testing.T) {
	store, err := config.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	_ = store.UpsertAccount(config.Account{ID: "a1", Provider: "codex", Email: "a1@example.com", Enabled: true, HealthState: config.AccountHealthReady})
	_ = store.UpsertAccount(config.Account{ID: "a2", Provider: "codex", Email: "a2@example.com", Enabled: true, HealthState: config.AccountHealthReady})
	planner := NewRetryPlanner(account.NewPool(store), "codex", nil)
	account, ok := planner.NextAccount(map[string]bool{"a1": true})
	if !ok {
		t.Fatalf("expected account")
	}
	if account.ID != "a2" {
		t.Fatalf("account id = %q, want a2", account.ID)
	}
}

func TestStreamBridge_OpenVerified(t *testing.T) {
	bridge := StreamBridge{}
	probe, err := bridge.OpenVerified(io.NopCloser(strings.NewReader("hello")), 50*time.Millisecond)
	if err != nil {
		t.Fatalf("OpenVerified: %v", err)
	}
	if !probe.UpstreamReadable {
		t.Fatalf("probe = %#v", probe)
	}
	data, err := io.ReadAll(probe.Reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("data = %q", string(data))
	}
}

func TestStreamBridge_OpenVerifiedImmediateEOF(t *testing.T) {
	bridge := StreamBridge{}
	_, err := bridge.OpenVerified(io.NopCloser(strings.NewReader("")), 50*time.Millisecond)
	if !errors.Is(err, ErrEmptyStream) {
		t.Fatalf("err = %v, want ErrEmptyStream", err)
	}
}

func TestStreamBridge_OpenVerifiedReadError(t *testing.T) {
	bridge := StreamBridge{}
	_, err := bridge.OpenVerified(errorReadCloser{err: io.ErrUnexpectedEOF}, 50*time.Millisecond)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("err = %v, want io.ErrUnexpectedEOF", err)
	}
}

type errorReadCloser struct{ err error }

func (e errorReadCloser) Read(_ []byte) (int, error) { return 0, e.err }
func (e errorReadCloser) Close() error               { return nil }
