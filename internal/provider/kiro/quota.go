package kiro

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cliro-go/internal/config"
)

const (
	kiroQuotaBaseURL = "https://codewhisperer.us-east-1.amazonaws.com"
)

type QuotaFetcher struct {
	httpClient *http.Client
}

func NewQuotaFetcher(httpClient *http.Client) *QuotaFetcher {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 25 * time.Second}
	}
	return &QuotaFetcher{httpClient: httpClient}
}

func (f *QuotaFetcher) FetchQuota(ctx context.Context, account config.Account, refreshCallback func(string) (config.Account, error)) (config.QuotaInfo, string, error) {
	currentAccount := account
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, kiroQuotaBaseURL+"/getUsageLimits?origin=AI_EDITOR&resourceType=AGENTIC_REQUEST&isEmailRequired=true", nil)
		if err != nil {
			return synthesizeQuota(currentAccount, err), "", err
		}
		applyKiroQuotaHeaders(req, currentAccount.AccessToken)

		resp, err := f.httpClient.Do(req)
		if err != nil {
			return synthesizeQuota(currentAccount, err), "", err
		}

		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			data, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if attempt == 0 && refreshCallback != nil {
				refreshedAccount, refreshErr := refreshCallback(currentAccount.ID)
				if refreshErr == nil {
					currentAccount = refreshedAccount
					continue
				}
			}
			err = fmt.Errorf("kiro quota request failed (%d): %s", resp.StatusCode, compactHTTPBody(data))
			return synthesizeQuota(currentAccount, err), "", err
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			data, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			err = fmt.Errorf("kiro quota request failed (%d): %s", resp.StatusCode, compactHTTPBody(data))
			return synthesizeQuota(currentAccount, err), "", err
		}

		var payload struct {
			UsageBreakdownList []struct {
				CurrentUsage float64 `json:"currentUsage"`
				UsageLimit   float64 `json:"usageLimit"`
			} `json:"usageBreakdownList"`
			SubscriptionInfo struct {
				SubscriptionName  string `json:"subscriptionName"`
				SubscriptionTitle string `json:"subscriptionTitle"`
			} `json:"subscriptionInfo"`
			UserInfo struct {
				Email string `json:"email"`
			} `json:"userInfo"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			_ = resp.Body.Close()
			return synthesizeQuota(currentAccount, err), "", err
		}
		_ = resp.Body.Close()

		used := 0
		total := 0
		if len(payload.UsageBreakdownList) > 0 {
			used = int(payload.UsageBreakdownList[0].CurrentUsage)
			total = int(payload.UsageBreakdownList[0].UsageLimit)
		}
		remaining := 0
		if total > 0 {
			remaining = maxInt(total-used, 0)
		}
		percent := 0
		if total > 0 {
			percent = int(float64(remaining) / float64(total) * 100)
		}

		bucket := config.QuotaBucket{
			Name:      "credits",
			Used:      used,
			Total:     total,
			Remaining: remaining,
			Percent:   percent,
			Status:    bucketStatus(config.QuotaBucket{Used: used, Total: total, Remaining: remaining}),
		}
		status := bucket.Status
		if status == "" {
			status = "healthy"
		}
		summary := firstNonEmpty(
			strings.TrimSpace(payload.SubscriptionInfo.SubscriptionTitle),
			strings.TrimSpace(payload.SubscriptionInfo.SubscriptionName),
		)
		if summary == "" && total > 0 {
			summary = fmt.Sprintf("%d/%d credits remaining", remaining, total)
		}
		if summary == "" {
			summary = "Kiro usage data loaded"
		}

		resolvedEmail := strings.TrimSpace(payload.UserInfo.Email)
		if resolvedEmail == "" {
			email, err := f.fetchUserEmail(ctx, currentAccount, refreshCallback)
			if err == nil {
				resolvedEmail = strings.TrimSpace(email)
			}
		}

		return config.QuotaInfo{
			Status:        status,
			Summary:       summary,
			Source:        "kiro/getUsageLimits",
			LastCheckedAt: time.Now().Unix(),
			Buckets:       []config.QuotaBucket{bucket},
		}, resolvedEmail, nil
	}

	return synthesizeQuota(currentAccount, fmt.Errorf("kiro quota request failed")), "", fmt.Errorf("kiro quota request failed")
}

func (f *QuotaFetcher) fetchUserEmail(ctx context.Context, account config.Account, refreshCallback func(string) (config.Account, error)) (string, error) {
	currentAccount := account
	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, kiroQuotaBaseURL+"/GetUserInfo", strings.NewReader(`{"origin":"KIRO_IDE"}`))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		applyKiroQuotaHeaders(req, currentAccount.AccessToken)

		resp, err := f.httpClient.Do(req)
		if err != nil {
			return "", err
		}

		data, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return "", err
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			if attempt == 0 && refreshCallback != nil {
				refreshedAccount, refreshErr := refreshCallback(currentAccount.ID)
				if refreshErr == nil {
					currentAccount = refreshedAccount
					continue
				}
			}
			return "", fmt.Errorf("kiro user info request failed (%d): %s", resp.StatusCode, compactHTTPBody(data))
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("kiro user info request failed (%d): %s", resp.StatusCode, compactHTTPBody(data))
		}

		var payload struct {
			Email    string `json:"email"`
			UserInfo struct {
				Email string `json:"email"`
			} `json:"userInfo"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return "", err
		}

		resolved := strings.TrimSpace(firstNonEmpty(payload.Email, payload.UserInfo.Email))
		if resolved == "" {
			return "", fmt.Errorf("kiro user info response missing email")
		}
		return resolved, nil
	}

	return "", fmt.Errorf("kiro user info request failed")
}

