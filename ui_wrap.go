package main

import "strings"

func wrapUIText(text string, width float32, measure func(string) float32) []string {
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		line := ""
		for _, word := range strings.Fields(paragraph) {
			candidate := word
			if line != "" {
				candidate = line + " " + word
			}
			if line != "" && measure(candidate) > width {
				lines = append(lines, line)
				line = ""
			}
			if line != "" {
				line += " "
			}
			for _, r := range word {
				if line != "" && measure(line+string(r)) > width {
					lines = append(lines, line)
					line = ""
				}
				line += string(r)
			}
		}
		lines = append(lines, line)
	}
	return lines
}
