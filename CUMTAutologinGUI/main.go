package main

import (
	"embed"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if !ensureSingleInstance() {
		return
	}
	defer releaseSingleInstance()
	initLogging()
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "CUMTAutologinGUI",
		Width:  1024,
		Height: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 16, B: 32, A: 1},
		OnStartup:        app.Startup,
		OnShutdown:       app.Shutdown,
		Windows: &windows.Options{
			// Keep the main window available but let tray control visibility.
			DisableWindowIcon: false,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
		log.Printf("[gui] wails run error: %v", err)
	}
}

func initLogging() {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("[gui] get executable path error: %v", err)
		return
	}
	exeDir := filepath.Dir(exe)
	logPath := filepath.Join(exeDir, "gui_app.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[gui] open log file error: %v", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stderr, f))
	log.Printf("[gui] logging to %s", logPath)
}
