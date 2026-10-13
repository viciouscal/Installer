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
		SetColor(g.StyleColorWindowBg, color.RGBA{12, 9, 18, 255}).
		SetColor(g.StyleColorPopupBg, color.RGBA{19, 14, 27, 250}).
		SetColor(g.StyleColorBorder, color.RGBA{235, 204, 255, 58}).
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
			purple := math.Exp(-((u-0.15)*(u-0.15)*3 + (v-0.25)*(v-0.25)*3)) * 0.30
			pink := math.Exp(-((u-0.95)*(u-0.95)*4 + (v-0.75)*(v-0.75)*3)) * 0.23
			reflection := math.Exp(-math.Pow((v+u*0.45-0.32)/0.12, 2)) * 5
			channel := func(value float64) uint8 { return uint8(math.Min(255, value)) }
			bitmap.SetRGBA(x, y, color.RGBA{channel(8 + 147*purple + 219*pink + reflection), channel(6 + 51*purple + 39*pink + reflection), channel(14 + 234*purple + 119*pink + reflection), 255})
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
		list := imgui.GetWindowDrawList()
		edge := imgui.Vec2{X: origin.X + size.X, Y: origin.Y + size.Y}
		list.AddRectFilled(origin, edge, imgui.Vec4{X: 1, Y: 0.94, Z: 1, W: 0.015}, 0, 0)
		list.AddLine(imgui.Vec2{X: origin.X, Y: origin.Y + 1}, imgui.Vec2{X: edge.X, Y: origin.Y + 1}, imgui.Vec4{X: 1, Y: 0.9, Z: 1, W: 0.12}, 1)
		list.AddRect(imgui.Vec2{X: origin.X + 1, Y: origin.Y + 1}, imgui.Vec2{X: edge.X - 1, Y: edge.Y - 1}, imgui.Vec4{X: 0.9, Y: 0.78, Z: 1, W: 0.07}, 0, 0, 1)

	})
}
