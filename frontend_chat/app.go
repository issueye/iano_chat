package main

import (
	"context"
	"runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return "Hello " + name + "!"
}

// GetOS returns the current operating system
func (a *App) GetOS() string {
	return runtime.GOOS
}

// Minimize minimizes the window
func (a *App) Minimize() {
	runtime.WindowMinimise(a.ctx)
}

// Maximize maximizes the window
func (a *App) Maximize() {
	runtime.WindowMaximise(a.ctx)
}

// Unmaximize restores the window from maximized state
func (a *App) Unmaximize() {
	runtime.WindowUnmaximise(a.ctx)
}

// Close closes the application
func (a *App) Close() {
	runtime.Quit(a.ctx)
}
