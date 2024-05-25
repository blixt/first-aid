package firstaid_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/blixt/first-aid/firstaid"
	"github.com/blixt/first-aid/tool"
)

func TestListFiles(t *testing.T) {
	// Setup a temporary directory with files and subdirectories
	tempDir := t.TempDir()
	os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("line1\nline2\n"), 0644)
	os.WriteFile(filepath.Join(tempDir, "subdir", "file2.txt"), []byte("line1\nline2\nline3\n"), 0644)

	result := firstaid.ListFiles.Run(tool.NopRunner, json.RawMessage(fmt.Sprintf(`{"path":%q}`, tempDir)))

	if result.Error() != nil {
		t.Fatalf("Expected no error, got %v", result.Error())
	}

	expected := map[string]firstaid.FileInfo{
		"file1.txt": {
			Type:  "file",
			Lines: 2,
		},
		"subdir": {
			Type:  "directory",
			Count: 1,
		},
		"subdir/file2.txt": {
			Type:  "file",
			Lines: 3,
		},
	}

	var actual map[string]firstaid.FileInfo
	if err := json.Unmarshal([]byte(result.String()), &actual); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected result:\n%v\nGot:\n%v", expected, actual)
	}
}
