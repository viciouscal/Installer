//go:build !cli

package main

import (
	_ "embed"
	g "github.com/AllenDang/giu"
	"image"
	"vencordinstaller/buildinfo"
)

//go:embed assets/IBMPlexSansArabic-Bold.ttf
var arabicFont []byte

//go:embed assets/IBMPlexSans-Bold.ttf
var englishFont []byte

var arabicFontInfo *g.FontInfo
var englishFontInfo *g.FontInfo

func setupUIFont() {
	g.SetDefaultFontFromBytes(arabicFont, 24)
	arabicFontInfo = g.AddFontFromBytes("IBMPlexSansArabic-Bold", arabicFont, 24)
	englishFontInfo = g.AddFontFromBytes("IBMPlexSans-Bold", englishFont, 24)
}

func uiFont(size float32) *g.FontInfo {
	if uiArabic {
		return arabicFontInfo.SetSize(size)
	}
	return englishFontInfo.SetSize(size)
}

func localizedWindowTitle() string {
	title := translatedUI("Vencord Arab Installer")
	if buildinfo.InstallerTag == buildinfo.VersionUnknown {
		title += " [" + translatedUI("Preview") + "]"
	}
	return title
}

func setUILanguage(arabic bool) {
	uiArabic = arabic
	win.SetTitle(localizedWindowTitle())
	g.Update()
}

func uiParagraph(text string) g.Widget {
	text = translatedUI(text)
	return g.Custom(func() {
		start := g.GetCursorPos()
		width, _ := g.GetAvailableRegion()
		for _, line := range wrapUIText(text, width, func(value string) float32 {
			w, _ := g.CalcTextSize(visualArabic(value))
			return w
		}) {
			if line == "" {
				g.Dummy(0, 12).Build()
				continue
			}
			visual := visualArabic(line)
			x := start.X
			if uiArabic {
				w, _ := g.CalcTextSize(visual)
				if width > w {
					x += int(width - w)
				}
			}
			g.SetCursorPos(image.Pt(x, g.GetCursorPos().Y))
			g.Label(visual).Build()
			g.SetCursorPos(image.Pt(start.X, g.GetCursorPos().Y))
		}
	})
}
