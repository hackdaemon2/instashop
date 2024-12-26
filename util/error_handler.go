package util

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ExtractValidationErrorMessage(err error, structType any) map[int]string {
	if err != nil {
		var validationError validator.ValidationErrors
		if errors.As(err, &validationError) {
			messages := make(map[int]string, len(validationError))
			for index, value := range validationError {
				jsonName := getJSONFieldName(structType, value.StructField())
				messages[index] = messageForTag(jsonName, value.Tag())
			}
			return messages
		}
	}
	return nil
}

func messageForTag(field, tag string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s is not a valid email", field)
	case "min":
		return fmt.Sprintf("Invalid min length for %s", field)
	case "max":
		return fmt.Sprintf("Invalid max length for %s", field)
	case "numeric":
		return fmt.Sprintf("%s is not a valid numeric value", field)
	default:
		return fmt.Sprintf("Invalid value passed for %s", field)
	}
}

func getJSONFieldName(structType interface{}, structFieldName string) string {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	field, found := t.FieldByName(structFieldName)
	if !found {
		return structFieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || strings.Contains(jsonTag, ",") {
		jsonTag = strings.Split(jsonTag, ",")[0]
	}

	return jsonTag
}
