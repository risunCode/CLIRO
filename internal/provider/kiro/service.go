package kiro

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cliro-go/internal/account"
	"cliro-go/internal/config"
	"cliro-go/internal/logger"
	"cliro-go/internal/platform"
	provider "cliro-go/internal/provider"
)

type Service struct {
	store             *config.Manager
	auth              accountAuth
	pool              *account.Pool
	log               *logger.Logger
	httpClient        *http.Client
	firstTokenTimeout time.Duration
}

type accountAuth interface {
	EnsureFreshAccount(accountID string) (config.Account, error)
	RefreshAccount(accountID string) (config.Account, error)
}

func NewService(store *config.Manager, authManager accountAuth, accountPool *account.Pool, log *logger.Logger, httpClient *http.Client) *Service {
	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	return &Service{
		store:             store,
		auth:              authManager,
		pool:              accountPool,
		log:               log,
		httpClient:        client,
		firstTokenTimeout: kiroFirstTokenTimeout,
	}
}

func (s *Service) Complete(ctx context.Context, req provider.ChatRequest) (provider.CompletionOutcome, int, string, error) {
	return s.CompleteWithCallback(ctx, req, nil)
}

func (s *Service) CompleteWithCallback(ctx context.Context, req provider.ChatRequest, eventCallback func(StreamEvent)) (provider.CompletionOutcome, int, string, error) {
	requestID := platform.RequestIDFromContext(ctx)
	if strings.TrimSpace(req.Model) == "" {
		s.recordRequestFailure()
		s.log.Warn("proxy", fmt.Sprintf("request_id=%q provider=%q route=%q phase=%q reason=%q", requestID, "kiro", strings.TrimSpace(req.RouteFamily), "rejected", "model is required"))
		return provider.CompletionOutcome{}, http.StatusBadRequest, "model is required", fmt.Errorf("model is required")
	}

	upstreamCandidates := s.pool.AvailableAccountsForProvider("kiro")
	if len(upstreamCandidates) == 0 {
		s.recordRequestFailure()
		reason := s.pool.ProviderUnavailableReason("kiro")
		s.log.Warn("proxy", fmt.Sprintf("request_id=%q provider=%q route=%q phase=%q reason=%q", requestID, "kiro", strings.TrimSpace(req.RouteFamily), "rejected", reason))
		return provider.CompletionOutcome{}, http.StatusServiceUnavailable, reason, fmt.Errorf(reason)
	}

	runtimeClient := newRuntimeClient(s.httpClient, s.firstTokenTimeout)
	var lastStatus int
	var lastMessage string

	for _, candidate := range upstreamCandidates {
		accountLabel := config.AccountLabel(candidate)
		s.log.Info("proxy", fmt.Sprintf("request_id=%q provider=%q route=%q phase=%q account=%q model=%q", requestID, "kiro", strings.TrimSpace(req.RouteFamily), "attempt", accountLabel, strings.TrimSpace(req.Model)))
		account, err := s.auth.EnsureFreshAccount(candidate.ID)
		if err != nil {
			decision := classifyHTTPFailure(http.StatusUnauthorized, []byte(err.Error()))
			s.applyFailureDecision(requestID, candidate.ID, accountLabel, decision)
			lastStatus = decision.Status
			lastMessage = decision.Message
			continue
		}
		accountLabel = config.AccountLabel(account)
		refreshedAfterFailure := false

		for {
			payload, err := buildRequestPayload(req, account)
			if err != nil {
				s.recordRequestFailure()
				return provider.CompletionOutcome{}, http.StatusBadRequest, err.Error(), err
			}
			body, err := json.Marshal(payload)
			if err != nil {
				s.recordRequestFailure()
				return provider.CompletionOutcome{}, http.StatusBadRequest, err.Error(), err
			}

			resp, _, err := runtimeClient.Do(ctx, account, body)
			if err != nil {
				decision := classifyTransportFailure(err)
				s.applyFailureDecision(requestID, account.ID, accountLabel, decision)
				lastStatus = decision.Status
				lastMessage = decision.Message
				break
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				data, _ := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				decision := classifyHTTPFailure(resp.StatusCode, data)
				if decision.Class == provider.FailureAuthRefreshable && !refreshedAfterFailure {
					refreshedAccount, refreshErr := s.auth.RefreshAccount(account.ID)
					if refreshErr == nil {
						account = refreshedAccount
						accountLabel = config.AccountLabel(account)
						refreshedAfterFailure = true
						s.log.Info("auth", fmt.Sprintf("request_id=%q provider=%q phase=%q account=%q", requestID, "kiro", "token_refreshed_retry", accountLabel))
						continue
					}
					decision = classifyHTTPFailure(http.StatusUnauthorized, []byte(refreshErr.Error()))
				}

				s.applyFailureDecision(requestID, account.ID, accountLabel, decision)
				lastStatus = decision.Status
				lastMessage = decision.Message
				if decision.Class == provider.FailureRequestShape {
					s.recordRequestFailure()
					return provider.CompletionOutcome{}, decision.Status, decision.Message, fmt.Errorf(decision.Message)
				}
				break
			}

			outcome, err := collectCompletionWithCallback(resp.Body, req, eventCallback)
			_ = resp.Body.Close()
			if err != nil {
				decision := classifyTransportFailure(err)
				s.applyFailureDecision(requestID, account.ID, accountLabel, decision)
				lastStatus = decision.Status
				lastMessage = decision.Message
				break
			}
			outcome.Provider = "kiro"
			outcome.AccountID = account.ID
			outcome.AccountLabel = accountLabel

			s.markSuccess(requestID, account.ID, accountLabel, outcome.Usage)
			return outcome, 0, "", nil
		}
	}

	snapshot := s.pool.AvailabilitySnapshot("kiro")
	if snapshot.ReadyCount == 0 {
		lastStatus = http.StatusServiceUnavailable
		lastMessage = s.pool.ProviderUnavailableReason("kiro")
	}
	if lastStatus == 0 {
		lastStatus = http.StatusServiceUnavailable
	}
	if strings.TrimSpace(lastMessage) == "" {
		lastMessage = "all kiro accounts failed"
	}
	s.recordRequestFailure()
	s.log.Warn("proxy", fmt.Sprintf("request_id=%q provider=%q route=%q phase=%q reason=%q", requestID, "kiro", strings.TrimSpace(req.RouteFamily), "failed", lastMessage))
	return provider.CompletionOutcome{}, lastStatus, lastMessage, fmt.Errorf(lastMessage)
}

