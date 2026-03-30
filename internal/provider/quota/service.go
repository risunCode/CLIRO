package quota

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"cliro-go/internal/auth"
	"cliro-go/internal/config"
	"cliro-go/internal/logger"
	codexprovider "cliro-go/internal/provider/codex"
	kiroprovider "cliro-go/internal/provider/kiro"
)

const fetchTimeout = 25 * time.Second

type Service struct {
	store        *config.Manager
	auth         *auth.Manager
	log          *logger.Logger
	codexFetcher *codexprovider.QuotaFetcher
	kiroFetcher  *kiroprovider.QuotaFetcher
}

func NewService(store *config.Manager, authManager *auth.Manager, log *logger.Logger, httpClient *http.Client) *Service {
	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: fetchTimeout}
	}
	return &Service{
		store:        store,
		auth:         authManager,
		log:          log,
		codexFetcher: codexprovider.NewQuotaFetcher(client),
		kiroFetcher:  kiroprovider.NewQuotaFetcher(client),
	}
}

func (s *Service) RefreshAccountWithQuota(accountID string) (config.Account, error) {
	account, ok := s.store.GetAccount(accountID)
	if !ok {
		return config.Account{}, fmt.Errorf("account not found")
	}
	if err := validateQuotaProvider(account); err != nil {
		return account, err
	}

	refreshed, err := s.auth.RefreshAccount(accountID)
	if err != nil {
		return refreshed, err
	}
	account = refreshed

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	quota, resolvedEmail, quotaErr := s.fetchQuotaForAccount(ctx, account)
	if err := s.applyQuotaSnapshot(accountID, quota, resolvedEmail); err != nil {
		return account, err
	}

	updated, _ := s.store.GetAccount(accountID)
	return updated, quotaErr
}

func (s *Service) RefreshQuota(accountID string) (config.Account, error) {
	account, ok := s.store.GetAccount(accountID)
	if !ok {
		return config.Account{}, fmt.Errorf("account not found")
	}
	if err := validateQuotaProvider(account); err != nil {
		return account, err
	}

	fresh, err := s.auth.EnsureFreshAccount(accountID)
	if err != nil {
		quota := synthesizeQuota(account, err)
		blockedMsg, blocked := blockedAccountMessageFromError(err)
		_ = s.store.UpdateAccount(accountID, func(a *config.Account) {
			a.Quota = quota
			if blocked {
				a.Enabled = false
				a.Banned = true
				a.BannedReason = blockedMsg
				a.LastError = blockedMsg
			}
		})
		return account, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	quota, resolvedEmail, quotaErr := s.fetchQuotaForAccount(ctx, fresh)
	if err := s.applyQuotaSnapshot(accountID, quota, resolvedEmail); err != nil {
		return fresh, err
	}
	updated, _ := s.store.GetAccount(accountID)
	return updated, quotaErr
}

func (s *Service) RefreshQuotaOnly(accountID string) error {
	_, err := s.RefreshQuota(accountID)
	return err
}

func (s *Service) RefreshAllQuotas() error {
	return s.refreshAllQuotas(false)
}

func (s *Service) ForceRefreshAllQuotas() error {
	return s.refreshAllQuotas(true)
}

func (s *Service) applyQuotaSnapshot(accountID string, quota config.QuotaInfo, resolvedEmail string) error {
	return s.store.UpdateAccount(accountID, func(a *config.Account) {
		a.Quota = quota
		if strings.TrimSpace(resolvedEmail) != "" {
			a.Email = strings.TrimSpace(resolvedEmail)
		}
		if blockedMsg, blocked := blockedAccountMessageFromQuota(quota); blocked {
			a.Enabled = false
			a.Banned = true
			a.BannedReason = blockedMsg
			a.HealthState = config.AccountHealthBanned
			a.HealthReason = blockedMsg
			a.LastError = blockedMsg
			return
		}
		if shouldApplyQuotaCooldown(quota) {
			cooldownUntil := quotaResetAt(quota)
			if cooldownUntil <= time.Now().Unix() {
				cooldownUntil = time.Now().Add(time.Hour).Unix()
			}
			a.CooldownUntil = cooldownUntil
			a.HealthState = config.AccountHealthCooldownQuota
			a.HealthReason = firstNonEmpty(strings.TrimSpace(quota.Summary), "Quota exhausted")
			a.LastFailureAt = time.Now().Unix()
			if strings.TrimSpace(a.LastError) == "" {
				a.LastError = firstNonEmpty(strings.TrimSpace(quota.Summary), "Quota exhausted")
			}
		} else if a.HealthState == config.AccountHealthCooldownQuota {
			a.HealthState = config.AccountHealthReady
			a.HealthReason = ""
			a.CooldownUntil = 0
			a.ConsecutiveFailures = 0
		}
	})
}

func (s *Service) fetchQuotaForAccount(ctx context.Context, account config.Account) (config.QuotaInfo, string, error) {
	if isKiroAccount(account) {
		return s.kiroFetcher.FetchQuota(ctx, account, func(accountID string) (config.Account, error) {
			return s.auth.RefreshAccount(accountID)
		})
	}
	if !isCodexAccount(account) {
		return config.QuotaInfo{}, "", fmt.Errorf("unsupported provider for quota refresh: %s", strings.TrimSpace(account.Provider))
	}
	quota, err := s.codexFetcher.FetchQuota(ctx, account)
	return quota, "", err
}

func (s *Service) refreshAllQuotas(force bool) error {
	accounts := s.store.Accounts()
	if len(accounts) == 0 {
		return nil
	}

	now := time.Now().Unix()
	eligible := make([]config.Account, 0, len(accounts))
	skipped := map[string]int{}
	for _, account := range accounts {
		if !force {
			if skip, reason := shouldSkipBatchQuotaRefresh(account, now); skip {
				skipped[reason]++
				continue
			}
		}
		eligible = append(eligible, account)
	}

	if len(eligible) == 0 {
		s.logQuotaRefreshBatch(force, len(accounts), 0, skipped)
		return nil
	}

	workerCount := 4
	if workerCount > len(eligible) {
		workerCount = len(eligible)
	}
	if workerCount <= 0 {
		workerCount = 1
	}

	jobs := make(chan config.Account)
	failures := make([]string, 0)
	var failuresMu sync.Mutex
	var wg sync.WaitGroup

	for worker := 0; worker < workerCount; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for account := range jobs {
				if _, err := s.RefreshQuota(account.ID); err != nil {
					failuresMu.Lock()
					failures = append(failures, firstNonEmpty(account.Email, account.ID)+": "+err.Error())
					failuresMu.Unlock()
				}
			}
		}()
	}

	for _, account := range eligible {
		jobs <- account
	}
	close(jobs)
	wg.Wait()

	s.logQuotaRefreshBatch(force, len(accounts), len(eligible), skipped)

	if len(failures) > 0 {
		return fmt.Errorf(strings.Join(failures, "; "))
	}
	return nil
}

