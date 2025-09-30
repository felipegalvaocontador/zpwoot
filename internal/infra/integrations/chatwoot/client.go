package chatwoot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

type Client struct {
	httpClient *http.Client
	logger     *logger.Logger
	baseURL    string
	token      string
	accountID  string
}

func NewClient(baseURL, token, accountID string, logger *logger.Logger) *Client {
	return &Client{
		baseURL:   baseURL,
		token:     token,
		accountID: accountID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

func (c *Client) CreateInbox(name, webhookURL string) (*ports.ChatwootInbox, error) {
	payload := map[string]interface{}{
		"name": name,
		"channel": map[string]interface{}{
			"type":        "api",
			"webhook_url": webhookURL,
		},
	}

	var inbox ports.ChatwootInbox
	err := c.makeRequest("POST", "/inboxes", payload, &inbox)
	if err != nil {
		return nil, fmt.Errorf("failed to create inbox: %w", err)
	}

	return &inbox, nil
}

func (c *Client) ListInboxes() ([]ports.ChatwootInbox, error) {
	var response struct {
		Payload []ports.ChatwootInbox `json:"payload"`
	}

	err := c.makeRequest("GET", "/inboxes", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list inboxes: %w", err)
	}

	return response.Payload, nil
}

func (c *Client) GetInbox(inboxID int) (*ports.ChatwootInbox, error) {
	var inbox ports.ChatwootInbox
	err := c.makeRequest("GET", fmt.Sprintf("/inboxes/%d", inboxID), nil, &inbox)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}

	return &inbox, nil
}

func (c *Client) UpdateInbox(inboxID int, updates map[string]interface{}) error {
	err := c.makeRequest("PATCH", fmt.Sprintf("/inboxes/%d", inboxID), updates, nil)
	if err != nil {
		return fmt.Errorf("failed to update inbox: %w", err)
	}

	return nil
}

func (c *Client) DeleteInbox(inboxID int) error {
	err := c.makeRequest("DELETE", fmt.Sprintf("/inboxes/%d", inboxID), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete inbox: %w", err)
	}

	return nil
}

func (c *Client) CreateContact(phone, name string, inboxID int) (*ports.ChatwootContact, error) {
	payload := map[string]interface{}{
		"name":         name,
		"phone_number": phone,
		"inbox_id":     inboxID,
	}

	var contact ports.ChatwootContact
	err := c.makeRequest("POST", "/contacts", payload, &contact)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return &contact, nil
}

func (c *Client) FindContact(phone string, inboxID int) (*ports.ChatwootContact, error) {
	var response struct {
		Payload []ports.ChatwootContact `json:"payload"`
	}

	encodedPhone := url.QueryEscape(phone)
	err := c.makeRequest("GET", fmt.Sprintf("/contacts/search?q=%s", encodedPhone), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to find contact: %w", err)
	}

	if len(response.Payload) == 0 {
		return nil, fmt.Errorf("contact not found")
	}

	contact := &response.Payload[0]
	return contact, nil
}

func (c *Client) ListContactConversations(contactID int) ([]ports.ChatwootConversation, error) {
	var response struct {
		Payload []ports.ChatwootConversation `json:"payload"`
	}

	err := c.makeRequest("GET", fmt.Sprintf("/contacts/%d/conversations", contactID), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list contact conversations: %w", err)
	}

	return response.Payload, nil
}

func (c *Client) GetContact(contactID int) (*ports.ChatwootContact, error) {
	var contact ports.ChatwootContact
	err := c.makeRequest("GET", fmt.Sprintf("/contacts/%d", contactID), nil, &contact)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return &contact, nil
}

func (c *Client) UpdateContactAttributes(contactID int, attributes map[string]interface{}) error {
	payload := map[string]interface{}{
		"custom_attributes": attributes,
	}

	err := c.makeRequest("PUT", fmt.Sprintf("/contacts/%d", contactID), payload, nil)
	if err != nil {
		return fmt.Errorf("failed to update contact attributes: %w", err)
	}

	return nil
}

func (c *Client) CreateConversation(contactID, inboxID int) (*ports.ChatwootConversation, error) {
	payload := map[string]interface{}{
		"contact_id": contactID,
		"inbox_id":   inboxID,
	}

	var conversation ports.ChatwootConversation
	err := c.makeRequest("POST", "/conversations", payload, &conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return &conversation, nil
}

func (c *Client) GetConversation(contactID, inboxID int) (*ports.ChatwootConversation, error) {
	var response struct {
		Payload []ports.ChatwootConversation `json:"payload"`
	}

	err := c.makeRequest("GET", fmt.Sprintf("/conversations?contact_id=%d&inbox_id=%d", contactID, inboxID), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	if len(response.Payload) == 0 {
		return nil, fmt.Errorf("conversation not found")
	}

	return &response.Payload[0], nil
}

func (c *Client) GetConversationByID(conversationID int) (*ports.ChatwootConversation, error) {
	c.logger.InfoWithFields("Getting Chatwoot conversation by ID", map[string]interface{}{
		"conversation_id": conversationID,
	})

	var conversation ports.ChatwootConversation
	err := c.makeRequest("GET", fmt.Sprintf("/conversations/%d", conversationID), nil, &conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	return &conversation, nil
}

func (c *Client) GetConversationSenderPhone(conversationID int) (string, error) {
	c.logger.InfoWithFields("Getting conversation sender phone", map[string]interface{}{
		"conversation_id": conversationID,
	})

	var resp struct {
		Meta struct {
			Sender struct {
				PhoneNumber string `json:"phone_number"`
			} `json:"sender"`
		} `json:"meta"`
	}

	err := c.makeRequest("GET", fmt.Sprintf("/conversations/%d", conversationID), nil, &resp)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation meta: %w", err)
	}
	return resp.Meta.Sender.PhoneNumber, nil
}

func (c *Client) UpdateConversationStatus(conversationID int, status string) error {
	c.logger.InfoWithFields("Updating Chatwoot conversation status", map[string]interface{}{
		"conversation_id": conversationID,
		"status":          status,
	})

	payload := map[string]interface{}{
		"status": status,
	}

	err := c.makeRequest("POST", fmt.Sprintf("/conversations/%d/toggle_status", conversationID), payload, nil)
	if err != nil {
		return fmt.Errorf("failed to update conversation status: %w", err)
	}

	return nil
}

func (c *Client) SendMessage(conversationID int, content string) (*ports.ChatwootMessage, error) {
	return c.SendMessageWithType(conversationID, content, "incoming")
}

func (c *Client) SendMessageWithType(conversationID int, content string, messageType string) (*ports.ChatwootMessage, error) {
	payload := map[string]interface{}{
		"content":      content,
		"message_type": messageType, // incoming (from client) or outgoing (from agent)
	}

	var message ports.ChatwootMessage
	err := c.makeRequest("POST", fmt.Sprintf("/conversations/%d/messages", conversationID), payload, &message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &message, nil
}

func (c *Client) SendMediaMessage(conversationID int, content string, attachment io.Reader, filename string) (*ports.ChatwootMessage, error) {
	return c.SendMessage(conversationID, content)
}

func (c *Client) GetMessages(conversationID int, before int) ([]ports.ChatwootMessage, error) {
	var response struct {
		Payload []ports.ChatwootMessage `json:"payload"`
	}

	url := fmt.Sprintf("/conversations/%d/messages", conversationID)
	if before > 0 {
		url += fmt.Sprintf("?before=%d", before)
	}

	err := c.makeRequest("GET", url, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return response.Payload, nil
}

func (c *Client) GetAccount() (*ports.ChatwootAccount, error) {
	var account ports.ChatwootAccount
	err := c.makeRequest("GET", "", nil, &account)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

func (c *Client) UpdateAccount(updates map[string]interface{}) error {
	err := c.makeRequest("PATCH", "", updates, nil)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

func (c *Client) makeRequest(method, endpoint string, payload interface{}, result interface{}) error {
	url := fmt.Sprintf("%s/api/v1/accounts/%s%s", c.baseURL, c.accountID, endpoint)

	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api_access_token", c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API request failed with status %d (failed to read response body: %w)", resp.StatusCode, err)
		}
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) MergeContacts(baseContactID, mergeContactID int) error {
	c.logger.InfoWithFields("Merging Chatwoot contacts", map[string]interface{}{
		"base_contact_id":  baseContactID,
		"merge_contact_id": mergeContactID,
	})

	requestBody := map[string]interface{}{
		"base_contact_id":   baseContactID,
		"mergee_contact_id": mergeContactID,
	}

	url := fmt.Sprintf("/api/v1/accounts/%s/actions/contact_merge", c.accountID)

	err := c.makeRequest("POST", url, requestBody, nil)
	if err != nil {
		c.logger.ErrorWithFields("Failed to merge contacts", map[string]interface{}{
			"base_contact_id":  baseContactID,
			"merge_contact_id": mergeContactID,
			"error":            err.Error(),
		})
		return fmt.Errorf("failed to merge contacts: %w", err)
	}

	c.logger.InfoWithFields("Successfully merged contacts", map[string]interface{}{
		"base_contact_id":  baseContactID,
		"merge_contact_id": mergeContactID,
	})

	return nil
}
