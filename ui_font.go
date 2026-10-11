//go:build !cli

package main

import (
	_ "embed"
	g "github.com/AllenDang/giu"
	"strings"
	"vencordinstaller/buildinfo"
)

// Amiri supports Arabic presentation forms and Latin. Bundle it so Arabic
// works on systems without Arabic fonts; assets/Amiri-OFL.txt accompanies it.
//
//go:embed assets/Amiri-Regular.ttf
var arabicFont []byte

func setupUIFont() {
	g.SetDefaultFontFromBytes(arabicFont, 24)
}

func localizedWindowTitle() string {
	title := translatedUI("Vencord Installer")
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

func uiActionRow(widgets ...g.Widget) g.Widget {
	if uiArabic {
		for i, j := 0, len(widgets)-1; i < j; i, j = i+1, j-1 {
			widgets[i], widgets[j] = widgets[j], widgets[i]
		}
	}
	return g.Row(widgets...)
}
