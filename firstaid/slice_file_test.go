package firstaid

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/blixt/first-aid/tool"
)

func TestSliceFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		params   SliceFileParams
		expected func(string) string
	}{
		{
			name: "Read entire file",
			content: `Line 1
Line 2
Line 3
Line 4
Line 5`,
			params: SliceFileParams{
				Start: 0,
			},
			expected: func(path string) string {
				return `// Below are the sliced lines of "` + path + `" as an array. The number in each comment is the zero-based index of the string after it.
[
  /*0:*/"Line 1",
  /*1:*/"Line 2",
  /*2:*/"Line 3",
  /*3:*/"Line 4",
  /*4:*/"Line 5",
]`
			},
		},
		{
			name: "Read from start index",
			content: `Line 1
Line 2
Line 3
Line 4
Line 5`,
			params: SliceFileParams{
				Start: 2,
			},
			expected: func(path string) string {
				return `// Below are the sliced lines of "` + path + `" as an array. The number in each comment is the zero-based index of the string after it.
[
  /*2:*/"Line 3",
  /*3:*/"Line 4",
  /*4:*/"Line 5",
]`
			},
		},
		{
			name: "Read with end index",
			content: `Line 1
Line 2
Line 3
Line 4
Line 5`,
			params: SliceFileParams{
				Start: 1,
				End:   intptr(3),
			},
			expected: func(path string) string {
				return `// Below are the sliced lines of "` + path + `" as an array. The number in each comment is the zero-based index of the string after it.
[
  /*1:*/"Line 2",
  /*2:*/"Line 3",
  // There's 2 lines more after this.
]`
			},
		},
		{
			name: "Read with negative start index",
			content: `Line 1
Line 2
Line 3
Line 4
Line 5`,
			params: SliceFileParams{
				Start: -2,
			},
			expected: func(path string) string {
				return `// Below are the sliced lines of "` + path + `" as an array. The number in each comment is the zero-based index of the string after it.
[
  /*3:*/"Line 4",
  /*4:*/"Line 5",
]`
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile, err := os.CreateTemp(tmpDir, "testfile-*.txt")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			if err := tmpFile.Close(); err != nil {
				t.Fatalf("failed to close temp file: %v", err)
			}

			params := tt.params
			params.Path = tmpFile.Name()

			paramsJSON, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("failed to marshal params: %v", err)
			}
			result := SliceFile.Run(tool.NopRunner, paramsJSON)
			if result.Error() != nil {
				t.Fatalf("unexpected error: %v", result.Error())
			}
			expected := tt.expected(tmpFile.Name())
			assert.Equal(t, expected, result.String())
		})
	}
}

func intptr(i int) *int {
	return &i
}
