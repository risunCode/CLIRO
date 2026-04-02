//go:build !windows

package tray

import (
	"context"
	"testing"
)

func TestNoopControllerCapabilities(t *testing.T) {
	c := NewController()
	if c.Supported() {
		t.Fatalf("Supported() = true, want false")
	}
	if c.Available() {
		t.Fatalf("Available() = true, want false")
	}
	if err := c.Start(context.Background(), MenuCallbacks{}); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	c.SetProxyRunning(true)
	if err := c.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
