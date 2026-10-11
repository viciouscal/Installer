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
