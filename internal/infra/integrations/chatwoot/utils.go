package chatwoot

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"zpwoot/platform/logger"
)

type Utils struct {
	logger *logger.Logger
}

func NewUtils(logger *logger.Logger) *Utils {
	return &Utils{
		logger: logger,
	}
}

func (u *Utils) ConvertJIDToPhone(jid string) string {
	u.logger.DebugWithFields("Converting JID to phone", map[string]interface{}{
		"jid": jid,
	})

	phone := strings.Split(jid, "@")[0]

	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	u.logger.DebugWithFields("Converted JID to phone", map[string]interface{}{
		"jid":   jid,
		"phone": phone,
	})

	return phone
}

func (u *Utils) ConvertPhoneToJID(phone string) string {
	u.logger.DebugWithFields("Converting phone to JID", map[string]interface{}{
		"phone": phone,
	})

	cleanPhone := u.NormalizePhoneNumber(phone)

	jid := cleanPhone + "@s.whatsapp.net"

	u.logger.DebugWithFields("Converted phone to JID", map[string]interface{}{
		"phone": phone,
		"jid":   jid,
	})

	return jid
}

func (u *Utils) NormalizePhoneNumber(phone string) string {
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.TrimPrefix(phone, "00")

	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	if len(phone) <= 11 && !strings.HasPrefix(phone, "55") {
		if len(phone) >= 10 {
			phone = "55" + phone
		}
	}

	return phone
}

func (u *Utils) FormatBrazilianPhone(phone string) string {
	u.logger.DebugWithFields("Formatting Brazilian phone", map[string]interface{}{
		"original": phone,
	})

	normalized := u.NormalizePhoneNumber(phone)

	if !strings.HasPrefix(normalized, "55") {
		return phone
	}

	localNumber := normalized[2:]

	if len(localNumber) == 10 {
		areaCode := localNumber[:2]
		number := localNumber[2:]

		if len(areaCode) == 2 {
			formatted := "55" + areaCode + "9" + number
			u.logger.DebugWithFields("Added 9 to Brazilian mobile", map[string]interface{}{
				"original":  phone,
				"formatted": formatted,
			})
			return formatted
		}
	}

	return normalized
}

func (u *Utils) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL must use http or https scheme")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must have a host")
	}

	return nil
}

func (u *Utils) ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	if len(token) < 10 {
		return fmt.Errorf("token is too short")
	}

	matched, err := regexp.MatchString(`^[a-zA-Z0-9]+$`, token)
	if err != nil {
		return fmt.Errorf("regex validation failed: %w", err)
	}
	if !matched {
		return fmt.Errorf("token contains invalid characters")
	}

	return nil
}

func (u *Utils) ValidateAccountID(accountID string) error {
	if accountID == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	matched, err := regexp.MatchString(`^\d+$`, accountID)
	if err != nil {
		return fmt.Errorf("regex validation failed: %w", err)
	}
	if !matched {
		return fmt.Errorf("account ID must be numeric")
	}

	return nil
}

func (u *Utils) RetryWithBackoff(operation func() error, maxRetries int, initialDelay time.Duration) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			attemptUint := uint(attempt)
			var exponent uint
			if attemptUint > 0 {
				exponent = attemptUint - 1
			}
			if exponent > 10 {
				exponent = 10
			}
			delay := time.Duration(1<<exponent) * initialDelay
			u.logger.InfoWithFields("Retrying operation", map[string]interface{}{
				"attempt": attempt,
				"delay":   delay.String(),
			})
			time.Sleep(delay)
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				u.logger.InfoWithFields("Operation succeeded after retry", map[string]interface{}{
					"attempts": attempt + 1,
				})
			}
			return nil
		}

		lastErr = err
		u.logger.WarnWithFields("Operation failed", map[string]interface{}{
			"attempt": attempt + 1,
			"error":   err.Error(),
		})
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, lastErr)
}

func (u *Utils) IsValidJID(jid string) bool {
	patterns := []string{
		`^\d+@s\.whatsapp\.net$`,
		`^\d+-\d+@g\.us$`,
		`^status@broadcast$`,
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, jid)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}

func (u *Utils) ExtractPhoneFromJID(jid string) string {
	if !u.IsValidJID(jid) {
		return ""
	}

	parts := strings.Split(jid, "@")
	if len(parts) == 0 {
		return ""
	}

	phone := parts[0]

	if strings.Contains(phone, "-") {
		groupParts := strings.Split(phone, "-")
		if len(groupParts) > 0 {
			phone = groupParts[0]
		}
	}

	return phone
}

func (u *Utils) GenerateWebhookURL(baseURL, sessionID string) string {
	baseURL = strings.TrimSuffix(baseURL, "/")

	webhookURL := fmt.Sprintf("%s/chatwoot/webhook/%s", baseURL, url.QueryEscape(sessionID))

	u.logger.DebugWithFields("Generated webhook URL", map[string]interface{}{
		"session_id":  sessionID,
		"webhook_url": webhookURL,
	})

	return webhookURL
}

func (u *Utils) SanitizeInboxName(name string) string {
	sanitized := regexp.MustCompile(`[^a-zA-Z0-9\s\-_]`).ReplaceAllString(name, "")
	sanitized = strings.TrimSpace(sanitized)

	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	if sanitized == "" {
		sanitized = "WhatsApp Inbox"
	}

	return sanitized
}

func (u *Utils) ParseChatwootURL(urlStr string) (*ChatwootURLInfo, error) {
	err := u.ValidateURL(urlStr)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	info := &ChatwootURLInfo{
		BaseURL: fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host),
		Host:    parsedURL.Host,
		Scheme:  parsedURL.Scheme,
		Port:    parsedURL.Port(),
		Path:    parsedURL.Path,
	}

	if info.Port == "" {
		if info.Scheme == "https" {
			info.Port = "443"
		} else {
			info.Port = "80"
		}
	}

	return info, nil
}

type ChatwootURLInfo struct {
	BaseURL string `json:"base_url"`
	Host    string `json:"host"`
	Scheme  string `json:"scheme"`
	Port    string `json:"port"`
	Path    string `json:"path"`
}

func (u *Utils) GetErrorCategory(err error) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())

	errorCategories := map[string][]string{
		"connection":     {"connection"},
		"timeout":        {"timeout"},
		"authentication": {"unauthorized", "401"},
		"authorization":  {"forbidden", "403"},
		"not_found":      {"not found", "404"},
		"rate_limit":     {"rate limit", "429"},
		"server_error":   {"server", "500"},
		"network":        {"network"},
		"parsing":        {"parse", "json"},
	}

	for category, keywords := range errorCategories {
		for _, keyword := range keywords {
			if strings.Contains(errStr, keyword) {
				return category
			}
		}
	}

	return "unknown"
}
