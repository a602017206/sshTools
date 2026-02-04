package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/build
var assets embed.FS

func main() {
	app := NewApp()

	AppMenu := menu.NewMenu()

	if runtime.GOOS == "darwin" {
		appMenu := AppMenu.AddSubmenu("AHaSSHTools")

		appMenu.AddText("关于 AHaSSHTools", nil, func(cd *menu.CallbackData) {
			if app.ctx != nil {
				app.ShowAboutDialog()
			}
		})

		appMenu.AddSeparator()
		appMenu.AddText("隐藏 AHaSSHTools", nil, nil)
		appMenu.AddSeparator()
		appMenu.AddText("退出 AHaSSHTools", nil, nil)
	}

	err := wails.Run(&options.App{
		Title:  "AHaSSHTools",
		Width:  1024,
		Height: 768,
		Menu:   AppMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
