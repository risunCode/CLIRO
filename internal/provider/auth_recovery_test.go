package provider

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cliro/internal/config"
)

func TestAuthRecoveryCoordinator_CoalescesConcurrentRefreshes(t *testing.T) {
	var calls atomic.Int32
	coordinator := NewAuthRecoveryCoordinator(func(accountID string) (config.Account, error) {
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return config.Account{ID: accountID, Enabled: true}, nil
	}, 2)

	var wg sync.WaitGroup
	statuses := make(chan RecoveryStatus, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, status, err := coordinator.Recover(context.Background(), "codex", "acct-1")
			if err != nil {
				t.Errorf("Recover: %v", err)
				return
			}
			statuses <- status
		}()
	}
	wg.Wait()
	close(statuses)

	if got := calls.Load(); got != 1 {
		t.Fatalf("refresh calls = %d, want 1", got)
	}
	seenWaiting := false
	for status := range statuses {
		if status == RecoveryStatusWaiting {
			seenWaiting = true
		}
	}
	if !seenWaiting {
		t.Fatalf("expected one waiting status")
	}
}

func TestAuthRecoveryCoordinator_BoundsConcurrentRefreshes(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{}, 2)
	coordinator := NewAuthRecoveryCoordinator(func(accountID string) (config.Account, error) {
		started <- accountID
		<-release
		return config.Account{ID: accountID, Enabled: true}, nil
	}, 1)

	go func() {
		_, _, _ = coordinator.Recover(context.Background(), "codex", "acct-1")
	}()
	first := <-started
	if first != "acct-1" {
		t.Fatalf("first account = %q", first)
	}

	done := make(chan struct{})
	go func() {
		_, _, _ = coordinator.Recover(context.Background(), "codex", "acct-2")
		close(done)
	}()

	select {
	case second := <-started:
		t.Fatalf("second refresh started early for %q", second)
	case <-time.After(15 * time.Millisecond):
	}

	release <- struct{}{}
	select {
	case second := <-started:
		if second != "acct-2" {
			t.Fatalf("second account = %q", second)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("second refresh did not start after release")
	}
	release <- struct{}{}
	<-done
}
