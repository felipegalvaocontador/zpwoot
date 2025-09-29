# zpwoot API Documentation

## Base URL
```
http://localhost:8080
```

## Authentication
```
Authorization: ZP_API_KEY
```

## Variables
Replace these variables in the examples:
- `SESSION_ID` - Your session UUID (e.g., `b4f3f798-4f80-4369-b602-ce09e8b0a33c`)
- `ZP_API_KEY` - API key from .env file (e.g., `a0b1125a0eb3364d98e2c49ec6f7d6ba`)
- `MESSAGE_ID` - WhatsApp message ID (e.g., `3EB06398DC0CB5E35C31CE`)
- `GROUP_JID` - WhatsApp group JID (e.g., `120363123456789012@g.us`)

## Health Check
- **GET** `/health` - API status
- **GET** `/health/wameow` - WhatsApp manager status

## Sessions
- **POST** `/sessions/create` - Create session (with optional QR code generation)
- **GET** `/sessions/list` - List sessions
- **GET** `/sessions/{sessionId}/info` - Session info
- **DELETE** `/sessions/{sessionId}/delete` - Delete session
- **POST** `/sessions/{sessionId}/connect` - Connect session (returns QR code if needed)
- **POST** `/sessions/{sessionId}/logout` - Logout session
- **GET** `/sessions/{sessionId}/qr` - Get QR code (with base64 image)
- **POST** `/sessions/{sessionId}/pair` - Pair phone

## Proxy
- **POST** `/sessions/{sessionId}/proxy/set` - Configure proxy
- **GET** `/sessions/{sessionId}/proxy/find` - Get proxy config

## Messages - Send
- **POST** `/sessions/{sessionId}/messages/send/text` - Send text
- **POST** `/sessions/{sessionId}/messages/send/image` - Send image
- **POST** `/sessions/{sessionId}/messages/send/audio` - Send audio
- **POST** `/sessions/{sessionId}/messages/send/video` - Send video
- **POST** `/sessions/{sessionId}/messages/send/document` - Send document
- **POST** `/sessions/{sessionId}/messages/send/sticker` - Send sticker
- **POST** `/sessions/{sessionId}/messages/send/location` - Send location
- **POST** `/sessions/{sessionId}/messages/send/contact` - Send contact
- **POST** `/sessions/{sessionId}/messages/send/poll` - Send poll
- **POST** `/sessions/{sessionId}/messages/send/reaction` - Send reaction
- **POST** `/sessions/{sessionId}/messages/send/presence` - Send presence
- **POST** `/sessions/{sessionId}/messages/send/media` - Send media (auto-detect)
- **POST** `/sessions/{sessionId}/messages/send/button` - Send button message
- **POST** `/sessions/{sessionId}/messages/send/list` - Send list message

## Messages - Management
- **POST** `/sessions/{sessionId}/messages/mark-read` - Mark as read
- **POST** `/sessions/{sessionId}/messages/edit` - Edit message
- **POST** `/sessions/{sessionId}/messages/revoke` - Revoke message
- **GET** `/sessions/{sessionId}/messages/poll/{messageId}/results` - Get poll results

## Contacts
- **POST** `/sessions/{sessionId}/contacts/check` - Check WhatsApp numbers
- **GET** `/sessions/{sessionId}/contacts/avatar?jid=...` - Get avatar
- **POST** `/sessions/{sessionId}/contacts/info` - Get contact info
- **GET** `/sessions/{sessionId}/contacts?limit=10` - List contacts
- **GET** `/sessions/{sessionId}/contacts/business?jid=...` - Get business profile
- **POST** `/sessions/{sessionId}/contacts/sync` - Sync contacts

## Groups
- **POST** `/sessions/{sessionId}/groups/create` - Create group
- **GET** `/sessions/{sessionId}/groups` - List groups
- **GET** `/sessions/{sessionId}/groups/info?jid=...` - Get group info
- **POST** `/sessions/{sessionId}/groups/participants` - Manage participants
- **PUT** `/sessions/{sessionId}/groups/name` - Set group name
- **PUT** `/sessions/{sessionId}/groups/description` - Set group description
- **PUT** `/sessions/{sessionId}/groups/photo` - Set group photo
- **GET** `/sessions/{sessionId}/groups/invite-link?jid=...` - Get invite link
- **POST** `/sessions/{sessionId}/groups/join` - Join group via link
- **POST** `/sessions/{sessionId}/groups/leave` - Leave group
- **PUT** `/sessions/{sessionId}/groups/settings` - Update group settings

## Group Requests
- **GET** `/sessions/{sessionId}/groups/requests?jid=...` - List join requests
- **POST** `/sessions/{sessionId}/groups/requests` - Approve/reject requests
- **PUT** `/sessions/{sessionId}/groups/join-approval` - Set approval mode
- **PUT** `/sessions/{sessionId}/groups/member-add-mode` - Set add member mode

