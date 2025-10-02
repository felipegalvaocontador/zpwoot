# 📚 ZPWoot API Routes Documentation

Esta documentação detalha todas as rotas disponíveis na API do ZPWoot, organizadas por módulos funcionais.

## 🔗 Base URL
```
http://localhost:8080
```

## 🔐 Autenticação
Todas as rotas (exceto públicas) requerem autenticação via API Key:
```
Authorization: Bearer YOUR_API_KEY
# ou
X-API-Key: YOUR_API_KEY
```

## 📋 Índice de Rotas

- [🔧 Sessions](#-sessions) - Gerenciamento de sessões WhatsApp
- [💬 Messages](#-messages) - Envio e gerenciamento de mensagens
- [👥 Groups](#-groups) - Gerenciamento de grupos
- [👤 Contacts](#-contacts) - Gerenciamento de contatos
- [🔗 Webhooks](#-webhooks) - Configuração de webhooks
- [📁 Media](#-media) - Gerenciamento de mídia
- [🤖 Chatwoot](#-chatwoot) - Integração Chatwoot
- [🏥 Health](#-health) - Status da aplicação

---

## 🔧 Sessions

### Gerenciamento de Sessões

#### `POST /sessions/create`
Cria uma nova sessão WhatsApp.

**Request Body:**
```json
{
  "sessionName": "my-session",
  "qrCode": true,
  "proxy": {
    "host": "proxy.example.com",
    "port": 8080,
    "username": "user",
    "password": "pass"
  }
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "sessionId": "my-session",
    "status": "created",
    "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "qrCodeUrl": "https://api.qrserver.com/v1/create-qr-code/?data=...",
    "expiresAt": "2024-01-01T12:00:00Z"
  },
  "message": "Session created successfully"
}
```

#### `GET /sessions/list`
Lista todas as sessões existentes.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "sessions": [
      {
        "sessionId": "my-session",
        "status": "connected",
        "deviceInfo": {
          "pushName": "My WhatsApp",
          "platform": "android",
          "deviceManufacturer": "Samsung"
        },
        "createdAt": "2024-01-01T10:00:00Z",
        "connectedAt": "2024-01-01T10:05:00Z"
      }
    ],
    "total": 1
  },
  "message": "Sessions retrieved successfully"
}
```

#### `GET /sessions/{sessionId}/info`
Obtém informações detalhadas de uma sessão.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "sessionId": "my-session",
    "status": "connected",
    "deviceInfo": {
      "pushName": "My WhatsApp",
      "platform": "android",
      "deviceManufacturer": "Samsung",
      "deviceModel": "Galaxy S21",
      "osVersion": "11",
      "waVersion": "2.23.24.76"
    },
    "proxy": {
      "enabled": true,
      "host": "proxy.example.com",
      "port": 8080
    },
    "createdAt": "2024-01-01T10:00:00Z",
    "connectedAt": "2024-01-01T10:05:00Z",
    "lastActivity": "2024-01-01T12:00:00Z"
  },
  "message": "Session info retrieved successfully"
}
```

#### `DELETE /sessions/{sessionId}/delete`
Remove uma sessão permanentemente.

**Response (200):**
```json
{
  "success": true,
  "data": null,
  "message": "Session deleted successfully"
}
```

### Conexão e Autenticação

#### `POST /sessions/{sessionId}/connect`
Conecta uma sessão ao WhatsApp.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "status": "connecting",
    "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "qrCodeUrl": "https://api.qrserver.com/v1/create-qr-code/?data=...",
    "expiresAt": "2024-01-01T12:05:00Z"
  },
  "message": "Connection initiated successfully"
}
```

#### `POST /sessions/{sessionId}/logout`
Desconecta uma sessão do WhatsApp.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "status": "disconnected"
  },
  "message": "Session logged out successfully"
}
```

#### `GET /sessions/{sessionId}/qr`
Obtém QR Code para conexão.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
    "qrCodeUrl": "https://api.qrserver.com/v1/create-qr-code/?data=...",
    "expiresAt": "2024-01-01T12:05:00Z"
  },
  "message": "QR Code generated successfully"
}
```

#### `POST /sessions/{sessionId}/pair`
Pareamento via código de telefone.

**Request Body:**
```json
{
  "phoneNumber": "+5511999999999"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "pairingCode": "ABCD-EFGH",
    "expiresAt": "2024-01-01T12:05:00Z"
  },
  "message": "Pairing code generated successfully"
}
```

### Configuração de Proxy

#### `POST /sessions/{sessionId}/proxy/set`
Configura proxy para a sessão.

**Request Body:**
```json
{
  "host": "proxy.example.com",
  "port": 8080,
  "username": "user",
  "password": "pass",
  "protocol": "http"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "proxyConfigured": true
  },
  "message": "Proxy configured successfully"
}
```

#### `GET /sessions/{sessionId}/proxy/find`
Obtém configuração atual do proxy.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "enabled": true,
    "host": "proxy.example.com",
    "port": 8080,
    "protocol": "http",
    "username": "user"
  },
  "message": "Proxy configuration retrieved successfully"
}
```

### Estatísticas

