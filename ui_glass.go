//go:build !cli

package main

import (
	"image"
	"image/color"
	"math"
	"sync/atomic"

	g "github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
)

func glassTheme() *g.StyleSetter {
	return g.Style().
		SetColor(g.StyleColorWindowBg, color.RGBA{0, 0, 0, 255}).
		SetColor(g.StyleColorPopupBg, color.RGBA{14, 10, 18, 255}).
		SetColor(g.StyleColorBorder, color.RGBA{176, 57, 111, 80}).
		SetColor(g.StyleColorFrameBg, color.RGBA{190, 205, 240, 22}).
		SetColor(g.StyleColorFrameBgHovered, color.RGBA{200, 215, 250, 38}).
		SetColor(g.StyleColorFrameBgActive, color.RGBA{200, 215, 250, 54}).
		SetColor(g.StyleColorButton, color.RGBA{193, 211, 255, 30}).
		SetColor(g.StyleColorButtonHovered, color.RGBA{210, 225, 255, 52}).
		SetColor(g.StyleColorButtonActive, color.RGBA{210, 225, 255, 72}).
		SetStyleFloat(g.StyleVarFrameBorderSize, 1).
		SetStyleFloat(g.StyleVarFrameRounding, 14).
		SetStyleFloat(g.StyleVarChildRounding, 14)
}

var glassTexture atomic.Pointer[g.Texture]

func prepareGlassBackground() {
	bitmap := image.NewRGBA(image.Rect(0, 0, 512, 512))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			u, v := float64(x)/512, float64(y)/512
			purple := math.Max(0, 1-math.Hypot(u-0.5, v-0.5)/0.5) * 0.05
			pink := math.Max(0, 1-math.Hypot(u-0.8, v-0.2)/0.5) * 0.05
			bitmap.SetRGBA(x, y, color.RGBA{uint8(147*purple + 219*pink), uint8(51*purple + 39*pink), uint8(234*purple + 119*pink), 255})
		}
	}
	g.EnqueueNewTextureFromRgba(bitmap, func(texture *g.Texture) { glassTexture.Store(texture); g.Update() })
}

func glassBackground() g.Widget {
	return g.Custom(func() {
		origin := imgui.WindowPos()
		size := imgui.WindowSize()
		imgui.PushClipRect(origin, imgui.Vec2{X: origin.X + size.X, Y: origin.Y + size.Y}, false)
		defer imgui.PopClipRect()
		if texture := glassTexture.Load(); texture != nil {
			g.GetCanvas().AddImage(texture, image.Pt(int(origin.X), int(origin.Y)), image.Pt(int(origin.X+size.X), int(origin.Y+size.Y)))
		}
	})
}
