package chatwoot

import (
	"regexp"
	"strings"

	"zpwoot/platform/logger"
)

type MessageFormatter struct {
	logger *logger.Logger
}

func NewMessageFormatter(logger *logger.Logger) *MessageFormatter {
	return &MessageFormatter{
		logger: logger,
	}
}

func (mf *MessageFormatter) FormatMarkdownForChatwoot(content string) string {
	mf.logger.DebugWithFields("Formatting markdown for Chatwoot", map[string]interface{}{
		"original_length": len(content),
	})

	content = mf.convertBoldMarkdown(content, "*", "**")

	content = mf.convertItalicMarkdown(content, "_", "*")

	content = mf.convertStrikethroughMarkdown(content, "~", "~~")

	mf.logger.DebugWithFields("Formatted markdown for Chatwoot", map[string]interface{}{
		"formatted_length": len(content),
	})

	return content
}

func (mf *MessageFormatter) FormatMarkdownForWhatsApp(content string) string {
	mf.logger.DebugWithFields("Formatting markdown for WhatsApp", map[string]interface{}{
		"original_length": len(content),
	})

	content = mf.convertBoldMarkdown(content, "**", "*")

	content = mf.convertItalicMarkdown(content, "*", "_")

	content = mf.convertStrikethroughMarkdown(content, "~~", "~")

	mf.logger.DebugWithFields("Formatted markdown for WhatsApp", map[string]interface{}{
		"formatted_length": len(content),
	})

	return content
}

func (mf *MessageFormatter) convertBoldMarkdown(content, from, to string) string {
	if from == "*" && to == "**" {
		re := regexp.MustCompile(`(?:^|[^*])\*([^*\s][^*]*[^*\s]|\S)\*(?:[^*]|$)`)
		content = re.ReplaceAllStringFunc(content, func(match string) string {
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
		re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
		content = re.ReplaceAllString(content, "*$1*")
	}

	return content
}

func (mf *MessageFormatter) convertItalicMarkdown(content, from, to string) string {
	if from == "_" && to == "*" {
		re := regexp.MustCompile(`_([^_\s][^_]*[^_\s]|\S)_`)
		content = re.ReplaceAllString(content, "*$1*")
	} else if from == "*" && to == "_" {
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

func (mf *MessageFormatter) convertStrikethroughMarkdown(content, from, to string) string {
	if from == "~" && to == "~~" {
		re := regexp.MustCompile(`~([^~\s][^~]*[^~\s]|\S)~`)
		content = re.ReplaceAllString(content, "~~$1~~")
	} else if from == "~~" && to == "~" {
		re := regexp.MustCompile(`~~([^~]+)~~`)
		content = re.ReplaceAllString(content, "~$1~")
	}

	return content
}

func (mf *MessageFormatter) FormatQuotedMessage(originalMessage, quotedContent string) string {
	mf.logger.DebugWithFields("Formatting quoted message", map[string]interface{}{
		"quoted_length":   len(quotedContent),
		"original_length": len(originalMessage),
	})

	quotedLines := strings.Split(quotedContent, "\n")
	var formattedQuote strings.Builder

	for _, line := range quotedLines {
		if strings.TrimSpace(line) != "" {
			formattedQuote.WriteString("> ")
			formattedQuote.WriteString(line)
			formattedQuote.WriteString("\n")
		}
	}

	result := formattedQuote.String() + "\n" + originalMessage

	return strings.TrimSpace(result)
}

func (mf *MessageFormatter) FormatReactionMessage(reaction, messageContent string) string {
	mf.logger.DebugWithFields("Formatting reaction message", map[string]interface{}{
		"reaction": reaction,
	})

	return "👍 Reacted with " + reaction + " to: \"" + messageContent + "\""
}

func (mf *MessageFormatter) FormatContactMessage(contactName, contactPhone string) string {
	mf.logger.DebugWithFields("Formatting contact message", map[string]interface{}{
		"contact_name":  contactName,
		"contact_phone": contactPhone,
	})

	return "📞 **Contact Shared**\n" +
		"**Name:** " + contactName + "\n" +
		"**Phone:** " + contactPhone
}

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

func (mf *MessageFormatter) FormatMediaCaption(mediaType, caption string) string {
	if caption == "" {
		return ""
	}

	mf.logger.DebugWithFields("Formatting media caption", map[string]interface{}{
		"media_type":     mediaType,
		"caption_length": len(caption),
	})

	formattedCaption := mf.FormatMarkdownForChatwoot(caption)

	return formattedCaption
}

func (mf *MessageFormatter) ExtractMentions(content string) []string {
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

func (mf *MessageFormatter) SanitizeContent(content string) string {
	re := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	content = re.ReplaceAllString(content, "")

	re = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	content = re.ReplaceAllString(content, "")

	return content
}

func (mf *MessageFormatter) TruncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	truncated := content[:maxLength-3] + "..."

	mf.logger.DebugWithFields("Content truncated", map[string]interface{}{
		"original_length":  len(content),
		"truncated_length": len(truncated),
		"max_length":       maxLength,
	})

	return truncated
}

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