func (s *Service) markSuccess(requestID string, accountID string, accountLabel string, usage config.ProxyStats) {
	now := time.Now().Unix()
	_ = s.store.UpdateAccount(accountID, func(account *config.Account) {
		account.RequestCount++
		account.PromptTokens += usage.PromptTokens
		account.CompletionTokens += usage.CompletionTokens
		account.TotalTokens += usage.TotalTokens
		account.LastUsed = now
		account.CooldownUntil = 0
		account.ConsecutiveFailures = 0
		account.Banned = false
		account.BannedReason = ""
		account.HealthState = config.AccountHealthReady
		account.HealthReason = ""
		account.LastError = ""
		if account.Quota.Status == "exhausted" || account.Quota.Status == "unknown" || account.Quota.Status == "degraded" {
			account.Quota.Status = "healthy"
			account.Quota.Summary = "Recent request succeeded."
			account.Quota.Source = firstNonEmpty(account.Quota.Source, "runtime")
			account.Quota.Error = ""
			account.Quota.LastCheckedAt = now
			for index := range account.Quota.Buckets {
				if account.Quota.Buckets[index].Status == "exhausted" || account.Quota.Buckets[index].Status == "unknown" {
					account.Quota.Buckets[index].Status = "healthy"
				}
			}
		}
	})

	_ = s.store.UpdateStats(func(stats *config.ProxyStats) {
		stats.TotalRequests++
		stats.SuccessRequests++
		stats.PromptTokens += usage.PromptTokens
		stats.CompletionTokens += usage.CompletionTokens
		stats.TotalTokens += usage.TotalTokens
		stats.LastRequestAt = now
	})
	s.log.Info("proxy", fmt.Sprintf("request_id=%q provider=%q phase=%q account=%q prompt_tokens=%d completion_tokens=%d total_tokens=%d", requestID, "kiro", "success", accountLabel, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens))
}

