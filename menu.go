package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/menu"
)

func buildAppMenu(isDarwin bool, ctx context.Context, showAbout func(), quitApp func()) *menu.Menu {
	appMenu := menu.NewMenu()

	if isDarwin {
		appSubmenu := appMenu.AddSubmenu("AHaSSHTools")

		appSubmenu.AddText("关于 AHaSSHTools", nil, func(cd *menu.CallbackData) {
			if showAbout != nil {
				showAbout()
			}
		})

		appSubmenu.AddSeparator()
		appSubmenu.AddText("隐藏 AHaSSHTools", nil, func(cd *menu.CallbackData) {
		})

		appSubmenu.AddSeparator()
		appSubmenu.AddText("退出 AHaSSHTools", nil, func(cd *menu.CallbackData) {
			if quitApp != nil {
				quitApp()
			}
		})
	}

	appMenu.Append(menu.EditMenu())

	return appMenu
}
