package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	mgr, err := NewManager(dir)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	return mgr
}

func TestMarkAccountBanned(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{
		ID:      "test-banned",
		Email:   "test@example.com",
		Enabled: true,
	}
	if err := mgr.UpsertAccount(account); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := mgr.MarkAccountBanned("test-banned", "Account deactivated"); err != nil {
		t.Fatalf("mark banned: %v", err)
	}

	acc, ok := mgr.GetAccount("test-banned")
	if !ok {
		t.Fatal("account not found after ban")
	}
	if !acc.Banned {
		t.Fatal("account should be banned")
	}
	if acc.BannedReason != "Account deactivated" {
		t.Fatalf("banned reason = %q", acc.BannedReason)
	}
	if acc.HealthState != AccountHealthBanned {
		t.Fatalf("health state = %q", acc.HealthState)
	}
	if acc.Enabled {
		t.Fatal("banned account should be disabled")
	}
}

func TestMarkAccountBanned_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	err := mgr.MarkAccountBanned("nonexistent", "reason")
	if !os.IsNotExist(err) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestMarkAccountHealthy(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{
		ID:                  "test-healthy",
		Email:               "test@example.com",
		Banned:              true,
		BannedReason:        "was banned",
		CooldownUntil:       time.Now().Add(1 * time.Hour).Unix(),
		ConsecutiveFailures: 5,
		HealthState:         AccountHealthCooldownTransient,
		HealthReason:        "transient cooldown",
		LastError:           "some error",
	}
	if err := mgr.UpsertAccount(account); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := mgr.MarkAccountHealthy("test-healthy"); err != nil {
		t.Fatalf("mark healthy: %v", err)
	}

	acc, _ := mgr.GetAccount("test-healthy")
	if acc.Banned {
		t.Fatal("account should not be banned")
	}
	if acc.CooldownUntil != 0 {
		t.Fatal("cooldown should be cleared")
	}
	if acc.ConsecutiveFailures != 0 {
		t.Fatal("consecutive failures should be cleared")
	}
	if acc.HealthState != AccountHealthReady {
		t.Fatalf("health state = %q", acc.HealthState)
	}
}

func TestMarkAccountTransientCooldown(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{ID: "test-transient", Email: "test@example.com"}
	mgr.UpsertAccount(account)

	cooldownTime := time.Now().Add(30 * time.Second).Unix()
	if err := mgr.MarkAccountTransientCooldown("test-transient", "rate limited", cooldownTime); err != nil {
		t.Fatalf("mark transient cooldown: %v", err)
	}

	acc, _ := mgr.GetAccount("test-transient")
	if acc.HealthState != AccountHealthCooldownTransient {
		t.Fatalf("health state = %q", acc.HealthState)
	}
	if acc.CooldownUntil != cooldownTime {
		t.Fatalf("cooldown until = %d, want %d", acc.CooldownUntil, cooldownTime)
	}
	if !strings.Contains(acc.HealthReason, "rate limited") {
		t.Fatalf("health reason = %q", acc.HealthReason)
	}
}

func TestMarkAccountQuotaCooldown(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{ID: "test-quota", Email: "test@example.com"}
	mgr.UpsertAccount(account)

	cooldownTime := time.Now().Add(1 * time.Hour).Unix()
	if err := mgr.MarkAccountQuotaCooldown("test-quota", "quota exhausted", cooldownTime); err != nil {
		t.Fatalf("mark quota cooldown: %v", err)
	}

	acc, _ := mgr.GetAccount("test-quota")
	if acc.HealthState != AccountHealthCooldownQuota {
		t.Fatalf("health state = %q", acc.HealthState)
	}
}

func TestMarkAccountDurablyDisabled(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{ID: "test-disabled", Email: "test@example.com", Enabled: true}
	mgr.UpsertAccount(account)

	if err := mgr.MarkAccountDurablyDisabled("test-disabled", "manual disable"); err != nil {
		t.Fatalf("mark durably disabled: %v", err)
	}

	acc, _ := mgr.GetAccount("test-disabled")
	if acc.HealthState != AccountHealthDisabledDurable {
		t.Fatalf("health state = %q", acc.HealthState)
	}
	if acc.Enabled {
		t.Fatal("disabled account should be enabled=false")
	}
}

func TestMarkAccountReloginRequired(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{ID: "test-relogin", Email: "test@example.com"}
	mgr.UpsertAccount(account)

	if err := mgr.MarkAccountReloginRequired("test-relogin", "refresh token expired"); err != nil {
		t.Fatalf("mark relogin required: %v", err)
	}

	acc, _ := mgr.GetAccount("test-relogin")
	if acc.HealthState != AccountHealthCooldownTransient {
		t.Fatalf("health state = %q", acc.HealthState)
	}
	if acc.HealthReason != "Need re-login" {
		t.Fatalf("health reason = %q", acc.HealthReason)
	}
	if acc.Quota.Status != "unknown" {
		t.Fatalf("quota status = %q", acc.Quota.Status)
	}
	if !strings.Contains(acc.Quota.Summary, "Authentication required") {
		t.Fatalf("quota summary = %q", acc.Quota.Summary)
	}
}

func TestClearTransientCooldown_Success(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{
		ID:                  "test-clear",
		Email:               "test@example.com",
		Enabled:             true,
		CooldownUntil:       time.Now().Add(1 * time.Hour).Unix(),
		HealthState:         AccountHealthCooldownTransient,
		ConsecutiveFailures: 3,
		LastError:           "some error",
	}
	mgr.UpsertAccount(account)

	if err := mgr.ClearTransientCooldown("test-clear"); err != nil {
		t.Fatalf("clear transient cooldown: %v", err)
	}

	acc, _ := mgr.GetAccount("test-clear")
	if acc.CooldownUntil != 0 {
		t.Fatalf("cooldown should be cleared, got CooldownUntil=%d", acc.CooldownUntil)
	}
	if acc.ConsecutiveFailures != 0 {
		t.Fatal("consecutive failures should be cleared")
	}
	if acc.LastError != "" {
		t.Fatal("last error should be cleared")
	}
}

func TestClearTransientCooldown_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	err := mgr.ClearTransientCooldown("nonexistent")
	if !os.IsNotExist(err) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	acc, ok := mgr.GetAccount("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
	if acc.ID != "" {
		t.Fatal("returned account should be zero value")
	}
}

func TestDeleteAccount(t *testing.T) {
	mgr := newTestManager(t)
	account := Account{ID: "test-delete", Email: "test@example.com"}
	mgr.UpsertAccount(account)

	if err := mgr.DeleteAccount("test-delete"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, ok := mgr.GetAccount("test-delete")
	if ok {
		t.Fatal("account should be deleted")
	}
}

func TestDeleteAccount_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	if err := mgr.DeleteAccount("nonexistent"); err != nil {
		t.Fatalf("delete nonexistent should not error: %v", err)
	}
}

func TestUpdateAccount_NotFound(t *testing.T) {
	mgr := newTestManager(t)
	err := mgr.UpdateAccount("nonexistent", func(a *Account) {
		a.Email = "updated"
	})
	if !os.IsNotExist(err) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}
