package tool

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Params defines a struct type with various fields for testing tool functionality.
type Params struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email,omitempty"` // Optional field
	IsAdmin bool   `json:"isAdmin"`
}

// TestGenerateSchema checks that the JSON schema is generated correctly from the Params struct.
func TestGenerateSchema(t *testing.T) {
	typ := reflect.TypeOf(Params{})
	schema := generateSchema("TestFunction", "Test function description", typ)

	expectedSchema := map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "TestFunction",
			"description": "Test function description",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name":    map[string]any{"type": "string"},
					"age":     map[string]any{"type": "integer"},
					"email":   map[string]any{"type": "string"}, // Email is optional
					"isAdmin": map[string]any{"type": "boolean"},
				},
				"required": []string{"name", "age", "isAdmin"}, // Email is not required due to omitempty
			},
		},
	}
	if !reflect.DeepEqual(schema, expectedSchema) {
		t.Errorf("Expected schema %v, got %v", expectedSchema, schema)
	}
}

// TestToolRun_CorrectData verifies that the tool functions correctly with valid input data.
func TestToolRun_CorrectData(t *testing.T) {
	testFunc := func(r Runner, p Params) Result {
		return Success("Test", fmt.Sprintf("Profile: %s, %d, %s, Admin: %t", p.Name, p.Age, p.Email, p.IsAdmin))
	}
	tool := Func("Test Tool", "Test function for Params", "test_tool", testFunc)

	params := json.RawMessage(`{"name":"Bob", "age":30, "email":"bob@example.com", "isAdmin":false}`)
	result := tool.Run(&runner{}, params)
	if result.Error() != nil {
		t.Fatalf("Expected no error, got %v", result.Error())
	}
	expectedResult := "Profile: Bob, 30, bob@example.com, Admin: false"
	if result.String() != expectedResult {
		t.Errorf("Expected result %q, got %q", expectedResult, result.String())
	}
}

// TestToolRun_OptionalFieldAbsent verifies that the tool handles the absence of optional fields correctly.
func TestToolRun_OptionalFieldAbsent(t *testing.T) {
	testFunc := func(r Runner, p Params) Result {
		return Success("Test", fmt.Sprintf("Name: %s, Age: %d, Email: %s", p.Name, p.Age, p.Email))
	}
	tool := Func("Test Tool", "Test function for Params", "test_tool", testFunc)

	params := json.RawMessage(`{"name":"Alice", "age":28, "isAdmin":true}`)
	result := tool.Run(&runner{}, params)
	if result.Error() != nil {
		t.Fatalf("Expected no error, got %v", result.Error())
	}
	expectedResult := "Name: Alice, Age: 28, Email: "
	if result.String() != expectedResult {
		t.Errorf("Expected result %q, got %q", expectedResult, result.String())
	}
}

// TestToolRun_MissingRequiredField verifies that the tool correctly handles missing required fields.
func TestToolRun_MissingRequiredField(t *testing.T) {
	testFunc := func(r Runner, p Params) Result {
		return Success("Test", fmt.Sprintf("Received: Name: %s, Age: %d", p.Name, p.Age))
	}
	tool := Func("Test Tool", "Test function for Params", "test_tool", testFunc)

	params := json.RawMessage(`{"name":"John"}`) // Missing 'age' and 'isAdmin', which are required
	result := tool.Run(&runner{}, params)
	if err := result.Error(); err == nil || !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("Expected an error for missing required fields 'age' and 'isAdmin', but got: %v", err)
	}
}

// TestToolRun_InvalidDataType checks that the tool correctly identifies incorrect data types in input.
func TestToolRun_InvalidDataType(t *testing.T) {
	testFunc := func(r Runner, p Params) Result {
		return Success("Test", "")
	}
	tool := Func("Test Tool", "Test function for Params", "test_tool", testFunc)

	// Invalid data type for 'isAdmin', expecting a boolean but providing a string
	params := json.RawMessage(`{"name":"Alice", "age":28, "isAdmin":"yes"}`)
	result := tool.Run(&runner{}, params)
	if err := result.Error(); err == nil || !strings.Contains(err.Error(), "type mismatch") {
		t.Fatalf("Expected a type mismatch error for 'isAdmin', but got: %v", err)
	}
}

