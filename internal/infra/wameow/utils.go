// Refactored: centralized utilities; improved validation; standardized error handling
package wameow

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
	"zpwoot/platform/logger"
)

// JIDValidator handles JID validation and normalization
type JIDValidator struct {
	phoneRegex *regexp.Regexp
}

// NewJIDValidator creates a new JID validator
func NewJIDValidator() *JIDValidator {
	return &JIDValidator{
		phoneRegex: regexp.MustCompile(`^\d+$`),
	}
}

// IsValid checks if a JID is valid after normalization
// Supports: +5511999999999, 5511999999999, 5511999999999@s.whatsapp.net, groups@g.us, newsletters@newsletter
func (v *JIDValidator) IsValid(jid string) bool {
	if jid == "" {
		return false
	}

	// Normalize first, then validate
	normalizedJID := v.Normalize(jid)

	// Check for WhatsApp JID format (individual, group, or newsletter)
	if strings.Contains(normalizedJID, "@s.whatsapp.net") ||
		strings.Contains(normalizedJID, "@g.us") ||
		strings.Contains(normalizedJID, "@newsletter") {
		return true
	}

	return false
}

// Normalize converts a JID to standard WhatsApp format
// Supports formats: +5511999999999, 5511999999999, 5511999999999@s.whatsapp.net, newsletters@newsletter
func (v *JIDValidator) Normalize(jid string) string {
	jid = strings.TrimSpace(jid)

	// If it's already a full JID (contains @), return as is
	if strings.Contains(jid, "@") {
		return jid
	}

	// Remove leading + if present
	jid = strings.TrimPrefix(jid, "+")

	// Remove any spaces or dashes
	jid = strings.ReplaceAll(jid, " ", "")
	jid = strings.ReplaceAll(jid, "-", "")
	jid = strings.ReplaceAll(jid, "(", "")
	jid = strings.ReplaceAll(jid, ")", "")

	// If it's just a phone number (digits only), add the WhatsApp suffix
	if v.phoneRegex.MatchString(jid) {
		return jid + "@s.whatsapp.net"
	}

	return jid
}

// Parse converts a string JID to types.JID with intelligent normalization
func (v *JIDValidator) Parse(jid string) (waTypes.JID, error) {
	if jid == "" {
		return waTypes.EmptyJID, fmt.Errorf("JID cannot be empty")
	}

	normalizedJID := v.Normalize(jid)

	if !v.IsValid(normalizedJID) {
		return waTypes.EmptyJID, fmt.Errorf("invalid JID format: %s (normalized: %s)", jid, normalizedJID)
	}

	parsedJID, err := waTypes.ParseJID(normalizedJID)
	if err != nil {
		return waTypes.EmptyJID, fmt.Errorf("failed to parse JID %s: %w", normalizedJID, err)
	}

	// Additional validation
	if parsedJID.User == "" {
		return waTypes.EmptyJID, fmt.Errorf("JID missing user part: %s", normalizedJID)
	}

	return parsedJID, nil
}

// Global validator instance for backward compatibility
var defaultValidator = NewJIDValidator()

// JIDValidatorAdapter adapts our JIDValidator to the domain interface
type JIDValidatorAdapter struct {
	validator *JIDValidator
}

// NewJIDValidatorAdapter creates a new adapter
func NewJIDValidatorAdapter() *JIDValidatorAdapter {
	return &JIDValidatorAdapter{
		validator: NewJIDValidator(),
	}
}

// IsValid implements the domain interface
func (a *JIDValidatorAdapter) IsValid(jid string) bool {
	return a.validator.IsValid(jid)
}

// Normalize implements the domain interface
func (a *JIDValidatorAdapter) Normalize(jid string) string {
	return a.validator.Normalize(jid)
}

// IsNewsletterJID checks if a JID is a newsletter JID
func (a *JIDValidatorAdapter) IsNewsletterJID(jid string) bool {
	return strings.Contains(jid, "@newsletter")
}

// IsValidJID checks if a JID is valid (alias for IsValid)
func (a *JIDValidatorAdapter) IsValidJID(jid string) bool {
	return a.validator.IsValid(jid)
}

// ParseJID parses a JID string to extract components
func (a *JIDValidatorAdapter) ParseJID(jid string) (string, error) {
	if jid == "" {
		return "", fmt.Errorf("JID cannot be empty")
	}

	normalizedJID := a.validator.Normalize(jid)
	if !a.validator.IsValid(normalizedJID) {
		return "", fmt.Errorf("invalid JID format: %s (normalized: %s)", jid, normalizedJID)
	}

	return normalizedJID, nil
}

