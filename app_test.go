package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cliro-go/internal/account"
	"cliro-go/internal/config"
	"cliro-go/internal/tray"
	"github.com/wailsapp/wails/v2/pkg/options"
)

func TestBuildSecondLaunchNotice(t *testing.T) {
	notice := buildSecondLaunchNotice(options.SecondInstanceData{
		Args:             []string{"--foo", "bar"},
		WorkingDirectory: `C:\Users\AceLova`,
	})

	if notice.Message == "" {
		t.Fatalf("expected non-empty message")
	}
	if notice.WorkingDirectory != `C:\Users\AceLova` {
		t.Fatalf("working directory = %q", notice.WorkingDirectory)
	}
	if len(notice.Args) != 2 {
		t.Fatalf("args length = %d, want 2", len(notice.Args))
	}
	if notice.ReceivedAt == 0 {
		t.Fatalf("expected received timestamp")
	}
}

func TestGetStateIncludesStartupWarnings(t *testing.T) {
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "accounts.json"), []byte(`{"accounts":[{"id":"legacy"}]}`), 0o600); err != nil {
		t.Fatalf("write legacy accounts: %v", err)
	}

	store, err := config.NewManager(dataDir)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	app := &App{store: store, pool: account.NewPool(store)}
	state := app.GetState()
	if len(state.StartupWarnings) == 0 {
		t.Fatalf("expected startup warnings in app state")
	}
}

func TestBeforeCloseGuardInterceptsByDefault(t *testing.T) {
	app := NewApp()
	app.ctx = context.Background()

	eventNames := make([]string, 0, 1)
	app.emitEvent = func(_ context.Context, name string, _ ...interface{}) {
		eventNames = append(eventNames, name)
	}

	preventClose := app.beforeCloseGuard(context.Background())
	if !preventClose {
		t.Fatalf("beforeCloseGuard() = false, want true")
	}
	if len(eventNames) != 1 || eventNames[0] != "app:close-requested" {
		t.Fatalf("events = %v, want [app:close-requested]", eventNames)
	}
}

func TestConfirmQuitAuthorizationIsOneShot(t *testing.T) {
	app := NewApp()
	app.ctx = context.Background()

	quitCalls := 0
	app.quitApp = func(context.Context) {
		quitCalls++
	}

	eventNames := make([]string, 0, 2)
	app.emitEvent = func(_ context.Context, name string, _ ...interface{}) {
		eventNames = append(eventNames, name)
	}

	if err := app.ConfirmQuit(); err != nil {
		t.Fatalf("ConfirmQuit() error = %v", err)
	}
	if quitCalls != 1 {
		t.Fatalf("quit calls = %d, want 1", quitCalls)
	}
	if app.beforeCloseGuard(context.Background()) {
		t.Fatalf("first guard after ConfirmQuit should allow close")
	}
	if !app.beforeCloseGuard(context.Background()) {
		t.Fatalf("second guard after ConfirmQuit should intercept")
	}
	if len(eventNames) == 0 || eventNames[len(eventNames)-1] != "app:close-requested" {
		t.Fatalf("expected close-requested event after one-shot consumed, got %v", eventNames)
	}
}

func TestGetStateIncludesTrayCapabilityWithoutStore(t *testing.T) {
	app := &App{tray: &fakeTrayController{supported: true, available: false}}

	state := app.GetState()
	if !state.TraySupported {
		t.Fatalf("TraySupported = false, want true")
	}
	if state.TrayAvailable {
		t.Fatalf("TrayAvailable = true, want false")
	}
}

func TestToggleProxyByStateChoosesStartOrStop(t *testing.T) {
	startCalls := 0
	stopCalls := 0
	start := func() error {
		startCalls++
		return nil
	}
	stop := func() error {
		stopCalls++
		return nil
	}

	if err := toggleProxyByState(false, start, stop); err != nil {
		t.Fatalf("toggleProxyByState(false) error = %v", err)
	}
	if startCalls != 1 || stopCalls != 0 {
		t.Fatalf("calls after start path: start=%d stop=%d", startCalls, stopCalls)
	}

	if err := toggleProxyByState(true, start, stop); err != nil {
		t.Fatalf("toggleProxyByState(true) error = %v", err)
	}
	if startCalls != 1 || stopCalls != 1 {
		t.Fatalf("calls after stop path: start=%d stop=%d", startCalls, stopCalls)
	}
}

type fakeTrayController struct {
	supported bool
	available bool
}

var _ tray.Controller = (*fakeTrayController)(nil)

func (f *fakeTrayController) Supported() bool { return f.supported }
func (f *fakeTrayController) Available() bool { return f.available }
func (f *fakeTrayController) Start(context.Context, tray.MenuCallbacks) error {
	return nil
}
func (f *fakeTrayController) SetProxyRunning(bool) {}
func (f *fakeTrayController) Close() error         { return nil }
