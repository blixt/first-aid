package firstaid

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blixt/first-aid/tool"
)

type SpliceFileParams struct {
	Path        string   `json:"path" description:"The path to the file to update."`
	Start       int      `json:"start" description:"The start index of the slice to delete and optionally replace."`
	DeleteCount int      `json:"deleteCount,omitempty" description:"The number of lines to delete from the slice."`
	InsertLines []string `json:"insertLines,omitempty" description:"The lines to insert at the start of the slice."`
}

var SpliceFile = tool.Func(
	"Update file",
	"Delete and/or replace a slice of the lines in the specified file, if we imagine the file as a zero-indexed array of lines.",
	"splice_file",
	func(r tool.Runner, p SpliceFileParams) tool.Result {
		p.Path = expandPath(p.Path)
		// Open or create the file if it doesn't exist.
		file, err := os.OpenFile(p.Path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return tool.Error(p.Path, fmt.Errorf("failed to open %q: %w", p.Path, err))
		}
		defer file.Close()

		var result strings.Builder

		scanner := bufio.NewScanner(file)
		var i int
		for {
			// Insert content at the requested index.
			if i == p.Start {
				for _, line := range p.InsertLines {
					result.WriteString(line + "\n")
				}
			}
			if scanner.Scan() {
				// Write one line from the source file as long as it's outside
				// the deleted range.
				if i < p.Start || i >= p.Start+p.DeleteCount {
					result.WriteString(scanner.Text() + "\n")
				}
			} else if err := scanner.Err(); err != nil {
				return tool.Error(p.Path, fmt.Errorf("failed to read file: %w", err))
			} else if i < p.Start {
				// If we get here it means we reached the end of the file but
				// the intended insert location is further ahead. We consider
				// this an erroneous usage of the API.
				return tool.Error(p.Path, fmt.Errorf("file has less than %d lines", p.Start+1))
			} else {
				// We finished all our work.
				break
			}
			i++
		}

		// Create a backup of the original file (if it wasn't empty).
		if i > 0 {
			backupPath := fmt.Sprintf("%s.%d.bak", p.Path, time.Now().Unix())
			if err := copyFile(p.Path, backupPath); err != nil {
				return tool.Error(p.Path, fmt.Errorf("failed to create backup: %w", err))
			}
		}

		if err := writeFileAtomically(p.Path, strings.NewReader(result.String())); err != nil {
			return tool.Error(p.Path, fmt.Errorf("failed to write updated content: %w", err))
		}

		var description string
		if p.DeleteCount > 0 && len(p.InsertLines) > 0 {
			if p.DeleteCount == len(p.InsertLines) {
				description = fmt.Sprintf("Replaced %s in %q", line(p.DeleteCount), p.Path)
			} else {
				description = fmt.Sprintf("Replaced %s with %s in %q", line(p.DeleteCount), line(len(p.InsertLines)), p.Path)
			}
		} else if p.DeleteCount > 0 {
			description = fmt.Sprintf("Deleted %s from %q", line(p.DeleteCount), p.Path)
		} else if len(p.InsertLines) > 0 {
			description = fmt.Sprintf("Added %s to %q", line(len(p.InsertLines)), p.Path)
		}
		return tool.Success(description, "File updated.")
	})

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	return writeFileAtomically(dst, srcFile)
}

func writeFileAtomically(dst string, content io.Reader) error {
	tmpDstFile, err := os.CreateTemp(filepath.Dir(dst), "tmp-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpDstFile.Name())

	if _, err = io.Copy(tmpDstFile, content); err != nil {
		return err
	}

	if err = tmpDstFile.Sync(); err != nil {
		return err
	}
	if err = tmpDstFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpDstFile.Name(), dst)
}
