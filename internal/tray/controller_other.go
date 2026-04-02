//go:build !windows

package tray

import "context"

type noopController struct{}

func newPlatformController() Controller {
	return &noopController{}
}

func (c *noopController) Supported() bool {
	return false
}

func (c *noopController) Available() bool {
	return false
}

func (c *noopController) Start(_ context.Context, _ MenuCallbacks) error {
	return nil
}

func (c *noopController) SetProxyRunning(_ bool) {}

func (c *noopController) Close() error {
	return nil
}
