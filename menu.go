package main

import "github.com/wailsapp/wails/v2/pkg/menu"

func buildAppMenu(isDarwin bool, showAbout func()) *menu.Menu {
	appMenu := menu.NewMenu()

	if isDarwin {
		appSubmenu := appMenu.AddSubmenu("AHaSSHTools")

		appSubmenu.AddText("关于 AHaSSHTools", nil, func(cd *menu.CallbackData) {
			if showAbout != nil {
				showAbout()
			}
		})

		appSubmenu.AddSeparator()
		appSubmenu.AddText("隐藏 AHaSSHTools", nil, nil)
		appSubmenu.AddSeparator()
		appSubmenu.AddText("退出 AHaSSHTools", nil, nil)
	}

	appMenu.Append(menu.EditMenu())

	return appMenu
}
