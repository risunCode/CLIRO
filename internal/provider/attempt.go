package provider

import (
	"strings"
	"time"

	"cliro/internal/logger"
)

type AttemptContext struct {
	RequestID string
	Provider  string
	Model     string
	Stream    bool
}

type AttemptResult struct {
	Attempt             int
	Status              int
	Message             string
	Err                 error
	Failure             FailureDecision
	ClientBytesSent     bool
	UpstreamReadable    bool
	EmptyStream         bool
	UpstreamOpen        time.Duration
	FirstClientChunk    time.Duration
	RecoveredAuth       bool
	RetryCause          string
	Final               bool
	Success             bool
	CompletionHasOutput bool
}

type AttemptDiagnostic struct {
	RequestID          string
	Provider           string
	AccountID          string
	AccountLabel       string
	Model              string
	Attempt            int
	Stream             bool
	UpstreamOpenMillis int64
	FirstClientChunkMs int64
	ClientBytesSent    bool
	UpstreamReadable   bool
	EmptyStream        bool
	RetryCause         string
	FailureClass       string
	Success            bool
	RecoveredAuth      bool
	Final              bool
}

func NewAttemptDiagnostic(ctx AttemptContext, accountID string, accountLabel string, result AttemptResult) AttemptDiagnostic {
	return AttemptDiagnostic{
		RequestID:          strings.TrimSpace(ctx.RequestID),
		Provider:           strings.TrimSpace(ctx.Provider),
		AccountID:          strings.TrimSpace(accountID),
		AccountLabel:       strings.TrimSpace(accountLabel),
		Model:              strings.TrimSpace(ctx.Model),
		Attempt:            result.Attempt,
		Stream:             ctx.Stream,
		UpstreamOpenMillis: durationMillis(result.UpstreamOpen),
		FirstClientChunkMs: durationMillis(result.FirstClientChunk),
		ClientBytesSent:    result.ClientBytesSent,
		UpstreamReadable:   result.UpstreamReadable,
		EmptyStream:        result.EmptyStream,
		RetryCause:         strings.TrimSpace(result.RetryCause),
		FailureClass:       string(result.Failure.Class),
		Success:            result.Success,
		RecoveredAuth:      result.RecoveredAuth,
		Final:              result.Final,
	}
}

func LogAttemptDiagnostic(log *logger.Logger, diag AttemptDiagnostic) {
	if log == nil {
		return
	}
	fields := []logger.Field{
		logger.String("request_id", strings.TrimSpace(diag.RequestID)),
		logger.String("provider", strings.TrimSpace(diag.Provider)),
		logger.String("account_id", strings.TrimSpace(diag.AccountID)),
		logger.String("account", strings.TrimSpace(diag.AccountLabel)),
		logger.String("model", strings.TrimSpace(diag.Model)),
		logger.Int("attempt", diag.Attempt),
		logger.Bool("stream", diag.Stream),
		logger.Int64("upstream_open_ms", diag.UpstreamOpenMillis),
		logger.Int64("first_client_chunk_ms", diag.FirstClientChunkMs),
		logger.Bool("client_bytes_sent", diag.ClientBytesSent),
		logger.Bool("upstream_readable", diag.UpstreamReadable),
		logger.Bool("empty_stream", diag.EmptyStream),
		logger.String("retry_cause", strings.TrimSpace(diag.RetryCause)),
		logger.String("failure_class", strings.TrimSpace(diag.FailureClass)),
		logger.Bool("recovered_auth", diag.RecoveredAuth),
		logger.Bool("success", diag.Success),
		logger.Bool("final", diag.Final),
	}
	if diag.Success {
		log.InfoEvent("proxy", "request.attempt_result", fields...)
		return
	}
	log.WarnEvent("proxy", "request.attempt_result", fields...)
}

func durationMillis(value time.Duration) int64 {
	if value <= 0 {
		return 0
	}
	return value.Milliseconds()
}

func CompletionHasVisibleOutput(outcome CompletionOutcome) bool {
	if strings.TrimSpace(outcome.Text) != "" {
		return true
	}
	if strings.TrimSpace(outcome.Thinking) != "" {
		return true
	}
	return len(outcome.ToolUses) > 0
}