func applyKiroQuotaHeaders(req *http.Request, accessToken string) {
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", kiroRuntimeUserAgent)
	req.Header.Set("x-amz-user-agent", kiroRuntimeAmzUserAgent)
}

func synthesizeQuota(account config.Account, err error) config.QuotaInfo {
	now := time.Now().Unix()
	info := config.QuotaInfo{
		Status:        "healthy",
		Summary:       "Quota endpoint not resolved yet; using local runtime state.",
		Source:        "runtime",
		LastCheckedAt: now,
	}
	if err != nil {
		info.Error = err.Error()
		info.Status = "unknown"
	}
	if account.CooldownUntil > now {
		info.Status = "exhausted"
		info.Summary = firstNonEmpty(account.LastError, "Quota cooldown is active.")
		info.Buckets = []config.QuotaBucket{{
			Name:    "session",
			ResetAt: account.CooldownUntil,
			Status:  "exhausted",
		}, {
			Name:   "weekly",
			Status: "unknown",
		}}
		return info
	}
	if account.LastError != "" {
		info.Status = "degraded"
		info.Summary = account.LastError
	}
	if len(account.Quota.Buckets) > 0 {
		info.Buckets = append([]config.QuotaBucket(nil), account.Quota.Buckets...)
	}
	return info
}

func bucketStatus(bucket config.QuotaBucket) string {
	if bucket.Status != "" {
		status := normalizeQuotaStatus(bucket.Status)
		if status != "" {
			return status
		}
		return strings.ToLower(strings.TrimSpace(bucket.Status))
	}
	now := time.Now().Unix()
	if bucket.Total > 0 {
		remaining := bucket.Remaining
		if remaining == 0 && bucket.Used > 0 && bucket.Used <= bucket.Total {
			remaining = maxInt(bucket.Total-bucket.Used, 0)
		}
		if remaining <= 0 {
			return "exhausted"
		}
		remainingPercent := int(float64(remaining) / float64(bucket.Total) * 100)
		if remainingPercent <= 15 {
			return "low"
		}
		return "healthy"
	}
	if bucket.ResetAt > now {
		if bucket.Remaining <= 0 {
			return "exhausted"
		}
		return "low"
	}
	return "unknown"
}

func normalizeQuotaStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "ready", "healthy", "ok":
		return "healthy"
	case "expiring", "warning":
		return "low"
	case "expired", "exhausted", "quota_exceeded", "insufficient_quota":
		return "exhausted"
	case "", "unknown":
		return ""
	default:
		return status
	}
}

func compactHTTPBody(body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "empty response"
	}
	if len(trimmed) > 180 {
		return trimmed[:180] + "..."
	}
	return trimmed
}
