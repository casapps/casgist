package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/go-playground/validator/v10"
)

// EchoValidator wraps go-playground/validator for Echo
type EchoValidator struct {
	validator *validator.Validate
}

// NewEchoValidator creates a new Echo validator
func NewEchoValidator() *EchoValidator {
	return &EchoValidator{
		validator: validator.New(),
	}
}

// Validate implements echo.Validator interface
func (ev *EchoValidator) Validate(i interface{}) error {
	if err := ev.validator.Struct(i); err != nil {
		// Return Echo HTTP error with validation details
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}