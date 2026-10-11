package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestWrapUITextFitsAndPreservesWords(t *testing.T) {
	text := "تم تثبيت فينكورد العرب بنجاح.\nأغلق Discord بالكامل، ثم أعد تشغيله للتحقق من التثبيت."
	measure := func(value string) float32 { return float32(utf8.RuneCountInString(value)) }
	lines := wrapUIText(text, 20, measure)
	for _, line := range lines {
		if measure(line) > 20 {
			t.Fatalf("line exceeds available width: %q", line)
		}
	}
	if strings.Join(strings.Fields(strings.Join(lines, " ")), " ") != strings.Join(strings.Fields(text), " ") {
		t.Fatal("text was lost or reordered")
	}
}

func TestWrapUITextLongTokensAndBlankLines(t *testing.T) {
	measure := func(value string) float32 { return float32(utf8.RuneCountInString(value)) }
	lines := wrapUIText("abcdefghijk\n\nالعربية", 4, measure)
	if strings.Join(lines, "") != "abcdefghijkالعربية" {
		t.Fatal("long token lost text")
	}
	blank := false
	for _, line := range lines {
		if line == "" {
			blank = true
		}
		if measure(line) > 4 || !utf8.ValidString(line) {
			t.Fatalf("invalid wrapped line: %q", line)
		}
	}
	if !blank {
		t.Fatal("blank paragraph was removed")
	}
}
