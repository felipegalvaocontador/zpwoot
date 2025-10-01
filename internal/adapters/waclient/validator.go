package waclient

import (
	"fmt"
	"regexp"
	"strings"

	"go.mau.fi/whatsmeow/types"

	"zpwoot/internal/core/session"
)

// Validator valida dados relacionados ao WhatsApp baseado no legacy
type Validator struct{}

// NewValidator cria novo validador
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateSessionName valida nome da sessão
func (v *Validator) ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("session name too long (max 100 characters)")
	}

	// Verificar caracteres válidos (alfanuméricos, hífen, underscore)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("session name contains invalid characters (only alphanumeric, hyphen, and underscore allowed)")
	}

	return nil
}

// ValidatePhoneNumber valida número de telefone
func (v *Validator) ValidatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return fmt.Errorf("phone number cannot be empty")
	}

	// Remover caracteres não numéricos para validação
	cleanNumber := v.CleanPhoneNumber(phoneNumber)

	// Verificar se tem pelo menos 10 dígitos
	if len(cleanNumber) < 10 {
		return fmt.Errorf("phone number too short (minimum 10 digits)")
	}

	// Verificar se tem no máximo 15 dígitos (padrão internacional)
	if len(cleanNumber) > 15 {
		return fmt.Errorf("phone number too long (maximum 15 digits)")
	}

	// Verificar se contém apenas dígitos
	for _, char := range cleanNumber {
		if char < '0' || char > '9' {
			return fmt.Errorf("phone number contains invalid characters")
		}
	}

	return nil
}

// ValidateJID valida JID do WhatsApp
func (v *Validator) ValidateJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("JID cannot be empty")
	}

	// Tentar fazer parse do JID
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return fmt.Errorf("invalid JID format: %w", err)
	}

	// Verificar se é um JID válido do WhatsApp
	if parsedJID.Server != types.DefaultUserServer &&
	   parsedJID.Server != types.GroupServer &&
	   parsedJID.Server != types.BroadcastServer {
		return fmt.Errorf("invalid WhatsApp JID server: %s", parsedJID.Server)
	}

	return nil
}

// ValidateProxyConfig valida configuração de proxy
func (v *Validator) ValidateProxyConfig(config *session.ProxyConfig) error {
	if config == nil {
		return nil // Proxy é opcional
	}

	if config.Host == "" {
		return fmt.Errorf("proxy host cannot be empty")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("proxy port must be between 1 and 65535")
	}

	// Validar tipo de proxy
	validTypes := []string{"http", "https", "socks5"}
	validType := false
	for _, validT := range validTypes {
		if config.Type == validT {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid proxy type: %s (allowed: http, https, socks5)", config.Type)
	}

	// Se tem username, deve ter password também
	if config.Username != "" && config.Password == "" {
		return fmt.Errorf("proxy password is required when username is provided")
	}

	return nil
}

// CleanPhoneNumber remove caracteres não numéricos do número de telefone
func (v *Validator) CleanPhoneNumber(phoneNumber string) string {
	cleaned := strings.ReplaceAll(phoneNumber, "+", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	return cleaned
}

// IsValidWhatsAppNumber verifica se é um número válido do WhatsApp
func (v *Validator) IsValidWhatsAppNumber(phoneNumber string) bool {
	err := v.ValidatePhoneNumber(phoneNumber)
	return err == nil
}

// IsValidJID verifica se é um JID válido
func (v *Validator) IsValidJID(jid string) bool {
	err := v.ValidateJID(jid)
	return err == nil
}

// IsGroupJID verifica se JID é de grupo
func (v *Validator) IsGroupJID(jid string) bool {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return false
	}
	return parsedJID.Server == types.GroupServer
}

// IsBroadcastJID verifica se JID é de broadcast
func (v *Validator) IsBroadcastJID(jid string) bool {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return false
	}
	return parsedJID.Server == types.BroadcastServer
}

// IsUserJID verifica se JID é de usuário individual
func (v *Validator) IsUserJID(jid string) bool {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return false
	}
	return parsedJID.Server == types.DefaultUserServer
}

// ValidateMessageContent valida conteúdo de mensagem
func (v *Validator) ValidateMessageContent(content string, messageType string) error {
	if content == "" && messageType == "text" {
		return fmt.Errorf("text message content cannot be empty")
	}

	// Verificar tamanho máximo (WhatsApp tem limite de ~65KB)
	if len(content) > 65000 {
		return fmt.Errorf("message content too long (max 65000 characters)")
	}

	return nil
}

// ValidateMediaURL valida URL de mídia
func (v *Validator) ValidateMediaURL(url string) error {
	if url == "" {
		return fmt.Errorf("media URL cannot be empty")
	}

	// Verificar se começa com http ou https
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("media URL must start with http:// or https://")
	}

	return nil
}

// ValidateLocation valida coordenadas de localização
func (v *Validator) ValidateLocation(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	return nil
}