// ConnectionError represents connection-related errors
type ConnectionError struct {
	SessionID string
	Operation string
	Err       error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error for session %s during %s: %v", e.SessionID, e.Operation, e.Err)
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

func newConnectionError(sessionID, operation string, err error) *ConnectionError {
	return &ConnectionError{
		SessionID: sessionID,
		Operation: operation,
		Err:       err,
	}
}

func ValidateClientAndStore(client *whatsmeow.Client, sessionID string) error {
	if client == nil {
		return fmt.Errorf("client is nil for session %s", sessionID)
	}

	if client.Store == nil {
		return fmt.Errorf("client store is nil for session %s", sessionID)
	}

	if client.Store.ID == nil {
		return fmt.Errorf("client store ID is nil for session %s", sessionID)
	}

	return nil
}

// DeviceStoreManager handles device store operations
type DeviceStoreManager struct {
	container *sqlstore.Container
	logger    *logger.Logger
}

// NewDeviceStoreManager creates a new device store manager
func NewDeviceStoreManager(container *sqlstore.Container, logger *logger.Logger) *DeviceStoreManager {
	return &DeviceStoreManager{
		container: container,
		logger:    logger,
	}
}

// GetOrCreateDeviceStore gets an existing device store or creates a new one
func (dsm *DeviceStoreManager) GetOrCreateDeviceStore(sessionID, expectedDeviceJID string) *store.Device {
	if expectedDeviceJID != "" {
		if deviceStore := dsm.getExistingDeviceStore(sessionID, expectedDeviceJID); deviceStore != nil {
			return deviceStore
		}
	}

	return dsm.createNewDeviceStore(sessionID)
}

func (dsm *DeviceStoreManager) getExistingDeviceStore(sessionID, expectedDeviceJID string) *store.Device {
	dsm.logger.InfoWithFields("Loading existing device store", map[string]interface{}{
		"session_id": sessionID,
		"device_jid": expectedDeviceJID,
	})

	jid, err := waTypes.ParseJID(expectedDeviceJID)
	if err != nil {
		dsm.logger.WarnWithFields("Failed to parse expected JID", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": expectedDeviceJID,
			"error":      err.Error(),
		})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deviceStore, err := dsm.container.GetDevice(ctx, jid)
	if err != nil {
		dsm.logger.WarnWithFields("Failed to get device store", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": expectedDeviceJID,
			"error":      err.Error(),
		})
		return nil
	}

	if deviceStore != nil {
		dsm.logger.InfoWithFields("Successfully loaded existing device store", map[string]interface{}{
			"session_id": sessionID,
			"device_jid": expectedDeviceJID,
		})
	}

	return deviceStore
}

func (dsm *DeviceStoreManager) createNewDeviceStore(sessionID string) *store.Device {
	dsm.logger.InfoWithFields("Creating new device store", map[string]interface{}{
		"session_id": sessionID,
	})

	deviceStore := dsm.container.NewDevice()
	if deviceStore == nil {
		dsm.logger.ErrorWithFields("Failed to create device store", map[string]interface{}{
			"session_id": sessionID,
		})
		return nil
	}

	dsm.logger.InfoWithFields("Device store ready", map[string]interface{}{
		"session_id": sessionID,
	})

	return deviceStore
}

// GetDeviceStoreForSession maintains backward compatibility
func GetDeviceStoreForSession(sessionID, expectedDeviceJID string, container *sqlstore.Container) *store.Device {
	// Create a temporary logger for backward compatibility
	tempLogger := &logger.Logger{}
	dsm := NewDeviceStoreManager(container, tempLogger)
	return dsm.GetOrCreateDeviceStore(sessionID, expectedDeviceJID)
}

// IsValidJID checks if a JID is valid (backward compatibility)
func IsValidJID(jidStr string) bool {
	return defaultValidator.IsValid(jidStr)
}

// NormalizeJID normalizes a JID (backward compatibility)
func NormalizeJID(jid string) string {
	return defaultValidator.Normalize(jid)
}

// ParseJID parses a JID (backward compatibility)
func ParseJID(jid string) (waTypes.JID, error) {
	return defaultValidator.Parse(jid)
}

