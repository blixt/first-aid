package firstaid

import (
	"encoding/json"
	"testing"

	"github.com/flitsinc/go-llms/content"
	"github.com/flitsinc/go-llms/tools"
	"github.com/stretchr/testify/require"
)

// Helper to extract JSON data from result content for testing
func extractJSONFromResult(t *testing.T, r tools.Result) json.RawMessage {
	t.Helper()
	require.NotNil(t, r.Content(), "Result content should not be nil")
	require.NotEmpty(t, r.Content(), "Result content should not be empty")
	jsonItem, ok := r.Content()[0].(*content.JSON)
	require.True(t, ok, "First content item should be JSON")
	return jsonItem.Data
}