func (s *Service) logQuotaRefreshBatch(force bool, total int, eligible int, skipped map[string]int) {
	if s == nil || s.log == nil {
		return
	}
	mode := "smart"
	if force {
		mode = "force"
	}
	parts := []string{fmt.Sprintf("quota batch refresh mode=%s total=%d eligible=%d", mode, total, eligible)}
	if skipped["quota_cooldown"] > 0 {
		parts = append(parts, fmt.Sprintf("skipped_quota_cooldown=%d", skipped["quota_cooldown"]))
	}
	if skipped["disabled"] > 0 {
		parts = append(parts, fmt.Sprintf("skipped_disabled=%d", skipped["disabled"]))
	}
	if skipped["banned"] > 0 {
		parts = append(parts, fmt.Sprintf("skipped_banned=%d", skipped["banned"]))
	}
	s.log.Info("quota", strings.Join(parts, " "))
}

func isKiroAccount(account config.Account) bool {
	return strings.EqualFold(strings.TrimSpace(account.Provider), "kiro")
}

func isCodexAccount(account config.Account) bool {
	return strings.EqualFold(strings.TrimSpace(account.Provider), "codex")
}

func validateQuotaProvider(account config.Account) error {
	if isKiroAccount(account) || isCodexAccount(account) {
		return nil
	}
	provider := strings.TrimSpace(account.Provider)
	if provider == "" {
		return fmt.Errorf("account provider is required")
	}
	return fmt.Errorf("unsupported provider for quota refresh: %s", provider)
}

func blockedAccountMessageFromQuota(quota config.QuotaInfo) (string, bool) {
	sourceMessage := firstNonEmpty(strings.TrimSpace(quota.Error), strings.TrimSpace(quota.Summary))
	if sourceMessage == "" {
		return "", false
	}
	return config.BlockedAccountReason(sourceMessage)
}

func blockedAccountMessageFromError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	return config.BlockedAccountReason(err.Error())
}

func shouldApplyQuotaCooldown(quota config.QuotaInfo) bool {
	status := strings.ToLower(strings.TrimSpace(quota.Status))
	if status == "exhausted" || status == "empty" {
		return true
	}
	for _, bucket := range quota.Buckets {
		bucketStatus := strings.ToLower(strings.TrimSpace(bucket.Status))
		if bucketStatus == "exhausted" || bucketStatus == "empty" {
			return true
		}
		if bucket.Total > 0 {
			remaining := bucket.Remaining
			if remaining == 0 && bucket.Used > 0 && bucket.Used <= bucket.Total {
				remaining = maxInt(bucket.Total-bucket.Used, 0)
			}
			if remaining <= 0 {
				return true
			}
		}
	}
	return false
}

func quotaResetAt(quota config.QuotaInfo) int64 {
	var latest int64
	for _, bucket := range quota.Buckets {
		if bucket.ResetAt > latest {
			latest = bucket.ResetAt
		}
	}
	return latest
}

func shouldSkipBatchQuotaRefresh(account config.Account, now int64) (bool, string) {
	if account.Banned || account.HealthState == config.AccountHealthBanned {
		return true, "banned"
	}
	if !account.Enabled || account.HealthState == config.AccountHealthDisabledDurable {
		return true, "disabled"
	}
	if shouldApplyQuotaCooldown(account.Quota) {
		if resetAt := quotaResetAt(account.Quota); resetAt > now {
			return true, "quota_cooldown"
		}
	}
	return false, ""
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
