//go:build !cli

package main

import (
	"image"
	"image/color"
	"strings"

	g "github.com/AllenDang/giu"
)

func warningWithLink(text string) g.Widget {
	var layout g.Layout
	for _, line := range strings.Split(text, "\n") {
		if line != "vencordar.com" {
			layout = append(layout, uiParagraph(line))
			continue
		}
		layout = append(layout, g.Custom(func() {
			start := g.GetCursorPos()
			available, _ := g.GetAvailableRegion()
			width, _ := g.CalcTextSize("vencordar.com")
			if uiArabic && available > width+16 {
				g.SetCursorPos(image.Pt(start.X+int(available-width-16), start.Y))
			}
			g.Style().SetColor(g.StyleColorButton, color.RGBA{}).
				SetColor(g.StyleColorText, color.RGBA{0, 62, 125, 255}).
				SetStyle(g.StyleVarFramePadding, 8, 4).
				To(g.Button("vencordar.com").OnClick(func() { g.OpenURL("https://vencordar.com") })).Build()
			g.SetCursorPos(image.Pt(start.X, g.GetCursorPos().Y))
		}))
	}
	return g.Column(layout...)
}
