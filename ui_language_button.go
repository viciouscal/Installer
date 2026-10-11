//go:build !cli

package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"golang.org/x/image/draw"
)

//go:embed assets/sa.png
var saFlagBytes []byte

//go:embed assets/us.png
var usFlagBytes []byte

var languageButton *g.ImageButtonWithRgbaWidget

func makeLanguageButton() {
	sa, err := png.Decode(bytes.NewReader(saFlagBytes))
	if err != nil {
		panic(err)
	}
	us, err := png.Decode(bytes.NewReader(usFlagBytes))
	if err != nil {
		panic(err)
	}
	icon := image.NewRGBA(image.Rect(0, 0, 104, 32))
	draw.BiLinear.Scale(icon, image.Rect(0, 0, 52, 32), sa, sa.Bounds(), draw.Src, nil)
	draw.BiLinear.Scale(icon, image.Rect(52, 0, 104, 32), us, us.Bounds(), draw.Src, nil)
	languageButton = g.ImageButtonWithRgba(icon).Size(104, 32).FramePadding(4).OnClick(func() { setUILanguage(!uiArabic) })
}

func fitPopup(widget g.Widget) g.Widget {
	return g.Custom(func() {
		w, _ := win.GetSize()
		width := float32(w - 96)
		if width > 760 {
			width = 760
		}
		if width < 320 {
			width = 320
		}
		imgui.SetNextWindowSize(imgui.Vec2{X: width, Y: 0})
		widget.Build()
	})
}
