package services

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func LoadValidator() {
	v = validator.New()
}

func ValidateStruct(i interface{}) (errs error) {
	return v.Struct(i)
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
