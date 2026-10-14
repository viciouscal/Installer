//go:build !cli

package main

import (
	_ "embed"
	g "github.com/AllenDang/giu"
	"image"
	"runtime"
	"vencordinstaller/buildinfo"
)

//go:embed assets/VencordARUI-Bold.ttf
var arabicFont []byte

//go:embed assets/VencordARUI-Regular.ttf
var warningFont []byte

var warningFontInfo *g.FontInfo
var arabicFontInfo *g.FontInfo
var englishFontInfo *g.FontInfo

func setupUIFont() {
	g.SetDefaultFontFromBytes(arabicFont, 24)
	arabicFontInfo = g.AddFontFromBytes("VencordARUI", arabicFont, 24)
	englishFontInfo = arabicFontInfo
	warningFontInfo = g.AddFontFromBytes("VencordARUIRegular", warningFont, 18)
	warningFontInfo.SetSize(18)
	for _, size := range []float32{18, 20, 22, 24, 25, 30, 32, 35, 40} {
		arabicFontInfo.SetSize(size)
		englishFontInfo.SetSize(size)
	}
	g.Label(uiGlyphSeed())
}

func warningUIFont() *g.FontInfo {
	if !uiArabic {
		return uiFont(20)
	}
	size := float32(18)
	if runtime.GOOS == "windows" && win != nil {
		size *= g.Context.GetPlatform().GetContentScale()
	}
	return warningFontInfo.SetSize(size)
}

func uiFont(size float32) *g.FontInfo {
	if runtime.GOOS == "windows" && win != nil {
		size *= g.Context.GetPlatform().GetContentScale()
	}
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
