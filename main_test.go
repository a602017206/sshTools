package main

import (
	"testing"

	"github.com/wailsapp/wails/v2/pkg/menu"
)

func TestBuildAppMenuIncludesEditMenu(t *testing.T) {
	appMenu := buildAppMenu(false, func() {})

	if !menuHasRole(appMenu, menu.EditMenuRole) {
		t.Fatalf("expected edit menu role to be present")
	}
}

func menuHasRole(appMenu *menu.Menu, role menu.Role) bool {
	for _, item := range appMenu.Items {
		if item.Role == role {
			return true
		}
	}
	return false
}
