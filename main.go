package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appService := NewApp()

	app := application.New(application.Options{
		Name:        "TimeSpan",
		Description: "Minimal task tracker with per-task timers",
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Services: []application.Service{
			application.NewService(appService),
		},
	})

	// Main window
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "TimeSpan",
		Width:            400,
		Height:           600,
		MinWidth:         320,
		MinHeight:        480,
		BackgroundColour: application.NewRGB(0x12, 0x12, 0x14),
		URL:              "/",
	})

	// Floating timer window
	timerWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "Timer",
		Width:          220,
		Height:         100,
		AlwaysOnTop:    true,
		Frameless:      true,
		DisableResize:  true,
		Hidden:         true,
		BackgroundType: application.BackgroundTypeTransparent,
		URL:            "/timer.html",
		HideOnEscape:   true,
	})

	appService.SetTimerWindow(timerWindow)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
