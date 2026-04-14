package config

import (
	"testing"
)

func TestAccountLabel_Email(t *testing.T) {
	account := Account{
		ID:    "acc-123",
		Email: "user@example.com",
	}
	label := AccountLabel(account)
	if label != "user@example.com" {
		t.Fatalf("label = %q, want %q", label, "user@example.com")
	}
}

func TestAccountLabel_AccountID(t *testing.T) {
	account := Account{
		ID:        "acc-123",
		AccountID: "account-456",
	}
	label := AccountLabel(account)
	if label != "account-456" {
		t.Fatalf("label = %q, want %q", label, "account-456")
	}
}

func TestAccountLabel_ID(t *testing.T) {
	account := Account{ID: "acc-123"}
	label := AccountLabel(account)
	if label != "acc-123" {
		t.Fatalf("label = %q, want %q", label, "acc-123")
	}
}

func TestAccountLabel_Trimmed(t *testing.T) {
	account := Account{
		ID:    "acc-123",
		Email: "  user@example.com  ",
	}
	label := AccountLabel(account)
	if label != "user@example.com" {
		t.Fatalf("label = %q, want trimmed email", label)
	}
}

func TestQuotaResetAt_Empty(t *testing.T) {
	quota := QuotaInfo{Buckets: []QuotaBucket{}}
	reset := QuotaResetAt(quota)
	if reset != 0 {
		t.Fatalf("reset = %d, want 0 for empty buckets", reset)
	}
}

func TestQuotaResetAt_SingleBucket(t *testing.T) {
	quota := QuotaInfo{
		Buckets: []QuotaBucket{
			{Name: "gpt-4", ResetAt: 1234567890},
		},
	}
	reset := QuotaResetAt(quota)
	if reset != 1234567890 {
		t.Fatalf("reset = %d, want 1234567890", reset)
	}
}

func TestQuotaResetAt_MultipleBuckets(t *testing.T) {
	quota := QuotaInfo{
		Buckets: []QuotaBucket{
			{Name: "gpt-4", ResetAt: 1000},
			{Name: "gpt-4o", ResetAt: 2000},
			{Name: "o1", ResetAt: 1500},
		},
	}
	reset := QuotaResetAt(quota)
	if reset != 2000 {
		t.Fatalf("reset = %d, want 2000 (latest)", reset)
	}
}

func TestQuotaResetAt_AllZero(t *testing.T) {
	quota := QuotaInfo{
		Buckets: []QuotaBucket{
			{Name: "gpt-4", ResetAt: 0},
			{Name: "gpt-4o", ResetAt: 0},
		},
	}
	reset := QuotaResetAt(quota)
	if reset != 0 {
		t.Fatalf("reset = %d, want 0", reset)
	}
}
