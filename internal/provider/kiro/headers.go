package kiro

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"cliro-go/internal/config"

	"github.com/google/uuid"
)

const (
	kiroPrimaryURL          = "https://q.us-east-1.amazonaws.com/generateAssistantResponse"
	kiroFallbackURL         = "https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse"
	kiroFallbackTarget      = "AmazonCodeWhispererStreamingService.GenerateAssistantResponse"
	kiroContentType         = "application/json"
	kiroAcceptStream        = "application/vnd.amazon.eventstream"
	kiroRuntimeUserAgent    = "aws-sdk-js/1.0.27 ua/2.1 os/linux lang/js md/nodejs#22.21.1 api/codewhispererstreaming#1.0.27 m/E KiroIDE-0.10.32"
	kiroRuntimeAmzUserAgent = "aws-sdk-js/1.0.27 KiroIDE 0.10.32"
	kiroAgentMode           = "vibe"
	kiroFirstTokenTimeout   = 10 * time.Second
	kiroFirstTokenRetries   = 5
	kiroMaxTimeout          = 30 * time.Second
	kiroDefaultOrigin       = "AI_EDITOR"
	kiroDefaultMaxTokens    = 32000
)

type endpointConfig struct {
	Name      string
	URL       string
	AmzTarget string
}

var endpointConfigs = []endpointConfig{
	{Name: "amazonq", URL: kiroPrimaryURL},
	// Fallback disabled - only use Amazon Q endpoint
	// {Name: "codewhisperer", URL: kiroFallbackURL, AmzTarget: kiroFallbackTarget},
}

func applyRuntimeHeaders(req *http.Request, account config.Account, endpoint endpointConfig, attempt int) {
	req.Header.Set("Content-Type", kiroContentType)
	req.Header.Set("Accept", kiroAcceptStream)
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(account.AccessToken))
	req.Header.Set("User-Agent", kiroRuntimeUserAgent)
	req.Header.Set("x-amz-user-agent", kiroRuntimeAmzUserAgent)
	req.Header.Set("x-amzn-kiro-agent-mode", kiroAgentMode)
	req.Header.Set("x-amzn-codewhisperer-optout", "true")
	req.Header.Set("amz-sdk-invocation-id", uuid.NewString())
	req.Header.Set("amz-sdk-request", "attempt="+strconv.Itoa(maxInt(attempt, 1))+"; max="+strconv.Itoa(kiroFirstTokenRetries*len(endpointConfigs)))
	if endpoint.AmzTarget != "" {
		req.Header.Set("X-Amz-Target", endpoint.AmzTarget)
	}
}
