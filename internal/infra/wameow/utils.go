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

type JIDValidator struct {
	phoneRegex *regexp.Regexp
}

func NewJIDValidator() *JIDValidator {
	return &JIDValidator{
		phoneRegex: regexp.MustCompile(`^\d+$`),
	}
}

func (v *JIDValidator) IsValid(jid string) bool {
	if jid == "" {
		return false
	}

	normalizedJID := v.Normalize(jid)

	if strings.Contains(normalizedJID, "@s.whatsapp.net") ||
		strings.Contains(normalizedJID, "@g.us") ||
		strings.Contains(normalizedJID, "@newsletter") {
		return true
	}

	return false
}

func (v *JIDValidator) Normalize(jid string) string {
	jid = strings.TrimSpace(jid)

	if strings.Contains(jid, "@") {
		return jid
	}

	jid = strings.TrimPrefix(jid, "+")

	jid = strings.ReplaceAll(jid, " ", "")
	jid = strings.ReplaceAll(jid, "-", "")
	jid = strings.ReplaceAll(jid, "(", "")
	jid = strings.ReplaceAll(jid, ")", "")

	if v.phoneRegex.MatchString(jid) {
		return jid + "@s.whatsapp.net"
	}

	return jid
}

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

	if parsedJID.User == "" {
		return waTypes.EmptyJID, fmt.Errorf("JID missing user part: %s", normalizedJID)
	}

	return parsedJID, nil
}

var defaultValidator = NewJIDValidator()

type JIDValidatorAdapter struct {
	validator *JIDValidator
}

func NewJIDValidatorAdapter() *JIDValidatorAdapter {
	return &JIDValidatorAdapter{
		validator: NewJIDValidator(),
	}
}

func (a *JIDValidatorAdapter) IsValid(jid string) bool {
	return a.validator.IsValid(jid)
}

func (a *JIDValidatorAdapter) Normalize(jid string) string {
	return a.validator.Normalize(jid)
}

func (a *JIDValidatorAdapter) IsNewsletterJID(jid string) bool {
	return strings.Contains(jid, "@newsletter")
}

func (a *JIDValidatorAdapter) IsValidJID(jid string) bool {
	return a.validator.IsValid(jid)
}

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

type ConnectionError struct {
	Err       error
	SessionID string
	Operation string
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

type DeviceStoreManager struct {
	container *sqlstore.Container
	logger    *logger.Logger
}

func NewDeviceStoreManager(container *sqlstore.Container, logger *logger.Logger) *DeviceStoreManager {
	return &DeviceStoreManager{
		container: container,
		logger:    logger,
	}
}

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

func GetDeviceStoreForSession(sessionID, expectedDeviceJID string, container *sqlstore.Container) *store.Device {
	tempLogger := &logger.Logger{}
	dsm := NewDeviceStoreManager(container, tempLogger)
	return dsm.GetOrCreateDeviceStore(sessionID, expectedDeviceJID)
}

func IsValidJID(jidStr string) bool {
	return defaultValidator.IsValid(jidStr)
}

func NormalizeJID(jid string) string {
	return defaultValidator.Normalize(jid)
}

func ParseJID(jid string) (waTypes.JID, error) {
	return defaultValidator.Parse(jid)
}

func GetBrazilianAlternativeNumber(phoneNumber string) string {
	cleaned := strings.ReplaceAll(phoneNumber, "+", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, "(", "")
	cleaned = strings.ReplaceAll(cleaned, ")", "")

	if len(cleaned) == 13 && strings.HasPrefix(cleaned, "55") {
		countryCode := cleaned[:2] // "55"
		ddd := cleaned[2:4]        // Area code (DDD)
		joker := cleaned[4:5]      // The "joker" digit (usually 9)
		number := cleaned[5:]      // The 8-digit number

		if countryCode == "55" {
			dddInt := 0
			if len(ddd) == 2 {
				if d, err := strconv.Atoi(ddd); err == nil {
					dddInt = d
				}
			}

			jokerInt := 0
			if len(joker) == 1 {
				if j, err := strconv.Atoi(joker); err == nil {
					jokerInt = j
				}
			}

			if jokerInt < 7 || dddInt < 31 {
				return "+" + countryCode + ddd + number
			}
		}
	}

	if len(cleaned) == 12 && strings.HasPrefix(cleaned, "55") {
		countryCode := cleaned[:2] // "55"
		ddd := cleaned[2:4]        // Area code (DDD)
		number := cleaned[4:]      // The 8-digit number

		return "+" + countryCode + ddd + "9" + number
	}

	return ""
}

func ParseJIDWithBrazilianFallback(phoneNumber string) ([]waTypes.JID, error) {
	var jids []waTypes.JID

	jid, err := defaultValidator.Parse(phoneNumber)
	if err == nil {
		jids = append(jids, jid)
	}

	if strings.HasPrefix(phoneNumber, "+55") {
		cleaned := strings.ReplaceAll(phoneNumber, "+55", "")
		cleaned = strings.ReplaceAll(cleaned, " ", "")
		cleaned = strings.ReplaceAll(cleaned, "-", "")

		if len(cleaned) >= 10 {
			areaCode := cleaned[:2]
			number := cleaned[2:]

			var alternativeNumber string

			if len(number) == 9 && strings.HasPrefix(number, "9") {
				alternativeNumber = "+55" + areaCode + number[1:] // Remove the 9
			}
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

	validSessionRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validSessionRegex.MatchString(sessionID) {
		return fmt.Errorf("session ID contains invalid characters (only alphanumeric, underscore, and hyphen allowed): %s", sessionID)
	}

	return nil
}

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
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	SessionID string                 `json:"session_id"`
	DeviceJID string                 `json:"device_jid,omitempty"`
	LastError string                 `json:"last_error,omitempty"`
	UpdatedAt int64                  `json:"updated_at"`
	Connected bool                   `json:"connected"`
	LoggedIn  bool                   `json:"logged_in"`
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

type TextNormalizer struct{}

func NewTextNormalizer() *TextNormalizer {
	return &TextNormalizer{}
}

func (t *TextNormalizer) NormalizeText(text string) string {
	if !utf8.ValidString(text) {
		text = strings.ToValidUTF8(text, "")
	}
	return text
}

func (t *TextNormalizer) IsEmojiText(text string) bool {
	for _, r := range text {
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

func NormalizeMessageText(text string) string {
	normalizer := NewTextNormalizer()
	return normalizer.NormalizeText(text)
}

var defaultTextNormalizer = NewTextNormalizer()

func IsEmojiMessage(text string) bool {
	return defaultTextNormalizer.IsEmojiText(text)
}
