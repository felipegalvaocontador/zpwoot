package chatwoot

import (
	"regexp"
	"strings"

	"zpwoot/platform/logger"
)

// MessageFormatter handles message formatting between WhatsApp and Chatwoot
type MessageFormatter struct {
	logger *logger.Logger
}

// NewMessageFormatter creates a new message formatter
func NewMessageFormatter(logger *logger.Logger) *MessageFormatter {
	return &MessageFormatter{
		logger: logger,
	}
}

// ============================================================================
// MAIN FORMATTING METHODS
// ============================================================================

// FormatMarkdownForChatwoot converts WhatsApp markdown to Chatwoot markdown
func (mf *MessageFormatter) FormatMarkdownForChatwoot(content string) string {
	mf.logger.DebugWithFields("Formatting markdown for Chatwoot", map[string]interface{}{
		"original_length": len(content),
	})

	// WhatsApp → Chatwoot conversions based on Evolution API
	// * → **
	content = mf.convertBoldMarkdown(content, "*", "**")

	// _ → *
	content = mf.convertItalicMarkdown(content, "_", "*")

	// ~ → ~~
	content = mf.convertStrikethroughMarkdown(content, "~", "~~")

	mf.logger.DebugWithFields("Formatted markdown for Chatwoot", map[string]interface{}{
		"formatted_length": len(content),
	})

	return content
}

// FormatMarkdownForWhatsApp converts Chatwoot markdown to WhatsApp markdown
func (mf *MessageFormatter) FormatMarkdownForWhatsApp(content string) string {
	mf.logger.DebugWithFields("Formatting markdown for WhatsApp", map[string]interface{}{
		"original_length": len(content),
	})

	// Chatwoot → WhatsApp conversions (reverse of above)
	// ** → *
	content = mf.convertBoldMarkdown(content, "**", "*")

	// * → _ (but not if it's part of **)
	content = mf.convertItalicMarkdown(content, "*", "_")

	// ~~ → ~
	content = mf.convertStrikethroughMarkdown(content, "~~", "~")

	mf.logger.DebugWithFields("Formatted markdown for WhatsApp", map[string]interface{}{
		"formatted_length": len(content),
	})

	return content
}

// ============================================================================
// MARKDOWN CONVERSION UTILITIES
// ============================================================================

// convertBoldMarkdown converts bold markdown formatting
func (mf *MessageFormatter) convertBoldMarkdown(content, from, to string) string {
	// Handle bold text conversion
	if from == "*" && to == "**" {
		// WhatsApp to Chatwoot: * → **
		// Match *text* but not **text**
		re := regexp.MustCompile(`(?:^|[^*])\*([^*\s][^*]*[^*\s]|\S)\*(?:[^*]|$)`)
		content = re.ReplaceAllStringFunc(content, func(match string) string {
			// Extract the text between asterisks
			start := strings.Index(match, "*")
			end := strings.LastIndex(match, "*")
			if start != -1 && end != -1 && start != end {
				prefix := match[:start]
				text := match[start+1 : end]
				suffix := match[end+1:]
				return prefix + "**" + text + "**" + suffix
			}
			return match
		})
	} else if from == "**" && to == "*" {
		// Chatwoot to WhatsApp: ** → *
		re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
		content = re.ReplaceAllString(content, "*$1*")
	}

	return content
}

// convertItalicMarkdown converts italic markdown formatting
func (mf *MessageFormatter) convertItalicMarkdown(content, from, to string) string {
	if from == "_" && to == "*" {
		// WhatsApp to Chatwoot: _ → *
		re := regexp.MustCompile(`_([^_\s][^_]*[^_\s]|\S)_`)
		content = re.ReplaceAllString(content, "*$1*")
	} else if from == "*" && to == "_" {
		// Chatwoot to WhatsApp: * → _ (but not if it's part of **)
		// Only convert single asterisks that are not part of double asterisks
		re := regexp.MustCompile(`(?:^|[^*])\*([^*\s][^*]*[^*\s]|\S)\*(?:[^*]|$)`)
		content = re.ReplaceAllStringFunc(content, func(match string) string {
			start := strings.Index(match, "*")
			end := strings.LastIndex(match, "*")
			if start != -1 && end != -1 && start != end {
				prefix := match[:start]
				text := match[start+1 : end]
				suffix := match[end+1:]
				return prefix + "_" + text + "_" + suffix
			}
			return match
		})
	}

	return content
}

// convertStrikethroughMarkdown converts strikethrough markdown formatting
func (mf *MessageFormatter) convertStrikethroughMarkdown(content, from, to string) string {
	if from == "~" && to == "~~" {
		// WhatsApp to Chatwoot: ~ → ~~
		re := regexp.MustCompile(`~([^~\s][^~]*[^~\s]|\S)~`)
		content = re.ReplaceAllString(content, "~~$1~~")
	} else if from == "~~" && to == "~" {
		// Chatwoot to WhatsApp: ~~ → ~
		re := regexp.MustCompile(`~~([^~]+)~~`)
		content = re.ReplaceAllString(content, "~$1~")
	}

	return content
}

