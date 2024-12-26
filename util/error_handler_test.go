package util

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=5"`
	Age      int    `json:"age" validate:"required,numeric"`
}

func TestExtractValidationErrorMessage(t *testing.T) {
	validate := validator.New()

	testObj := TestStruct{
		Email:    "invalid-email",
		Username: "usr",
		Age:      -1,
	}

	err := validate.Struct(testObj)
	if err == nil {
		t.Fatal("Expected validation error, but got nil")
	}

	messages := ExtractValidationErrorMessage(err, TestStruct{})

	assert.NotNil(t, messages)

	assert.Contains(t, messages, 0)
	assert.Contains(t, messages[0], "is not a valid email")
}

func TestMessageForTag(t *testing.T) {
	tests := []struct {
		field    string
		tag      string
		expected string
	}{
		{"email", "required", "email is required"},
		{"username", "email", "username is not a valid email"},
		{"password", "min", "Invalid min length for password"},
		{"age", "max", "Invalid max length for age"},
		{"amount", "numeric", "amount is not a valid numeric value"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			actual := messageForTag(tt.field, tt.tag)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetJSONFieldName(t *testing.T) {
	// Test cases for getJSONFieldName function
	tests := []struct {
		structFieldName string
		expected        string
	}{
		{"Email", "email"},
		{"Username", "username"},
		{"Age", "age"},
	}

	for _, tt := range tests {
		t.Run(tt.structFieldName, func(t *testing.T) {
			actual := getJSONFieldName(TestStruct{}, tt.structFieldName)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestExtractValidationErrorMessageNoValidationErrors(t *testing.T) {
	validTestObj := TestStruct{
		Email:    "valid@example.com",
		Username: "validuser",
		Age:      25,
	}

	validate := validator.New()
	err := validate.Struct(validTestObj)
	assert.Nil(t, err, "Expected no validation errors")

	messages := ExtractValidationErrorMessage(err, TestStruct{})
	assert.Nil(t, messages, "Expected no error messages")
}

func TestExtractValidationErrorMessageEmptyStruct(t *testing.T) {
	messages := ExtractValidationErrorMessage(nil, TestStruct{})
	assert.Nil(t, messages, "Expected no error messages when error is nil")
}
