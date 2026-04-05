package config

// health.go — account health helpers: label formatting and quota reset time.
// Auth error signal parsing lives in auth_errors.go.

import "strings"

func AccountLabel(account Account) string {
	if email := strings.TrimSpace(account.Email); email != "" {
		return email
	}
	if accountID := strings.TrimSpace(account.AccountID); accountID != "" {
		return accountID
	}
	return strings.TrimSpace(account.ID)
}

func QuotaResetAt(quota QuotaInfo) int64 {
	var latest int64
	for _, bucket := range quota.Buckets {
		if bucket.ResetAt > latest {
			latest = bucket.ResetAt
		}
	}
	return latest
}
