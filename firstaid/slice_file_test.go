package firstaid

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flitsinc/go-llms/tools"
)

func TestSliceFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		params   SliceFileParams
		expected map[string]any
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
			expected: map[string]any{
				"slicedLines": []map[string]any{
					{"0": "Line 1"},
					{"1": "Line 2"},
					{"2": "Line 3"},
					{"3": "Line 4"},
					{"4": "Line 5"},
				},
				"remainingLines": 0,
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
			expected: map[string]any{
				"slicedLines": []map[string]any{
					{"2": "Line 3"},
					{"3": "Line 4"},
					{"4": "Line 5"},
				},
				"remainingLines": 0,
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
			expected: map[string]any{
				"slicedLines": []map[string]any{
					{"1": "Line 2"},
					{"2": "Line 3"},
				},
				"remainingLines": 2,
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
			expected: map[string]any{
				"slicedLines": []map[string]any{
					{"3": "Line 4"},
					{"4": "Line 5"},
				},
				"remainingLines": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile, err := os.CreateTemp(tmpDir, "testfile-*.txt")
			require.NoError(t, err, "Failed to create temp file")
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err, "Failed to write to temp file")
			require.NoError(t, tmpFile.Close(), "Failed to close temp file")

			params := tt.params
			params.Path = tmpFile.Name()

			paramsJSON, err := json.Marshal(params)
			require.NoError(t, err, "Failed to marshal params")

			result := SliceFile.Run(tools.NopRunner, paramsJSON)
			require.NoError(t, result.Error(), "Unexpected error")

			// Create a copy of tt.expected and add filePath to it
			expected := make(map[string]any)
			for k, v := range tt.expected {
				expected[k] = v
			}
			expected["filePath"] = tmpFile.Name()

			expectedJSON, err := json.Marshal(expected)
			require.NoError(t, err, "Failed to marshal expected result")

			resultJSON := extractJSONFromResult(t, result)
			assert.JSONEq(t, string(expectedJSON), string(resultJSON), "Result mismatch")
		})
	}
}

func intptr(i int) *int {
	return &i
}
