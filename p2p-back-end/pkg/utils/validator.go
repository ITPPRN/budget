package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct validates a struct and returns a human-readable error if validation fails
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			var errMsgs []string
			for _, e := range ve {
				errMsg := fmt.Sprintf("Field '%s' failed on the '%s' tag", e.Field(), e.Tag())
				if e.Param() != "" {
					errMsg += fmt.Sprintf(" (param: %s)", e.Param())
				}
				errMsgs = append(errMsgs, errMsg)
			}
			return fmt.Errorf("%s", strings.Join(errMsgs, ", "))
		}
		return err
	}
	return nil
}
