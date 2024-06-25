package firstaid

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blixt/first-aid/tool"
)

type ListFilesParams struct {
	Path  string `json:"path"`
	Depth int    `json:"depth,omitempty"`
}

type FileInfo struct {
	Type            string `json:"type"`
	Lines           int    `json:"lines,omitempty"`
	Count           int    `json:"count,omitempty"`
	ContentsSkipped bool   `json:"contentsSkipped,omitempty"`
}

var ListFiles = tool.Func(
	"List files",
	"Lists some of the contents in the specified directory. Don't use this on files. Don't use a depth higher than 2 unless you're really sure.",
	"list_files",
	func(r tool.Runner, p ListFilesParams) tool.Result {
		if p.Depth < 1 {
			p.Depth = 1
		}
		p.Path = expandPath(p.Path)

		items := make(map[string]FileInfo)
		entries := 0

		err := filepath.WalkDir(p.Path, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(p.Path, path)
			if relPath == "." {
				// Skip the root directory itself.
				return nil
			}
			depth := len(strings.Split(relPath, string(os.PathSeparator)))
			if depth > p.Depth {
				return filepath.SkipDir
			}
			entries++
			if entries >= 1_000 {
				if d.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
			if d.IsDir() {
				subItems, _ := os.ReadDir(path)
				items[relPath] = FileInfo{
					Type:            "directory",
					Count:           len(subItems),
					ContentsSkipped: depth == p.Depth,
				}
				switch d.Name() {
				case ".git", "node_modules":
					return filepath.SkipDir
				}
			} else {
				file, _ := os.Open(path)
				scanner := bufio.NewScanner(file)
				lines := 0
				for scanner.Scan() {
					lines++
				}
				items[relPath] = FileInfo{
					Type:  "file",
					Lines: lines,
				}
				file.Close()
			}
			return nil
		})
		label := fmt.Sprintf("List files in `%s`", p.Path)
		if err != nil {
			return tool.Error(label, err)
		}
		result := map[string]any{
			"items":        items,
			"totalEntries": entries,
		}
		if entries > 1_000 {
			result["note"] = fmt.Sprintf("There were %d entries, but we could only include 1000.", entries)
		}
		return tool.Success(label, result)
	},
)
