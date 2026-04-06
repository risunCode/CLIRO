package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "CLIRO",
		Width:             1200,
		Height:            700,
		HideWindowOnClose: false,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "cliro-single-instance-v1",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 10, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeCloseGuard,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
