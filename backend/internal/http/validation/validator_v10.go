package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	// Use JSON tag names in error messages when available.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" {
			return strings.ToLower(fld.Name)
		}
		return name
	})

	return v
}()

// ValidateStruct validates a struct using validator/v10.
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// FormatValidationErrors returns a human-friendly string for validator errors.
func FormatValidationErrors(err error) string {
	if err == nil {
		return ""
	}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}

	messages := make([]string, 0, len(verrs))
	for _, ferr := range verrs {
		field := ferr.Field()
		tag := ferr.Tag()
		param := ferr.Param()

		switch tag {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", field))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s", field, param))
		case "max":
			messages = append(messages, fmt.Sprintf("%s must be at most %s", field, param))
		case "gte":
			messages = append(messages, fmt.Sprintf("%s must be greater than or equal to %s", field, param))
		case "lte":
			messages = append(messages, fmt.Sprintf("%s must be less than or equal to %s", field, param))
		case "oneof":
			messages = append(messages, fmt.Sprintf("%s must be one of [%s]", field, param))
		case "uuid4":
			messages = append(messages, fmt.Sprintf("%s must be a valid UUID", field))
		default:
			messages = append(messages, fmt.Sprintf("%s is invalid", field))
		}
	}

	return strings.Join(messages, "; ")
}
