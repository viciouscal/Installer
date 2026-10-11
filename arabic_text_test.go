package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"golang.org/x/image/font/sfnt"
)

func TestArabicJoiningAndDirection(t *testing.T) {
	cases := map[string]string{
		"بب":                    "\ufe90\ufe91",
		"باب":                   "\ufe8f\ufe8e\ufe91",
		"لا":                    "\ufefb",
		"بلا":                   "\ufefc\ufe91",
		"Vencord Arab 403":      "Vencord Arab 403",
		`C:\Users\Test\Discord`: `C:\Users\Test\Discord`,
	}
	for logical, expected := range cases {
		if actual := visualArabic(logical); actual != expected {
			t.Errorf("%q: got %U, expected %U", logical, []rune(actual), []rune(expected))
		}
	}
}

func TestArabicMixedTextAndLines(t *testing.T) {
	actual := visualArabic("تثبيت Vencord Arab 403\nإلغاء")
	if !strings.Contains(actual, "Vencord Arab 403") || strings.Contains(actual, "dro cneV") || strings.Count(actual, "\n") != 1 {
		t.Fatalf("mixed text or line order corrupted: %q", actual)
	}
	if !strings.HasSuffix(actual, "\n"+visualArabic("إلغاء")) {
		t.Fatal("line order changed")
	}
}

func TestArabicTranslationsAndFontCoverage(t *testing.T) {
	bytes, err := os.ReadFile("assets/IBMPlexSansArabic-Bold.ttf")
	if err != nil {
		t.Fatal(err)
	}
	font, err := sfnt.Parse(bytes)
	if err != nil {
		t.Fatal(err)
	}
	var buffer sfnt.Buffer
	for key, translation := range arabicTranslations {
		if translation == "" {
			t.Errorf("empty translation for %q", key)
		}
		for _, r := range visualArabic(translation) {
			if unicode.IsSpace(r) {
				continue
			}
			glyph, err := font.GlyphIndex(&buffer, r)
			if err != nil || glyph == 0 {
				t.Errorf("missing glyph %U in %q: %v", r, key, err)
			}
		}
	}
}

func TestEnglishFontCoverage(t *testing.T) {
	data, err := os.ReadFile("assets/IBMPlexSans-Bold.ttf")
	if err != nil {
		t.Fatal(err)
	}
	font, err := sfnt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buffer sfnt.Buffer
	texts := []string{"US - AR", "vencordar.com"}
	for key := range arabicTranslations {
		texts = append(texts, key)
	}
	for _, text := range texts {
		for _, r := range text {
			if unicode.IsSpace(r) {
				continue
			}
			glyph, err := font.GlyphIndex(&buffer, r)
			if err != nil || glyph == 0 {
				t.Errorf("missing English glyph %U: %v", r, err)
			}
		}
	}
}

func TestLanguageSwitch(t *testing.T) {
	original := uiArabic
	t.Cleanup(func() { uiArabic = original })
	uiArabic = true
	if translatedUI("Install") != "تثبيت" || uiText("Install") != visualArabic("تثبيت") {
		t.Fatal("Arabic translation missing")
	}
	if !strings.Contains(translatedUI("Discord (Vencord Arab Installed)"), "تم تثبيت") {
		t.Fatal("dynamic status not translated")
	}
	uiArabic = false
	if uiText("Install") != "Install" {
		t.Fatal("English selection ignored")
	}
}

func TestProjectBrandAndWarning(t *testing.T) {
	original := uiArabic
	t.Cleanup(func() { uiArabic = original })
	uiArabic = true
	if translatedUI("Vencord Arab Installer") != "مثبت فينكورد العرب" {
		t.Fatal("incorrect Arabic project name")
	}
	for _, text := range []string{"فينكورد العرب", "غير تابع", "أزلها", "البرمجيات الخبيثة", "كلمة مرور Discord"} {
		if !strings.Contains(safetyWarning(), text) {
			t.Errorf("Arabic warning missing %q", text)
		}
	}
	for key, translation := range arabicTranslations {
		if strings.Contains(translation, "Vencord") {
			t.Errorf("English project name in Arabic translation %q", key)
		}
	}
	uiArabic = false
	for _, text := range []string{"Vencord Arab", "independent", "uninstall", "malware", "Discord password"} {
		if !strings.Contains(safetyWarning(), text) {
			t.Errorf("English warning missing %q", text)
		}
	}
}

func literalUIString(expr ast.Expr) (string, bool) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		if value.Kind == token.STRING {
			text, err := strconv.Unquote(value.Value)
			return text, err == nil
		}
	case *ast.BinaryExpr:
		if value.Op == token.ADD {
			left, okLeft := literalUIString(value.X)
			right, okRight := literalUIString(value.Y)
			return left + right, okLeft && okRight
		}
	}
	return "", false
}

func TestStaticGUITranslationCoverage(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "gui.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name, ok := call.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		var expressions []ast.Expr
		switch name.Name {
		case "uiText", "uiParagraph", "Tooltip":
			if len(call.Args) > 0 {
				expressions = call.Args[:1]
			}
		case "InfoModal", "InfoModalExtra", "RawInfoModal":
			if len(call.Args) > 2 {
				expressions = call.Args[1:3]
			}
		}
		for _, expr := range expressions {
			text, constant := literalUIString(expr)
			if !constant || text == "" || text == "OpenAsar" {
				continue
			}
			if _, exists := arabicTranslations[text]; !exists {
				t.Errorf("missing GUI translation: %q", text)
			}
		}
		return true
	})
}
