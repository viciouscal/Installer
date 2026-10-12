package main

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

//go:embed assets/ar.json
var arabicTranslationsJSON []byte

var uiArabic = true
var arabicTranslations map[string]string
var renderedArabic sync.Map

func uiGlyphSeed() string {
	var seed strings.Builder
	for r := rune(32); r <= 255; r++ {
		seed.WriteRune(r)
	}
	for r := rune(0x600); r <= 0x6ff; r++ {
		seed.WriteRune(r)
	}
	for _, forms := range arabicForms {
		for _, r := range forms {
			if r != 0 {
				seed.WriteRune(r)
			}
		}
	}
	for _, forms := range lamAlefForms {
		for _, r := range forms {
			seed.WriteRune(r)
		}
	}
	for key, value := range arabicTranslations {
		seed.WriteString(key)
		seed.WriteString(value)
		seed.WriteString(visualArabic(value))
	}
	return seed.String()
}

func init() {
	if err := json.Unmarshal(arabicTranslationsJSON, &arabicTranslations); err != nil {
		panic(err)
	}
}

func translatedUI(text string) string {
	if !uiArabic {
		return text
	}
	if translated, ok := arabicTranslations[text]; ok {
		return translated
	}
	if strings.HasPrefix(text, "Error: Failed to create: ") {
		return "خطأ: تعذر إنشاء المجلد:\n" + strings.TrimPrefix(text, "Error: Failed to create: ")
	}
	return strings.ReplaceAll(text, " (Vencord Arab Installed)", " (تم تثبيت فينكورد العرب)")
}

func uiText(text string) string {
	text = translatedUI(text)
	if cached, ok := renderedArabic.Load(text); ok {
		return cached.(string)
	}
	rendered := visualArabic(text)
	renderedArabic.Store(text, rendered)
	return rendered
}
