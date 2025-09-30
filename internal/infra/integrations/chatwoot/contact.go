package chatwoot

import (
	"fmt"
	"regexp"
	"strings"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type ContactSync struct {
	logger *logger.Logger
	client ports.ChatwootClient
}

func NewContactSync(logger *logger.Logger, client ports.ChatwootClient) *ContactSync {
	return &ContactSync{
		logger: logger,
		client: client,
	}
}

func (cs *ContactSync) CreateOrUpdateContact(phone, name string, inboxID int, mergeBrazilContacts bool) (*ports.ChatwootContact, error) {
	normalizedPhone := cs.normalizePhoneNumber(phone)

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

	existingContact, err := cs.client.FindContact(normalizedPhone, inboxID)
	if err == nil {
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

	contact, err := cs.client.CreateContact(normalizedPhone, name, inboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return contact, nil
}

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

func (cs *ContactSync) mergeBrazilianContacts(phone string) string {
	cleanPhone := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	if !strings.HasPrefix(cleanPhone, "55") {
		return phone
	}

	localNumber := cleanPhone[2:]

	if len(localNumber) == 11 {
		areaCode := localNumber[:2]
		number := localNumber[2:]

		if len(number) == 9 && strings.HasPrefix(number, "9") {
			return "55" + localNumber
		}

		if len(number) == 8 {
			return "55" + areaCode + "9" + number
		}
	}

	return phone
}

func (cs *ContactSync) normalizePhoneNumber(phone string) string {
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.TrimPrefix(phone, "00")

	phone = regexp.MustCompile(`\D`).ReplaceAllString(phone, "")

	if len(phone) <= 11 && !strings.HasPrefix(phone, "55") {
		phone = "55" + phone
	}

	return phone
}

func (cs *ContactSync) GetContactByPhone(phone string, inboxID int) (*ports.ChatwootContact, error) {
	normalizedPhone := cs.normalizePhoneNumber(phone)
	return cs.client.FindContact(normalizedPhone, inboxID)
}

func (cs *ContactSync) UpdateContactAttributes(contactID int, attributes map[string]interface{}) error {
	return cs.client.UpdateContactAttributes(contactID, attributes)
}

type ContactImportData struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Phone      string                 `json:"phone"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email,omitempty"`
}

type ContactImportResult struct {
	Phone     string `json:"phone"`
	Name      string `json:"name"`
	Error     string `json:"error,omitempty"`
	ContactID int    `json:"contact_id,omitempty"`
	Success   bool   `json:"success"`
}

func (cs *ContactSync) ValidatePhoneNumber(phone string) bool {
	normalized := cs.normalizePhoneNumber(phone)

	if len(normalized) < 10 {
		return false
	}

	matched, err := regexp.MatchString(`^\d+$`, normalized)
	if err != nil {
		return false
	}
	return matched
}

func (cs *ContactSync) FormatPhoneForDisplay(phone string) string {
	normalized := cs.normalizePhoneNumber(phone)

	if strings.HasPrefix(normalized, "55") && len(normalized) >= 12 {
		areaCode := normalized[2:4]
		if len(normalized) == 13 {
			number := normalized[4:]
			return fmt.Sprintf("+55 (%s) %s-%s", areaCode, number[:5], number[5:])
		} else if len(normalized) == 12 {
			number := normalized[4:]
			return fmt.Sprintf("+55 (%s) %s-%s", areaCode, number[:4], number[4:])
		}
	}

	return "+" + normalized
}
