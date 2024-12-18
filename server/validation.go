package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	vdr = validator.New()
)

type MultipleErrorsErr struct {
	StatusCode int
	Message    string
	Errors     []string
}

// error method in validation error
func (e MultipleErrorsErr) Error() string {
	return fmt.Sprintf("Errors: %s", strings.Join(e.Errors, ", "))
}

func CreateMultipleErrorsErr(status int, message string, errors ...string) *MultipleErrorsErr {
	return &MultipleErrorsErr{
		StatusCode: status,
		Message:    message,
		Errors:     errors,
	}
}

func ParseAndValidate(c *fiber.Ctx, target any) error {
	if target == nil {
		return errors.New("validation target is nil")
	}
	if err := c.BodyParser(target); err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(err)
	}
	err := vdr.Struct(target)
	if err == nil {
		return nil
	}
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}
	ve := &MultipleErrorsErr{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    "Validation failed",
	}
	for _, err := range err.(validator.ValidationErrors) {
		ns := strings.TrimPrefix(err.Namespace(), "P.")
		errStr := fmt.Sprintf("%s failed on the '%s' tag", ns, err.Tag())
		ve.Errors = append(ve.Errors, errStr)
	}
	return ve
}

func ValidateVar(fieldName string, fieldValue interface{}, tag string) error {
	err := vdr.Var(fieldValue, tag)
	if err == nil {
		return nil
	}
	if _, ok := err.(*validator.InvalidValidationError); ok {
		return err
	}
	ve := &MultipleErrorsErr{}
	for _, err := range err.(validator.ValidationErrors) {
		errStr := fmt.Sprintf("%s failed on the '%s' tag", fieldName, err.Tag())
		ve.Errors = append(ve.Errors, errStr)
	}
	return ve
}

func RegisterValidation(tag string, fn validator.Func) error {
	return vdr.RegisterValidation(tag, fn)
}

func RegisterCustomTypeFunc(fn validator.CustomTypeFunc, types ...interface{}) {
	vdr.RegisterCustomTypeFunc(fn, types...)
}
