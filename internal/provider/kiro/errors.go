package kiro

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"cliro-go/internal/config"
	provider "cliro-go/internal/provider"
)

var ErrFirstTokenTimeout = errors.New("kiro first token timeout")

func classifyHTTPFailure(statusCode int, body []byte) provider.FailureDecision {
	message := upstreamErrorMessage(statusCode, body)
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		trimmed = http.StatusText(statusCode)
	}

	if blockedMessage, blocked := config.BlockedAccountReason(trimmed); blocked {
		return provider.FailureDecision{Class: provider.FailureDurableDisabled, Message: blockedMessage, BanAccount: true, Disable: true, Status: http.StatusUnauthorized}
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		if refreshableMessage, refreshable := config.RefreshableAuthReason(trimmed); refreshable {
			return provider.FailureDecision{Class: provider.FailureAuthRefreshable, Message: refreshableMessage, RetryAllowed: true, Status: http.StatusUnauthorized}
		}
		return provider.FailureDecision{Class: provider.FailureAuthRefreshable, Message: trimmed, RetryAllowed: true, Status: http.StatusUnauthorized}
	}

	lowerMessage := strings.ToLower(trimmed)
	if statusCode == http.StatusTooManyRequests || strings.Contains(lowerMessage, "usage limit") || strings.Contains(lowerMessage, "quota") || strings.Contains(lowerMessage, "credit") || strings.Contains(lowerMessage, "rate limit") {
		return provider.FailureDecision{Class: provider.FailureQuotaCooldown, Message: trimmed, Cooldown: time.Hour, Status: http.StatusTooManyRequests}
	}

	if statusCode == http.StatusBadRequest || statusCode == http.StatusUnprocessableEntity || strings.Contains(lowerMessage, "improperly formed request") || strings.Contains(lowerMessage, "validationexception") || strings.Contains(lowerMessage, "validation exception") || strings.Contains(lowerMessage, "invalid schema") || strings.Contains(lowerMessage, "invalid tool") || strings.Contains(lowerMessage, "malformed") {
		return provider.FailureDecision{Class: provider.FailureRequestShape, Message: trimmed, Status: http.StatusBadRequest}
	}

	if statusCode == http.StatusInternalServerError || statusCode == http.StatusBadGateway || statusCode == http.StatusServiceUnavailable || statusCode == http.StatusGatewayTimeout {
		return provider.FailureDecision{Class: provider.FailureRetryableTransport, Message: trimmed, Cooldown: 15 * time.Second, RetryAllowed: true, Status: http.StatusBadGateway}
	}

	return provider.FailureDecision{Class: provider.FailureProviderFatal, Message: trimmed, Cooldown: 30 * time.Second, Status: http.StatusBadGateway}
}

func classifyTransportFailure(err error) provider.FailureDecision {
	if errors.Is(err, ErrFirstTokenTimeout) {
		return provider.FailureDecision{Class: provider.FailureRetryableTransport, Message: ErrFirstTokenTimeout.Error(), Cooldown: 5 * time.Second, RetryAllowed: true, Status: http.StatusGatewayTimeout}
	}
	if err == nil {
		return provider.FailureDecision{Class: provider.FailureRetryableTransport, Message: "transport error", Cooldown: 15 * time.Second, RetryAllowed: true, Status: http.StatusBadGateway}
	}
	message := strings.TrimSpace(err.Error())
	status := http.StatusBadGateway
	if strings.Contains(strings.ToLower(message), "timeout") || errors.Is(err, context.DeadlineExceeded) {
		status = http.StatusGatewayTimeout
	}
	return provider.FailureDecision{Class: provider.FailureRetryableTransport, Message: message, Cooldown: 15 * time.Second, RetryAllowed: true, Status: status}
}

func upstreamErrorMessage(statusCode int, body []byte) string {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return http.StatusText(statusCode)
	}
	for _, key := range []string{"message", "Message", "errorMessage"} {
		var object map[string]any
		if err := json.Unmarshal(body, &object); err == nil {
			if value, ok := object[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
			if nested, ok := object["error"].(map[string]any); ok {
				if value, ok := nested[key].(string); ok && strings.TrimSpace(value) != "" {
					return strings.TrimSpace(value)
				}
			}
		}
	}
	return trimmed
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
