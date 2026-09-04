//go:build windows

package service

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const cfHDROP = 15

func readClipboardFilePaths() ([]string, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	shell32 := windows.NewLazySystemDLL("shell32.dll")
	isClipboardFormatAvailable := user32.NewProc("IsClipboardFormatAvailable")
	openClipboard := user32.NewProc("OpenClipboard")
	closeClipboard := user32.NewProc("CloseClipboard")
	getClipboardData := user32.NewProc("GetClipboardData")
	dragQueryFile := shell32.NewProc("DragQueryFileW")

	available, _, _ := isClipboardFormatAvailable.Call(cfHDROP)
	if available == 0 {
		return nil, nil
	}
	opened, _, err := openClipboard.Call(0)
	if opened == 0 {
		return nil, err
	}
	defer closeClipboard.Call()

	handle, _, _ := getClipboardData.Call(cfHDROP)
	if handle == 0 {
		return nil, nil
	}

	count, _, _ := dragQueryFile.Call(handle, 0xFFFFFFFF, 0, 0)
	if count == 0 {
		return nil, nil
	}
	paths := make([]string, 0, int(count))
	for i := uintptr(0); i < count; i++ {
		n, _, _ := dragQueryFile.Call(handle, i, 0, 0)
		if n == 0 {
			continue
		}
		buf := make([]uint16, n+1)
		dragQueryFile.Call(handle, i, uintptr(unsafe.Pointer(&buf[0])), n+1)
		paths = append(paths, syscall.UTF16ToString(buf))
	}
	return paths, nil
}
