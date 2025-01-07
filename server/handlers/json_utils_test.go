package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAllowedJSONFields(t *testing.T) {
	// Test Case 1: Flat Struct
	t.Run("Flat Struct", func(t *testing.T) {
		type FlatStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		expected := []string{"name", "age"}
		result := allowedJSONFields(FlatStruct{})

		if len(result) != len(expected) {
			t.Fatalf("expected %d fields, got %d. Fields: %v", len(expected), len(result), result)
		}
	})

	// Test Case 2: Nested Struct
	t.Run("Nested Struct", func(t *testing.T) {
		type NestedStruct struct {
			InnerField string `json:"innerField"`
		}
		type ParentStruct struct {
			OuterField string       `json:"outerField"`
			Nested     NestedStruct `json:"nested"`
		}

		expected := []string{"outerField", "nested.innerField"}
		result := allowedJSONFields(ParentStruct{})

		if len(result) != len(expected) {
			t.Fatalf("expected %d fields, got %d. Fields: %v", len(expected), len(result), result)
		}
	})

	// Test Case 3: Struct with Slices
	t.Run("Struct with Slices", func(t *testing.T) {
		type SliceStruct struct {
			Items []struct {
				Item string `json:"item"`
			} `json:"items"`
		}

		expected := []string{"items"}
		result := allowedJSONFields(SliceStruct{})

		if len(result) != len(expected) {
			t.Fatalf("expected %d fields, got %d. Fields: %v", len(expected), len(result), result)
		}
	})

	// Test Case 4: Deeply Nested Struct
	t.Run("Deeply Nested Struct", func(t *testing.T) {
		type DeepNested struct {
			Deep string `json:"deep"`
		}
		type NestedStruct struct {
			Nested DeepNested `json:"nested"`
		}
		type ComplexStruct struct {
			Field1 string       `json:"field1"`
			Struct NestedStruct `json:"struct"`
		}

		expected := []string{"field1", "struct.nested.deep"}
		result := allowedJSONFields(ComplexStruct{})

		if len(result) != len(expected) {
			t.Fatalf("expected %d fields, got %d. Fields: %v", len(expected), len(result), result)
		}
	})

	// Test Case 5: Ignored and Untagged Fields
	t.Run("Ignored and Untagged Fields", func(t *testing.T) {
		type MixedStruct struct {
			Valid   string `json:"valid"`
			Ignored string `json:"-"`
			NoTag   string
		}

		expected := []string{"valid"}
		result := allowedJSONFields(MixedStruct{})

		if len(result) != len(expected) {
			t.Fatalf("expected %d fields, got %d. Fields: %v", len(expected), len(result), result)
		}
	})
}

func TestAllowedJSONFields_NonStructInput(t *testing.T) {
	result := allowedJSONFields("not a struct")
	if len(result) != 0 {
		t.Errorf("expected empty slice for non-struct input, got %v", result)
	}
}