func (s *Service) markTransientFailure(requestID string, accountID string, accountLabel string, err error) {
	breakerEnabled := s.store.CircuitBreaker()
	steps := s.store.CircuitSteps()
	now := time.Now().Unix()
	appliedCooldown := time.Duration(0)
	appliedStep := 0
	_ = s.store.UpdateAccount(accountID, func(account *config.Account) {
		account.ErrorCount++
		account.LastError = err.Error()
		account.LastFailureAt = now
		account.Quota.Status = firstNonEmpty(account.Quota.Status, "degraded")
		account.Quota.Summary = err.Error()
		account.Quota.Source = firstNonEmpty(account.Quota.Source, "runtime")
		account.Quota.Error = err.Error()
		account.Quota.LastCheckedAt = now
		if !breakerEnabled {
			account.ConsecutiveFailures = 0
			account.CooldownUntil = 0
			account.HealthState = config.AccountHealthReady
			account.HealthReason = ""
			return
		}
		nextFailures := account.ConsecutiveFailures + 1
		appliedCooldown = provider.CircuitCooldown(steps, nextFailures)
		appliedStep = nextFailures
		account.ConsecutiveFailures = nextFailures
		account.CooldownUntil = now + int64(appliedCooldown/time.Second)
		account.HealthState = config.AccountHealthCooldownTransient
		account.HealthReason = err.Error()
	})
	if breakerEnabled && appliedCooldown > 0 {
		cappedStep := appliedStep
		if cappedStep > len(steps) {
			cappedStep = len(steps)
		}
		s.log.Warn("proxy", fmt.Sprintf("request_id=%q provider=%q phase=%q account=%q reason=%q circuit_step=%d cooldown_seconds=%d", requestID, "kiro", "attempt_failed", accountLabel, err.Error(), cappedStep, int(appliedCooldown/time.Second)))
		return
	}
	s.log.Warn("proxy", fmt.Sprintf("request_id=%q provider=%q phase=%q account=%q reason=%q circuit_breaker=%t", requestID, "kiro", "attempt_failed", accountLabel, err.Error(), breakerEnabled))
}

func (s *Service) markBanned(requestID string, accountID string, accountLabel string, reason string) {
	_ = s.store.MarkAccountBanned(accountID, reason)
	s.log.Warn("auth", fmt.Sprintf("request_id=%q provider=%q phase=%q account=%q reason=%q", requestID, "kiro", "banned", accountLabel, reason))
}

func (s *Service) applyFailureDecision(requestID string, accountID string, accountLabel string, decision provider.FailureDecision) {
	switch decision.Class {
	case provider.FailureRequestShape:
		s.log.Warn("proxy", fmt.Sprintf("request_id=%q provider=%q phase=%q account=%q reason=%q", requestID, "kiro", "request_shape", accountLabel, decision.Message))
	case provider.FailureDurableDisabled:
		if decision.BanAccount {
			s.markBanned(requestID, accountID, accountLabel, decision.Message)
			return
		}
		_ = s.store.MarkAccountDurablyDisabled(accountID, decision.Message)
		s.log.Warn("auth", fmt.Sprintf("request_id=%q provider=%q phase=%q account=%q reason=%q", requestID, "kiro", "durable_disabled", accountLabel, decision.Message))
	case provider.FailureQuotaCooldown:
		cooldownUntil := time.Now().Add(decision.Cooldown).Unix()
		_ = s.store.UpdateAccount(accountID, func(account *config.Account) {
			account.ErrorCount++
			account.CooldownUntil = cooldownUntil
			account.HealthState = config.AccountHealthCooldownQuota
			account.HealthReason = decision.Message
			account.LastFailureAt = time.Now().Unix()
			account.LastError = decision.Message
			account.Quota = config.QuotaInfo{
				Status:        "exhausted",
				Summary:       decision.Message,
				Source:        "runtime",
				Error:         decision.Message,
				LastCheckedAt: time.Now().Unix(),
				Buckets:       []config.QuotaBucket{{Name: "session", ResetAt: cooldownUntil, Status: "exhausted"}},
			}
		})
		s.log.Warn("quota", fmt.Sprintf("request_id=%q provider=%q account=%q phase=%q reason=%q cooldown_until=%d", requestID, "kiro", accountLabel, "cooldown", decision.Message, cooldownUntil))
	default:
		s.markTransientFailure(requestID, accountID, accountLabel, fmt.Errorf(decision.Message))
	}
}

func (s *Service) recordRequestFailure() {
	now := time.Now().Unix()
	_ = s.store.UpdateStats(func(stats *config.ProxyStats) {
		stats.TotalRequests++
		stats.FailedRequests++
		stats.LastRequestAt = now
	})
}
