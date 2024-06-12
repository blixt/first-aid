package firstaid

import (
	"fmt"
	"os"
	"path/filepath"
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

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}
	path = os.ExpandEnv(path)
	// Prefer a relative path when it goes deeper into the current directory.
	cwd, _ := os.Getwd()
	relPath, err := filepath.Rel(cwd, path)
	if err != nil || relPath == ".." || strings.HasPrefix(relPath, "../") {
		return path
	}
	return relPath
}
