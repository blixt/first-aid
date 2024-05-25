package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// generateSchema initializes and returns the main structure of a function's JSON Schema
func generateSchema(name, description string, typ reflect.Type) map[string]any {
	parameters := generateObjectSchema(typ)
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        name,
			"description": description,
			"parameters":  parameters,
		},
	}
}

// fieldTypeToJSONSchema maps Go data types to corresponding JSON Schema properties consistently
func fieldTypeToJSONSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": fieldTypeToJSONSchema(t.Elem())}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": fieldTypeToJSONSchema(t.Elem())}
	case reflect.Struct:
		return generateObjectSchema(t)
	case reflect.Ptr:
		return fieldTypeToJSONSchema(t.Elem())
	default:
		return map[string]any{"type": "unknown"}
	}
}

// generateObjectSchema constructs a JSON Schema for structs
func generateObjectSchema(typ reflect.Type) map[string]any {
	properties := make(map[string]any)
	required := []string{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" { // Skip unexported fields
			continue
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" { // Field is explicitly ignored
			continue
		}
		parts := strings.Split(jsonTag, ",")
		fieldName := field.Name
		if parts[0] != "" {
			fieldName = parts[0]
		}

		fieldSchema := fieldTypeToJSONSchema(field.Type)
		if description := field.Tag.Get("description"); description != "" {
			fieldSchema["description"] = description
		}
		properties[fieldName] = fieldSchema
		if len(parts) == 1 || (len(parts) > 1 && parts[1] != "omitempty") {
			required = append(required, fieldName)
		}
	}
	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// validateJSON checks if jsonData conforms to the structure defined in the schema from generateSchema
func validateJSON(schema map[string]any, jsonData json.RawMessage) error {
	function, ok := schema["function"].(map[string]any)
	if !ok {
		return errors.New("schema error: expected 'function' key in schema")
	}

	parameters, ok := function["parameters"].(map[string]any)
	if !ok {
		return errors.New("schema error: expected 'parameters' key in function schema")
	}

	return validateParameters(parameters, jsonData)
}

// validateParameters validates JSON data against the provided parameters schema
func validateParameters(schema map[string]any, jsonData json.RawMessage) error {
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return errors.New("schema error: properties must be a map")
	}
	requiredFields, _ := schema["required"].([]string)

	var dataMap map[string]any
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return errors.New("invalid JSON format")
	}

	for key, val := range dataMap {
		fieldSchema, found := properties[key]
		if !found {
			continue // Ignoring extra fields
		}
		if err := validateField(fieldSchema, val); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}

	for _, field := range requiredFields {
		if _, exists := dataMap[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	return nil
}

// validateField checks a single field against its schema
func validateField(fieldSchema any, data any) error {
	spec, ok := fieldSchema.(map[string]any)
	if !ok {
		return errors.New("schema error: field schema must be a map")
	}

	dataType, ok := spec["type"].(string)
	if !ok {
		return errors.New("schema error: missing type specification")
	}

	switch dataType {
	case "integer":
		num, ok := data.(float64)
		if !ok || num != float64(int(num)) {
			return fmt.Errorf("type mismatch: expected integer, got %T", data)
		}
	case "number":
		if _, ok := data.(float64); !ok {
			return fmt.Errorf("type mismatch: expected number, got %T", data)
		}
	case "string":
		if _, ok := data.(string); !ok {
			return fmt.Errorf("type mismatch: expected string, got %T", data)
		}
	case "boolean":
		if _, ok := data.(bool); !ok {
			return fmt.Errorf("type mismatch: expected boolean, got %T", data)
		}
	case "array":
		items, ok := data.([]any)
		if !ok {
			return fmt.Errorf("type mismatch: expected array, got %T", data)
		}
		itemSchema, ok := spec["items"].(map[string]any)
		if !ok {
			return errors.New("schema error: missing item schema for array")
		}
		for _, item := range items {
			if err := validateField(itemSchema, item); err != nil {
				return err
			}
		}
	case "object":
		properties, ok := data.(map[string]any)
		if !ok {
			return fmt.Errorf("type mismatch: expected object, got %T", data)
		}
		jsonData, err := json.Marshal(properties)
		if err != nil {
			return errors.New("failed to marshal object data for validation")
		}
		return validateParameters(spec, json.RawMessage(jsonData))
	default:
		return fmt.Errorf("unsupported type: %s", dataType)
	}
	return nil
}
