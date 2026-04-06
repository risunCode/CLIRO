package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"cliro/internal/config"
)

type RecoveryStatus string

const (
	RecoveryStatusRefreshed RecoveryStatus = "refreshed"
	RecoveryStatusWaiting   RecoveryStatus = "waiting_for_peer"
	RecoveryStatusRejected  RecoveryStatus = "rejected"
)

type recoveryCall struct {
	done   chan struct{}
	value  config.Account
	err    error
	status RecoveryStatus
	ready  bool
}

type AuthRecoveryCoordinator struct {
	refresh func(accountID string) (config.Account, error)
	mu      sync.Mutex
	active  map[string]*recoveryCall
	sem     chan struct{}
}

func NewAuthRecoveryCoordinator(refresh func(accountID string) (config.Account, error), maxConcurrent int) *AuthRecoveryCoordinator {
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}
	return &AuthRecoveryCoordinator{
		refresh: refresh,
		active:  make(map[string]*recoveryCall),
		sem:     make(chan struct{}, maxConcurrent),
	}
}

func (c *AuthRecoveryCoordinator) Recover(ctx context.Context, provider string, accountID string) (config.Account, RecoveryStatus, error) {
	if c == nil || c.refresh == nil {
		return config.Account{}, RecoveryStatusRejected, fmt.Errorf("auth recovery is unavailable")
	}
	trimmedID := strings.TrimSpace(accountID)
	if trimmedID == "" {
		return config.Account{}, RecoveryStatusRejected, fmt.Errorf("account id is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	c.mu.Lock()
	if existing, ok := c.active[trimmedID]; ok {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return config.Account{}, RecoveryStatusRejected, ctx.Err()
		case <-existing.done:
			return existing.value, RecoveryStatusWaiting, existing.err
		}
	}

	call := &recoveryCall{done: make(chan struct{})}
	c.active[trimmedID] = call
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.active, trimmedID)
		c.mu.Unlock()
		close(call.done)
	}()

	select {
	case c.sem <- struct{}{}:
		defer func() { <-c.sem }()
	case <-ctx.Done():
		call.err = ctx.Err()
		call.status = RecoveryStatusRejected
		call.ready = true
		return config.Account{}, call.status, call.err
	}

	account, err := c.refresh(trimmedID)
	call.value = account
	call.err = err
	call.ready = true
	if err != nil {
		call.status = RecoveryStatusRejected
		return account, call.status, err
	}
	if !account.Enabled || account.Banned {
		call.status = RecoveryStatusRejected
		return account, call.status, fmt.Errorf("%s account is not ready after refresh", firstNonEmpty(provider, "provider"))
	}
	call.status = RecoveryStatusRefreshed
	return account, call.status, nil
}