## Newsletters (WhatsApp Channels)
- **POST** `/sessions/{sessionId}/newsletters/create` - Create newsletter/channel
- **GET** `/sessions/{sessionId}/newsletters/info?jid=...` - Get newsletter info
- **POST** `/sessions/{sessionId}/newsletters/info-from-invite` - Get info via invite
- **POST** `/sessions/{sessionId}/newsletters/follow` - Follow newsletter
- **POST** `/sessions/{sessionId}/newsletters/unfollow` - Unfollow newsletter
- **GET** `/sessions/{sessionId}/newsletters` - List subscribed newsletters

## Webhooks
- **POST** `/sessions/{sessionId}/webhook/set` - Configure webhook
- **GET** `/sessions/{sessionId}/webhook/find` - Get webhook config

## Chatwoot
- **POST** `/sessions/{sessionId}/chatwoot/set` - Configure Chatwoot
- **GET** `/sessions/{sessionId}/chatwoot/find` - Get Chatwoot config
- **POST** `/sessions/{sessionId}/chatwoot/contacts/sync` - Sync contacts
- **POST** `/sessions/{sessionId}/chatwoot/conversations/sync` - Sync conversations

## Request Examples

### Create Session with QR Code
```bash
# Create session and get QR code immediately
curl -X POST "http://localhost:8080/sessions/create" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-session",
    "qrCode": true,
    "proxyConfig": {
      "host": "proxy.example.com",
      "password": "proxypass123",
      "port": 8080,
      "type": "http",
      "username": "proxyuser"
    }
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "1b2e424c-a2a0-41a4-b992-15b7ec06b9bc",
    "name": "my-session",
    "isConnected": false,
    "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "code": "2@abc123def456...",
    "createdAt": "2024-01-01T00:00:00Z"
  }
}
```

### Create Session without QR Code
```bash
# Create session without QR code (traditional flow)
curl -X POST "http://localhost:8080/sessions/create" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-session",
    "qrCode": false
  }'
```

### Connect Session and Get QR Code
```bash
# Connect session and get QR code if needed
curl -X POST "http://localhost:8080/sessions/my-session/connect" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba"
```

**Response:**
```json
{
  "success": true,
  "message": "Session connection initiated successfully",
  "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
  "code": "2@abc123def456..."
}
```

### Send Text Message
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/text" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "body": "Hello World"}'
```

### Send Image
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/image" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "file": "https://picsum.photos/400/300.jpg", "caption": "Test image"}'
```

### Send Audio
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/audio" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "file": "https://www.soundjay.com/misc/sounds/bell-ringing-05.wav", "ptt": true}'
```

### Send Video
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/video" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "file": "https://cdn1.hongtaocdn3.com/video/m3u8/202401/24/49b02fdd58b9/49b02fdd58b9.mp4", "caption": "Test video"}'
```

### Send Document
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/document" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "file": "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf", "filename": "document.pdf"}'
```

### Send Sticker (Base64)
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/sticker" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "file": "data:image/webp;base64,UklGRlIAAABXRUJQVlA4IEYAAAAwAQCdASoQABAAAgA0JaQAA3AA/vuqAAA="}'
```

### Send Location
```bash
curl -X POST "http://localhost:8080/sessions/b4f3f798-4f80-4369-b602-ce09e8b0a33c/messages/send/location" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "latitude": -23.5505, "longitude": -46.6333, "address": "São Paulo"}'
```

### Send Contact
```bash
curl -X POST "http://localhost:8080/sessions/b4f3f798-4f80-4369-b602-ce09e8b0a33c/messages/send/contact" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "contactName": "John Doe", "contactPhone": "+5511987654321"}'
```

### Send Poll
```bash
curl -X POST "http://localhost:8080/sessions/b4f3f798-4f80-4369-b602-ce09e8b0a33c/messages/send/poll" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "name": "Favorite color?", "options": ["Red", "Blue", "Green"], "selectableOptionCount": 1}'
```

### Send Reaction
```bash
curl -X POST "http://localhost:8080/sessions/b4f3f798-4f80-4369-b602-ce09e8b0a33c/messages/send/reaction" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "messageId": "3EB06398DC0CB5E35C31CE", "reaction": "👍"}'
```

### Send Media (Auto-detect)
```bash
curl -X POST "http://localhost:8080/sessions/b4f3f798-4f80-4369-b602-ce09e8b0a33c/messages/send/media" \
  -H "Authorization: a0b1125a0eb3364d98e2c49ec6f7d6ba" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "file": "https://picsum.photos/400/300.jpg", "mimeType": "image/jpeg", "caption": "Auto-detected image"}'
```

### Mark Message as Read
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/mark-read" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "messageId": "MESSAGE_ID"}'
```

### Edit Message
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/edit" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "5511999999999@s.whatsapp.net", "messageId": "MESSAGE_ID", "newBody": "Edited message"}'
```

### Check WhatsApp Numbers
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/contacts/check" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"phoneNumbers": ["5511999999999", "5511888888888"]}'
```

### Get Contact Avatar
```bash
curl "http://localhost:8080/sessions/SESSION_ID/contacts/avatar?jid=5511999999999@s.whatsapp.net" \
  -H "Authorization: ZP_API_KEY"
