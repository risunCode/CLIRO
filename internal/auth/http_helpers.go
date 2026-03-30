package auth

import "strings"

const kiroQuotaBaseURL = "https://codewhisperer.us-east-1.amazonaws.com"

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
