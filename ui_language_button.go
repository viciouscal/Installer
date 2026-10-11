//go:build !cli

package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"math"

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
	bounds := image.Rect(0, 0, 72, 40)
	saImage := image.NewRGBA(bounds)
	usImage := image.NewRGBA(bounds)
	draw.BiLinear.Scale(saImage, bounds, sa, sa.Bounds(), draw.Src, nil)
	draw.BiLinear.Scale(usImage, bounds, us, us.Bounds(), draw.Src, nil)
	icon := image.NewRGBA(bounds)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if x*bounds.Dy()+y*bounds.Dx() < bounds.Dx()*bounds.Dy() {
				icon.SetRGBA(x, y, saImage.RGBAAt(x, y))
			} else {
				icon.SetRGBA(x, y, usImage.RGBAAt(x, y))
			}
			px, py := float64(x)+0.5, float64(y)+0.5
			cx := math.Max(10, math.Min(float64(bounds.Dx())-10, px))
			cy := math.Max(10, math.Min(float64(bounds.Dy())-10, py))
			coverage := math.Max(0, math.Min(1, 10.5-math.Hypot(px-cx, py-cy)))
			pixel := icon.RGBAAt(x, y)
			pixel.R = uint8(float64(pixel.R) * coverage)
			pixel.G = uint8(float64(pixel.G) * coverage)
			pixel.B = uint8(float64(pixel.B) * coverage)
			pixel.A = uint8(float64(pixel.A) * coverage)
			icon.SetRGBA(x, y, pixel)
		}
	}
	languageButton = g.ImageButtonWithRgba(icon).Size(72, 40).FramePadding(4).OnClick(func() { setUILanguage(!uiArabic) })
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