```

### List Contacts
```bash
curl "http://localhost:8080/sessions/SESSION_ID/contacts?limit=10&offset=0" \
  -H "Authorization: ZP_API_KEY"
```

### Create Group
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/groups/create" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Group", "participants": ["5511999999999@s.whatsapp.net"], "description": "Group description"}'
```

### List Groups
```bash
curl "http://localhost:8080/sessions/SESSION_ID/groups" \
  -H "Authorization: ZP_API_KEY"
```

### Get Group Info
```bash
curl "http://localhost:8080/sessions/SESSION_ID/groups/info?jid=GROUP_JID" \
  -H "Authorization: ZP_API_KEY"
```

### Set Group Name
```bash
curl -X PUT "http://localhost:8080/sessions/SESSION_ID/groups/name" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"groupJid": "GROUP_JID", "name": "New Group Name"}'
```

### Add Group Participants
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/groups/participants" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"groupJid": "GROUP_JID", "action": "add", "participants": ["5511999999999@s.whatsapp.net"]}'
```

### Get Group Invite Link
```bash
curl "http://localhost:8080/sessions/SESSION_ID/groups/invite-link?jid=GROUP_JID" \
  -H "Authorization: ZP_API_KEY"
```

### Join Group via Link
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/groups/join" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"inviteLink": "https://chat.whatsapp.com/ABC123DEF456"}'
```

### Leave Group
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/groups/leave" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"groupJid": "GROUP_JID"}'
```

### Create Newsletter
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/newsletters/create" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Channel", "description": "Channel description"}'
```

### Get Newsletter Info
```bash
curl "http://localhost:8080/sessions/SESSION_ID/newsletters/info?newsletterJid=120363123456789012@newsletter" \
  -H "Authorization: ZP_API_KEY"
```

### Get Newsletter Info with Invite
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/newsletters/info-from-invite" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"inviteKey": "https://whatsapp.com/channel/0029VaAqUqGCha30a5twXb2j"}'
```

### Follow Newsletter
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/newsletters/follow" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"newsletterJid": "120363123456789012@newsletter"}'
```

### Unfollow Newsletter
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/newsletters/unfollow" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"newsletterJid": "120363123456789012@newsletter"}'
```

### List Subscribed Newsletters
```bash
curl "http://localhost:8080/sessions/SESSION_ID/newsletters" \
  -H "Authorization: ZP_API_KEY"
```

### Configure Proxy
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/proxy/set" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"host": "proxy.example.com", "port": 8080, "username": "user", "password": "pass"}'
```

### Configure Webhook
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/webhook/set" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://your-domain.com/webhook", "events": ["message", "status"]}'
```

### Configure Chatwoot
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/chatwoot/set" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"baseUrl": "https://chatwoot.example.com", "accountId": "1", "token": "your-token"}'
```

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    "id": "MESSAGE_ID",
    "status": "sent",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error description",
  "details": "Additional error details"
}
```

## Media Requirements

### Images
- Formats: JPG, PNG, WebP
- Max size: 10MB (recommended)

### Audio
- Formats: MP3, WAV, OGG
- Max size: 16MB
- PTT: true for voice notes

### Video
- Formats: MP4, WebM
- Max size: 100MB

### Documents
- Formats: PDF, TXT, DOC, XLS, etc.
- Max size: 100MB

### Stickers
- Format: WebP only
- Max size: 100KB (static), 500KB (animated)
- Recommended: Use base64 for reliable delivery

## JID Format
The API supports multiple JID formats:

### Individual Contacts
- **Full JID**: `5511999999999@s.whatsapp.net`
- **Phone with +**: `+5511999999999`
- **Phone only**: `5511999999999`
- **Formatted**: `+55 (11) 99999-9999`

### Groups
- **Group JID**: `120363123456789012@g.us`

### Newsletters (Channels)
- **Newsletter JID**: `120363123456789012@newsletter`

All formats are automatically normalized to WhatsApp JID format.

## Sending Emojis
To send emojis in messages, use one of these methods:

### Method 1: JSON File (Recommended)
```bash
# Create file with emoji
echo '{"remoteJid": "+5511999999999", "body": "Hello! 👋 How are you? 😊🎉"}' > message.json

# Send file
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/text" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  --data @message.json
```

### Method 2: Unicode Escape
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/text" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"remoteJid": "+5511999999999", "body": "Party time! \\ud83c\\udf89"}'
```

### Method 3: UTF-8 with --data-raw
```bash
curl -X POST "http://localhost:8080/sessions/SESSION_ID/messages/send/text" \
  -H "Authorization: ZP_API_KEY" \
  -H "Content-Type: application/json; charset=utf-8" \
  --data-raw '{"remoteJid": "+5511999999999", "body": "Hello 🌟"}'
```

## Swagger Documentation
Access interactive API documentation at: http://localhost:8080/swagger/
