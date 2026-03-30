package kiro

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"cliro-go/internal/config"
)

type runtimeClient struct {
	httpClient        *http.Client
	firstTokenTimeout time.Duration
}

func newRuntimeClient(httpClient *http.Client, firstTokenTimeout time.Duration) *runtimeClient {
	client := httpClient
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}
	if firstTokenTimeout <= 0 {
		firstTokenTimeout = kiroFirstTokenTimeout
	}
	return &runtimeClient{httpClient: client, firstTokenTimeout: firstTokenTimeout}
}

func (c *runtimeClient) Do(ctx context.Context, account config.Account, body []byte) (*http.Response, endpointConfig, error) {
	var lastErr error
	for attempt := 1; attempt <= kiroFirstTokenRetries; attempt++ {
		// Exponential backoff: 10s, 20s, 30s (capped), 30s, 30s
		timeout := time.Duration(attempt) * c.firstTokenTimeout
		if timeout > kiroMaxTimeout {
			timeout = kiroMaxTimeout
		}

		for _, endpoint := range endpointConfigs {
			resp, err := c.doOnceWithTimeout(ctx, account, body, endpoint, attempt, timeout)
			if err == nil {
				return resp, endpoint, nil
			}
			lastErr = err
			if err != ErrFirstTokenTimeout {
				return nil, endpoint, err
			}
		}
	}
	return nil, endpointConfig{}, lastErr
}

func (c *runtimeClient) doOnce(ctx context.Context, account config.Account, body []byte, endpoint endpointConfig, attempt int) (*http.Response, error) {
	return c.doOnceWithTimeout(ctx, account, body, endpoint, attempt, c.firstTokenTimeout)
}

func (c *runtimeClient) doOnceWithTimeout(ctx context.Context, account config.Account, body []byte, endpoint endpointConfig, attempt int, timeout time.Duration) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	applyRuntimeHeaders(httpReq, account, endpoint, attempt)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, nil
	}

	wrappedBody, err := waitForFirstToken(resp.Body, timeout)
	if err != nil {
		_ = resp.Body.Close()
		return nil, err
	}
	resp.Body = wrappedBody
	return resp, nil
}

func waitForFirstToken(body io.ReadCloser, timeout time.Duration) (io.ReadCloser, error) {
	if timeout <= 0 {
		return body, nil
	}

	type readResult struct {
		data []byte
		err  error
	}

	buf := make([]byte, 4096)
	resultCh := make(chan readResult, 1)
	go func() {
		n, err := body.Read(buf)
		resultCh <- readResult{data: append([]byte(nil), buf[:n]...), err: err}
	}()

	select {
	case result := <-resultCh:
		if len(result.data) > 0 {
			reader := io.MultiReader(bytes.NewReader(result.data), body)
			return readCloser{Reader: reader, Closer: body}, nil
		}
		if result.err == nil {
			return body, nil
		}
		return nil, result.err
	case <-time.After(timeout):
		_ = body.Close()
		return nil, ErrFirstTokenTimeout
	}
}

type readCloser struct {
	io.Reader
	io.Closer
}
