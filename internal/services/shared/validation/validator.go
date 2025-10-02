package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)


type Validator struct {
	validate *validator.Validate
}


func New() *Validator {
	validate := validator.New()


	registerCustomValidations(validate)


	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: validate,
	}
}


func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}


func (v *Validator) ValidateVar(field interface{}, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}


func (v *Validator) formatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string

		for _, fieldError := range validationErrors {
			message := v.getErrorMessage(fieldError)
			messages = append(messages, message)
		}

		return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
	}

	return err
}


func (v *Validator) getErrorMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "hostname_rfc1123":
		return fmt.Sprintf("%s must be a valid hostname", field)
	case "e164":
		return fmt.Sprintf("%s must be a valid phone number in E.164 format", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "session_name":
		return fmt.Sprintf("%s contains invalid characters (only alphanumeric, dash and underscore allowed)", field)
	case "proxy_type":
		return fmt.Sprintf("%s must be either 'http' or 'socks5'", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}


func registerCustomValidations(validate *validator.Validate) {

	validate.RegisterValidation("session_name", validateSessionName)


	validate.RegisterValidation("proxy_type", validateProxyType)


	validate.RegisterValidation("e164", validateE164)
}


func validateSessionName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if name == "" {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}


func validateProxyType(fl validator.FieldLevel) bool {
	proxyType := fl.Field().String()
	return proxyType == "http" || proxyType == "socks5"
}


func validateE164(fl validator.FieldLevel) bool {
	phone := fl.Field().String()


	if !strings.HasPrefix(phone, "+") {
		return false
	}


	digits := phone[1:]
	if len(digits) < 7 || len(digits) > 15 {
		return false
	}

	for _, char := range digits {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}


func IsValidSessionName(name string) bool {
	validator := New()
	return validator.ValidateVar(name, "session_name") == nil
}