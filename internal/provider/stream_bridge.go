package provider

import (
	"bytes"
	"errors"
	"io"
	"time"
)

var ErrEmptyStream = errors.New("empty upstream stream")
var ErrStreamProbeTimeout = errors.New("stream probe timeout")

type StreamProbeResult struct {
	Reader           io.ReadCloser
	UpstreamReadable bool
	EmptyStream      bool
	OpenDuration     time.Duration
}

type StreamBridge struct {
	ProbeSize int
}

func (b StreamBridge) OpenVerified(body io.ReadCloser, timeout time.Duration) (StreamProbeResult, error) {
	if body == nil {
		return StreamProbeResult{EmptyStream: true}, ErrEmptyStream
	}
	probeSize := b.ProbeSize
	if probeSize <= 0 {
		probeSize = 4096
	}

	started := time.Now()
	type readResult struct {
		data []byte
		err  error
	}
	resultCh := make(chan readResult, 1)
	go func() {
		buf := make([]byte, probeSize)
		n, err := body.Read(buf)
		resultCh <- readResult{data: append([]byte(nil), buf[:n]...), err: err}
	}()

	var result readResult
	if timeout > 0 {
		select {
		case result = <-resultCh:
		case <-time.After(timeout):
			_ = body.Close()
			return StreamProbeResult{OpenDuration: time.Since(started)}, ErrStreamProbeTimeout
		}
	} else {
		result = <-resultCh
	}

	probe := StreamProbeResult{OpenDuration: time.Since(started)}
	if len(result.data) > 0 {
		probe.UpstreamReadable = true
		probe.Reader = readCloser{Reader: io.MultiReader(bytes.NewReader(result.data), body), Closer: body}
		return probe, nil
	}
	if errors.Is(result.err, io.EOF) || result.err == nil {
		_ = body.Close()
		probe.EmptyStream = true
		return probe, ErrEmptyStream
	}
	_ = body.Close()
	return probe, result.err
}

func CanRetryStreamFailure(clientBytesSent bool) bool {
	return !clientBytesSent
}

type readCloser struct {
	io.Reader
	io.Closer
}
