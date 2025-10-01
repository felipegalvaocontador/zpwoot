package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wrapper para go-playground/validator com funcionalidades customizadas
type Validator struct {
	validate *validator.Validate
}

// New cria nova instância do validador
func New() *Validator {
	validate := validator.New()

	// Registrar validações customizadas
	registerCustomValidations(validate)

	// Configurar nomes de campos usando tags JSON
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

// ValidateStruct valida uma struct
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.validate.Struct(s); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// ValidateVar valida uma variável individual
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	if err := v.validate.Var(field, tag); err != nil {
		return v.formatValidationError(err)
	}
	return nil
}

// formatValidationError formata erros de validação para serem mais legíveis
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

// getErrorMessage retorna mensagem de erro personalizada para cada tipo de validação
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

// registerCustomValidations registra validações customizadas
func registerCustomValidations(validate *validator.Validate) {
	// Validação para nome de sessão
	validate.RegisterValidation("session_name", validateSessionName)

	// Validação para tipo de proxy
	validate.RegisterValidation("proxy_type", validateProxyType)

	// Validação para formato E.164 de telefone
	validate.RegisterValidation("e164", validateE164)
}

// validateSessionName valida nome de sessão (apenas alfanuméricos, hífen e underscore)
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

// validateProxyType valida tipo de proxy
func validateProxyType(fl validator.FieldLevel) bool {
	proxyType := fl.Field().String()
	return proxyType == "http" || proxyType == "socks5"
}

// validateE164 valida formato E.164 para números de telefone
func validateE164(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Deve começar com +
	if !strings.HasPrefix(phone, "+") {
		return false
	}

	// Remover o + e verificar se o resto são apenas dígitos
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

// IsValidSessionName verifica se nome de sessão é válido
func IsValidSessionName(name string) bool {
	validator := New()
	return validator.ValidateVar(name, "session_name") == nil
}