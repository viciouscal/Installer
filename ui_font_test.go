//go:build !cli

package main

import (
	"golang.org/x/image/font/sfnt"
	"testing"
)

func TestArabicStartupFontContainsShapedLetters(t *testing.T) {
	font, err := sfnt.Parse(arabicFont)
	if err != nil {
		t.Fatal(err)
	}
	var buffer sfnt.Buffer
	for key, text := range arabicTranslations {
		for _, r := range visualArabic(text) {
			if r < 0x600 || r == 0x200e || r == 0x200f {
				continue
			}
			glyph, err := font.GlyphIndex(&buffer, r)
			if err != nil || glyph == 0 {
				t.Fatalf("font missing %U for %s: %v", r, key, err)
			}
		}
	}
}
