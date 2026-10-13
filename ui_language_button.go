//go:build !cli

package main

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync/atomic"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"golang.org/x/image/draw"
)

//go:embed assets/sa.png
var saFlagBytes []byte

//go:embed assets/us.png
var usFlagBytes []byte

var saLanguageButton atomic.Pointer[g.ImageButtonWidget]
var usLanguageButton atomic.Pointer[g.ImageButtonWidget]

func makeLanguageButton() {
	sa, err := png.Decode(bytes.NewReader(saFlagBytes))
	if err != nil {
		panic(err)
	}
	us, err := png.Decode(bytes.NewReader(usFlagBytes))
	if err != nil {
		panic(err)
	}
	g.EnqueueNewTextureFromRgba(roundedFlag(sa), func(texture *g.Texture) {
		saLanguageButton.Store(g.ImageButton(texture).Size(48, 28).FramePadding(0).
			BgColor(color.RGBA{}).OnClick(func() { setUILanguage(true) }))
		g.Update()
	})
	g.EnqueueNewTextureFromRgba(roundedFlag(us), func(texture *g.Texture) {
		usLanguageButton.Store(g.ImageButton(texture).Size(48, 28).FramePadding(0).
			BgColor(color.RGBA{}).OnClick(func() { setUILanguage(false) }))
		g.Update()
	})
}

func roundedFlag(flag image.Image) image.Image {
	bounds := image.Rect(0, 0, 48, 28)
	icon := image.NewRGBA(bounds)
	draw.BiLinear.Scale(icon, bounds, flag, flag.Bounds(), draw.Src, nil)
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			cx := math.Max(7, math.Min(float64(bounds.Dx())-7, px))
			cy := math.Max(7, math.Min(float64(bounds.Dy())-7, py))
			coverage := math.Max(0, math.Min(1, 7.5-math.Hypot(px-cx, py-cy)))
			pixel := icon.RGBAAt(x, y)
			pixel.R = uint8(float64(pixel.R) * coverage)
			pixel.G = uint8(float64(pixel.G) * coverage)
			pixel.B = uint8(float64(pixel.B) * coverage)
			pixel.A = uint8(float64(pixel.A) * coverage)
			icon.SetRGBA(x, y, pixel)
		}
	}
	return icon
}

func languageSelector() g.Widget {
	return g.Style().SetFont(uiFont(20)).
		SetStyleFloat(g.StyleVarFrameBorderSize, 0).
		SetColor(g.StyleColorButton, color.RGBA{}).
		SetColor(g.StyleColorButtonHovered, color.RGBA{}).
		SetColor(g.StyleColorButtonActive, color.RGBA{}).
		To(g.Custom(func() {
			start := g.GetCursorPos()
			available, _ := g.GetAvailableRegion()
			width, height := g.CalcTextSize("US - AR")
			x := start.X
			if available > width+60 {
				x += int(available - width - 60)
			}
			g.SetCursorPos(image.Pt(x, start.Y+int((28-height)/2)))
			g.Label("US - AR").Build()
			g.SetCursorPos(image.Pt(x+int(width)+12, start.Y))
			button := saLanguageButton.Load()
			if uiArabic {
				button = usLanguageButton.Load()
			}
			if button != nil {
				button.Build()
			}
			_, spacing := g.GetItemSpacing()
			g.SetCursorPos(image.Pt(start.X, start.Y+28+int(spacing)))
		}))
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