func TestMissingJSONFields(t *testing.T) {
	// Test Case 1: Flat Struct with Missing Fields
	t.Run("Flat Struct with Missing Fields", func(t *testing.T) {
		type FlatStruct struct {
			Name  string `json:"name" validate:"required"`
			Email string `json:"email"`
			Age   int    `json:"age" validate:"required"`
		}

		testObj := FlatStruct{
			Name: "",
			Age:  0,
		}

		expected := []string{"name", "age"}
		missing := missingJSONFields(testObj)

		if len(missing) != len(expected) {
			t.Fatalf("expected %d missing fields, got %d. Fields: %v", len(expected), len(missing), missing)
		}
	})

	// Test Case 2: Nested Struct with Missing Fields
	t.Run("Nested Struct with Missing Fields", func(t *testing.T) {
		type NestedStruct struct {
			InnerField string `json:"innerField" validate:"required"`
		}
		type ParentStruct struct {
			Name   string       `json:"name" validate:"required"`
			Nested NestedStruct `json:"nested"`
		}

		testObj := ParentStruct{
			Name: "",
			Nested: NestedStruct{
				InnerField: "",
			},
		}

		expected := []string{"name", "nested.innerField"}
		missing := missingJSONFields(testObj)

		if len(missing) != len(expected) {
			t.Fatalf("expected %d missing fields, got %d. Fields: %v", len(expected), len(missing), missing)
		}
	})

	// Test Case 3: Struct with Slices
	t.Run("Struct with Slice of Structs", func(t *testing.T) {
		type SliceStruct struct {
			Items []struct {
				Item string `json:"item" validate:"required"`
			} `json:"items"`
		}

		testObj := SliceStruct{
			Items: []struct {
				Item string `json:"item" validate:"required"`
			}{},
		}

		// We only expect the top-level key
		expected := []string{"items"}
		missing := missingJSONFields(testObj)

		if len(missing) != len(expected) {
			t.Fatalf("expected %d missing fields, got %d. Fields: %v", len(expected), len(missing), missing)
		}
	})

	// Test Case 4: Deeply Nested Struct
	t.Run("Deeply Nested Struct", func(t *testing.T) {
		type DeepNested struct {
			Deep string `json:"deep" validate:"required"`
		}
		type NestedStruct struct {
			Nested DeepNested `json:"nested"`
		}
		type ComplexStruct struct {
			Struct NestedStruct `json:"struct"`
		}

		testObj := ComplexStruct{
			Struct: NestedStruct{
				Nested: DeepNested{
					Deep: "",
				},
			},
		}

		expected := []string{"struct.nested.deep"}
		missing := missingJSONFields(testObj)

		if len(missing) != len(expected) {
			t.Fatalf("expected %d missing fields, got %d. Fields: %v", len(expected), len(missing), missing)
		}
	})

	// Test Case 5: Ignored and Untagged Fields
	t.Run("Ignored and Untagged Fields", func(t *testing.T) {
		type MixedStruct struct {
			Valid   string `json:"valid" validate:"required"`
			Ignored string `json:"-"`
			NoTag   string
		}

		testObj := MixedStruct{
			Valid: "",
		}

		expected := []string{"valid"}
		missing := missingJSONFields(testObj)

		if len(missing) != len(expected) {
			t.Fatalf("expected %d missing fields, got %d. Fields: %v", len(expected), len(missing), missing)
		}
	})
}

func TestMissingJSONFields_NonStructInput(t *testing.T) {
	result := missingJSONFields(123)
	if len(result) != 0 {
		t.Errorf("expected empty slice for non-struct input, got %v", result)
	}
}

func TestStrictDecodeJSON(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name" validate:"required"`
		Age   int    `json:"age"`
		Email string `json:"email" validate:"required"`
	}

	testCases := []struct {
		name          string // Test case description
		jsonBody      string // JSON payload for the request
		expectError   bool   // Whether we expect an error
		expectedError string // Expected error message
	}{
		{
			name:        "Valid JSON",
			jsonBody:    `{"name": "John", "age": 30, "email": "john@example.com"}`,
			expectError: false,
		},
		{
			name:          "Missing required field (email)",
			jsonBody:      `{"name": "John", "age": 30}`,
			expectError:   true,
			expectedError: "missing required fields: [email]",
		},
		{
			name:          "Unknown field",
			jsonBody:      `{"name": "John", "age": 30, "email": "john@example.com", "extraField": "unexpected"}`,
			expectError:   true,
			expectedError: "unknown fields in the request. Valid fields are:",
		},
		{
			name:          "Malformed JSON",
			jsonBody:      `{"name": "John", "age": 30, "email": "john@example.com"`,
			expectError:   true,
			expectedError: "failed to parse JSON: unexpected EOF",
		},
		{
			name:          "Trailing extra data",
			jsonBody:      `{"name": "John", "age": 30, "email": "john@example.com"} {"extra": "data"}`,
			expectError:   true,
			expectedError: "invalid request body: extra data detected after valid JSON payload",
		},
		{
			name:          "Type mismatch (int instead of string)",
			jsonBody:      `{"name": 123, "age": 30, "email": "john@example.com"}`,
			expectError:   true,
			expectedError: "invalid type for field 'name'. Expected string",
		},
		{
			name:          "Type mismatch (string instead of int)",
			jsonBody:      `{"name": "John", "age": "thirty", "email": "john@example.com"}`,
			expectError:   true,
			expectedError: "invalid type for field 'age'. Expected int",
		},
		{
			name:          "Completely invalid JSON",
			jsonBody:      `invalid-json`,
			expectError:   true,
			expectedError: "invalid JSON syntax at byte offset",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(tc.jsonBody)))
			var testObj TestStruct

			err := StrictDecodeJSON(req, &testObj)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}

				if tc.expectedError != "" && !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error to contain '%s', but got '%s'", tc.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				// Validate proper decoding in the valid case
				if testObj.Name != "John" || testObj.Age != 30 || testObj.Email != "john@example.com" {
					t.Errorf("decoded struct does not match expected values: %+v", testObj)
				}
			}
		})
	}
}

