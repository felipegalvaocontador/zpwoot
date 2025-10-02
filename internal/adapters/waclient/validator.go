package waclient

import (
	"fmt"
	"regexp"
	"strings"

	"go.mau.fi/whatsmeow/types"

	"zpwoot/internal/core/session"
)


type Validator struct{}


func NewValidator() *Validator {
	return &Validator{}
}


func (v *Validator) ValidateSessionName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("session name too long (max 100 characters)")
	}


	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("session name contains invalid characters (only alphanumeric, hyphen, and underscore allowed)")
	}

	return nil
}


func (v *Validator) ValidatePhoneNumber(phoneNumber string) error {
	if phoneNumber == "" {
		return fmt.Errorf("phone number cannot be empty")
	}


	cleanNumber := v.CleanPhoneNumber(phoneNumber)


	if len(cleanNumber) < 10 {
		return fmt.Errorf("phone number too short (minimum 10 digits)")
	}


	if len(cleanNumber) > 15 {
		return fmt.Errorf("phone number too long (maximum 15 digits)")
	}


	for _, char := range cleanNumber {
		if char < '0' || char > '9' {
			return fmt.Errorf("phone number contains invalid characters")
		}
	}

	return nil
}


func (v *Validator) ValidateJID(jid string) error {
	if jid == "" {
		return fmt.Errorf("JID cannot be empty")
	}


	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return fmt.Errorf("invalid JID format: %w", err)
	}


	if parsedJID.Server != types.DefaultUserServer &&
	   parsedJID.Server != types.GroupServer &&
	   parsedJID.Server != types.BroadcastServer {
		return fmt.Errorf("invalid WhatsApp JID server: %s", parsedJID.Server)
	}

	return nil
}


func (v *Validator) ValidateProxyConfig(config *session.ProxyConfig) error {
	if config == nil {
		return nil
	}

	if config.Host == "" {
		return fmt.Errorf("proxy host cannot be empty")
	}

	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("proxy port must be between 1 and 65535")
	}


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


	if config.Username != "" && config.Password == "" {
		return fmt.Errorf("proxy password is required when username is provided")
	}

	return nil
}


func (v *Validator) CleanPhoneNumber(phoneNumber string) string {
	cleaned := strings.ReplaceAll(phoneNumber, "+", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")
	return cleaned
}


func (v *Validator) IsValidWhatsAppNumber(phoneNumber string) bool {
	err := v.ValidatePhoneNumber(phoneNumber)
	return err == nil
}


func (v *Validator) IsValidJID(jid string) bool {
	err := v.ValidateJID(jid)
	return err == nil
}


func (v *Validator) IsGroupJID(jid string) bool {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return false
	}
	return parsedJID.Server == types.GroupServer
}


func (v *Validator) IsBroadcastJID(jid string) bool {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return false
	}
	return parsedJID.Server == types.BroadcastServer
}


func (v *Validator) IsUserJID(jid string) bool {
	parsedJID, err := types.ParseJID(jid)
	if err != nil {
		return false
	}
	return parsedJID.Server == types.DefaultUserServer
}


func (v *Validator) ValidateMessageContent(content string, messageType string) error {
	if content == "" && messageType == "text" {
		return fmt.Errorf("text message content cannot be empty")
	}


	if len(content) > 65000 {
		return fmt.Errorf("message content too long (max 65000 characters)")
	}

	return nil
}


func (v *Validator) ValidateMediaURL(url string) error {
	if url == "" {
		return fmt.Errorf("media URL cannot be empty")
	}


	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("media URL must start with http:// or https://")
	}

	return nil
}


func (v *Validator) ValidateLocation(latitude, longitude float64) error {
	if latitude < -90 || latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}

	if longitude < -180 || longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	return nil
}
