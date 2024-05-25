package firstaid

import (
	"fmt"
	"regexp"
	"strings"
)

var reWhitespace = regexp.MustCompile(`\s+`)

func line(n int) string {
	if n == 1 {
		return fmt.Sprintf("%d line", n)
	}
	return fmt.Sprintf("%d lines", n)
}

func FirstLine(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	firstLineString := reWhitespace.ReplaceAllString(strings.TrimSpace(lines[0]), " ")
	if len(firstLineString) > 50 {
		firstLineString = firstLineString[:49] + "…"
	}
	if len(lines) > 1 {
		return fmt.Sprintf("`%s` (+%s)", firstLineString, line(len(lines)-1))
	}
	return fmt.Sprintf("`%s`", firstLineString)
}

func FirstLineBytes(data []byte) string {
	return FirstLineString(string(data))
}

func FirstLineString(s string) string {
	return FirstLine(strings.Split(strings.TrimSpace(s), "\n"))
}
