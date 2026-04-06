package provider

import (
	"strings"
	"time"

	"cliro/internal/account"
	"cliro/internal/config"
)

type StatusPhase string

const (
	StatusPhaseNone              StatusPhase = "none"
	StatusPhaseRefreshOnce       StatusPhase = "refresh_once"
	StatusPhaseCooldownThenRetry StatusPhase = "cooldown_then_retry"
)

type StatusPolicy struct {
	Phase StatusPhase
	Final FailureClass
}

type RetryDecision struct {
	Retry          bool
	RefreshAuth    bool
	Cooldown       time.Duration
	Disable        bool
	Ban            bool
	Cause          string
	ExcludeAccount bool
	FinalStatus    int
	FinalMessage   string
	FailureClass   FailureClass
}

type RetryPlanner struct {
	pool     *account.Pool
	provider string
	policies map[int]StatusPolicy
}

func DefaultStatusPolicies() map[int]StatusPolicy {
	return map[int]StatusPolicy{
		401: {Phase: StatusPhaseRefreshOnce, Final: FailureAuthRefreshable},
		403: {Phase: StatusPhaseRefreshOnce, Final: FailureAuthRefreshable},
		429: {Phase: StatusPhaseCooldownThenRetry, Final: FailureQuotaCooldown},
	}
}

func NewRetryPlanner(pool *account.Pool, provider string, policies map[int]StatusPolicy) RetryPlanner {
	resolved := make(map[int]StatusPolicy)
	for status, policy := range DefaultStatusPolicies() {
		resolved[status] = policy
	}
	for status, policy := range policies {
		resolved[status] = policy
	}
	return RetryPlanner{pool: pool, provider: strings.ToLower(strings.TrimSpace(provider)), policies: resolved}
}

func (p RetryPlanner) NextAccount(excluded map[string]bool) (config.Account, bool) {
	if p.pool == nil {
		return config.Account{}, false
	}
	for _, candidate := range p.pool.AvailableAccountsForProvider(p.provider) {
		if excluded != nil && excluded[candidate.ID] {
			continue
		}
		return candidate, true
	}
	return config.Account{}, false
}

func (p RetryPlanner) Decide(result AttemptResult) RetryDecision {
	decision := result.Failure
	statusPolicy, hasPolicy := p.policies[result.Status]
	if hasPolicy && statusPolicy.Final != "" && decision.Class != FailureEmptyStream {
		decision.Class = statusPolicy.Final
	}
	if decision.Class == FailureEmptyStream {
		if !result.ClientBytesSent {
			return RetryDecision{
				Retry:          true,
				Cause:          firstNonEmpty(result.RetryCause, "empty_stream"),
				ExcludeAccount: true,
				FinalStatus:    decision.Status,
				FinalMessage:   firstNonEmpty(result.Message, decision.Message),
				FailureClass:   decision.Class,
			}
		}
		return RetryDecision{FinalStatus: decision.Status, FinalMessage: firstNonEmpty(result.Message, decision.Message), FailureClass: decision.Class}
	}

	if hasPolicy {
		switch statusPolicy.Phase {
		case StatusPhaseRefreshOnce:
			if !result.RecoveredAuth && decision.Class == FailureAuthRefreshable {
				return RetryDecision{
					Retry:        true,
					RefreshAuth:  true,
					Cause:        "status_phase_refresh_once",
					FinalStatus:  decision.Status,
					FinalMessage: firstNonEmpty(result.Message, decision.Message),
					FailureClass: decision.Class,
				}
			}
		case StatusPhaseCooldownThenRetry:
			if !result.ClientBytesSent {
				return RetryDecision{
					Retry:          true,
					Cooldown:       decision.Cooldown,
					Cause:          "status_phase_cooldown_then_retry",
					ExcludeAccount: true,
					FinalStatus:    decision.Status,
					FinalMessage:   firstNonEmpty(result.Message, decision.Message),
					FailureClass:   decision.Class,
				}
			}
		}
	}

	if decision.Class == FailureAuthRefreshable && !result.RecoveredAuth {
		return RetryDecision{
			Retry:        true,
			RefreshAuth:  true,
			Cause:        firstNonEmpty(result.RetryCause, "auth_refresh"),
			FinalStatus:  decision.Status,
			FinalMessage: firstNonEmpty(result.Message, decision.Message),
			FailureClass: decision.Class,
		}
	}

	if decision.RetryAllowed && !result.ClientBytesSent {
		return RetryDecision{
			Retry:          true,
			Cooldown:       decision.Cooldown,
			Cause:          firstNonEmpty(result.RetryCause, string(decision.Class)),
			ExcludeAccount: true,
			FinalStatus:    decision.Status,
			FinalMessage:   firstNonEmpty(result.Message, decision.Message),
			FailureClass:   decision.Class,
		}
	}

	return RetryDecision{
		Cooldown:     decision.Cooldown,
		Disable:      decision.Disable,
		Ban:          decision.BanAccount,
		FinalStatus:  decision.Status,
		FinalMessage: firstNonEmpty(result.Message, decision.Message),
		FailureClass: decision.Class,
	}
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
