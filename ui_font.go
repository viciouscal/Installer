//go:build !cli

package main

import (
	_ "embed"
	g "github.com/AllenDang/giu"
	"strings"
	"vencordinstaller/buildinfo"
)

//go:embed assets/IBMPlexSansArabic-Bold.ttf
var arabicFont []byte

func setupUIFont() {
	g.SetDefaultFontFromBytes(arabicFont, 24)
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
	var lines g.Layout
	for _, line := range strings.Split(text, "\n") {
		if line == "" {
			lines = append(lines, g.Dummy(0, 12))
			continue
		}
		alignment := g.AlignLeft
		if uiArabic {
			alignment = g.AlignRight
		}
		lines = append(lines, g.Align(alignment).To(g.Label(visualArabic(line))))
	}
	return g.Column(lines...)
}
