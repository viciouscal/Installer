//go:build !cli

package main

import (
	g "github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"runtime"
	"testing"
)

func TestGlassBackgroundDrawsWithoutAssertion(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	context := imgui.CreateContext(nil)
	defer context.Destroy()
	io := imgui.CurrentIO()
	io.SetIniFilename("")
	io.SetDisplaySize(imgui.Vec2{X: 1200, Y: 800})
	io.SetDeltaTime(1.0 / 60)
	io.Fonts().TextureDataRGBA32()
	previous := glassTexture.Load()
	defer glassTexture.Store(previous)
	for frame := 0; frame < 2; frame++ {
		glassTexture.Store(nil)
		if frame == 1 {
			glassTexture.Store(g.ToTexture(imgui.TextureID(1)))
		}
		imgui.NewFrame()
		imgui.SetNextWindowSize(imgui.Vec2{X: 1200, Y: 800})
		glassTheme().To(g.Custom(func() {
			imgui.Begin("Glass preview")
			glassBackground().Build()
			imgui.End()
		})).Build()
		imgui.Render()
		if !imgui.RenderedDrawData().Valid() {
			t.Fatal("invalid rendered frame")
		}
	}
}