// TestToolRun_UnexpectedFields verifies that the tool ignores fields that are not defined in the schema.
func TestToolRun_UnexpectedFields(t *testing.T) {
	testFunc := func(r Runner, p Params) Result {
		return Success("Test", fmt.Sprintf("Name: %s", p.Name))
	}
	tool := Func("Test Tool", "Test function for Params", "test_tool", testFunc)

	// Including an unexpected 'location' field
	params := json.RawMessage(`{"name":"Alice", "age":28, "isAdmin":true, "location":"unknown"}`)
	result := tool.Run(&runner{}, params)
	if result.Error() != nil {
		t.Fatalf("Expected no error for unexpected field, got %v", result.Error())
	}
	expectedResult := "Name: Alice"
	if result.String() != expectedResult {
		t.Errorf("Expected result %q, got %q", expectedResult, result.String())
	}
}

type AdvancedParams struct {
	ID       int      `json:"id"`
	Features []string `json:"features"`
	Profile  struct {
		Username string `json:"username"`
		Active   bool   `json:"active"`
	} `json:"profile"`
}

// TestValidateJSONWithArrayAndObject tests validation of both array and nested object fields.
func TestValidateJSONWithArrayAndObject(t *testing.T) {
	testFunc := func(r Runner, p AdvancedParams) Result {
		return Success("Test", fmt.Sprintf("ID: %d, Features: %v, Profile: %s, Active: %t", p.ID, p.Features, p.Profile.Username, p.Profile.Active))
	}
	tool := Func("Advanced Tool", "Test function for Advanced Params", "advanced_tool", testFunc)

	// Valid input JSON that includes array and nested object
	validParams := json.RawMessage(`{"id":101, "features":["fast", "reliable", "secure"], "profile":{"username":"user01", "active":true}}`)
	result := tool.Run(&runner{}, validParams)
	if result.Error() != nil {
		t.Fatalf("Expected no error, got %v", result.Error())
	}
	expectedResult := "ID: 101, Features: [fast reliable secure], Profile: user01, Active: true"
	if result.String() != expectedResult {
		t.Errorf("Expected result %q, got %q", expectedResult, result.String())
	}

	// Invalid input JSON that has wrong data types in the array and nested object
	invalidParams := json.RawMessage(`{"id":101, "features":"fast", "profile":{"username":123, "active":"yes"}}`)
	result = tool.Run(&runner{}, invalidParams)
	if err := result.Error(); err == nil || !(strings.Contains(err.Error(), "type mismatch") || strings.Contains(err.Error(), "validation error")) {
		t.Fatalf("Expected a type mismatch or validation error, but got: %v", err)
	}
}

func TestToolFunctionErrorHandling(t *testing.T) {
	testFunc := func(r Runner, p AdvancedParams) Result {
		if p.ID == 0 {
			return Error("Test", fmt.Errorf("ID cannot be zero"))
		}
		return Success("Test", fmt.Sprintf("ID: %d, Features: %v, Profile: %s, Active: %t", p.ID, p.Features, p.Profile.Username, p.Profile.Active))
	}
	tool := Func("Error Handling Tool", "Test function for error handling in Params", "error_handling_tool", testFunc)

	// Input JSON with ID set to zero to trigger an error
	errorParams := json.RawMessage(`{"id":0, "features":["fast", "reliable"], "profile":{"username":"user01", "active":true}}`)
	result := tool.Run(&runner{}, errorParams)
	if err := result.Error(); err == nil || !strings.Contains(err.Error(), "ID cannot be zero") {
		t.Fatalf("Expected error 'ID cannot be zero', but got: %v", err)
	}

	// Valid input JSON to ensure no error is returned
	validParams := json.RawMessage(`{"id":101, "features":["fast", "reliable"], "profile":{"username":"user01", "active":true}}`)
	result = tool.Run(&runner{}, validParams)
	if result.Error() != nil {
		t.Fatalf("Expected no error, got %v", result.Error())
	}
	expectedResult := "ID: 101, Features: [fast reliable], Profile: user01, Active: true"
	if result.String() != expectedResult {
		t.Errorf("Expected result %q, got %q", expectedResult, result.String())
	}
}

func TestToolFunctionReport(t *testing.T) {
	reportCalled := false
	runner := &runner{
		report: func(status string) {
			reportCalled = true
			if status != "running" {
				t.Errorf("Expected status %q, got %q", "running", status)
			}
		},
	}

	testFunc := func(r Runner, p Params) Result {
		r.Report("running")
		return Success("Test", fmt.Sprintf("Profile: %s, %d, %s, Admin: %t", p.Name, p.Age, p.Email, p.IsAdmin))
	}
	tool := Func("Report Tool", "Test function for report functionality", "report_tool", testFunc)

	params := json.RawMessage(`{"name":"Alice", "age":28, "email":"alice@example.com", "isAdmin":true}`)
	result := tool.Run(runner, params)
	if result.Error() != nil {
		t.Fatalf("Expected no error, got %v", result.Error())
	}
	if !reportCalled {
		t.Errorf("Expected report function to be called, but it was not")
	}
}
