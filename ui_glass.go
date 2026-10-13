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
		SetColor(g.StyleColorWindowBg, color.RGBA{20, 23, 36, 255}).
		SetColor(g.StyleColorPopupBg, color.RGBA{31, 35, 51, 250}).
		SetColor(g.StyleColorBorder, color.RGBA{218, 229, 255, 40}).
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
			blue := math.Exp(-((u-0.12)*(u-0.12) + (v-0.12)*(v-0.12)) * 8)
			pink := math.Exp(-((u-0.93)*(u-0.93) + (v-0.82)*(v-0.82)) * 9)
			bitmap.SetRGBA(x, y, color.RGBA{uint8((0.065 + blue*0.08 + pink*0.12) * 255), uint8((0.078 + blue*0.1 + pink*0.04) * 255), uint8((0.12 + blue*0.19 + pink*0.11) * 255), 255})
		}
	}
	g.EnqueueNewTextureFromRgba(bitmap, func(texture *g.Texture) { glassTexture.Store(texture); g.Update() })
}

func glassBackground() g.Widget {
	return g.Custom(func() {
		origin := imgui.WindowPos()
		size := imgui.WindowSize()
		list := imgui.GetWindowDrawList()
		if texture := glassTexture.Load(); texture != nil {
			g.GetCanvas().AddImage(texture, image.Pt(int(origin.X), int(origin.Y)), image.Pt(int(origin.X+size.X), int(origin.Y+size.Y)))
		}
		min := imgui.Vec2{X: origin.X + 20, Y: origin.Y + 20}
		max := imgui.Vec2{X: origin.X + size.X - 20, Y: origin.Y + size.Y - 20}
		list.AddRectFilled(min, max, imgui.Vec4{X: 0.82, Y: 0.87, Z: 1, W: 0.035}, 22, 15)
		list.AddRect(min, max, imgui.Vec4{X: 0.86, Y: 0.91, Z: 1, W: 0.14}, 22, 15, 1)
	})
}
