package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/build
var assets embed.FS

func main() {
	app := NewApp()

	var appContext context.Context

	onStartup := func(ctx context.Context) {
		appContext = ctx
		app.startup(ctx)
	}

	quitApp := func() {
		if appContext != nil {
			confirmed, err := runtime.MessageDialog(appContext, runtime.MessageDialogOptions{
				Type:          runtime.QuestionDialog,
				Title:         "确认退出",
				Message:       "确定要退出 AHaSSHTools 吗？",
				Buttons:       []string{"取消", "退出"},
				DefaultButton: "取消",
			})
			if err == nil && confirmed == "退出" {
				runtime.Quit(appContext)
			}
		}
	}

	AppMenu := buildAppMenu(true, appContext, func() {
		if appContext != nil {
			app.ShowAboutDialog()
		}
	}, quitApp)

	err := wails.Run(&options.App{
		Title:  "AHaSSHTools",
		Width:  1024,
		Height: 768,
		Menu:   AppMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        onStartup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
