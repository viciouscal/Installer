//go:build !cli

package main

import (
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
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

func TestArabicAndLatinVisibleHeightsAreComparable(t *testing.T) {
	typeface, err := sfnt.Parse(arabicFont)
	if err != nil {
		t.Fatal(err)
	}
	var buffer sfnt.Buffer
	heights := make([]float64, 0, 2)
	for _, r := range []rune{'A', 0xfe95} {
		glyph, err := typeface.GlyphIndex(&buffer, r)
		if err != nil || glyph == 0 {
			t.Fatalf("missing glyph %U", r)
		}
		bounds, _, err := typeface.GlyphBounds(&buffer, glyph, fixed.I(100), font.HintingNone)
		if err != nil {
			t.Fatal(err)
		}
		heights = append(heights, float64(bounds.Max.Y-bounds.Min.Y))
	}
	ratio := heights[1] / heights[0]
	if ratio < 0.85 || ratio > 1.10 {
		t.Fatalf("Arabic/Latin height ratio %.2f", ratio)
	}
}
