package main

import (
	"embed"
	"log"

	"tool-hub/backend/app"
	appPkg "tool-hub/backend/app"
	"tool-hub/backend/hub"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/public
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

const appName = "tool-hub"

func main() {
	// Create an instance of the app structure
	app := app.NewApp()
	wailsLogger, err := appPkg.InitLogger(appName)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	// Create application with options
	err = wails.Run(&options.App{
		Title:             appName,
		Width:             1024,
		Height:            768,
		MinWidth:          1024,
		MinHeight:         768,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Menu:               nil,
		Logger:             wailsLogger,
		LogLevel:           logger.DEBUG,
		LogLevelProduction: logger.DEBUG,
		OnStartup:          app.Startup,
		OnDomReady:         app.DomReady,
		OnBeforeClose:      app.BeforeClose,
		OnShutdown:         app.Shutdown,
		WindowStartState:   options.Normal,
		Bind: []any{
			app,
			hub.ExposeModel(),
		},
		EnumBind: []any{
			genStringEnumBinds(),
			genIntEnumBinds(),
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
			ZoomFactor:          1.0,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "tool-hub",
				Message: "",
				Icon:    icon,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

type StringValues string

type StringEnumItem struct {
	Value  StringValues
	TSName string
}

func genStringEnumBinds() []StringEnumItem {
	return []StringEnumItem{
		{StringValues("white"), "ColorOfIcebear"},
		{StringValues("pink"), "ColorOfMeiMeiBear"},
		{StringValues("MeiMeiBear"), "MyFavoriteBear"},
	}
}

type IntValues int

type IntEnumItem struct {
	Value  IntValues
	TSName string
}

func genIntEnumBinds() []IntEnumItem {
	return []IntEnumItem{
		{IntValues(4), "NumberOfBears"},
	}
}
