package services

import (
	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func LoadValidator() {
	v = validator.New()
}

func ValidateStruct(i interface{}) (errs error) {
	return v.Struct(i)
}