func TestWriteJSONError(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		errorMessage   string
		fieldErrors    map[string]string
		expectedBody   ErrorResponse
		expectedHeader string
	}{
		{
			name:         "Valid error response with field errors",
			statusCode:   http.StatusBadRequest,
			errorMessage: "Invalid data",
			fieldErrors: map[string]string{
				"field1": "This field is required",
			},
			expectedBody: ErrorResponse{
				StatusCode:   http.StatusBadRequest,
				ErrorMessage: "Invalid data",
				FieldErrors: map[string]string{
					"field1": "This field is required",
				},
			},
			expectedHeader: "application/json",
		},
		{
			name:         "Error response without field errors",
			statusCode:   http.StatusInternalServerError,
			errorMessage: "Internal server error",
			fieldErrors:  nil,
			expectedBody: ErrorResponse{
				StatusCode:   http.StatusInternalServerError,
				ErrorMessage: "Internal server error",
				FieldErrors:  nil,
			},
			expectedHeader: "application/json",
		},
		{
			name:         "Error response with empty field errors map",
			statusCode:   http.StatusForbidden,
			errorMessage: "Forbidden access",
			fieldErrors:  map[string]string{},
			expectedBody: ErrorResponse{
				StatusCode:   http.StatusForbidden,
				ErrorMessage: "Forbidden access",
				FieldErrors:  map[string]string{},
			},
			expectedHeader: "application/json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			// Call the function
			WriteJSONError(rr, tc.statusCode, tc.errorMessage, tc.fieldErrors)

			// Assert HTTP status code
			if rr.Code != tc.statusCode {
				t.Fatalf("expected status %d, got %d", tc.statusCode, rr.Code)
			}

			// Assert Content-Type header
			if contentType := rr.Header().Get("Content-Type"); contentType != tc.expectedHeader {
				t.Errorf("expected header 'Content-Type' to be '%s', got '%s'", tc.expectedHeader, contentType)
			}

			// Assert response body
			var resp ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response JSON: %v", err)
			}

			if resp.StatusCode != tc.expectedBody.StatusCode {
				t.Errorf("expected StatusCode '%d', got '%d'", tc.expectedBody.StatusCode, resp.StatusCode)
			}

			if resp.ErrorMessage != tc.expectedBody.ErrorMessage {
				t.Errorf("expected ErrorMessage '%s', got '%s'", tc.expectedBody.ErrorMessage, resp.ErrorMessage)
			}

			if len(resp.FieldErrors) != len(tc.expectedBody.FieldErrors) {
				t.Errorf("expected %d field errors, got %d", len(tc.expectedBody.FieldErrors), len(resp.FieldErrors))
			}

			for key, expected := range tc.expectedBody.FieldErrors {
				if actual, exists := resp.FieldErrors[key]; !exists || actual != expected {
					t.Errorf("expected field error for '%s' to be '%s', got '%s'", key, expected, actual)
				}
			}
		})
	}
}
