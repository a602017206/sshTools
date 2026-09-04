package main

import (
	"encoding/binary"
	"os"
	"testing"
)

// Wails 默认 icon.ico 里 256×256 PNG 正好是 12989 字节（黑底白圆角上的斜体 W）。
const defaultWails256PNGSize = 12989

func TestWindowsIconIsNotDefaultWailsLogo(t *testing.T) {
	data, err := os.ReadFile("build/windows/icon.ico")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 40000 {
		t.Fatalf("build/windows/icon.ico 仍像 Wails 默认图标（%d 字节，默认约 21KB）；应从 build/appicon.png 重新生成", len(data))
	}

	png, err := largestPNGFromICO(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(png) == defaultWails256PNGSize {
		t.Fatal("icon.ico 内嵌 256 图仍是 Wails 默认 W，Windows 打包会继续显示默认图标")
	}
	if len(png) < 40000 {
		t.Fatalf("icon.ico 内嵌最大 PNG 只有 %d 字节，不像当前 1024 的 appicon.png", len(png))
	}
}

func largestPNGFromICO(data []byte) ([]byte, error) {
	if len(data) < 6 {
		return nil, errInvalidICO("header")
	}
	count := int(binary.LittleEndian.Uint16(data[4:6]))
	var best []byte
	for i := 0; i < count; i++ {
		entry := 6 + i*16
		if len(data) < entry+16 {
			return nil, errInvalidICO("entry")
		}
		size := int(binary.LittleEndian.Uint32(data[entry+8 : entry+12]))
		offset := int(binary.LittleEndian.Uint32(data[entry+12 : entry+16]))
		if offset < 0 || size < 8 || offset+size > len(data) {
			return nil, errInvalidICO("payload")
		}
		chunk := data[offset : offset+size]
		if len(chunk) >= 8 && string(chunk[:8]) == "\x89PNG\r\n\x1a\n" && len(chunk) > len(best) {
			best = chunk
		}
	}
	if len(best) == 0 {
		return nil, errInvalidICO("no png")
	}
	return best, nil
}

type invalidICO string

func (e invalidICO) Error() string { return "invalid ico: " + string(e) }

func errInvalidICO(part string) error { return invalidICO(part) }
