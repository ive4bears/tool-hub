package app

import (
	"context"
	"fmt"

	"tool-hub/backend/hub"

	"github.com/wailsapp/wails/v2/pkg/logger"
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

// Startup is called at application startup
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	isProduction := runtime.Environment(ctx).BuildType == "production"
	if isProduction {
		runtime.LogSetLogLevel(ctx, logger.INFO)
		disableStdout()
	}
	runtime.LogInfof(ctx, "env: %v", runtime.Environment(ctx))
	hub.InitDB(ctx, isProduction)
}

// DomReady is called after front-end resources have been loaded
func (a *App) DomReady(ctx context.Context) {
	// Add your action here
}

// BeforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) BeforeClose(ctx context.Context) (prevent bool) {
	return false
}

// Shutdown is called at application termination
func (a *App) Shutdown(ctx context.Context) {
	if logger, ok := a.ctx.Value("logger").(*WailsAdapter); ok {
		logger.Out.Close()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
