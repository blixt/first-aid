package firstaid

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/flitsinc/go-llms/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFiles(t *testing.T) {
	// Setup a temporary directory with files and subdirectories
	tempDir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(tempDir, "subdir"), 0755))
	require.NoError(t, os.Mkdir(filepath.Join(tempDir, "subdir", "subsubdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("line1\nline2\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "subdir", "file2.txt"), []byte("line1\nline2\nline3\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "subdir", "subsubdir", "file3.txt"), []byte("line1\nline2\nline3\nline4\n"), 0644))

	result := ListFiles.Run(tools.NopRunner, json.RawMessage(fmt.Sprintf(`{"path":%q,"depth":2}`, tempDir)))

	require.NoError(t, result.Error())

	expectedItems := map[string]FileInfo{
		"file1.txt": {
			Type:  "file",
			Lines: 2,
		},
		"subdir": {
			Type:  "directory",
			Count: 2,
		},
		"subdir/file2.txt": {
			Type:  "file",
			Lines: 3,
		},
		"subdir/subsubdir": {
			Type:            "directory",
			Count:           1,
			ContentsSkipped: true,
		},
	}

	expectedResult := map[string]any{
		"items":        expectedItems,
		"totalEntries": 4,
	}

	var actual map[string]any
	resultJSON := extractJSONFromResult(t, result)
	err := json.Unmarshal(resultJSON, &actual)
	require.NoError(t, err)

	expectedJSON, err := json.Marshal(expectedResult)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)

	assert.JSONEq(t, string(expectedJSON), string(actualJSON))
}
