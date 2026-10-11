package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/bidi"
)

// arabicForms is generated from Unicode 15.0 UnicodeData.txt. The table and
// Unicode license are included with this fork. ImGui draws text without shaping,
// so convert logical Arabic to contextual presentation forms at display time.
func shapeArabic(text string) string {
	runes := []rune(text)
	shaped := make([]rune, 0, len(runes))
	for i, current := range runes {
		forms, ok := arabicForms[current]
		if !ok {
			shaped = append(shaped, current)
			continue
		}
		previous := i - 1
		for previous >= 0 && unicode.Is(unicode.Mn, runes[previous]) {
			previous--
		}
		next := i + 1
		for next < len(runes) && unicode.Is(unicode.Mn, runes[next]) {
			next++
		}
		joinPrevious := previous >= 0 && arabicForms[runes[previous]][2] != 0 && forms[1] != 0
		joinNext := next < len(runes) && forms[2] != 0 && arabicForms[runes[next]][1] != 0
		// Lam/alef ligatures avoid a gap between the two joined letters.
		if current == '\u0644' && i+1 < len(runes) {
			if ligature, ok := lamAlefForms[runes[i+1]]; ok {
				form := 0
				if joinPrevious {
					form = 1
				}
				shaped = append(shaped, ligature[form])
				runes[i+1] = 0 // Skip the consumed alef, which cannot connect forward.
				continue
			}
		}
		index := 0
		switch {
		case joinPrevious && joinNext:
			index = 3
		case joinPrevious:
			index = 1
		case joinNext:
			index = 2
		}
		replacement := forms[index]
		if replacement == 0 {
			replacement = current
		}
		shaped = append(shaped, replacement)
	}
	return strings.ReplaceAll(string(shaped), "\x00", "")
}

// visualArabic preserves Latin names, paths, numbers and each line's position.
// Keep stored strings logical; only pass this result to the ImGui renderer.
func visualArabic(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		hasArabic := false
		for _, r := range line {
			if r >= 0x0600 && r <= 0x06ff {
				hasArabic = true
				break
			}
		}
		if !hasArabic {
			continue
		}
		var paragraph bidi.Paragraph
		if _, err := paragraph.SetString(shapeArabic(line), bidi.DefaultDirection(bidi.RightToLeft)); err != nil {
			continue
		}
		order, err := paragraph.Order()
		if err != nil {
			continue
		}
		var result strings.Builder
		for j := order.NumRuns() - 1; j >= 0; j-- {
			run := order.Run(j)
			value := run.String()
			if run.Direction() == bidi.RightToLeft {
				value = bidi.ReverseString(value)
			}
			result.WriteString(value)
		}
		lines[i] = result.String()
	}
	return strings.Join(lines, "\n")
}