// ============================================================================
// SPECIAL MESSAGE FORMATTERS
// ============================================================================

// FormatQuotedMessage formats a quoted message for Chatwoot
func (mf *MessageFormatter) FormatQuotedMessage(originalMessage, quotedContent string) string {
	mf.logger.DebugWithFields("Formatting quoted message", map[string]interface{}{
		"quoted_length":   len(quotedContent),
		"original_length": len(originalMessage),
	})

	// Format as blockquote in Chatwoot
	quotedLines := strings.Split(quotedContent, "\n")
	var formattedQuote strings.Builder

	for _, line := range quotedLines {
		if strings.TrimSpace(line) != "" {
			formattedQuote.WriteString("> ")
			formattedQuote.WriteString(line)
			formattedQuote.WriteString("\n")
		}
	}

	// Add the original message after the quote
	result := formattedQuote.String() + "\n" + originalMessage

	return strings.TrimSpace(result)
}

// FormatReactionMessage formats a reaction message
func (mf *MessageFormatter) FormatReactionMessage(reaction, messageContent string) string {
	mf.logger.DebugWithFields("Formatting reaction message", map[string]interface{}{
		"reaction": reaction,
	})

	return "👍 Reacted with " + reaction + " to: \"" + messageContent + "\""
}

// FormatContactMessage formats a contact message for Chatwoot
func (mf *MessageFormatter) FormatContactMessage(contactName, contactPhone string) string {
	mf.logger.DebugWithFields("Formatting contact message", map[string]interface{}{
		"contact_name":  contactName,
		"contact_phone": contactPhone,
	})

	return "📞 **Contact Shared**\n" +
		"**Name:** " + contactName + "\n" +
		"**Phone:** " + contactPhone
}

// FormatLocationMessage formats a location message for Chatwoot
func (mf *MessageFormatter) FormatLocationMessage(latitude, longitude, address string) string {
	mf.logger.DebugWithFields("Formatting location message", map[string]interface{}{
		"latitude":  latitude,
		"longitude": longitude,
		"address":   address,
	})

	message := "📍 **Location Shared**\n"
	if address != "" {
		message += "**Address:** " + address + "\n"
	}
	message += "**Coordinates:** " + latitude + ", " + longitude + "\n"
	message += "**Map:** https://maps.google.com/?q=" + latitude + "," + longitude

	return message
}

// FormatMediaCaption formats media caption with content
func (mf *MessageFormatter) FormatMediaCaption(mediaType, caption string) string {
	if caption == "" {
		return ""
	}

	mf.logger.DebugWithFields("Formatting media caption", map[string]interface{}{
		"media_type":     mediaType,
		"caption_length": len(caption),
	})

	// Format caption for Chatwoot
	formattedCaption := mf.FormatMarkdownForChatwoot(caption)

	return formattedCaption
}

// ExtractMentions extracts mentions from message content
func (mf *MessageFormatter) ExtractMentions(content string) []string {
	// Extract @mentions from content
	re := regexp.MustCompile(`@(\w+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	mentions := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			mentions = append(mentions, match[1])
		}
	}

	return mentions
}

// SanitizeContent sanitizes message content for safe display
func (mf *MessageFormatter) SanitizeContent(content string) string {
	// Remove potentially harmful content
	// This is a basic implementation - extend as needed

	// Remove script tags
	re := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	content = re.ReplaceAllString(content, "")

	// Remove on* event handlers
	re = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	content = re.ReplaceAllString(content, "")

	return content
}

// TruncateContent truncates content to a maximum length
func (mf *MessageFormatter) TruncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	// Truncate and add ellipsis
	truncated := content[:maxLength-3] + "..."

	mf.logger.DebugWithFields("Content truncated", map[string]interface{}{
		"original_length":  len(content),
		"truncated_length": len(truncated),
		"max_length":       maxLength,
	})

	return truncated
}

// FormatSystemMessage formats system messages for Chatwoot
func (mf *MessageFormatter) FormatSystemMessage(messageType, content string) string {
	switch messageType {
	case "group_create":
		return "👥 Group created: " + content
	case "group_add":
		return "➕ Added to group: " + content
	case "group_remove":
		return "➖ Removed from group: " + content
	case "group_leave":
		return "👋 Left the group"
	case "group_subject":
		return "📝 Group name changed to: " + content
	case "group_description":
		return "📄 Group description changed: " + content
	default:
		return "ℹ️ " + content
	}
}
