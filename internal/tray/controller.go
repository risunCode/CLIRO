package tray

import "context"

type MenuCallbacks struct {
	OnReady        func()
	OnOpen         func()
	OnToggleProxy  func() error
	OnExit         func()
	IsProxyRunning func() bool
}

type Controller interface {
	Supported() bool
	Available() bool
	Start(ctx context.Context, callbacks MenuCallbacks) error
	SetProxyRunning(running bool)
	Close() error
}

func NewController() Controller {
	return newPlatformController()
}
