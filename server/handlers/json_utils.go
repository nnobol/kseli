package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

func allowedJSONFields(out interface{}) []string {
	val := reflect.ValueOf(out)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return []string{} // Only structs are processed
	}

	t := val.Type()
	var fields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// Skip ignored or unnamed JSON fields
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue := val.Field(i)
		fieldType := field.Type

		// Handle nested structs
		if fieldType.Kind() == reflect.Struct {
			nestedFields := allowedJSONFields(fieldValue.Interface())
			for _, nestedField := range nestedFields {
				fields = append(fields, tag+"."+nestedField)
			}
		} else if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
			// Treat slices as a single top-level key
			fields = append(fields, tag)
		} else {
			// Add the top-level field
			fields = append(fields, tag)
		}
	}
	return fields
}

func missingJSONFields(out interface{}) []string {
	val := reflect.ValueOf(out)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return []string{} // Only structs are processed
	}

	t := val.Type()
	var fields []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		validateTag := field.Tag.Get("validate")
		jsonTag := field.Tag.Get("json")

		// Skip ignored or unnamed JSON fields
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldValue := val.Field(i)
		fieldType := field.Type

		// Check if the field is required and zero-valued
		if validateTag == "required" && fieldValue.IsZero() {
			fields = append(fields, jsonTag)
		}

		// Handle nested structs
		if fieldType.Kind() == reflect.Struct {
			// Recursively check nested fields
			nestedFields := missingJSONFields(fieldValue.Interface())
			for _, nestedField := range nestedFields {
				fields = append(fields, jsonTag+"."+nestedField)
			}
		} else if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
			// Treat slices as a single top-level key
			fields = append(fields, jsonTag)
		}
	}

	return fields
}

func StrictDecodeJSON(r *http.Request, out interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(out); err != nil {
		// Handle specific unmarshaling errors
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("invalid JSON syntax at byte offset %d", syntaxError.Offset)

		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("invalid type for field '%s'. Expected %s", unmarshalTypeError.Field, unmarshalTypeError.Type)

		case strings.Contains(err.Error(), "unknown field"):
			allowedFields := allowedJSONFields(out)
			return fmt.Errorf("unknown fields in the request. Valid fields are: %v", allowedFields)

		default:
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	}

	missingFields := missingJSONFields(out)
	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %v", missingFields)
	}

	if decoder.More() {
		return fmt.Errorf("invalid request body: extra data detected after valid JSON payload")
	}

	return nil
}

func WriteJSONError(w http.ResponseWriter, statusCode int, message string, fieldErrors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		StatusCode:   statusCode,
		ErrorMessage: message,
		FieldErrors:  fieldErrors,
	}

	json.NewEncoder(w).Encode(errorResponse)
}
