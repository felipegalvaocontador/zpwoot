package chatwoot

import (
	"fmt"
	"regexp"
	"strings"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

// ContactSync handles contact synchronization between WhatsApp and Chatwoot
type ContactSync struct {
	logger *logger.Logger
	client ports.ChatwootClient
}

// NewContactSync creates a new contact synchronizer
func NewContactSync(logger *logger.Logger, client ports.ChatwootClient) *ContactSync {
	return &ContactSync{
		logger: logger,
		client: client,
	}
}

// CreateOrUpdateContact creates or updates a contact in Chatwoot
func (cs *ContactSync) CreateOrUpdateContact(phone, name string, inboxID int, mergeBrazilContacts bool) (*ports.ChatwootContact, error) {
	// Normalize phone number
	normalizedPhone := cs.normalizePhoneNumber(phone)

	// Handle Brazilian contact merging
	if mergeBrazilContacts {
		mergedPhone := cs.mergeBrazilianContacts(normalizedPhone)
		if mergedPhone != normalizedPhone {
			cs.logger.InfoWithFields("Merged Brazilian contact", map[string]interface{}{
				"original": normalizedPhone,
				"merged":   mergedPhone,
			})
			normalizedPhone = mergedPhone
		}
	}

	// Try to find existing contact
	existingContact, err := cs.client.FindContact(normalizedPhone, inboxID)
	if err == nil {
		// Contact exists, update if needed
		if existingContact.Name != name && name != "" {
			err = cs.client.UpdateContactAttributes(existingContact.ID, map[string]interface{}{
				"name": name,
			})
			if err == nil {
				existingContact.Name = name
			}
		}
		return existingContact, nil
	}

	// Contact doesn't exist, create new one
	contact, err := cs.client.CreateContact(normalizedPhone, name, inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return contact, nil
}

// ImportContacts imports contacts from WhatsApp to Chatwoot
func (cs *ContactSync) ImportContacts(contacts []ContactImportData, inboxID int, mergeBrazilContacts bool) ([]ContactImportResult, error) {
	results := make([]ContactImportResult, 0, len(contacts))

	for _, contactData := range contacts {
		result := ContactImportResult{
			Phone:   contactData.Phone,
			Name:    contactData.Name,
			Success: false,
		}

		contact, err := cs.CreateOrUpdateContact(contactData.Phone, contactData.Name, inboxID, mergeBrazilContacts)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
			result.ContactID = contact.ID
		}

		results = append(results, result)
	}

	return results, nil
}

// MergeBrazilianContacts merges Brazilian phone numbers according to Evolution API logic
func (cs *ContactSync) mergeBrazilianContacts(phone string) string {
	// Remove all non-digit characters
	cleanPhone := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Check if it's a Brazilian number (+55)
	if !strings.HasPrefix(cleanPhone, "55") {
		return phone
	}

	// Remove country code for processing
	localNumber := cleanPhone[2:]

	// Brazilian mobile numbers have 11 digits (including area code)
	// Fixed numbers have 10 digits (including area code)
	if len(localNumber) == 11 {
		// Mobile number - check if it has the extra 9
		areaCode := localNumber[:2]
		number := localNumber[2:]

		// If the number starts with 9 and has 9 digits, it's the new format
		if len(number) == 9 && strings.HasPrefix(number, "9") {
			// Keep as is - this is the correct format
			return "55" + localNumber
		}

		// If it has 8 digits, add the 9
		if len(number) == 8 {
			return "55" + areaCode + "9" + number
		}
	}

	// Return original if no transformation needed
	return phone
}

// normalizePhoneNumber normalizes phone number format
func (cs *ContactSync) normalizePhoneNumber(phone string) string {
	// Remove common prefixes and formatting
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.TrimPrefix(phone, "00")

	// Remove all non-digit characters
	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	// Add default country code if needed (Brazil = 55)
	if len(phone) <= 11 && !strings.HasPrefix(phone, "55") {
		// Assume Brazilian number if no country code
		phone = "55" + phone
	}

	return phone
}

// GetContactByPhone gets a contact by phone number
func (cs *ContactSync) GetContactByPhone(phone string, inboxID int) (*ports.ChatwootContact, error) {
	normalizedPhone := cs.normalizePhoneNumber(phone)
	return cs.client.FindContact(normalizedPhone, inboxID)
}

// UpdateContactAttributes updates contact custom attributes
func (cs *ContactSync) UpdateContactAttributes(contactID int, attributes map[string]interface{}) error {
	return cs.client.UpdateContactAttributes(contactID, attributes)
}

// ContactImportData represents data for importing a contact
type ContactImportData struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Phone      string                 `json:"phone"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email,omitempty"`
}

// ContactImportResult represents the result of importing a contact
type ContactImportResult struct {
	Phone     string `json:"phone"`
	Name      string `json:"name"`
	Error     string `json:"error,omitempty"`
	ContactID int    `json:"contact_id,omitempty"`
	Success   bool   `json:"success"`
}

// ValidatePhoneNumber validates if a phone number is valid
func (cs *ContactSync) ValidatePhoneNumber(phone string) bool {
	normalized := cs.normalizePhoneNumber(phone)

	// Basic validation - should have at least 10 digits
	if len(normalized) < 10 {
		return false
	}

	// Should contain only digits
	matched, err := regexp.MatchString(`^\d+$`, normalized)
	if err != nil {
		// If regex fails, assume invalid phone number
		return false
	}
	return matched
}

// FormatPhoneForDisplay formats phone number for display
func (cs *ContactSync) FormatPhoneForDisplay(phone string) string {
	normalized := cs.normalizePhoneNumber(phone)

	// Format Brazilian numbers
	if strings.HasPrefix(normalized, "55") && len(normalized) >= 12 {
		// +55 (11) 99999-9999
		areaCode := normalized[2:4]
		if len(normalized) == 13 {
			// Mobile with 9
			number := normalized[4:]
			return fmt.Sprintf("+55 (%s) %s-%s", areaCode, number[:5], number[5:])
		} else if len(normalized) == 12 {
			// Fixed line
			number := normalized[4:]
			return fmt.Sprintf("+55 (%s) %s-%s", areaCode, number[:4], number[4:])
		}
	}

	// Default format with +
	return "+" + normalized
}
