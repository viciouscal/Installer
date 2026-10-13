//go:build !cli && windows

package main

import (
	"golang.org/x/sys/windows"
	"unsafe"
)

func initialWindowSize() (int, int) {
	width, height := 1200, 700
	user32 := windows.NewLazySystemDLL("user32.dll")
	var area struct{ Left, Top, Right, Bottom int32 }
	result, _, _ := user32.NewProc("SystemParametersInfoW").Call(0x30, 0, uintptr(unsafe.Pointer(&area)), 0)
	if result == 0 {
		return width, height
	}
	scale := 1.0
	dpiProc := user32.NewProc("GetDpiForSystem")
	if dpiProc.Find() == nil {
		dpi, _, _ := dpiProc.Call()
		if dpi >= 96 {
			scale = float64(dpi) / 96
		}
	}
	maxWidth := int(float64(area.Right-area.Left)/scale) - 32
	maxHeight := int(float64(area.Bottom-area.Top)/scale) - 64
	if maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	if maxHeight > 0 && height > maxHeight {
		height = maxHeight
	}
	return width, height
}