#### `GET /sessions/{sessionId}/stats`
Obtém estatísticas da sessão.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "messagesSent": 150,
    "messagesReceived": 89,
    "uptime": "2h 30m",
    "lastActivity": "2024-01-01T12:00:00Z",
    "connectionQuality": "excellent"
  },
  "message": "Session statistics retrieved successfully"
}
```

---

## 💬 Messages

### CRUD de Mensagens

#### `POST /sessions/{sessionId}/messages`
Cria/registra uma nova mensagem no sistema.

#### `GET /sessions/{sessionId}/messages`
Lista mensagens da sessão.

#### `GET /sessions/{sessionId}/messages/{messageId}`
Obtém detalhes de uma mensagem específica.

#### `DELETE /sessions/{sessionId}/messages/{messageId}`
Remove uma mensagem do sistema.

### Envio de Mensagens

#### `POST /sessions/{sessionId}/messages/send/text`
Envia mensagem de texto.

**Request Body:**
```json
{
  "to": "5511999999999@s.whatsapp.net",
  "text": "Hello, World!",
  "options": {
    "linkPreview": true,
    "mentionedJid": ["5511888888888@s.whatsapp.net"]
  }
}
```

#### `POST /sessions/{sessionId}/messages/send/media`
Envia mensagem de mídia.

**Request Body:**
```json
{
  "to": "5511999999999@s.whatsapp.net",
  "mediaUrl": "https://example.com/image.jpg",
  "type": "image",
  "caption": "Check this out!",
  "filename": "image.jpg"
}
```

#### `POST /sessions/{sessionId}/messages/send/image`
Envia imagem.

#### `POST /sessions/{sessionId}/messages/send/audio`
Envia áudio.

#### `POST /sessions/{sessionId}/messages/send/video`
Envia vídeo.

#### `POST /sessions/{sessionId}/messages/send/document`
Envia documento.

### Mensagens Interativas

#### `POST /sessions/{sessionId}/messages/send/button`
Envia mensagem com botões.

#### `POST /sessions/{sessionId}/messages/send/list`
Envia mensagem com lista.

#### `POST /sessions/{sessionId}/messages/send/poll`
Envia enquete.

### Ações de Mensagem

#### `POST /sessions/{sessionId}/messages/edit`
Edita uma mensagem.

#### `POST /sessions/{sessionId}/messages/revoke`
Revoga uma mensagem.

#### `POST /sessions/{sessionId}/messages/mark-read`
Marca mensagens como lidas.

#### `POST /sessions/{sessionId}/messages/send/reaction`
Envia reação a uma mensagem.

---

## 👥 Groups

### Gerenciamento Básico

#### `POST /sessions/{sessionId}/groups`
Cria um novo grupo.

#### `GET /sessions/{sessionId}/groups`
Lista grupos da sessão.

#### `GET /sessions/{sessionId}/groups/info`
Obtém informações de um grupo.

### Participantes

#### `POST /sessions/{sessionId}/groups/participants`
Gerencia participantes do grupo.

### Configurações

#### `PUT /sessions/{sessionId}/groups/name`
Altera nome do grupo.

#### `PUT /sessions/{sessionId}/groups/description`
Altera descrição do grupo.

#### `PUT /sessions/{sessionId}/groups/photo`
Altera foto do grupo.

---

## 👤 Contacts

### Verificação

#### `POST /sessions/{sessionId}/contacts/check`
Verifica se contatos estão no WhatsApp.

#### `POST /sessions/{sessionId}/contacts/is-on-whatsapp`
Verifica se um número específico está no WhatsApp.

### Informações

#### `GET /sessions/{sessionId}/contacts/avatar`
Obtém avatar de um contato.

#### `POST /sessions/{sessionId}/contacts/info`
Obtém informações de contatos.

#### `GET /sessions/{sessionId}/contacts/profile-picture-info`
Obtém informações da foto de perfil.

### Listagem

#### `GET /sessions/{sessionId}/contacts`
Lista contatos.

#### `GET /sessions/{sessionId}/contacts/all`
Obtém todos os contatos.

---

## 🔗 Webhooks

#### `POST /sessions/{sessionId}/webhook/set`
Configura webhook para a sessão.

#### `GET /sessions/{sessionId}/webhook/find`
Obtém configuração atual do webhook.

#### `POST /sessions/{sessionId}/webhook/test`
Testa configuração do webhook.

---

## 📁 Media

#### `POST /sessions/{sessionId}/media/download`
Faz download de mídia.

#### `GET /sessions/{sessionId}/media/info`
Obtém informações de mídia.

#### `GET /sessions/{sessionId}/media/list`
Lista mídia em cache.

#### `POST /sessions/{sessionId}/media/clear-cache`
Limpa cache de mídia.

#### `GET /sessions/{sessionId}/media/stats`
Obtém estatísticas de mídia.

---

## 🤖 Chatwoot

#### `POST /sessions/{sessionId}/chatwoot/set`
Configura integração Chatwoot.

#### `GET /sessions/{sessionId}/chatwoot`
Obtém configuração Chatwoot.

#### `PUT /sessions/{sessionId}/chatwoot`
Atualiza configuração Chatwoot.

#### `DELETE /sessions/{sessionId}/chatwoot`
Remove configuração Chatwoot.

#### `POST /sessions/{sessionId}/chatwoot/test`
Testa conexão Chatwoot.

#### `GET /sessions/{sessionId}/chatwoot/stats`
Obtém estatísticas Chatwoot.

---

## 🏥 Health

#### `GET /health`
Verifica status da aplicação.

**Response (200):**
```json
{
  "status": "ok",
  "service": "zpwoot",
  "version": "2.0.0"
}
```

---

## 📝 Códigos de Status HTTP

- `200` - OK
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error

## 🔍 Filtros e Paginação

Muitas rotas de listagem suportam parâmetros de query:

- `limit` - Número máximo de resultados (padrão: 50)
- `offset` - Número de resultados para pular (padrão: 0)
- `sort` - Campo para ordenação
- `order` - Direção da ordenação (asc/desc)

Exemplo:
```
GET /sessions/my-session/messages?limit=20&offset=40&sort=createdAt&order=desc
```