// GetBrazilianAlternativeNumber returns the alternative format for Brazilian numbers
// Based on Evolution API logic for handling WhatsApp ID inconsistency in Brazil
func GetBrazilianAlternativeNumber(phoneNumber string) string {
	// Remove + and clean the number
	cleaned := strings.ReplaceAll(phoneNumber, "+", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	// Check if it matches Brazilian pattern: 55 + 2 digits (DDD) + 1 digit + 8 digits
	// Regex equivalent: /^(\d{2})(\d{2})\d{1}(\d{8})$/
	if len(cleaned) == 13 && strings.HasPrefix(cleaned, "55") {
		countryCode := cleaned[:2] // "55"
		ddd := cleaned[2:4]        // Area code (DDD)
		joker := cleaned[4:5]      // The "joker" digit (usually 9)
		number := cleaned[5:]      // The 8-digit number

		if countryCode == "55" {
			// Convert DDD to int for comparison
			dddInt := 0
			if len(ddd) == 2 {
				if d, err := strconv.Atoi(ddd); err == nil {
					dddInt = d
				}
			}

			// Convert joker to int for comparison
			jokerInt := 0
			if len(joker) == 1 {
				if j, err := strconv.Atoi(joker); err == nil {
					jokerInt = j
				}
			}

			// Evolution API logic: if (joker < 7 || ddd < 31)
			// This means: remove the 9 for certain area codes or joker digits
			if jokerInt < 7 || dddInt < 31 {
				// Return without the joker digit (8-digit format)
				return "+" + countryCode + ddd + number
			}
		}
	}

	// Check if it's a 12-digit Brazilian number (without the 9)
	if len(cleaned) == 12 && strings.HasPrefix(cleaned, "55") {
		countryCode := cleaned[:2] // "55"
		ddd := cleaned[2:4]        // Area code (DDD)
		number := cleaned[4:]      // The 8-digit number

		// Add the 9 to create the 13-digit format
		return "+" + countryCode + ddd + "9" + number
	}

	return ""
}

// ParseJIDWithBrazilianFallback tries to parse a JID and if it's a Brazilian number,
// also tries the alternative format (with/without the 9th digit)
func ParseJIDWithBrazilianFallback(phoneNumber string) ([]waTypes.JID, error) {
	var jids []waTypes.JID

	// Try the original number first
	jid, err := defaultValidator.Parse(phoneNumber)
	if err == nil {
		jids = append(jids, jid)
	}

	// Check if it's a Brazilian mobile number (+55)
	if strings.HasPrefix(phoneNumber, "+55") {
		// Extract the area code and number
		cleaned := strings.ReplaceAll(phoneNumber, "+55", "")
		cleaned = strings.ReplaceAll(cleaned, " ", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")

		if len(cleaned) >= 10 {
			areaCode := cleaned[:2]
			number := cleaned[2:]

			var alternativeNumber string

			// If it has 9 digits (mobile with 9), try without the 9
			if len(number) == 9 && strings.HasPrefix(number, "9") {
				alternativeNumber = "+55" + areaCode + number[1:] // Remove the 9
			}
			// If it has 8 digits (mobile without 9), try with the 9
			if len(number) == 8 {
				alternativeNumber = "+55" + areaCode + "9" + number // Add the 9
			}

			if alternativeNumber != "" && alternativeNumber != phoneNumber {
				altJid, altErr := defaultValidator.Parse(alternativeNumber)
				if altErr == nil {
					jids = append(jids, altJid)
				}
			}
		}
	}

	if len(jids) == 0 {
		return nil, fmt.Errorf("failed to parse JID for %s: %w", phoneNumber, err)
	}

	return jids, nil
}

func FormatJID(jid waTypes.JID) string {
	if jid.IsEmpty() {
		return ""
	}
	return jid.String()
}

func GetClientInfo(client *whatsmeow.Client) map[string]interface{} {
	if client == nil {
		return map[string]interface{}{
			"client":    "nil",
			"connected": false,
		}
	}

	info := map[string]interface{}{
		"connected": client.IsConnected(),
		"logged_in": client.IsLoggedIn(),
	}

	if client.Store != nil && client.Store.ID != nil {
		info["device_jid"] = FormatJID(*client.Store.ID)
	}

	return info
}

// ValidateSessionID validates a session ID with improved rules
func ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	if len(sessionID) < 3 {
		return fmt.Errorf("session ID too short (min 3 characters): %s", sessionID)
	}

	if len(sessionID) > 100 {
		return fmt.Errorf("session ID too long (max 100 characters): %s", sessionID)
	}

	// Check for valid characters (alphanumeric, underscore, hyphen)
	validSessionRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validSessionRegex.MatchString(sessionID) {
		return fmt.Errorf("session ID contains invalid characters (only alphanumeric, underscore, and hyphen allowed): %s", sessionID)
	}

	return nil
}

// SafeClientOperation executes an operation safely with validation and panic recovery
func SafeClientOperation(client *whatsmeow.Client, sessionID string, operation func() error, logger *logger.Logger) error {
	if err := ValidateClientAndStore(client, sessionID); err != nil {
		return newConnectionError(sessionID, "validate", err)
	}

	if err := ValidateSessionID(sessionID); err != nil {
		return newConnectionError(sessionID, "validate_session", err)
	}

	defer func() {
		if r := recover(); r != nil {
			if logger != nil {
				logger.ErrorWithFields("Panic in client operation", map[string]interface{}{
					"session_id": sessionID,
					"panic":      r,
				})
			}
		}
	}()

	return operation()
}

func GetStoreInfo(deviceStore *store.Device) map[string]interface{} {
	if deviceStore == nil {
		return map[string]interface{}{
			"store": "nil",
		}
	}

	info := map[string]interface{}{
		"exists": true,
	}

	if deviceStore.ID != nil {
		info["device_jid"] = FormatJID(*deviceStore.ID)
	}

	return info
}

type ConnectionStatus struct {
	SessionID string                 `json:"session_id"`
	Connected bool                   `json:"connected"`
	LoggedIn  bool                   `json:"logged_in"`
	DeviceJID string                 `json:"device_jid,omitempty"`
	LastError string                 `json:"last_error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt int64                  `json:"updated_at"`
}

func GetConnectionStatus(client *whatsmeow.Client, sessionID string) *ConnectionStatus {
	status := &ConnectionStatus{
		SessionID: sessionID,
		UpdatedAt: getCurrentTimestamp(),
		Metadata:  make(map[string]interface{}),
	}

	if client == nil {
		status.LastError = "client is nil"
		return status
	}

	status.Connected = client.IsConnected()
	status.LoggedIn = client.IsLoggedIn()

	if client.Store != nil && client.Store.ID != nil {
		status.DeviceJID = FormatJID(*client.Store.ID)
	}

	status.Metadata = GetClientInfo(client)

	return status
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func IsRecoverableError(err error) bool {
	return err != nil
}

// GetErrorCategory categorizes errors for better handling
func GetErrorCategory(err error) string {
	if err == nil {
		return "none"
	}

	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "connection"):
		return "connection"
	case strings.Contains(errStr, "auth") || strings.Contains(errStr, "login"):
		return "authentication"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "network"):
		return "network"
	case strings.Contains(errStr, "context"):
		return "context"
	default:
		return "unknown"
	}
}

// TextNormalizer handles text normalization for emojis and special characters
type TextNormalizer struct{}

// NewTextNormalizer creates a new text normalizer
func NewTextNormalizer() *TextNormalizer {
	return &TextNormalizer{}
}

// NormalizeText ensures text is properly encoded for WhatsApp
func (t *TextNormalizer) NormalizeText(text string) string {
	// Ensure valid UTF-8
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "")
	}
	return text
}

// IsEmojiText checks if text contains emojis
func (t *TextNormalizer) IsEmojiText(text string) bool {
	for _, r := range text {
		// Check for emoji ranges
		if (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
			(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols
			(r >= 0x1F680 && r <= 0x1F6FF) || // Transport
			(r >= 0x1F1E0 && r <= 0x1F1FF) || // Flags
			(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
			(r >= 0x2700 && r <= 0x27BF) { // Dingbats
			return true
		}
	}
	return false
}

// NormalizeMessageText is a global helper for message text normalization
func NormalizeMessageText(text string) string {
	normalizer := NewTextNormalizer()
	return normalizer.NormalizeText(text)
}

// Global text normalizer instance
var defaultTextNormalizer = NewTextNormalizer()

// IsEmojiMessage checks if a message contains emojis (global helper)
func IsEmojiMessage(text string) bool {
	return defaultTextNormalizer.IsEmojiText(text)
}
